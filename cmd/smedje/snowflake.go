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
	flags.AddOutputFlags(snowflakeCmd)
	flags.AddBulkFlags(snowflakeCmd)
	flags.AddBenchFlag(snowflakeCmd)
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

		of := flags.GetOutputFlags(cmd)
		worker, _ := cmd.Flags().GetInt("worker")

		return flags.RunGenerate(cmd.Context(), flags.RunOptions{
			Generator: g,
			Opts: forge.Options{
				Count:  1,
				Format: of.ResolveFormat(),
				Params: map[string]string{"worker": fmt.Sprintf("%d", worker)},
			},
			Count:  flags.GetCount(cmd),
			Format: of.ResolveFormat(),
			Writer: os.Stdout,
		})
	},
}
