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
	tlsCmd.AddCommand(tlsCAChainCmd)
	tlsCmd.AddCommand(tlsMTLSCmd)
	tlsCmd.AddCommand(tlsRSACmd)
	tlsCmd.AddCommand(tlsECDSACmd)
	tlsCmd.AddCommand(tlsCSRCmd)

	tlsSelfSignedCmd.Flags().String("cn", "localhost", "Common name for the certificate")
	tlsSelfSignedCmd.Flags().Int("days", 825, "Certificate validity in days")
	tlsSelfSignedCmd.Flags().StringSlice("san", nil, "Subject alternative names (DNS or IP)")
	flags.AddOutputFlags(tlsSelfSignedCmd)
	flags.AddBulkFlags(tlsSelfSignedCmd)
	flags.AddBenchFlag(tlsSelfSignedCmd)
	flags.AddWhyFlag(tlsSelfSignedCmd)

	tlsCAChainCmd.Flags().String("cn", "My CA", "Common name base for the CA chain")
	tlsCAChainCmd.Flags().Int("days", 825, "Certificate validity in days")
	tlsCAChainCmd.Flags().Int("depth", 3, "Chain depth (2=root+leaf, 3=root+intermediate+leaf)")
	tlsCAChainCmd.Flags().StringSlice("san", nil, "Leaf SANs (DNS or IP, comma-separated)")
	flags.AddOutputFlags(tlsCAChainCmd)
	flags.AddBenchFlag(tlsCAChainCmd)

	tlsMTLSCmd.Flags().String("cn", "My CA", "Common name base for the mTLS bundle")
	tlsMTLSCmd.Flags().Int("days", 825, "Certificate validity in days")
	tlsMTLSCmd.Flags().StringSlice("san", nil, "Leaf SANs (DNS or IP, comma-separated)")
	flags.AddOutputFlags(tlsMTLSCmd)
	flags.AddBenchFlag(tlsMTLSCmd)

	tlsRSACmd.Flags().String("cn", "localhost", "Common name for the certificate")
	tlsRSACmd.Flags().Int("days", 825, "Certificate validity in days")
	tlsRSACmd.Flags().StringSlice("san", nil, "Subject alternative names (DNS or IP)")
	tlsRSACmd.Flags().Int("bits", 2048, "RSA key size (2048 or 4096)")
	flags.AddOutputFlags(tlsRSACmd)
	flags.AddBulkFlags(tlsRSACmd)
	flags.AddBenchFlag(tlsRSACmd)

	tlsECDSACmd.Flags().String("cn", "localhost", "Common name for the certificate")
	tlsECDSACmd.Flags().Int("days", 825, "Certificate validity in days")
	tlsECDSACmd.Flags().StringSlice("san", nil, "Subject alternative names (DNS or IP)")
	tlsECDSACmd.Flags().String("curve", "p256", "ECDSA curve (p256 or p384)")
	flags.AddOutputFlags(tlsECDSACmd)
	flags.AddBulkFlags(tlsECDSACmd)
	flags.AddBenchFlag(tlsECDSACmd)

	tlsCSRCmd.Flags().String("cn", "localhost", "Common name for the CSR")
	tlsCSRCmd.Flags().StringSlice("san", nil, "Subject alternative names (DNS or IP)")
	tlsCSRCmd.Flags().String("algo", "ed25519", "Key algorithm (ed25519, rsa, ecdsa)")
	tlsCSRCmd.Flags().String("org", "", "Organization name (O=)")
	tlsCSRCmd.Flags().String("country", "", "Two-letter country code (C=)")
	tlsCSRCmd.Flags().String("state", "", "State or province (ST=)")
	tlsCSRCmd.Flags().String("locality", "", "City or locality (L=)")
	flags.AddOutputFlags(tlsCSRCmd)
	flags.AddBulkFlags(tlsCSRCmd)
	flags.AddBenchFlag(tlsCSRCmd)
}

var tlsCmd = &cobra.Command{
	Use:   "tls",
	Short: "Generate TLS certificates",
}

var tlsCAChainCmd = &cobra.Command{
	Use:   "ca-chain",
	Short: "Generate a TLS CA chain (root → intermediate → leaf)",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, ok := forge.Get(forge.CategoryCrypto, "ca-chain")
		if !ok {
			return fmt.Errorf("generator not found: crypto/ca-chain")
		}

		if flags.GetBench(cmd) {
			return runBench(cmd, g)
		}

		of := flags.GetOutputFlags(cmd)
		cn, _ := cmd.Flags().GetString("cn")
		days, _ := cmd.Flags().GetInt("days")
		depth, _ := cmd.Flags().GetInt("depth")
		sans, _ := cmd.Flags().GetStringSlice("san")

		params := map[string]string{
			"cn":    cn,
			"days":  fmt.Sprintf("%d", days),
			"depth": fmt.Sprintf("%d", depth),
		}
		if len(sans) > 0 {
			params["san"] = strings.Join(sans, ",")
		}

		opts := forge.Options{Count: 1, Format: of.ResolveFormat(), Params: params}

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

		opts := forge.Options{Count: 1, Format: of.ResolveFormat(), Params: params}
		if handled, err := flags.RunWhy(cmd, g, opts); handled {
			return err
		}

		return flags.RunGenerate(cmd.Context(), flags.RunOptions{
			Generator: g,
			Opts:      opts,
			Count:     flags.GetCount(cmd),
			Format:    of.ResolveFormat(),
			OutputDir: of.OutputDir,
			Writer:    os.Stdout,
		})
	},
}

