package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	_ "github.com/smedje/smedje/pkg/forge/tls"

	"github.com/smedje/smedje/internal/flags"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	rootCmd.AddCommand(tlsCmd)
	tlsCmd.AddCommand(tlsSelfSignedCmd)

	tlsSelfSignedCmd.Flags().String("cn", "localhost", "Common name for the certificate")
	tlsSelfSignedCmd.Flags().Int("days", 365, "Certificate validity in days")
	tlsSelfSignedCmd.Flags().StringSlice("san", nil, "Subject alternative names (DNS or IP)")
	flags.AddOutputFlags(tlsSelfSignedCmd)
	flags.AddBulkFlags(tlsSelfSignedCmd)
	flags.AddBenchFlag(tlsSelfSignedCmd)
}

var tlsCmd = &cobra.Command{
	Use:   "tls",
	Short: "Generate TLS certificates",
}

var tlsSelfSignedCmd = &cobra.Command{
	Use:   "self-signed",
	Short: "Generate a self-signed TLS certificate",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, ok := forge.Get(forge.CategoryCrypto, "self-signed")
		if !ok {
			return fmt.Errorf("generator not found: crypto/self-signed")
		}

		if flags.GetBench(cmd) {
			return runBench(cmd, g)
		}

		of := flags.GetOutputFlags(cmd)
		cn, _ := cmd.Flags().GetString("cn")
		days, _ := cmd.Flags().GetInt("days")
		sans, _ := cmd.Flags().GetStringSlice("san")

		params := map[string]string{
			"cn":   cn,
			"days": fmt.Sprintf("%d", days),
		}
		if len(sans) > 0 {
			params["san"] = strings.Join(sans, ",")
		}

		return flags.RunGenerate(cmd.Context(), flags.RunOptions{
			Generator: g,
			Opts: forge.Options{
				Count:  1,
				Format: of.ResolveFormat(),
				Params: params,
			},
			Count:  flags.GetCount(cmd),
			Format: of.ResolveFormat(),
			Writer: os.Stdout,
		})
	},
}
