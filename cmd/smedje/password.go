package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	_ "github.com/smedje/smedje/pkg/forge/secret"

	"github.com/smedje/smedje/internal/flags"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	rootCmd.AddCommand(passwordCmd)

	passwordCmd.Flags().Int("length", 24, "Password length (8-256)")
	passwordCmd.Flags().String("charset", "full", "Character set: full, alpha, alphanum, digits")
	flags.AddOutputFlags(passwordCmd)
	flags.AddBulkFlags(passwordCmd)
	flags.AddBenchFlag(passwordCmd)
}

var passwordCmd = &cobra.Command{
	Use:   "password",
	Short: "Generate a random password",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, ok := forge.Get(forge.CategorySecret, "password")
		if !ok {
			return fmt.Errorf("generator not found: secret/password")
		}

		if flags.GetBench(cmd) {
			return runBench(cmd, g)
		}

		of := flags.GetOutputFlags(cmd)
		length, _ := cmd.Flags().GetInt("length")
		charset, _ := cmd.Flags().GetString("charset")

		return flags.RunGenerate(cmd.Context(), flags.RunOptions{
			Generator: g,
			Opts: forge.Options{
				Count:  1,
				Format: of.ResolveFormat(),
				Params: map[string]string{
					"length":  fmt.Sprintf("%d", length),
					"charset": charset,
				},
			},
			Count:  flags.GetCount(cmd),
			Format: of.ResolveFormat(),
			Writer: os.Stdout,
		})
	},
}
