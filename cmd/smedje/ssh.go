package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	_ "github.com/smedje/smedje/pkg/forge/ssh"

	"github.com/smedje/smedje/internal/flags"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	rootCmd.AddCommand(sshCmd)
	sshCmd.AddCommand(sshEd25519Cmd)
	sshCmd.AddCommand(sshRSACmd)
	sshCmd.AddCommand(sshECDSACmd)

	flags.AddOutputFlags(sshEd25519Cmd)
	flags.AddBulkFlags(sshEd25519Cmd)
	flags.AddBenchFlag(sshEd25519Cmd)
	flags.AddWhyFlag(sshEd25519Cmd)

	sshRSACmd.Flags().Int("bits", 4096, "RSA key size (2048 or 4096)")
	flags.AddOutputFlags(sshRSACmd)
	flags.AddBulkFlags(sshRSACmd)
	flags.AddBenchFlag(sshRSACmd)

	sshECDSACmd.Flags().String("curve", "p256", "ECDSA curve (p256 or p384)")
	flags.AddOutputFlags(sshECDSACmd)
	flags.AddBulkFlags(sshECDSACmd)
	flags.AddBenchFlag(sshECDSACmd)
}

var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "Generate SSH keys",
}

var sshEd25519Cmd = &cobra.Command{
	Use:   "ed25519",
	Short: "Generate an Ed25519 OpenSSH keypair",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, ok := forge.Get(forge.CategoryCrypto, "ed25519")
		if !ok {
			return fmt.Errorf("generator not found: crypto/ed25519")
		}

		if flags.GetBench(cmd) {
			return runBench(cmd, g)
		}

		of := flags.GetOutputFlags(cmd)
		opts := forge.Options{Count: 1, Format: of.ResolveFormat()}

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

var sshRSACmd = &cobra.Command{
	Use:   "rsa",
	Short: "Generate an RSA OpenSSH keypair",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, ok := forge.Get(forge.CategoryCrypto, "ssh-rsa")
		if !ok {
			return fmt.Errorf("generator not found: crypto/ssh-rsa")
		}

		if flags.GetBench(cmd) {
			return runBench(cmd, g)
		}

		bits, _ := cmd.Flags().GetInt("bits")
		of := flags.GetOutputFlags(cmd)
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

var sshECDSACmd = &cobra.Command{
	Use:   "ecdsa",
	Short: "Generate an ECDSA OpenSSH keypair",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, ok := forge.Get(forge.CategoryCrypto, "ssh-ecdsa")
		if !ok {
			return fmt.Errorf("generator not found: crypto/ssh-ecdsa")
		}

		if flags.GetBench(cmd) {
			return runBench(cmd, g)
		}

		curve, _ := cmd.Flags().GetString("curve")
		of := flags.GetOutputFlags(cmd)
		opts := forge.Options{
			Count:  1,
			Format: of.ResolveFormat(),
			Params: map[string]string{"curve": curve},
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
