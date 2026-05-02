package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	_ "github.com/smedje/smedje/pkg/forge/wireguard"

	"github.com/smedje/smedje/internal/flags"
	"github.com/smedje/smedje/internal/output"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	rootCmd.AddCommand(wireguardCmd)
	wireguardCmd.AddCommand(wireguardKeypairCmd)

	flags.AddOutputFlags(wireguardKeypairCmd)
	flags.AddBulkFlags(wireguardKeypairCmd)
	flags.AddBenchFlag(wireguardKeypairCmd)
}

var wireguardCmd = &cobra.Command{
	Use:   "wireguard",
	Short: "Generate WireGuard keys",
}

var wireguardKeypairCmd = &cobra.Command{
	Use:   "keypair",
	Short: "Generate a WireGuard Curve25519 keypair",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, ok := forge.Get(forge.CategoryCrypto, "keypair")
		if !ok {
			return fmt.Errorf("generator not found: crypto/keypair")
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
