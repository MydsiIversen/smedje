package id

import (
	"context"
	"fmt"
	"math/big"

	"github.com/smedje/smedje/internal/bench"
	"github.com/smedje/smedje/internal/config"
	"github.com/smedje/smedje/internal/entropy"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	forge.Register(&NanoID{})
}

const defaultNanoIDAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_-"
const defaultNanoIDLength = 21

// NanoID generates URL-safe unique identifiers with configurable length and
// alphabet. Defaults to 21 characters using the standard NanoID alphabet
// (A-Za-z0-9_-), providing ~126 bits of entropy.
type NanoID struct{}

func (n *NanoID) Name() string             { return "nanoid" }
func (n *NanoID) Group() string            { return "nanoid" }
func (n *NanoID) Description() string      { return "Generate a NanoID (URL-safe, configurable)" }
func (n *NanoID) Category() forge.Category { return forge.CategoryID }

func (n *NanoID) Generate(ctx context.Context, opts forge.Options) (*forge.Output, error) {
	length := defaultNanoIDLength
	if def := config.GetDefault("nanoid.length"); def != "" {
		fmt.Sscanf(def, "%d", &length)
	}
	if v, ok := opts.Params["length"]; ok {
		fmt.Sscanf(v, "%d", &length)
	}
	if length < 1 || length > 256 {
		return nil, fmt.Errorf("nanoid: length must be 1-256, got %d", length)
	}

	alphabet := defaultNanoIDAlphabet
	if v, ok := opts.Params["alphabet"]; ok && v != "" {
		alphabet = v
	}
	if len(alphabet) < 2 {
		return nil, fmt.Errorf("nanoid: alphabet must have at least 2 characters")
	}

	id, err := generateNanoID(length, alphabet)
	if err != nil {
		return nil, err
	}

	return &forge.Output{
		Name: "nanoid",
		Fields: []forge.Field{
			{Key: "value", Value: id},
		},
	}, nil
}

func (n *NanoID) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.RunLegacy(ctx, n, 0)
}

func generateNanoID(length int, alphabet string) (string, error) {
	max := big.NewInt(int64(len(alphabet)))
	result := make([]byte, length)
	for i := range result {
		b := make([]byte, 1)
		if _, err := entropy.Read(b); err != nil {
			return "", fmt.Errorf("nanoid: entropy: %w", err)
		}
		// Rejection sampling to avoid modulo bias.
		idx := new(big.Int).SetBytes(b)
		idx.Mod(idx, max)
		result[i] = alphabet[idx.Int64()]
	}
	return string(result), nil
}
