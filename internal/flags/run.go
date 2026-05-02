package flags

import (
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/smedje/smedje/internal/config"
	"github.com/smedje/smedje/internal/output"
	"github.com/smedje/smedje/internal/progress"
	"github.com/smedje/smedje/pkg/forge"
)

// RunOptions configures a generator run.
type RunOptions struct {
	Generator forge.Generator
	Opts      forge.Options
	Count     int
	Format    string
	SQLTable  string
	Writer    io.Writer
}

// RunGenerate executes a generator Count times and renders output in the
// appropriate batch format. Enforces bulk.max-count from config.
func RunGenerate(ctx context.Context, r RunOptions) error {
	maxCount := 100000000
	if v := config.GetDefault("bulk.max-count"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxCount = n
		}
	}

	if r.Count <= 0 {
		return fmt.Errorf("--count must be >= 1")
	}
	if r.Count > maxCount {
		return fmt.Errorf("--count %d exceeds maximum %d (configure bulk.max-count to increase)", r.Count, maxCount)
	}

	// Single item: render directly.
	if r.Count == 1 && r.Format != "json" && r.Format != "csv" && r.Format != "sql" {
		out, err := r.Generator.Generate(ctx, r.Opts)
		if err != nil {
			return err
		}
		return output.Render(r.Writer, out, r.Format)
	}

	// Batch: collect and render.
	quiet := r.Format == "quiet"
	prog := progress.New(forge.Address(r.Generator), r.Count, quiet)
	outputs := make([]*forge.Output, 0, r.Count)
	for i := range r.Count {
		out, err := r.Generator.Generate(ctx, r.Opts)
		if err != nil {
			prog.Done()
			return err
		}
		outputs = append(outputs, out)
		prog.Update(i + 1)
	}
	prog.Done()

	return output.RenderBatch(r.Writer, outputs, r.Format, output.BatchOptions{
		SQLTable: r.SQLTable,
	})
}
