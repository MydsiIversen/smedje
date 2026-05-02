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
