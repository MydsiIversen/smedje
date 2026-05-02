package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/smedje/smedje/internal/config"
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configValidateCmd)

	configShowCmd.Flags().Bool("explain", false, "Show the source of each value")
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Smedje configuration",
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Write a commented default config to the user config directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := config.UserConfigPath()
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("config file already exists: %s", path)
		}

		defaults := config.Defaults
		sections := make(map[string][]string)
		for k := range defaults {
			parts := strings.SplitN(k, ".", 2)
			if len(parts) == 2 {
				sections[parts[0]] = append(sections[parts[0]], k)
			}
		}

		if err := os.MkdirAll(config.UserConfigDir(), 0o755); err != nil {
			return err
		}

		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()

		fmt.Fprintln(f, "# Smedje default configuration")
		fmt.Fprintln(f, "# Values here override built-in defaults.")
		fmt.Fprintln(f, "# Project-level overrides go in .smedje.toml in your project root.")
		fmt.Fprintln(f)

		sectionNames := make([]string, 0, len(sections))
		for s := range sections {
			sectionNames = append(sectionNames, s)
		}
		sort.Strings(sectionNames)

		for _, sec := range sectionNames {
			keys := sections[sec]
			sort.Strings(keys)
			fmt.Fprintf(f, "[%s]\n", sec)
			for _, k := range keys {
				name := strings.TrimPrefix(k, sec+".")
				fmt.Fprintf(f, "# %s = %q\n", name, defaults[k])
			}
			fmt.Fprintln(f)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Config written to %s\n", path)
		return nil
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show effective configuration values",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(config.LoadOptions{})
		if err != nil {
			return err
		}

		explain, _ := cmd.Flags().GetBool("explain")
		all := cfg.All()

		keys := make([]string, 0, len(all))
		for k := range all {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			e := all[k]
			if explain {
				fmt.Fprintf(cmd.OutOrStdout(), "%-25s = %-20s [%s]\n", k, e.Value, e.Source)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "%-25s = %s\n", k, e.Value)
			}
		}
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get KEY",
	Short: "Get the effective value of a config key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(config.LoadOptions{})
		if err != nil {
			return err
		}
		entry, ok := cfg.GetEntry(args[0])
		if !ok {
			keys := cfg.Keys()
			return fmt.Errorf("unknown config key: %s\n\nValid keys:\n  %s", args[0], strings.Join(keys, "\n  "))
		}
		fmt.Fprintln(cmd.OutOrStdout(), entry.Value)
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set KEY VALUE",
	Short: "Set a value in the user config file",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := config.UserConfigPath()

		existing := make(map[string]string)
		if _, err := os.Stat(path); err == nil {
			cfg, err := config.Load(config.LoadOptions{})
			if err != nil {
				return err
			}
			for _, k := range cfg.Keys() {
				e, _ := cfg.GetEntry(k)
				if e.Source == config.SourceUserConfig {
					existing[k] = e.Value
				}
			}
		}

		existing[args[0]] = args[1]

		grouped := make(map[string]map[string]string)
		for k, v := range existing {
			parts := strings.SplitN(k, ".", 2)
			sec, name := "", k
			if len(parts) == 2 {
				sec, name = parts[0], parts[1]
			}
			if grouped[sec] == nil {
				grouped[sec] = make(map[string]string)
			}
			grouped[sec][name] = v
		}

		if err := os.MkdirAll(config.UserConfigDir(), 0o755); err != nil {
			return err
		}
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()

		secs := make([]string, 0, len(grouped))
		for s := range grouped {
			secs = append(secs, s)
		}
		sort.Strings(secs)

		for _, sec := range secs {
			if sec != "" {
				fmt.Fprintf(f, "[%s]\n", sec)
			}
			keys := make([]string, 0, len(grouped[sec]))
			for k := range grouped[sec] {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				fmt.Fprintf(f, "%s = %q\n", k, grouped[sec][k])
			}
			fmt.Fprintln(f)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Set %s = %s in %s\n", args[0], args[1], path)
		return nil
	},
}

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(config.LoadOptions{})
		if err != nil {
			return err
		}
		errs := cfg.Validate()
		if len(errs) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "Configuration is valid.")
			return nil
		}
		for _, e := range errs {
			fmt.Fprintf(cmd.ErrOrStderr(), "  ✗ %s\n", e)
		}
		return fmt.Errorf("%d validation error(s)", len(errs))
	},
}
