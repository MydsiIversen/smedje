package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/smedje/smedje/pkg/forge"
)

func runBench(cmd *cobra.Command, g forge.Generator) error {
	result, err := g.Bench(cmd.Context())
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%s: %d ops in %s (%.0f ops/sec)\n",
		result.Generator, result.Iterations, result.Duration, result.OpsPerSec)
	return nil
}
