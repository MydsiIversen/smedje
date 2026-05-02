// Package forge defines the Generator interface and the global registry
// that all generators self-register into via init().
package forge

import (
	"context"
	"fmt"
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

// Generator is the interface every forge generator implements.
type Generator interface {
	// Name returns the CLI-facing name (e.g., "v7", "ed25519").
	Name() string

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
