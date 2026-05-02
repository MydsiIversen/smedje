package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/smedje/smedje/internal/output"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	rootCmd.AddCommand(snowflakeCmd)

	snowflakeCmd.Flags().Int("worker", 0, "Worker ID (0-1023)")
	snowflakeCmd.Flags().Bool("json", false, "Output as JSON")
	snowflakeCmd.Flags().Bool("quiet", false, "Output only the value")
	snowflakeCmd.Flags().Bool("bench", false, "Run a benchmark instead of generating")
}

var snowflakeCmd = &cobra.Command{
	Use:   "snowflake",
	Short: "Generate a Twitter-style Snowflake ID (64-bit)",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, ok := forge.Get(forge.CategoryID, "snowflake")
		if !ok {
			return fmt.Errorf("generator not found: id/snowflake")
		}

		benchFlag, _ := cmd.Flags().GetBool("bench")
		if benchFlag {
			result, err := g.Bench(cmd.Context())
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s: %d ops in %s (%.0f ops/sec)\n",
				result.Generator, result.Iterations, result.Duration, result.OpsPerSec)
			return nil
		}

		format := "text"
		if j, _ := cmd.Flags().GetBool("json"); j {
			format = "json"
		} else if q, _ := cmd.Flags().GetBool("quiet"); q {
			format = "quiet"
		}

		worker, _ := cmd.Flags().GetInt("worker")
		opts := forge.Options{
			Count:  1,
			Format: format,
			Params: map[string]string{
				"worker": fmt.Sprintf("%d", worker),
			},
		}

		out, err := g.Generate(cmd.Context(), opts)
		if err != nil {
			return err
		}
		return output.Render(os.Stdout, out, format)
	},
}
