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

	macCmd.Flags().String("style", "colon", "Separator style: colon, dash, dot")

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
		style, _ := cmd.Flags().GetString("style")
		opts := forge.Options{Count: 1, Format: of.ResolveFormat(), Params: map[string]string{"style": style}}

		if handled, err := flags.RunWhy(cmd, g, opts); handled {
			return err
		}

		return flags.RunGenerate(cmd.Context(), flags.RunOptions{
			Generator: g,
			Opts:      opts,
			Count:     flags.GetCount(cmd),
			Format:    of.ResolveFormat(),
			OutputDir: of.OutputDir,
			Writer:    os.Stdout,
		})
	},
}
