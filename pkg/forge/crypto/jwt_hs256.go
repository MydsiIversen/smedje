package crypto

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/smedje/smedje/internal/bench"
	"github.com/smedje/smedje/internal/entropy"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	forge.Register(&JWTHS256{})
}

// JWTHS256 generates a random symmetric secret suitable for use as a JWT HS256 key.
type JWTHS256 struct{}

func (j *JWTHS256) Name() string             { return "hs256" }
func (j *JWTHS256) Group() string            { return "jwt" }
func (j *JWTHS256) Description() string      { return "Generate a JWT HS256 symmetric secret" }
func (j *JWTHS256) Category() forge.Category { return forge.CategoryCrypto }

// Generate returns a base64-encoded random secret. The default length is 32
// bytes (256 bits), matching the HS256 key size recommendation in RFC 7518 §3.2.
func (j *JWTHS256) Generate(ctx context.Context, opts forge.Options) (*forge.Output, error) {
	length := 32
	if v, ok := opts.Params["length"]; ok {
		fmt.Sscanf(v, "%d", &length)
	}
	if length < 16 {
		return nil, fmt.Errorf("jwt: hs256 key length must be >= 16 bytes")
	}

	b := make([]byte, length)
	if _, err := entropy.Read(b); err != nil {
		return nil, fmt.Errorf("jwt: entropy: %w", err)
	}

	return forge.SingleArtifact("jwt-hs256",
		forge.Field{Key: "secret", Value: base64.StdEncoding.EncodeToString(b), Sensitive: true},
	), nil
}

// Flags implements forge.FlagDescriber.
func (j *JWTHS256) Flags() []forge.FlagDef {
	return []forge.FlagDef{
		{Name: "length", Type: "int", Default: "32", Description: "Secret length in bytes"},
	}
}

func (j *JWTHS256) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.RunLegacy(ctx, j, 0)
}
