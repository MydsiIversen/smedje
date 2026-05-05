package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/smedje/smedje/internal/flags"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	// ipsec-psk
	rootCmd.AddCommand(ipsecPSKCmd)
	ipsecPSKCmd.Flags().Int("length", 32, "Key length in bytes (16-128)")
	flags.AddOutputFlags(ipsecPSKCmd)
	flags.AddBulkFlags(ipsecPSKCmd)
	flags.AddBenchFlag(ipsecPSKCmd)

	// radius-secret
	rootCmd.AddCommand(radiusSecretCmd)
	radiusSecretCmd.Flags().Int("length", 24, "Secret length in bytes")
	flags.AddOutputFlags(radiusSecretCmd)
	flags.AddBulkFlags(radiusSecretCmd)
	flags.AddBenchFlag(radiusSecretCmd)

	// snmp-community
	rootCmd.AddCommand(snmpCommunityCmd)
	snmpCommunityCmd.Flags().Int("length", 16, "Community string length")
	flags.AddOutputFlags(snmpCommunityCmd)
	flags.AddBulkFlags(snmpCommunityCmd)
	flags.AddBenchFlag(snmpCommunityCmd)

	// openvpn-tls-auth
	rootCmd.AddCommand(openvpnTLSAuthCmd)
	openvpnTLSAuthCmd.Flags().Int("bits", 2048, "Key size in bits")
	flags.AddOutputFlags(openvpnTLSAuthCmd)
	flags.AddBulkFlags(openvpnTLSAuthCmd)
	flags.AddBenchFlag(openvpnTLSAuthCmd)
}

var ipsecPSKCmd = &cobra.Command{
	Use:   "ipsec-psk",
	Short: "Generate an IPsec pre-shared key",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, ok := forge.Get(forge.CategorySecret, "ipsec-psk")
		if !ok {
			return fmt.Errorf("generator not found: secret/ipsec-psk")
		}
		if flags.GetBench(cmd) {
			return runBench(cmd, g)
		}
		of := flags.GetOutputFlags(cmd)
		length, _ := cmd.Flags().GetInt("length")
		opts := forge.Options{
			Count:  1,
			Format: of.ResolveFormat(),
			Params: map[string]string{"length": fmt.Sprintf("%d", length)},
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

var radiusSecretCmd = &cobra.Command{
	Use:   "radius-secret",
	Short: "Generate a RADIUS shared secret",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, ok := forge.Get(forge.CategorySecret, "radius-secret")
		if !ok {
			return fmt.Errorf("generator not found: secret/radius-secret")
		}
		if flags.GetBench(cmd) {
			return runBench(cmd, g)
		}
		of := flags.GetOutputFlags(cmd)
		length, _ := cmd.Flags().GetInt("length")
		opts := forge.Options{
			Count:  1,
			Format: of.ResolveFormat(),
			Params: map[string]string{"length": fmt.Sprintf("%d", length)},
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

var snmpCommunityCmd = &cobra.Command{
	Use:   "snmp-community",
	Short: "Generate an SNMP community string",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, ok := forge.Get(forge.CategorySecret, "snmp-community")
		if !ok {
			return fmt.Errorf("generator not found: secret/snmp-community")
		}
		if flags.GetBench(cmd) {
			return runBench(cmd, g)
		}
		of := flags.GetOutputFlags(cmd)
		length, _ := cmd.Flags().GetInt("length")
		opts := forge.Options{
			Count:  1,
			Format: of.ResolveFormat(),
			Params: map[string]string{"length": fmt.Sprintf("%d", length)},
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

var openvpnTLSAuthCmd = &cobra.Command{
	Use:   "openvpn-tls-auth",
	Short: "Generate an OpenVPN tls-auth key",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, ok := forge.Get(forge.CategoryCrypto, "openvpn-tls-auth")
		if !ok {
			return fmt.Errorf("generator not found: crypto/openvpn-tls-auth")
		}
		if flags.GetBench(cmd) {
			return runBench(cmd, g)
		}
		of := flags.GetOutputFlags(cmd)
		bits, _ := cmd.Flags().GetInt("bits")
		opts := forge.Options{
			Count:  1,
			Format: of.ResolveFormat(),
			Params: map[string]string{"bits": fmt.Sprintf("%d", bits)},
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
