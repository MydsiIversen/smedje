package flags

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/smedje/smedje/internal/entropy"
	"github.com/smedje/smedje/pkg/forge"
)

// AddSeedFlags registers --seed and --time-source on a command.
func AddSeedFlags(cmd *cobra.Command) {
	cmd.Flags().String("seed", "", "Deterministic seed (disables crypto/rand)")
	cmd.Flags().String("time-source", "", "Time source when seeded: frozen (default) or real")
}

// ApplySeed configures the entropy source and returns a time function
// for forge.Options based on --seed and --time-source flags. Call
// CleanupSeed when done.
func ApplySeed(cmd *cobra.Command) func() time.Time {
	seed, _ := cmd.Flags().GetString("seed")
	if seed == "" {
		return nil
	}

	entropy.SetSeed(seed)

	ts, _ := cmd.Flags().GetString("time-source")
	if ts == "real" {
		return nil
	}

	frozen := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	return func() time.Time { return frozen }
}

// CleanupSeed restores the default entropy source.
func CleanupSeed(cmd *cobra.Command) {
	seed, _ := cmd.Flags().GetString("seed")
	if seed != "" {
		entropy.Reset()
	}
}

// ApplySeedToOpts applies seed-derived time function to forge.Options.
func ApplySeedToOpts(cmd *cobra.Command, opts *forge.Options) {
	timeFn := ApplySeed(cmd)
	if timeFn != nil {
		opts.Time = timeFn
	}
}
