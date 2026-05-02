package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	_ "github.com/smedje/smedje/pkg/forge/id"

	"github.com/smedje/smedje/internal/flags"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	rootCmd.AddCommand(uuidCmd)
	uuidCmd.AddCommand(uuidV7Cmd)

	flags.AddOutputFlags(uuidV7Cmd)
	flags.AddBulkFlags(uuidV7Cmd)
	flags.AddBenchFlag(uuidV7Cmd)
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
