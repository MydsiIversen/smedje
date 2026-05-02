package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/smedje/smedje/internal/recommend"
)

func init() {
	rootCmd.AddCommand(recommendCmd)

	recommendCmd.Flags().String("use-case", "", "Filter to a specific use case")
	recommendCmd.Flags().Bool("json", false, "Output as JSON")
	recommendCmd.Flags().Bool("markdown", false, "Output as Markdown")
}

var recommendCmd = &cobra.Command{
	Use:   "recommend <topic>",
	Short: "Opinionated recommendations for common use cases",
	Long: `Available topics: id, ssh-key, tls-cert, password, hash, jwt, secret, vpn-key

Examples:
  smedje recommend id
  smedje recommend id --use-case "user-facing API"
  smedje recommend ssh-key --json`,
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"id", "ssh-key", "tls-cert", "password", "hash", "jwt", "secret", "vpn-key"},
	RunE: func(cmd *cobra.Command, args []string) error {
		topic := args[0]
		recs, ok := recommend.Recommendations[topic]
		if !ok {
			return fmt.Errorf("unknown topic %q. Available: %s", topic, strings.Join(recommend.Topics(), ", "))
		}

		useCase, _ := cmd.Flags().GetString("use-case")
		if useCase != "" {
			filtered := recommend.FilterByUseCase(recs, useCase)
			if len(filtered) == 0 {
				var cases []string
				for _, r := range recs {
					cases = append(cases, r.UseCase)
				}
				return fmt.Errorf("no use case matching %q. Available:\n  %s",
					useCase, strings.Join(cases, "\n  "))
			}
			recs = filtered
		}

		jsonFlag, _ := cmd.Flags().GetBool("json")
		mdFlag, _ := cmd.Flags().GetBool("markdown")

		if jsonFlag {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(recs)
		}
		if mdFlag {
			return renderRecommendationsMD(topic, recs)
		}
		return renderRecommendationsText(topic, recs)
	},
}

func renderRecommendationsText(topic string, recs []recommend.Recommendation) error {
	fmt.Printf("Recommendations: %s\n", topic)
	fmt.Println(strings.Repeat("─", 60))
	for i, r := range recs {
		if i > 0 {
			fmt.Println()
		}
		fmt.Printf("  Use case: %s\n", r.UseCase)
		fmt.Printf("  Recommended: %s\n", r.Primary)
		fmt.Printf("  Why: %s\n", r.Why)
		fmt.Printf("  Command: %s\n", r.Command)
		if len(r.Alternatives) > 0 {
			fmt.Printf("  Alternatives:\n")
			for _, a := range r.Alternatives {
				fmt.Printf("    - %s — %s\n", a.Name, a.When)
			}
		}
		if len(r.Avoid) > 0 {
			fmt.Printf("  Avoid: %s\n", strings.Join(r.Avoid, "; "))
		}
	}
	return nil
}

func renderRecommendationsMD(topic string, recs []recommend.Recommendation) error {
	fmt.Printf("# Recommendations: %s\n\n", topic)
	for _, r := range recs {
		fmt.Printf("## %s\n\n", r.UseCase)
		fmt.Printf("**Recommended:** %s\n\n", r.Primary)
		fmt.Printf("**Why:** %s\n\n", r.Why)
		fmt.Printf("**Command:** `%s`\n\n", r.Command)
		if len(r.Alternatives) > 0 {
			fmt.Printf("**Alternatives:**\n\n")
			for _, a := range r.Alternatives {
				fmt.Printf("- %s — %s\n", a.Name, a.When)
			}
			fmt.Println()
		}
		if len(r.Avoid) > 0 {
			fmt.Printf("**Avoid:** %s\n\n", strings.Join(r.Avoid, "; "))
		}
	}
	return nil
}
