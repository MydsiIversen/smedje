package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/smedje/smedje/internal/flags"
	"github.com/smedje/smedje/internal/output"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	rootCmd.AddCommand(totpCmd)

	totpCmd.Flags().String("issuer", "Smedje", "TOTP issuer label")
	totpCmd.Flags().String("account", "user@example.com", "TOTP account label")
	totpCmd.Flags().Int("digits", 6, "Code length (6 or 8)")
	totpCmd.Flags().Int("period", 30, "Time step in seconds")
	flags.AddOutputFlags(totpCmd)
	flags.AddBulkFlags(totpCmd)
	flags.AddBenchFlag(totpCmd)
}

var totpCmd = &cobra.Command{
	Use:   "totp",
	Short: "Generate a TOTP secret and otpauth URI",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, ok := forge.Get(forge.CategorySecret, "totp")
		if !ok {
			return fmt.Errorf("generator not found: secret/totp")
		}

		if flags.GetBench(cmd) {
			result, err := g.Bench(cmd.Context())
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s: %d ops in %s (%.0f ops/sec)\n",
				result.Generator, result.Iterations, result.Duration, result.OpsPerSec)
			return nil
		}

		of := flags.GetOutputFlags(cmd)
		count := flags.GetCount(cmd)
		issuer, _ := cmd.Flags().GetString("issuer")
		account, _ := cmd.Flags().GetString("account")
		digits, _ := cmd.Flags().GetInt("digits")
		period, _ := cmd.Flags().GetInt("period")

		for i := range count {
			out, err := g.Generate(cmd.Context(), forge.Options{
				Count:  1,
				Format: of.ResolveFormat(),
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
			if err := output.Render(os.Stdout, out, of.ResolveFormat()); err != nil {
				return err
			}
			_ = i
		}
		return nil
	},
}
