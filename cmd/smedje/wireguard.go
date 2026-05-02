package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	_ "github.com/smedje/smedje/pkg/forge/wireguard"

	"github.com/smedje/smedje/internal/flags"
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
			return runBench(cmd, g)
		}

		of := flags.GetOutputFlags(cmd)
		return flags.RunGenerate(cmd.Context(), flags.RunOptions{
			Generator: g,
			Opts:      forge.Options{Count: 1, Format: of.ResolveFormat()},
			Count:     flags.GetCount(cmd),
			Format:    of.ResolveFormat(),
			Writer:    os.Stdout,
		})
	},
}
