package main

import (
	"fmt"
	"runtime"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print build version and system information",
	Run: func(cmd *cobra.Command, args []string) {
		goVersion := runtime.Version()
		if info, ok := debug.ReadBuildInfo(); ok && info.GoVersion != "" {
			goVersion = info.GoVersion
		}

		fmt.Fprintf(cmd.OutOrStdout(), "smedje %s\n", version)
		fmt.Fprintf(cmd.OutOrStdout(), "  commit:  %s\n", commit)
		fmt.Fprintf(cmd.OutOrStdout(), "  built:   %s\n", date)
		fmt.Fprintf(cmd.OutOrStdout(), "  go:      %s\n", goVersion)
		fmt.Fprintf(cmd.OutOrStdout(), "  os/arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}
