package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/smedje/smedje/internal/explain"
)

func init() {
	rootCmd.AddCommand(explainCmd)
}

func recommendHint(format string) string {
	switch {
	case contains(format, "UUID", "ULID", "Snowflake", "NanoID"):
		return "Run `smedje recommend id` for guidance on which ID format to use."
	case contains(format, "SSH"):
		return "Run `smedje recommend ssh-key` for guidance on which key type to use."
	case contains(format, "PEM", "X.509", "Certificate"):
		return "Run `smedje recommend tls-cert` for guidance on certificate generation."
	case contains(format, "JWT"):
		return "Run `smedje recommend jwt` for guidance on which JWT algorithm to use."
	case contains(format, "age"):
		return "Run `smedje recommend age` for guidance on file encryption."
	case contains(format, "WireGuard"):
		return "Run `smedje recommend vpn-key` for guidance on VPN key generation."
	case contains(format, "MAC"):
		return "Run `smedje recommend storage-id` for guidance on network identifiers."
	case contains(format, "IQN", "iSCSI"):
		return "Run `smedje recommend storage-id` for guidance on storage identifiers."
	default:
		return ""
	}
}

func contains(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if len(s) >= len(sub) {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
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
		if hint := recommendHint(m.Format); hint != "" {
			fmt.Fprintln(cmd.OutOrStdout(), hint)
		}
		return nil
	},
}
