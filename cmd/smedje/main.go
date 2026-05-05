package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/smedje/smedje/internal/config"
)

var envFileFlag string

var rootCmd = &cobra.Command{
	Use:   "smedje",
	Short: "Forge keys, IDs, certs, and configs from scratch",
	Long:  "Smedje is a toolkit for generating cryptographic keys, identifiers, certificates, and network configurations.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(config.LoadOptions{
			EnvFilePath: envFileFlag,
		})
		if err != nil {
			return err
		}
		config.SetGlobal(cfg)
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&envFileFlag, "env-file", "", "Load environment overrides from a .env file")
}

func main() {
	registerCompletions()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
