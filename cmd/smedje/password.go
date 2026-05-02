package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	_ "github.com/smedje/smedje/pkg/forge/secret"

	"github.com/smedje/smedje/internal/flags"
	"github.com/smedje/smedje/internal/output"
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
		length, _ := cmd.Flags().GetInt("length")
		charset, _ := cmd.Flags().GetString("charset")

		for i := range count {
			out, err := g.Generate(cmd.Context(), forge.Options{
				Count:  1,
				Format: of.ResolveFormat(),
				Params: map[string]string{
					"length":  fmt.Sprintf("%d", length),
					"charset": charset,
				},
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
