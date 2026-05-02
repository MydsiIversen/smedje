// Package forge defines the Generator interface and the global registry
// that all generators self-register into via init().
package forge

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// Category groups generators in CLI help and documentation.
type Category string

const (
	CategoryID      Category = "id"
	CategoryCrypto  Category = "crypto"
	CategorySecret  Category = "secret"
	CategoryNetwork Category = "network"
)

// Options carries flags from the CLI layer into a generator.
type Options struct {
	// Count is the number of items to generate (default 1).
	Count int

	// Format controls output rendering: "text", "json", "quiet".
	Format string

	// Params holds generator-specific key-value options (e.g., "length"
	// for passwords, "worker" for Snowflake).
	Params map[string]string

	// Time returns the current time. When nil, generators use time.Now().
	// Set to a fixed-value function for deterministic output with --seed.
	Time func() time.Time
}

// Output is the result of a single Generate call.
type Output struct {
	// Name identifies what was generated (e.g., "uuidv7", "ed25519-keypair").
	Name string

	// Fields holds the generated values in display order.
	// For single-value generators, use one entry with key "value".
	Fields []Field
}

// Field is a single named value in generator output.
type Field struct {
	Key   string
	Value string

	// Sensitive marks values that should not be logged or displayed
	// in non-interactive contexts.
	Sensitive bool
}

// BenchResult holds the outcome of a generator's self-benchmark.
type BenchResult struct {
	Generator  string
	Iterations int
	Duration   time.Duration
	OpsPerSec  float64
}

// Explainer is optionally implemented by generators that support --why.
type Explainer interface {
	Why() string
}

// Generator is the interface every forge generator implements.
type Generator interface {
	// Name returns the CLI-facing name (e.g., "v7", "ed25519").
	Name() string

	// Group returns the CLI command group (e.g., "uuid", "ssh", "tls").
	// Single-variant generators return their own name as the group.
	Group() string

	// Description returns a one-line summary for help text.
	Description() string

	// Category returns the generator's category.
	Category() Category

	// Generate produces output using the given options.
	Generate(ctx context.Context, opts Options) (*Output, error)

	// Bench runs a self-benchmark and returns the result.
	Bench(ctx context.Context) (*BenchResult, error)
}

var (
	mu       sync.RWMutex
	registry = make(map[string]Generator)
)

// Register adds a generator to the global registry. It panics on duplicate
// names because duplicates indicate a wiring bug that should fail at startup.
func Register(g Generator) {
	key := string(g.Category()) + "/" + g.Name()
	mu.Lock()
	defer mu.Unlock()
	if _, exists := registry[key]; exists {
		panic(fmt.Sprintf("forge: duplicate generator %q", key))
	}
	registry[key] = g
}

// Get returns a registered generator by category and name.
func Get(category Category, name string) (Generator, bool) {
	key := string(category) + "/" + name
	mu.RLock()
	defer mu.RUnlock()
	g, ok := registry[key]
	return g, ok
}

// All returns every registered generator.
func All() []Generator {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]Generator, 0, len(registry))
	for _, g := range registry {
		out = append(out, g)
	}
	return out
}

// ByCategory returns all generators in a given category.
func ByCategory(c Category) []Generator {
	mu.RLock()
	defer mu.RUnlock()
	var out []Generator
	for _, g := range registry {
		if g.Category() == c {
			out = append(out, g)
		}
	}
	return out
}

// Address returns the dotted address for a generator (e.g., "uuid.v7", "ulid").
func Address(g Generator) string {
	if g.Group() == g.Name() {
		return g.Name()
	}
	return g.Group() + "." + g.Name()
}

// Resolve finds generators matching a dotted address string.
// Supports "uuid.v7" (exact), "ulid" (bare name), and "uuid" (group with
// multiple variants returns an error listing them all).
func Resolve(address string) ([]Generator, error) {
	mu.RLock()
	defer mu.RUnlock()

	parts := strings.SplitN(address, ".", 2)

	if len(parts) == 2 {
		group, variant := parts[0], parts[1]
		for _, g := range registry {
			if g.Group() == group && g.Name() == variant {
				return []Generator{g}, nil
			}
		}
		// Check if group exists at all for a better error.
		var inGroup []string
		for _, g := range registry {
			if g.Group() == group {
				inGroup = append(inGroup, Address(g))
			}
		}
		if len(inGroup) > 0 {
			sort.Strings(inGroup)
			return nil, fmt.Errorf("no variant %q in group %q. Available: %s",
				variant, group, strings.Join(inGroup, ", "))
		}
		return nil, fmt.Errorf("no generator group %q found", group)
	}

	// Bare name: try as group.
	name := parts[0]
	var inGroup []Generator
	for _, g := range registry {
		if g.Group() == name {
			inGroup = append(inGroup, g)
		}
	}

	if len(inGroup) == 1 {
		return inGroup, nil
	}
	if len(inGroup) > 1 {
		var addrs []string
		for _, g := range inGroup {
			addrs = append(addrs, Address(g))
		}
		sort.Strings(addrs)
		return nil, fmt.Errorf("%q has multiple variants. Did you mean one of:\n  %s",
			name, strings.Join(addrs, ", "))
	}

	// Try as exact name match across all generators.
	for _, g := range registry {
		if g.Name() == name {
			inGroup = append(inGroup, g)
		}
	}
	if len(inGroup) == 1 {
		return inGroup, nil
	}

	return nil, fmt.Errorf("no generator named %q found", name)
}

// Addresses returns sorted dotted addresses for all registered generators.
func Addresses() []string {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]string, 0, len(registry))
	for _, g := range registry {
		out = append(out, Address(g))
	}
	sort.Strings(out)
	return out
}
