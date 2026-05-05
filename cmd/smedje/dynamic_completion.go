package main

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/smedje/smedje/pkg/forge"
)

// generatorGroupCompletion returns group names matching the prefix.
func generatorGroupCompletion(_ *cobra.Command, _ []string, toComplete string) []string {
	seen := map[string]bool{}
	var completions []string
	for _, g := range forge.All() {
		group := g.Group()
		if seen[group] {
			continue
		}
		seen[group] = true
		if strings.HasPrefix(group, toComplete) {
			completions = append(completions, group)
		}
	}
	return completions
}

// generatorVariantCompletion returns a completion function for variants
// within a generator group.
func generatorVariantCompletion(group string) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var completions []string
		for _, g := range forge.All() {
			if g.Group() != group {
				continue
			}
			name := g.Name()
			if strings.HasPrefix(name, toComplete) {
				completions = append(completions, name+"\t"+g.Description())
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

// flagValueCompletion returns completions for a flag value by reading
// from the FlagDescriber interface.
func flagValueCompletion(g forge.Generator, flagName string) []string {
	fd, ok := g.(forge.FlagDescriber)
	if !ok {
		return nil
	}
	for _, f := range fd.Flags() {
		if f.Name == flagName {
			return f.Options
		}
	}
	return nil
}

// registerCompletions wires up tab completion on generator commands
// by reading the FlagDescriber interface for flag-value completions.
func registerCompletions() {
	for _, cmd := range rootCmd.Commands() {
		group := cmd.Name()

		// Find generators for this group.
		var groupGens []forge.Generator
		for _, g := range forge.All() {
			if g.Group() == group {
				groupGens = append(groupGens, g)
			}
		}
		if len(groupGens) == 0 {
			continue
		}

		// Register flag-value completions on subcommands.
		for _, sub := range cmd.Commands() {
			for _, g := range groupGens {
				if g.Name() == sub.Name() {
					registerFlagCompletions(sub, g)
					break
				}
			}
		}

		// For single-variant groups (password, ulid, etc.), register on the command itself.
		if len(groupGens) == 1 && groupGens[0].Name() == group {
			registerFlagCompletions(cmd, groupGens[0])
		}
	}
}

func registerFlagCompletions(cmd *cobra.Command, g forge.Generator) {
	fd, ok := g.(forge.FlagDescriber)
	if !ok {
		return
	}
	for _, f := range fd.Flags() {
		if len(f.Options) > 0 {
			opts := f.Options
			_ = cmd.RegisterFlagCompletionFunc(f.Name, func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
				return opts, cobra.ShellCompDirectiveNoFileComp
			})
		}
	}
}
