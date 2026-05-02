package flags

import "github.com/spf13/cobra"

// AddBulkFlags registers --count on a command.
func AddBulkFlags(cmd *cobra.Command) {
	cmd.Flags().IntP("count", "n", 1, "Number of items to generate")
}

// GetCount reads the --count flag value. Returns the raw value; validation
// happens in RunGenerate.
func GetCount(cmd *cobra.Command) int {
	n, _ := cmd.Flags().GetInt("count")
	return n
}
