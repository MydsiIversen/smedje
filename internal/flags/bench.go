package flags

import "github.com/spf13/cobra"

// AddBenchFlag registers --bench on a command.
func AddBenchFlag(cmd *cobra.Command) {
	cmd.Flags().Bool("bench", false, "Run a benchmark instead of generating")
}

// GetBench reads the --bench flag.
func GetBench(cmd *cobra.Command) bool {
	b, _ := cmd.Flags().GetBool("bench")
	return b
}
