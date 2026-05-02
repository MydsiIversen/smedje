package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	_ "github.com/smedje/smedje/pkg/forge/id"

	"github.com/smedje/smedje/internal/output"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	rootCmd.AddCommand(uuidCmd)
	uuidCmd.AddCommand(uuidV7Cmd)

	uuidV7Cmd.Flags().Bool("json", false, "Output as JSON")
	uuidV7Cmd.Flags().Bool("quiet", false, "Output only the value")
	uuidV7Cmd.Flags().Bool("bench", false, "Run a benchmark instead of generating")
}

var uuidCmd = &cobra.Command{
	Use:   "uuid",
	Short: "Generate UUIDs",
}

var uuidV7Cmd = &cobra.Command{
	Use:   "v7",
	Short: "Generate a UUIDv7 (RFC 9562, time-ordered)",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, ok := forge.Get(forge.CategoryID, "v7")
		if !ok {
			return fmt.Errorf("generator not found: id/v7")
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

		out, err := g.Generate(cmd.Context(), forge.Options{Count: 1, Format: format})
		if err != nil {
			return err
		}
		return output.Render(os.Stdout, out, format)
	},
}
