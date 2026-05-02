// Package bench provides a benchmarking harness for generators.
package bench

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/smedje/smedje/pkg/forge"
)

// Result holds the outcome of a benchmark run.
type Result struct {
	Generator   string      `json:"generator"`
	Duration    Duration    `json:"duration_ns"`
	Operations  int64       `json:"operations"`
	OpsPerSec   float64     `json:"ops_per_sec"`
	NsPerOp     float64     `json:"ns_per_op"`
	AllocsPerOp float64     `json:"allocs_per_op"`
	BytesPerOp  int64       `json:"bytes_per_op"`
	Repeats     int         `json:"repeats,omitempty"`
	VarianceCV  float64     `json:"variance_cv,omitempty"`
	Machine     MachineInfo `json:"machine"`
	GoVersion   string      `json:"go_version"`
}

// Duration wraps time.Duration for JSON as nanoseconds.
type Duration time.Duration

func (d Duration) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%d", time.Duration(d).Nanoseconds())), nil
}

func (d Duration) String() string {
	return time.Duration(d).String()
}

// MachineInfo describes the system running the benchmark.
type MachineInfo struct {
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	CPUModel string `json:"cpu_model,omitempty"`
	Cores    int    `json:"cores"`
}

// Options configures a benchmark run.
type Options struct {
	Duration time.Duration
	Warmup   time.Duration
	Repeat   int
	Cores    int
}

// DefaultOptions returns sensible benchmark defaults.
func DefaultOptions() Options {
	return Options{
		Duration: 2 * time.Second,
		Warmup:   500 * time.Millisecond,
		Repeat:   1,
		Cores:    runtime.NumCPU(),
	}
}

// Run benchmarks a generator and returns structured results.
func Run(ctx context.Context, g forge.Generator, opts Options) (*Result, error) {
	if opts.Duration == 0 {
		opts.Duration = DefaultOptions().Duration
	}
	if opts.Warmup == 0 {
		opts.Warmup = DefaultOptions().Warmup
	}
	if opts.Repeat == 0 {
		opts.Repeat = 1
	}

	genOpts := forge.Options{Count: 1, Format: "quiet"}

	// Warmup
	warmEnd := time.Now().Add(opts.Warmup)
	for time.Now().Before(warmEnd) {
		if _, err := g.Generate(ctx, genOpts); err != nil {
			return nil, err
		}
	}

	var results []singleResult
	for r := range opts.Repeat {
		_ = r
		res, err := runOnce(ctx, g, genOpts, opts.Duration)
		if err != nil {
			return nil, err
		}
		results = append(results, res)
	}

	best := aggregate(results)
	machine := getMachineInfo()

	goVer := runtime.Version()
	if info, ok := debug.ReadBuildInfo(); ok && info.GoVersion != "" {
		goVer = info.GoVersion
	}

	return &Result{
		Generator:   g.Name(),
		Duration:    Duration(best.duration),
		Operations:  best.ops,
		OpsPerSec:   best.opsPerSec,
		NsPerOp:     best.nsPerOp,
		AllocsPerOp: best.allocsPerOp,
		BytesPerOp:  best.bytesPerOp,
		Repeats:     opts.Repeat,
		VarianceCV:  best.cv,
		Machine:     machine,
		GoVersion:   goVer,
	}, nil
}

// RunLegacy provides backwards compatibility with the old bench.Run(ctx, g, d) signature.
func RunLegacy(ctx context.Context, g forge.Generator, d time.Duration) (*forge.BenchResult, error) {
	if d == 0 {
		d = 2 * time.Second
	}
	r, err := Run(ctx, g, Options{Duration: d, Warmup: 0, Repeat: 1})
	if err != nil {
		return nil, err
	}
	return &forge.BenchResult{
		Generator:  r.Generator,
		Iterations: int(r.Operations),
		Duration:   time.Duration(r.Duration),
		OpsPerSec:  r.OpsPerSec,
	}, nil
}

type singleResult struct {
	duration    time.Duration
	ops         int64
	opsPerSec   float64
	nsPerOp     float64
	allocsPerOp float64
	bytesPerOp  int64
}

func runOnce(ctx context.Context, g forge.Generator, opts forge.Options, d time.Duration) (singleResult, error) {
	runtime.GC()

	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	var ops int64
	start := time.Now()
	deadline := start.Add(d)

	for time.Now().Before(deadline) {
		if err := ctx.Err(); err != nil {
			return singleResult{}, err
		}
		if _, err := g.Generate(ctx, opts); err != nil {
			return singleResult{}, err
		}
		ops++
	}

	elapsed := time.Since(start)

	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	allocsPerOp := float64(0)
	bytesPerOp := int64(0)
	if ops > 0 {
		totalAllocs := memAfter.Mallocs - memBefore.Mallocs
		totalBytes := memAfter.TotalAlloc - memBefore.TotalAlloc
		allocsPerOp = float64(totalAllocs) / float64(ops)
		bytesPerOp = int64(totalBytes) / ops
	}

	return singleResult{
		duration:    elapsed,
		ops:         ops,
		opsPerSec:   float64(ops) / elapsed.Seconds(),
		nsPerOp:     float64(elapsed.Nanoseconds()) / float64(ops),
		allocsPerOp: allocsPerOp,
		bytesPerOp:  bytesPerOp,
	}, nil
}

type aggregated struct {
	duration    time.Duration
	ops         int64
	opsPerSec   float64
	nsPerOp     float64
	allocsPerOp float64
	bytesPerOp  int64
	cv          float64
}

func aggregate(results []singleResult) aggregated {
	if len(results) == 1 {
		r := results[0]
		return aggregated{
			duration:    r.duration,
			ops:         r.ops,
			opsPerSec:   r.opsPerSec,
			nsPerOp:     r.nsPerOp,
			allocsPerOp: r.allocsPerOp,
			bytesPerOp:  r.bytesPerOp,
		}
	}

	var sumOps float64
	var sumNs float64
	for _, r := range results {
		sumOps += r.opsPerSec
		sumNs += r.nsPerOp
	}
	meanOps := sumOps / float64(len(results))
	meanNs := sumNs / float64(len(results))

	var sumSqDev float64
	for _, r := range results {
		dev := r.opsPerSec - meanOps
		sumSqDev += dev * dev
	}
	stddev := math.Sqrt(sumSqDev / float64(len(results)))
	cv := 0.0
	if meanOps > 0 {
		cv = (stddev / meanOps) * 100
	}

	best := results[0]
	for _, r := range results[1:] {
		if r.opsPerSec > best.opsPerSec {
			best = r
		}
	}

	return aggregated{
		duration:    best.duration,
		ops:         best.ops,
		opsPerSec:   best.opsPerSec,
		nsPerOp:     meanNs,
		allocsPerOp: best.allocsPerOp,
		bytesPerOp:  best.bytesPerOp,
		cv:          cv,
	}
}

func getMachineInfo() MachineInfo {
	return MachineInfo{
		OS:    runtime.GOOS,
		Arch:  runtime.GOARCH,
		Cores: runtime.NumCPU(),
	}
}
