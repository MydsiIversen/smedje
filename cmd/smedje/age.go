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
	rootCmd.AddCommand(ageCmd)
	ageCmd.AddCommand(ageX25519Cmd)

	flags.AddOutputFlags(ageX25519Cmd)
	flags.AddBulkFlags(ageX25519Cmd)
	flags.AddBenchFlag(ageX25519Cmd)
}

var ageCmd = &cobra.Command{
	Use:   "age",
	Short: "Generate age encryption keys",
}

var ageX25519Cmd = &cobra.Command{
	Use:   "x25519",
	Short: "Generate an age X25519 keypair",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, ok := forge.Get(forge.CategoryCrypto, "x25519")
		if !ok {
			return fmt.Errorf("generator not found: crypto/x25519")
		}

		if flags.GetBench(cmd) {
			return runBench(cmd, g)
		}

		of := flags.GetOutputFlags(cmd)
		opts := forge.Options{Count: 1, Format: of.ResolveFormat()}

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
