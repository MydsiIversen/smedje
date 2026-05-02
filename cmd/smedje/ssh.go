package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	_ "github.com/smedje/smedje/pkg/forge/ssh"

	"github.com/smedje/smedje/internal/flags"
	"github.com/smedje/smedje/internal/output"
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

		for i := range count {
			out, err := g.Generate(cmd.Context(), forge.Options{
				Count:  1,
				Format: of.ResolveFormat(),
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
