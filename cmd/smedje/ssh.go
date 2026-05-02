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

	flags.AddOutputFlags(sshEd25519Cmd)
	flags.AddBulkFlags(sshEd25519Cmd)
	flags.AddBenchFlag(sshEd25519Cmd)
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
		return flags.RunGenerate(cmd.Context(), flags.RunOptions{
			Generator: g,
			Opts:      forge.Options{Count: 1, Format: of.ResolveFormat()},
			Count:     flags.GetCount(cmd),
			Format:    of.ResolveFormat(),
			Writer:    os.Stdout,
		})
	},
}
