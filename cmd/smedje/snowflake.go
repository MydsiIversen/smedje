package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/smedje/smedje/internal/flags"
	"github.com/smedje/smedje/internal/output"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	rootCmd.AddCommand(snowflakeCmd)

	snowflakeCmd.Flags().Int("worker", 0, "Worker ID (0-1023)")
	flags.AddOutputFlags(snowflakeCmd)
	flags.AddBulkFlags(snowflakeCmd)
	flags.AddBenchFlag(snowflakeCmd)
}

var snowflakeCmd = &cobra.Command{
	Use:   "snowflake",
	Short: "Generate a Twitter-style Snowflake ID (64-bit)",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, ok := forge.Get(forge.CategoryID, "snowflake")
		if !ok {
			return fmt.Errorf("generator not found: id/snowflake")
		}

		if flags.GetBench(cmd) {
			result, err := g.Bench(cmd.Context())
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s: %d ops in %s (%.0f ops/sec)\n",
				result.Generator, result.Iterations, result.Duration, result.OpsPerSec)
			return nil
		}

		of := flags.GetOutputFlags(cmd)
		count := flags.GetCount(cmd)
		worker, _ := cmd.Flags().GetInt("worker")

		for i := range count {
			out, err := g.Generate(cmd.Context(), forge.Options{
				Count:  1,
				Format: of.ResolveFormat(),
				Params: map[string]string{
					"worker": fmt.Sprintf("%d", worker),
				},
			})
			if err != nil {
				return err
			}
			if err := output.Render(os.Stdout, out, of.ResolveFormat()); err != nil {
				return err
			}
			_ = i
		}
		return nil
	},
}
