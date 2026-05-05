package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/smedje/smedje/internal/flags"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	rootCmd.AddCommand(nanoidCmd)

	nanoidCmd.Flags().Int("length", 21, "ID length (1-256)")
	nanoidCmd.Flags().String("alphabet", "", "Custom alphabet (default: A-Za-z0-9_-)")
	flags.AddOutputFlags(nanoidCmd)
	flags.AddBulkFlags(nanoidCmd)
	flags.AddBenchFlag(nanoidCmd)
	flags.AddSeedFlags(nanoidCmd)
	flags.AddWhyFlag(nanoidCmd)
}

var nanoidCmd = &cobra.Command{
	Use:   "nanoid",
	Short: "Generate a NanoID (URL-safe, configurable)",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, ok := forge.Get(forge.CategoryID, "nanoid")
		if !ok {
			return fmt.Errorf("generator not found: id/nanoid")
		}

		if flags.GetBench(cmd) {
			return runBench(cmd, g)
		}

		timeFn := flags.ApplySeed(cmd)
		defer flags.CleanupSeed(cmd)
		of := flags.GetOutputFlags(cmd)
		length, _ := cmd.Flags().GetInt("length")
		alphabet, _ := cmd.Flags().GetString("alphabet")

		params := map[string]string{
			"length": fmt.Sprintf("%d", length),
		}
		if alphabet != "" {
			params["alphabet"] = alphabet
		}

		opts := forge.Options{Count: 1, Format: of.ResolveFormat(), Params: params, Time: timeFn}

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
