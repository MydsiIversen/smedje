package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/smedje/smedje/internal/bench"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	rootCmd.AddCommand(benchCmd)
	benchCmd.AddCommand(benchAllCmd)
	benchCmd.AddCommand(benchCompareCmd)
	benchCmd.AddCommand(benchListCmd)

	for _, cmd := range []*cobra.Command{benchCmd, benchAllCmd, benchCompareCmd} {
		cmd.Flags().Duration("duration", 2*time.Second, "Benchmark duration per generator")
		cmd.Flags().Duration("warmup", 500*time.Millisecond, "Warmup duration")
		cmd.Flags().Int("repeat", 1, "Number of repeated measurements")
		cmd.Flags().Int("cores", 0, "Number of CPU cores (0 = all)")
		cmd.Flags().Bool("json", false, "Output as JSON")
		cmd.Flags().Bool("markdown", false, "Output as Markdown table")
	}
}

var benchCmd = &cobra.Command{
	Use:   "bench [generator]",
	Short: "Benchmark generators",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}

		generators, err := resolveGenerators(args[0])
		if err != nil {
			return err
		}

		opts := benchOptsFromCmd(cmd)
		results, err := runBenchmarks(cmd, generators, opts)
		if err != nil {
			return err
		}
		return renderBenchResults(cmd, results)
	},
}

var benchAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Benchmark all generators",
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := benchOptsFromCmd(cmd)
		results, err := runBenchmarks(cmd, forge.All(), opts)
		if err != nil {
			return err
		}
		return renderBenchResults(cmd, results)
	},
}

var benchCompareCmd = &cobra.Command{
	Use:   "compare <gen1> <gen2> [gen3...]",
	Short: "Benchmark multiple generators side by side",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		var generators []forge.Generator
		for _, name := range args {
			found, err := resolveGenerators(name)
			if err != nil {
				return err
			}
			generators = append(generators, found...)
		}

		opts := benchOptsFromCmd(cmd)
		results, err := runBenchmarks(cmd, generators, opts)
		if err != nil {
			return err
		}
		return renderBenchResults(cmd, results)
	},
}

func benchOptsFromCmd(cmd *cobra.Command) bench.Options {
	d, _ := cmd.Flags().GetDuration("duration")
	w, _ := cmd.Flags().GetDuration("warmup")
	r, _ := cmd.Flags().GetInt("repeat")
	c, _ := cmd.Flags().GetInt("cores")
	return bench.Options{Duration: d, Warmup: w, Repeat: r, Cores: c}
}

var benchListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all addressable generator names",
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, addr := range forge.Addresses() {
			fmt.Println(addr)
		}
		return nil
	},
}

func resolveGenerators(pattern string) ([]forge.Generator, error) {
	return forge.Resolve(pattern)
}

func runBenchmarks(cmd *cobra.Command, generators []forge.Generator, opts bench.Options) ([]*bench.Result, error) {
	var results []*bench.Result
	for _, g := range generators {
		addr := forge.Address(g)
		fmt.Fprintf(cmd.ErrOrStderr(), "Benchmarking %s...\n", addr)
		r, err := bench.Run(cmd.Context(), g, opts)
		if err != nil {
			return nil, fmt.Errorf("bench %s: %w", addr, err)
		}
		results = append(results, r)
	}
	return results, nil
}

func renderBenchResults(cmd *cobra.Command, results []*bench.Result) error {
	jsonFlag, _ := cmd.Flags().GetBool("json")
	mdFlag, _ := cmd.Flags().GetBool("markdown")

	if jsonFlag {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(results)
	}

	if mdFlag {
		return renderMarkdownTable(results)
	}

	return renderTextTable(results)
}

func renderTextTable(results []*bench.Result) error {
	sort.Slice(results, func(i, j int) bool {
		return results[i].OpsPerSec > results[j].OpsPerSec
	})

	fmt.Printf("%-20s %12s %12s %10s %10s", "Generator", "ops/sec", "ns/op", "allocs/op", "B/op")
	if results[0].Repeats > 1 {
		fmt.Printf(" %8s", "CV")
	}
	fmt.Println()
	fmt.Println(strings.Repeat("─", 72))

	for _, r := range results {
		fmt.Printf("%-20s %12.0f %12.1f %10.1f %10d",
			r.Generator, r.OpsPerSec, r.NsPerOp, r.AllocsPerOp, r.BytesPerOp)
		if r.Repeats > 1 {
			fmt.Printf(" %7.1f%%", r.VarianceCV)
		}
		fmt.Println()
	}
	return nil
}

func renderMarkdownTable(results []*bench.Result) error {
	sort.Slice(results, func(i, j int) bool {
		return results[i].OpsPerSec > results[j].OpsPerSec
	})

	fmt.Println("| Generator | ops/sec | ns/op | allocs/op | B/op |")
	fmt.Println("|-----------|---------|-------|-----------|------|")
	for _, r := range results {
		fmt.Printf("| %s | %.0f | %.1f | %.1f | %d |\n",
			r.Generator, r.OpsPerSec, r.NsPerOp, r.AllocsPerOp, r.BytesPerOp)
	}
	return nil
}
