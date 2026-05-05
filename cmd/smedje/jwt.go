package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	_ "github.com/smedje/smedje/pkg/forge/crypto"

	"github.com/smedje/smedje/internal/flags"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	rootCmd.AddCommand(jwtCmd)
	jwtCmd.AddCommand(jwtHS256Cmd)
	jwtCmd.AddCommand(jwtRS256Cmd)
	jwtCmd.AddCommand(jwtES256Cmd)
	jwtCmd.AddCommand(jwtEdDSACmd)

	jwtHS256Cmd.Flags().Int("length", 32, "Secret length in bytes")
	flags.AddOutputFlags(jwtHS256Cmd)
	flags.AddBulkFlags(jwtHS256Cmd)
	flags.AddBenchFlag(jwtHS256Cmd)

	jwtRS256Cmd.Flags().Int("bits", 2048, "RSA key size in bits (2048 or 4096)")
	jwtRS256Cmd.Flags().String("kid", "", "Key ID (auto-generated if omitted)")
	flags.AddOutputFlags(jwtRS256Cmd)
	flags.AddBenchFlag(jwtRS256Cmd)

	jwtES256Cmd.Flags().String("kid", "", "Key ID (auto-generated if omitted)")
	flags.AddOutputFlags(jwtES256Cmd)
	flags.AddBenchFlag(jwtES256Cmd)

	jwtEdDSACmd.Flags().String("kid", "", "Key ID (auto-generated if omitted)")
	flags.AddOutputFlags(jwtEdDSACmd)
	flags.AddBenchFlag(jwtEdDSACmd)
}

var jwtCmd = &cobra.Command{
	Use:   "jwt",
	Short: "Generate JWT keys and secrets",
}

var jwtHS256Cmd = &cobra.Command{
	Use:   "hs256",
	Short: "Generate a JWT HS256 symmetric secret",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, ok := forge.Get(forge.CategoryCrypto, "hs256")
		if !ok {
			return fmt.Errorf("generator not found: crypto/hs256")
		}

		if flags.GetBench(cmd) {
			return runBench(cmd, g)
		}

		length, _ := cmd.Flags().GetInt("length")
		of := flags.GetOutputFlags(cmd)
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

var jwtRS256Cmd = &cobra.Command{
	Use:   "rs256",
	Short: "Generate a JWT RS256 RSA keypair with JWKS",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, ok := forge.Get(forge.CategoryCrypto, "rs256")
		if !ok {
			return fmt.Errorf("generator not found: crypto/rs256")
		}

		if flags.GetBench(cmd) {
			return runBench(cmd, g)
		}

		bits, _ := cmd.Flags().GetInt("bits")
		kid, _ := cmd.Flags().GetString("kid")
		of := flags.GetOutputFlags(cmd)
		params := map[string]string{"bits": fmt.Sprintf("%d", bits)}
		if kid != "" {
			params["kid"] = kid
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

var jwtES256Cmd = &cobra.Command{
	Use:   "es256",
	Short: "Generate a JWT ES256 ECDSA P-256 keypair with JWKS",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, ok := forge.Get(forge.CategoryCrypto, "es256")
		if !ok {
			return fmt.Errorf("generator not found: crypto/es256")
		}

		if flags.GetBench(cmd) {
			return runBench(cmd, g)
		}

		kid, _ := cmd.Flags().GetString("kid")
		of := flags.GetOutputFlags(cmd)
		params := map[string]string{}
		if kid != "" {
			params["kid"] = kid
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

var jwtEdDSACmd = &cobra.Command{
	Use:   "eddsa",
	Short: "Generate a JWT EdDSA Ed25519 keypair with JWKS",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, ok := forge.Get(forge.CategoryCrypto, "eddsa")
		if !ok {
			return fmt.Errorf("generator not found: crypto/eddsa")
		}

		if flags.GetBench(cmd) {
			return runBench(cmd, g)
		}

		kid, _ := cmd.Flags().GetString("kid")
		of := flags.GetOutputFlags(cmd)
		params := map[string]string{}
		if kid != "" {
			params["kid"] = kid
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
