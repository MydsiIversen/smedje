package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	_ "github.com/smedje/smedje/pkg/forge/wireguard"

	"github.com/smedje/smedje/internal/flags"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	rootCmd.AddCommand(wireguardCmd)
	wireguardCmd.AddCommand(wireguardKeypairCmd)
	wireguardCmd.AddCommand(wireguardMeshCmd)

	flags.AddOutputFlags(wireguardKeypairCmd)
	flags.AddBulkFlags(wireguardKeypairCmd)
	flags.AddBenchFlag(wireguardKeypairCmd)
	flags.AddWhyFlag(wireguardKeypairCmd)

	wireguardMeshCmd.Flags().Int("peers", 3, "Number of peers in the mesh (2-255)")
	wireguardMeshCmd.Flags().String("subnet", "10.0.0.0/24", "Mesh subnet (e.g. 10.10.0.0/24)")
	wireguardMeshCmd.Flags().Int("port", 51820, "WireGuard listen port")
	wireguardMeshCmd.Flags().String("endpoint", "", "Peer endpoints (comma-separated host:port list)")
	wireguardMeshCmd.Flags().String("dns", "", "DNS server for the Interface section")
	wireguardMeshCmd.Flags().Int("keepalive", 0, "PersistentKeepalive interval in seconds (0 = disabled)")
	flags.AddOutputFlags(wireguardMeshCmd)
	flags.AddBenchFlag(wireguardMeshCmd)
}

var wireguardCmd = &cobra.Command{
	Use:   "wireguard",
	Short: "Generate WireGuard keys and mesh configurations",
}

var wireguardMeshCmd = &cobra.Command{
	Use:   "mesh",
	Short: "Generate a WireGuard mesh configuration for N peers",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, ok := forge.Get(forge.CategoryCrypto, "mesh")
		if !ok {
			return fmt.Errorf("generator not found: crypto/mesh")
		}

		if flags.GetBench(cmd) {
			return runBench(cmd, g)
		}

		peers, _ := cmd.Flags().GetInt("peers")
		subnet, _ := cmd.Flags().GetString("subnet")
		port, _ := cmd.Flags().GetInt("port")
		endpoint, _ := cmd.Flags().GetString("endpoint")
		dns, _ := cmd.Flags().GetString("dns")
		keepalive, _ := cmd.Flags().GetInt("keepalive")

		of := flags.GetOutputFlags(cmd)
		opts := forge.Options{
			Count:  1,
			Format: of.ResolveFormat(),
			Params: map[string]string{
				"peers":     fmt.Sprintf("%d", peers),
				"subnet":    subnet,
				"port":      fmt.Sprintf("%d", port),
				"endpoint":  endpoint,
				"dns":       dns,
				"keepalive": fmt.Sprintf("%d", keepalive),
			},
		}

		return flags.RunGenerate(cmd.Context(), flags.RunOptions{
			Generator: g,
			Opts:      opts,
			Count:     1,
			Format:    of.ResolveFormat(),
			OutputDir: of.OutputDir,
			Writer:    os.Stdout,
		})
	},
}

var wireguardKeypairCmd = &cobra.Command{
	Use:   "keypair",
	Short: "Generate a WireGuard Curve25519 keypair",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, ok := forge.Get(forge.CategoryCrypto, "keypair")
		if !ok {
			return fmt.Errorf("generator not found: crypto/keypair")
		}

		if flags.GetBench(cmd) {
			return runBench(cmd, g)
		}

		of := flags.GetOutputFlags(cmd)
		opts := forge.Options{Count: 1, Format: of.ResolveFormat()}

		if handled, err := flags.RunWhy(cmd, g, opts); handled {
			return err
		}

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
