package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/smedje/smedje/internal/output"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	rootCmd.AddCommand(totpCmd)

	totpCmd.Flags().String("issuer", "Smedje", "TOTP issuer label")
	totpCmd.Flags().String("account", "user@example.com", "TOTP account label")
	totpCmd.Flags().Int("digits", 6, "Code length (6 or 8)")
	totpCmd.Flags().Int("period", 30, "Time step in seconds")
	totpCmd.Flags().Bool("json", false, "Output as JSON")
	totpCmd.Flags().Bool("quiet", false, "Output only the values")
	totpCmd.Flags().Bool("bench", false, "Run a benchmark instead of generating")
}

var totpCmd = &cobra.Command{
	Use:   "totp",
	Short: "Generate a TOTP secret and otpauth URI",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, ok := forge.Get(forge.CategorySecret, "totp")
		if !ok {
			return fmt.Errorf("generator not found: secret/totp")
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

		issuer, _ := cmd.Flags().GetString("issuer")
		account, _ := cmd.Flags().GetString("account")
		digits, _ := cmd.Flags().GetInt("digits")
		period, _ := cmd.Flags().GetInt("period")

		out, err := g.Generate(cmd.Context(), forge.Options{
			Count:  1,
			Format: format,
			Params: map[string]string{
				"issuer":  issuer,
				"account": account,
				"digits":  fmt.Sprintf("%d", digits),
				"period":  fmt.Sprintf("%d", period),
			},
		})
		if err != nil {
			return err
		}
		return output.Render(os.Stdout, out, format)
	},
}
