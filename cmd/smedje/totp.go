package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/smedje/smedje/internal/flags"
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
	flags.AddSeedFlags(totpCmd)
	flags.AddWhyFlag(totpCmd)
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
			return runBench(cmd, g)
		}

		flags.ApplySeed(cmd)
		defer flags.CleanupSeed(cmd)
		of := flags.GetOutputFlags(cmd)
		issuer, _ := cmd.Flags().GetString("issuer")
		account, _ := cmd.Flags().GetString("account")
		digits, _ := cmd.Flags().GetInt("digits")
		period, _ := cmd.Flags().GetInt("period")

		opts := forge.Options{
			Count:  1,
			Format: of.ResolveFormat(),
			Params: map[string]string{
				"issuer":  issuer,
				"account": account,
				"digits":  fmt.Sprintf("%d", digits),
				"period":  fmt.Sprintf("%d", period),
			},
		}
		if handled, err := flags.RunWhy(cmd, g, opts); handled {
			return err
		}

		return flags.RunGenerate(cmd.Context(), flags.RunOptions{
			Generator: g,
			Opts:      opts,
			Count:     flags.GetCount(cmd),
			Format:    of.ResolveFormat(),
			Writer:    os.Stdout,
		})
	},
}
