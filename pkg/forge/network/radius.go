package network

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/smedje/smedje/internal/bench"
	"github.com/smedje/smedje/internal/entropy"
	"github.com/smedje/smedje/pkg/forge"
)

func init() { forge.Register(&RADIUSSecret{}) }

// RADIUSSecret generates a random base64-encoded RADIUS shared secret.
// RFC 2865 does not mandate a length, but 24 bytes (192 bits) is a
// reasonable default that exceeds what most RADIUS servers require.
type RADIUSSecret struct{}

func (r *RADIUSSecret) Name() string             { return "radius-secret" }
func (r *RADIUSSecret) Group() string            { return "network" }
func (r *RADIUSSecret) Description() string      { return "Generate a RADIUS shared secret" }
func (r *RADIUSSecret) Category() forge.Category { return forge.CategorySecret }

// Generate returns a base64-encoded random byte sequence of the requested length.
func (r *RADIUSSecret) Generate(ctx context.Context, opts forge.Options) (*forge.Output, error) {
	length := 24
	if v, ok := opts.Params["length"]; ok {
		fmt.Sscanf(v, "%d", &length)
	}
	if length < 8 || length > 128 {
		return nil, fmt.Errorf("radius: length must be 8-128, got %d", length)
	}
	b := make([]byte, length)
	if _, err := entropy.Read(b); err != nil {
		return nil, fmt.Errorf("radius: %w", err)
	}
	return forge.SingleArtifact("radius-secret",
		forge.Field{Key: "value", Value: base64.StdEncoding.EncodeToString(b), Sensitive: true},
	), nil
}

func (r *RADIUSSecret) Flags() []forge.FlagDef {
	return []forge.FlagDef{
		{Name: "length", Type: "int", Default: "24", Description: "Secret length in bytes (8-128). 24 = 192 bits, exceeds RADIUS requirements"},
	}
}

func (r *RADIUSSecret) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.RunLegacy(ctx, r, 0)
}
