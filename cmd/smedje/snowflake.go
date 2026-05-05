package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/smedje/smedje/internal/flags"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	rootCmd.AddCommand(snowflakeCmd)

	snowflakeCmd.Flags().Int("worker", 0, "Worker ID (0-1023)")
	snowflakeCmd.Flags().String("epoch", "2024-01-01", "Custom epoch as YYYY-MM-DD")
	flags.AddOutputFlags(snowflakeCmd)
	flags.AddBulkFlags(snowflakeCmd)
	flags.AddBenchFlag(snowflakeCmd)
	flags.AddSeedFlags(snowflakeCmd)
	flags.AddWhyFlag(snowflakeCmd)
}

var snowflakeCmd = &cobra.Command{
	Use:   "snowflake",
	Short: "Generate a Twitter-style Snowflake ID (64-bit)",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, ok := forge.Get(forge.CategoryID, "snowflake")
		if !ok {
			return fmt.Errorf("generator not found: id/snowflake")
		}

		if flags.GetBench(cmd) {
			return runBench(cmd, g)
		}

		timeFn := flags.ApplySeed(cmd)
		defer flags.CleanupSeed(cmd)
		of := flags.GetOutputFlags(cmd)
		worker, _ := cmd.Flags().GetInt("worker")
		epoch, _ := cmd.Flags().GetString("epoch")

		opts := forge.Options{
			Count:  1,
			Format: of.ResolveFormat(),
			Params: map[string]string{
				"worker": fmt.Sprintf("%d", worker),
				"epoch":  epoch,
			},
			Time: timeFn,
		}

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
