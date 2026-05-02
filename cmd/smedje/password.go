package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	_ "github.com/smedje/smedje/pkg/forge/secret"

	"github.com/smedje/smedje/internal/output"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	rootCmd.AddCommand(passwordCmd)

	passwordCmd.Flags().Int("length", 24, "Password length (8-256)")
	passwordCmd.Flags().String("charset", "full", "Character set: full, alpha, alphanum, digits")
	passwordCmd.Flags().Bool("json", false, "Output as JSON")
	passwordCmd.Flags().Bool("quiet", false, "Output only the value")
	passwordCmd.Flags().Bool("bench", false, "Run a benchmark instead of generating")
}

var passwordCmd = &cobra.Command{
	Use:   "password",
	Short: "Generate a random password",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, ok := forge.Get(forge.CategorySecret, "password")
		if !ok {
			return fmt.Errorf("generator not found: secret/password")
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

		length, _ := cmd.Flags().GetInt("length")
		charset, _ := cmd.Flags().GetString("charset")

		out, err := g.Generate(cmd.Context(), forge.Options{
			Count:  1,
			Format: format,
			Params: map[string]string{
				"length":  fmt.Sprintf("%d", length),
				"charset": charset,
			},
		})
		if err != nil {
			return err
		}
		return output.Render(os.Stdout, out, format)
	},
}
