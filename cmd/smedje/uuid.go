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

	uuidV1Cmd := newUUIDSubcmd("v1", "Generate a UUIDv1 (RFC 9562, time-based with random node)")
	uuidV4Cmd := newUUIDSubcmd("v4", "Generate a UUIDv4 (RFC 9562, random)")
	uuidV6Cmd := newUUIDSubcmd("v6", "Generate a UUIDv6 (RFC 9562, reordered time)")
	uuidV7Cmd := newUUIDSubcmd("v7", "Generate a UUIDv7 (RFC 9562, time-ordered)")
	uuidV8Cmd := newUUIDSubcmd("v8", "Generate a UUIDv8 (RFC 9562, custom)")
	uuidNilCmd := newUUIDSubcmd("nil", "Output the nil UUID (all zeros)")
	uuidMaxCmd := newUUIDSubcmd("max", "Output the max UUID (all ones)")

	uuidCmd.AddCommand(uuidV1Cmd, uuidV4Cmd, uuidV6Cmd, uuidV7Cmd, uuidV8Cmd, uuidNilCmd, uuidMaxCmd)

	flags.AddOutputFlags(uuidCmd)
	flags.AddBulkFlags(uuidCmd)
	flags.AddBenchFlag(uuidCmd)
	flags.AddSeedFlags(uuidCmd)
}

var uuidCmd = &cobra.Command{
	Use:   "uuid",
	Short: "Generate UUIDs",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Bare "smedje uuid" defaults to v7.
		g, ok := forge.Get(forge.CategoryID, "v7")
		if !ok {
			return fmt.Errorf("generator not found: id/v7")
		}
		timeFn := flags.ApplySeed(cmd)
		defer flags.CleanupSeed(cmd)
		of := flags.GetOutputFlags(cmd)
		opts := forge.Options{Count: 1, Format: of.ResolveFormat(), Time: timeFn}
		return flags.RunGenerate(cmd.Context(), flags.RunOptions{
			Generator: g,
			Opts:      opts,
			Count:     flags.GetCount(cmd),
			Format:    of.ResolveFormat(),
			Writer:    os.Stdout,
		})
	},
}

func newUUIDSubcmd(name, desc string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: desc,
		RunE: func(cmd *cobra.Command, args []string) error {
			g, ok := forge.Get(forge.CategoryID, name)
			if !ok {
				return fmt.Errorf("generator not found: id/%s", name)
			}

			if flags.GetBench(cmd) {
				return runBench(cmd, g)
			}

			timeFn := flags.ApplySeed(cmd)
			defer flags.CleanupSeed(cmd)
			of := flags.GetOutputFlags(cmd)
			opts := forge.Options{Count: 1, Format: of.ResolveFormat(), Time: timeFn}
			return flags.RunGenerate(cmd.Context(), flags.RunOptions{
				Generator: g,
				Opts:      opts,
				Count:     flags.GetCount(cmd),
				Format:    of.ResolveFormat(),
				Writer:    os.Stdout,
			})
		},
	}
	flags.AddOutputFlags(cmd)
	flags.AddBulkFlags(cmd)
	flags.AddBenchFlag(cmd)
	flags.AddSeedFlags(cmd)
	return cmd
}
