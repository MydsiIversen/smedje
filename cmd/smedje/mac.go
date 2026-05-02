package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	_ "github.com/smedje/smedje/pkg/forge/network"

	"github.com/smedje/smedje/internal/flags"
	"github.com/smedje/smedje/internal/output"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	rootCmd.AddCommand(macCmd)

	flags.AddOutputFlags(macCmd)
	flags.AddBulkFlags(macCmd)
	flags.AddBenchFlag(macCmd)
}

var macCmd = &cobra.Command{
	Use:   "mac",
	Short: "Generate a random locally-administered MAC address",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, ok := forge.Get(forge.CategoryNetwork, "mac")
		if !ok {
			return fmt.Errorf("generator not found: network/mac")
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

		for i := range count {
			out, err := g.Generate(cmd.Context(), forge.Options{
				Count:  1,
				Format: of.ResolveFormat(),
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
