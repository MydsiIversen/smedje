package flags

import "github.com/spf13/cobra"

// OutputFlags holds values from the shared output flag group.
type OutputFlags struct {
	Format  string
	Quiet   bool
	JSON    bool
	NoColor bool
}

// AddOutputFlags registers output-related flags on a command.
func AddOutputFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("format", "f", "text", "Output format: text, json, csv, env, sql")
	cmd.Flags().BoolP("quiet", "q", false, "Output only the raw value(s)")
	cmd.Flags().Bool("json", false, "Shorthand for --format json")
	cmd.Flags().Bool("no-color", false, "Disable colored output")
}

// GetOutputFlags reads output flag values from a command.
func GetOutputFlags(cmd *cobra.Command) OutputFlags {
	format, _ := cmd.Flags().GetString("format")
	quiet, _ := cmd.Flags().GetBool("quiet")
	jsonFlag, _ := cmd.Flags().GetBool("json")
	noColor, _ := cmd.Flags().GetBool("no-color")

	if jsonFlag {
		format = "json"
	}
	if quiet {
		format = "quiet"
	}

	return OutputFlags{
		Format:  format,
		Quiet:   quiet,
		JSON:    jsonFlag,
		NoColor: noColor,
	}
}

// ResolveFormat returns the effective format string.
func (o OutputFlags) ResolveFormat() string {
	return o.Format
}
