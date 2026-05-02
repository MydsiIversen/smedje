// Package bench provides a shared harness for generator self-benchmarks.
package bench

import (
	"context"
	"time"

	"github.com/smedje/smedje/pkg/forge"
)

// DefaultDuration is how long a benchmark runs.
const DefaultDuration = 2 * time.Second

// Run benchmarks a generator by calling Generate in a tight loop for the
// given duration and returns the result.
func Run(ctx context.Context, g forge.Generator, d time.Duration) (*forge.BenchResult, error) {
	if d == 0 {
		d = DefaultDuration
	}

	opts := forge.Options{Count: 1, Format: "quiet"}
	iterations := 0

	start := time.Now()
	deadline := start.Add(d)

	for time.Now().Before(deadline) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if _, err := g.Generate(ctx, opts); err != nil {
			return nil, err
		}
		iterations++
	}

	elapsed := time.Since(start)
	return &forge.BenchResult{
		Generator:  g.Name(),
		Iterations: iterations,
		Duration:   elapsed,
		OpsPerSec:  float64(iterations) / elapsed.Seconds(),
	}, nil
}
