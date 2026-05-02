package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "smedje",
	Short: "Forge keys, IDs, certs, and configs from scratch",
	Long:  "Smedje is a toolkit for generating cryptographic keys, identifiers, certificates, and network configurations.",
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
