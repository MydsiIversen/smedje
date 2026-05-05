package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	_ "github.com/smedje/smedje/pkg/forge/email"

	"github.com/smedje/smedje/internal/flags"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	rootCmd.AddCommand(emailCmd)
	emailCmd.AddCommand(emailDKIMCmd)
	emailCmd.AddCommand(emailDMARCCmd)

	emailDKIMCmd.Flags().String("selector", "", "DKIM selector (e.g., mail) [required]")
	emailDKIMCmd.Flags().String("domain", "", "Domain name (e.g., example.com) [required]")
	emailDKIMCmd.Flags().Int("bits", 2048, "RSA key size")
	flags.AddOutputFlags(emailDKIMCmd)
	flags.AddBenchFlag(emailDKIMCmd)

	emailDMARCCmd.Flags().String("domain", "", "Domain name [required]")
	emailDMARCCmd.Flags().String("policy", "none", "DMARC policy (none, quarantine, reject)")
	emailDMARCCmd.Flags().String("rua", "", "Aggregate report email address")
	emailDMARCCmd.Flags().String("ruf", "", "Forensic report email address")
	flags.AddOutputFlags(emailDMARCCmd)
	flags.AddBenchFlag(emailDMARCCmd)
}

var emailCmd = &cobra.Command{
	Use:   "email",
	Short: "Generate email authentication records (DKIM, DMARC)",
}

var emailDKIMCmd = &cobra.Command{
	Use:   "dkim",
	Short: "Generate a DKIM keypair with DNS TXT record",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, ok := forge.Get(forge.CategoryCrypto, "dkim")
		if !ok {
			return fmt.Errorf("generator not found: crypto/dkim")
		}

		if flags.GetBench(cmd) {
			return runBench(cmd, g)
		}

		selector, _ := cmd.Flags().GetString("selector")
		domain, _ := cmd.Flags().GetString("domain")
		bits, _ := cmd.Flags().GetInt("bits")
		of := flags.GetOutputFlags(cmd)
		opts := forge.Options{
			Count:  1,
			Format: of.ResolveFormat(),
			Params: map[string]string{
				"selector": selector,
				"domain":   domain,
				"bits":     fmt.Sprintf("%d", bits),
			},
		}

		return flags.RunGenerate(cmd.Context(), flags.RunOptions{
			Generator: g,
			Opts:      opts,
			Count:     1,
			Format:    of.ResolveFormat(),
			OutputDir: of.OutputDir,
			Writer:    os.Stdout,
		})
	},
}

var emailDMARCCmd = &cobra.Command{
	Use:   "dmarc",
	Short: "Generate a DMARC DNS TXT record",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, ok := forge.Get(forge.CategoryNetwork, "dmarc")
		if !ok {
			return fmt.Errorf("generator not found: network/dmarc")
		}

		if flags.GetBench(cmd) {
			return runBench(cmd, g)
		}

		domain, _ := cmd.Flags().GetString("domain")
		policy, _ := cmd.Flags().GetString("policy")
		rua, _ := cmd.Flags().GetString("rua")
		ruf, _ := cmd.Flags().GetString("ruf")
		of := flags.GetOutputFlags(cmd)
		params := map[string]string{
			"domain": domain,
			"policy": policy,
		}
		if rua != "" {
			params["rua"] = rua
		}
		if ruf != "" {
			params["ruf"] = ruf
		}
		opts := forge.Options{
			Count:  1,
			Format: of.ResolveFormat(),
			Params: params,
		}

		return flags.RunGenerate(cmd.Context(), flags.RunOptions{
			Generator: g,
			Opts:      opts,
			Count:     1,
			Format:    of.ResolveFormat(),
			OutputDir: of.OutputDir,
			Writer:    os.Stdout,
		})
	},
}
