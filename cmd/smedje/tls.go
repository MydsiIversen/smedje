package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	_ "github.com/smedje/smedje/pkg/forge/tls"

	"github.com/smedje/smedje/internal/output"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	rootCmd.AddCommand(tlsCmd)
	tlsCmd.AddCommand(tlsSelfSignedCmd)

	tlsSelfSignedCmd.Flags().String("cn", "localhost", "Common name for the certificate")
	tlsSelfSignedCmd.Flags().Int("days", 365, "Certificate validity in days")
	tlsSelfSignedCmd.Flags().StringSlice("san", nil, "Subject alternative names (DNS or IP)")
	tlsSelfSignedCmd.Flags().Bool("json", false, "Output as JSON")
	tlsSelfSignedCmd.Flags().Bool("quiet", false, "Output only the values")
	tlsSelfSignedCmd.Flags().Bool("bench", false, "Run a benchmark instead of generating")
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

		benchFlag, _ := cmd.Flags().GetBool("bench")
		if benchFlag {
			result, err := g.Bench(cmd.Context())
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s: %d ops in %s (%.0f ops/sec)\n",
				result.Generator, result.Iterations, result.Duration, result.OpsPerSec)
			return nil
		}

		format := "text"
		if j, _ := cmd.Flags().GetBool("json"); j {
			format = "json"
		} else if q, _ := cmd.Flags().GetBool("quiet"); q {
			format = "quiet"
		}

		cn, _ := cmd.Flags().GetString("cn")
		days, _ := cmd.Flags().GetInt("days")
		sans, _ := cmd.Flags().GetStringSlice("san")

		params := map[string]string{
			"cn":   cn,
			"days": fmt.Sprintf("%d", days),
		}
		if len(sans) > 0 {
			joined := ""
			for i, s := range sans {
				if i > 0 {
					joined += ","
				}
				joined += s
			}
			params["san"] = joined
		}

		out, err := g.Generate(cmd.Context(), forge.Options{
			Count:  1,
			Format: format,
			Params: params,
		})
		if err != nil {
			return err
		}
		return output.Render(os.Stdout, out, format)
	},
}
