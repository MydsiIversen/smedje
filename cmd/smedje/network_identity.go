package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/smedje/smedje/internal/flags"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	// oui-mac
	rootCmd.AddCommand(ouiMACCmd)
	ouiMACCmd.Flags().String("oui", "", "OUI prefix (e.g. 00:50:56); random if omitted")
	ouiMACCmd.Flags().String("style", "colon", "MAC address notation: colon, dash, dot")
	flags.AddOutputFlags(ouiMACCmd)
	flags.AddBulkFlags(ouiMACCmd)
	flags.AddBenchFlag(ouiMACCmd)

	// iqn
	rootCmd.AddCommand(iqnCmd)
	iqnCmd.Flags().String("authority", "", "DNS authority in forward order (e.g. com.example)")
	iqnCmd.Flags().String("target", "", "Target name (e.g. storage.lun0)")
	iqnCmd.Flags().String("date", "", "Date in YYYY-MM format; defaults to current month")
	flags.AddOutputFlags(iqnCmd)
	flags.AddBulkFlags(iqnCmd)
	flags.AddBenchFlag(iqnCmd)

	// wwpn
	rootCmd.AddCommand(wwpnCmd)
	flags.AddOutputFlags(wwpnCmd)
	flags.AddBulkFlags(wwpnCmd)
	flags.AddBenchFlag(wwpnCmd)
}

var ouiMACCmd = &cobra.Command{
	Use:   "oui-mac",
	Short: "Generate a MAC address with a real vendor OUI prefix",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, ok := forge.Get(forge.CategoryNetwork, "oui")
		if !ok {
			return fmt.Errorf("generator not found: network/oui")
		}
		if flags.GetBench(cmd) {
			return runBench(cmd, g)
		}
		of := flags.GetOutputFlags(cmd)
		oui, _ := cmd.Flags().GetString("oui")
		style, _ := cmd.Flags().GetString("style")
		opts := forge.Options{
			Count:  1,
			Format: of.ResolveFormat(),
			Params: map[string]string{
				"oui":    oui,
				"format": style,
			},
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

var iqnCmd = &cobra.Command{
	Use:   "iqn",
	Short: "Generate an iSCSI Qualified Name (IQN)",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, ok := forge.Get(forge.CategoryNetwork, "iqn")
		if !ok {
			return fmt.Errorf("generator not found: network/iqn")
		}
		if flags.GetBench(cmd) {
			return runBench(cmd, g)
		}
		of := flags.GetOutputFlags(cmd)
		authority, _ := cmd.Flags().GetString("authority")
		target, _ := cmd.Flags().GetString("target")
		date, _ := cmd.Flags().GetString("date")
		opts := forge.Options{
			Count:  1,
			Format: of.ResolveFormat(),
			Params: map[string]string{
				"authority": authority,
				"target":    target,
				"date":      date,
			},
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

var wwpnCmd = &cobra.Command{
	Use:   "wwpn",
	Short: "Generate a Fibre Channel World Wide Port Name (WWPN)",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, ok := forge.Get(forge.CategoryNetwork, "wwpn")
		if !ok {
			return fmt.Errorf("generator not found: network/wwpn")
		}
		if flags.GetBench(cmd) {
			return runBench(cmd, g)
		}
		of := flags.GetOutputFlags(cmd)
		opts := forge.Options{
			Count:  1,
			Format: of.ResolveFormat(),
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
