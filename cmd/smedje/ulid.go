package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/smedje/smedje/internal/flags"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	rootCmd.AddCommand(ulidCmd)

	flags.AddOutputFlags(ulidCmd)
	flags.AddBulkFlags(ulidCmd)
	flags.AddBenchFlag(ulidCmd)
	flags.AddSeedFlags(ulidCmd)
	flags.AddWhyFlag(ulidCmd)
}

var ulidCmd = &cobra.Command{
	Use:   "ulid",
	Short: "Generate a ULID (Crockford Base32, time-sortable)",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, ok := forge.Get(forge.CategoryID, "ulid")
		if !ok {
			return fmt.Errorf("generator not found: id/ulid")
		}

		if flags.GetBench(cmd) {
			return runBench(cmd, g)
		}

		timeFn := flags.ApplySeed(cmd)
		defer flags.CleanupSeed(cmd)
		of := flags.GetOutputFlags(cmd)
		opts := forge.Options{Count: 1, Format: of.ResolveFormat(), Time: timeFn}

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
