package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	_ "github.com/smedje/smedje/pkg/forge/network"

	"github.com/smedje/smedje/internal/output"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	rootCmd.AddCommand(macCmd)

	macCmd.Flags().Bool("json", false, "Output as JSON")
	macCmd.Flags().Bool("quiet", false, "Output only the value")
	macCmd.Flags().Bool("bench", false, "Run a benchmark instead of generating")
}

var macCmd = &cobra.Command{
	Use:   "mac",
	Short: "Generate a random locally-administered MAC address",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, ok := forge.Get(forge.CategoryNetwork, "mac")
		if !ok {
			return fmt.Errorf("generator not found: network/mac")
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