var tlsMTLSCmd = &cobra.Command{
	Use:   "mtls",
	Short: "Generate a mutual TLS bundle (CA + server + client certs)",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, ok := forge.Get(forge.CategoryCrypto, "mtls")
		if !ok {
			return fmt.Errorf("generator not found: crypto/mtls")
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

		opts := forge.Options{Count: 1, Format: of.ResolveFormat(), Params: params}

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

var tlsRSACmd = &cobra.Command{
	Use:   "rsa",
	Short: "Generate an RSA self-signed TLS certificate",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, ok := forge.Get(forge.CategoryCrypto, "rsa")
		if !ok {
			return fmt.Errorf("generator not found: crypto/rsa")
		}

		if flags.GetBench(cmd) {
			return runBench(cmd, g)
		}

		of := flags.GetOutputFlags(cmd)
		cn, _ := cmd.Flags().GetString("cn")
		days, _ := cmd.Flags().GetInt("days")
		bits, _ := cmd.Flags().GetInt("bits")
		sans, _ := cmd.Flags().GetStringSlice("san")

		params := map[string]string{
			"cn":   cn,
			"days": fmt.Sprintf("%d", days),
			"bits": fmt.Sprintf("%d", bits),
		}
		if len(sans) > 0 {
			params["san"] = strings.Join(sans, ",")
		}

		opts := forge.Options{Count: 1, Format: of.ResolveFormat(), Params: params}

		return flags.RunGenerate(cmd.Context(), flags.RunOptions{
			Generator: g,
			Opts:      opts,
			Count:     flags.GetCount(cmd),
			Format:    of.ResolveFormat(),
			OutputDir: of.OutputDir,
			Writer:    os.Stdout,
		})
	},
}

var tlsECDSACmd = &cobra.Command{
	Use:   "ecdsa",
	Short: "Generate an ECDSA self-signed TLS certificate",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, ok := forge.Get(forge.CategoryCrypto, "ecdsa")
		if !ok {
			return fmt.Errorf("generator not found: crypto/ecdsa")
		}

		if flags.GetBench(cmd) {
			return runBench(cmd, g)
		}

		of := flags.GetOutputFlags(cmd)
		cn, _ := cmd.Flags().GetString("cn")
		days, _ := cmd.Flags().GetInt("days")
		curve, _ := cmd.Flags().GetString("curve")
		sans, _ := cmd.Flags().GetStringSlice("san")

		params := map[string]string{
			"cn":    cn,
			"days":  fmt.Sprintf("%d", days),
			"curve": curve,
		}
		if len(sans) > 0 {
			params["san"] = strings.Join(sans, ",")
		}

		opts := forge.Options{Count: 1, Format: of.ResolveFormat(), Params: params}

		return flags.RunGenerate(cmd.Context(), flags.RunOptions{
			Generator: g,
			Opts:      opts,
			Count:     flags.GetCount(cmd),
			Format:    of.ResolveFormat(),
			OutputDir: of.OutputDir,
			Writer:    os.Stdout,
		})
	},
}

var tlsCSRCmd = &cobra.Command{
	Use:   "csr",
	Short: "Generate a TLS certificate signing request",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, ok := forge.Get(forge.CategoryCrypto, "csr")
		if !ok {
			return fmt.Errorf("generator not found: crypto/csr")
		}

		if flags.GetBench(cmd) {
			return runBench(cmd, g)
		}

		of := flags.GetOutputFlags(cmd)
		cn, _ := cmd.Flags().GetString("cn")
		algo, _ := cmd.Flags().GetString("algo")
		sans, _ := cmd.Flags().GetStringSlice("san")
		org, _ := cmd.Flags().GetString("org")
		country, _ := cmd.Flags().GetString("country")
		state, _ := cmd.Flags().GetString("state")
		locality, _ := cmd.Flags().GetString("locality")

		params := map[string]string{
			"cn":   cn,
			"algo": algo,
		}
		if len(sans) > 0 {
			params["san"] = strings.Join(sans, ",")
		}
		if org != "" {
			params["org"] = org
		}
		if country != "" {
			params["country"] = country
		}
		if state != "" {
			params["state"] = state
		}
		if locality != "" {
			params["locality"] = locality
		}

		opts := forge.Options{Count: 1, Format: of.ResolveFormat(), Params: params}

		return flags.RunGenerate(cmd.Context(), flags.RunOptions{
			Generator: g,
			Opts:      opts,
			Count:     flags.GetCount(cmd),
			Format:    of.ResolveFormat(),
			OutputDir: of.OutputDir,
			Writer:    os.Stdout,
		})
	},
}
