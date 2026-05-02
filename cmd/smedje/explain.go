package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/smedje/smedje/internal/explain"
)

func init() {
	rootCmd.AddCommand(explainCmd)
}

var explainCmd = &cobra.Command{
	Use:   "explain <value>",
	Short: "Identify the format of a value and decode embedded data",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		m := explain.Identify(args[0])
		if m == nil {
			fmt.Fprintln(cmd.OutOrStdout(), "Unknown format — could not identify this value.")
			return nil
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Format: %s\n", m.Format)
		for k, v := range m.Fields {
			if k == "format" {
				continue
			}
			fmt.Fprintf(cmd.OutOrStdout(), "  %s: %s\n", k, v)
		}
		fmt.Fprintln(cmd.OutOrStdout())
		fmt.Fprintln(cmd.OutOrStdout(), "Run `smedje recommend id` for guidance on which ID format to use.")
		return nil
	},
}
