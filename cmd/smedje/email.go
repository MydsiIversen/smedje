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
	emailDKIMCmd.Flags().String("hash", "sha256", "Hash algorithm (sha256, sha1)")
	flags.AddOutputFlags(emailDKIMCmd)
	flags.AddBenchFlag(emailDKIMCmd)

	emailDMARCCmd.Flags().String("domain", "", "Domain name [required]")
	emailDMARCCmd.Flags().String("policy", "none", "DMARC policy (none, quarantine, reject)")
	emailDMARCCmd.Flags().String("sp", "", "Subdomain policy (defaults to main policy)")
	emailDMARCCmd.Flags().String("adkim", "", "DKIM alignment: r (relaxed) or s (strict)")
	emailDMARCCmd.Flags().String("aspf", "", "SPF alignment: r (relaxed) or s (strict)")
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
		hash, _ := cmd.Flags().GetString("hash")
		of := flags.GetOutputFlags(cmd)
		opts := forge.Options{
			Count:  1,
			Format: of.ResolveFormat(),
			Params: map[string]string{
				"selector": selector,
				"domain":   domain,
				"bits":     fmt.Sprintf("%d", bits),
				"hash":     hash,
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
		sp, _ := cmd.Flags().GetString("sp")
		adkim, _ := cmd.Flags().GetString("adkim")
		aspf, _ := cmd.Flags().GetString("aspf")
		rua, _ := cmd.Flags().GetString("rua")
		ruf, _ := cmd.Flags().GetString("ruf")
		of := flags.GetOutputFlags(cmd)
		params := map[string]string{
			"domain": domain,
			"policy": policy,
		}
		if sp != "" {
			params["sp"] = sp
		}
		if adkim != "" {
			params["adkim"] = adkim
		}
		if aspf != "" {
			params["aspf"] = aspf
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
