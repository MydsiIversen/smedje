// Package secret provides secret generators (passwords, TOTP).
package secret

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
	forge.Register(&Password{})
}

const (
	charsetLower   = "abcdefghijklmnopqrstuvwxyz"
	charsetUpper   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	charsetDigits  = "0123456789"
	charsetSymbols = "!@#$%^&*()-_=+[]{}|;:,.<>?"
)

// Password generates a random password with configurable length and charset.
//
// Options:
//
//	length:  password length (default 24, min 8, max 256)
//	charset: "full" (default), "alpha", "alphanum", "digits"
type Password struct{}

func (p *Password) Name() string             { return "password" }
func (p *Password) Group() string            { return "password" }
func (p *Password) Description() string      { return "Generate a random password" }
func (p *Password) Category() forge.Category { return forge.CategorySecret }

func (p *Password) Generate(ctx context.Context, opts forge.Options) (*forge.Output, error) {
	length := 24
	if def := config.GetDefault("password.length"); def != "" {
		fmt.Sscanf(def, "%d", &length)
	}
	if v, ok := opts.Params["length"]; ok {
		fmt.Sscanf(v, "%d", &length)
	}
	if length < 8 {
		return nil, fmt.Errorf("password: length must be >= 8, got %d", length)
	}
	if length > 256 {
		return nil, fmt.Errorf("password: length must be <= 256, got %d", length)
	}

	charsetName := "full"
	if def := config.GetDefault("password.charset"); def != "" {
		charsetName = def
	}
	if v, ok := opts.Params["charset"]; ok {
		charsetName = v
	}

	charset := charsetLower + charsetUpper + charsetDigits + charsetSymbols
	if v := charsetName; v != "full" {
		switch v {
		case "alpha":
			charset = charsetLower + charsetUpper
		case "alphanum":
			charset = charsetLower + charsetUpper + charsetDigits
		case "digits", "numeric":
			charset = charsetDigits
		case "hex":
			charset = charsetDigits + "abcdef"
		case "full":
			// default
		default:
			return nil, fmt.Errorf("password: unknown charset %q", v)
		}
	}

	pw, err := randomString(length, charset)
	if err != nil {
		return nil, err
	}

	return forge.SingleArtifact("password", forge.Field{Key: "value", Value: pw, Sensitive: true}), nil
}

func (p *Password) Flags() []forge.FlagDef {
	return []forge.FlagDef{
		{Name: "length", Type: "int", Default: "24", Description: "Password length (8-256)"},
		{Name: "charset", Type: "string", Default: "full", Description: "Character set",
			Options: []string{"full", "alphanum", "alpha", "digits", "hex"}},
	}
}

func (p *Password) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.RunLegacy(ctx, p, 0)
}

func randomString(length int, charset string) (string, error) {
	max := big.NewInt(int64(len(charset)))
	result := make([]byte, length)
	for i := range result {
		idx, err := randInt(max)
		if err != nil {
			return "", err
		}
		result[i] = charset[idx.Int64()]
	}
	return string(result), nil
}

func randInt(max *big.Int) (*big.Int, error) {
	b := make([]byte, max.BitLen()/8+1)
	if _, err := entropy.Read(b); err != nil {
		return nil, err
	}
	n := new(big.Int).SetBytes(b)
	n.Mod(n, max)
	return n, nil
}
