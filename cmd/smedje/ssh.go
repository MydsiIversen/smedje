package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	_ "github.com/smedje/smedje/pkg/forge/ssh"

	"github.com/smedje/smedje/internal/output"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	rootCmd.AddCommand(sshCmd)
	sshCmd.AddCommand(sshEd25519Cmd)

	sshEd25519Cmd.Flags().Bool("json", false, "Output as JSON")
	sshEd25519Cmd.Flags().Bool("quiet", false, "Output only the values")
	sshEd25519Cmd.Flags().Bool("bench", false, "Run a benchmark instead of generating")
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

		benchFlag, _ := cmd.Flags().GetBool("bench")
		if benchFlag {
			result, err := g.Bench(cmd.Context())
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s: %d ops in %s (%.0f ops/sec)\n",
				result.Generator, result.Iterations, result.Duration, result.OpsPerSec)
			return nil
		}

		format := "text"
		if j, _ := cmd.Flags().GetBool("json"); j {
			format = "json"
		} else if q, _ := cmd.Flags().GetBool("quiet"); q {
			format = "quiet"
		}

		out, err := g.Generate(cmd.Context(), forge.Options{Count: 1, Format: format})
		if err != nil {
			return err
		}
		return output.Render(os.Stdout, out, format)
	},
}
