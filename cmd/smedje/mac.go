package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	_ "github.com/smedje/smedje/pkg/forge/network"

	"github.com/smedje/smedje/internal/flags"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	rootCmd.AddCommand(macCmd)

	flags.AddOutputFlags(macCmd)
	flags.AddBulkFlags(macCmd)
	flags.AddBenchFlag(macCmd)
	flags.AddSeedFlags(macCmd)
	flags.AddWhyFlag(macCmd)
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
			return runBench(cmd, g)
		}

		flags.ApplySeed(cmd)
		defer flags.CleanupSeed(cmd)
		of := flags.GetOutputFlags(cmd)
		opts := forge.Options{Count: 1, Format: of.ResolveFormat()}

		if handled, err := flags.RunWhy(cmd, g, opts); handled {
			return err
		}

		return flags.RunGenerate(cmd.Context(), flags.RunOptions{
			Generator: g,
			Opts:      opts,
			Count:     flags.GetCount(cmd),
			Format:    of.ResolveFormat(),
			Writer:    os.Stdout,
		})
	},
}
