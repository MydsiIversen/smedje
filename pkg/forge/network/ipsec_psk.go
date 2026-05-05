package network

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/smedje/smedje/internal/bench"
	"github.com/smedje/smedje/internal/entropy"
	"github.com/smedje/smedje/pkg/forge"
)

func init() { forge.Register(&IPsecPSK{}) }

// IPsecPSK generates a random hex-encoded IPsec pre-shared key.
// The default length is 32 bytes (256 bits), which is sufficient for
// IKEv2 with AES-256. Valid range is 16-128 bytes.
type IPsecPSK struct{}

func (i *IPsecPSK) Name() string             { return "ipsec-psk" }
func (i *IPsecPSK) Group() string            { return "network" }
func (i *IPsecPSK) Description() string      { return "Generate an IPsec pre-shared key" }
func (i *IPsecPSK) Category() forge.Category { return forge.CategorySecret }

// Generate returns a hex-encoded random byte sequence of the requested length.
func (i *IPsecPSK) Generate(ctx context.Context, opts forge.Options) (*forge.Output, error) {
	length := 32
	if v, ok := opts.Params["length"]; ok {
		fmt.Sscanf(v, "%d", &length)
	}
	if length < 16 || length > 128 {
		return nil, fmt.Errorf("ipsec-psk: length must be 16-128, got %d", length)
	}
	b := make([]byte, length)
	if _, err := entropy.Read(b); err != nil {
		return nil, fmt.Errorf("ipsec-psk: %w", err)
	}
	return forge.SingleArtifact("ipsec-psk",
		forge.Field{Key: "value", Value: hex.EncodeToString(b), Sensitive: true},
	), nil
}

func (i *IPsecPSK) Flags() []forge.FlagDef {
	return []forge.FlagDef{
		{Name: "length", Type: "int", Default: "32", Description: "Key length in bytes (16-128). 32 = 256 bits, suitable for IKEv2/AES-256"},
	}
}

func (i *IPsecPSK) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.RunLegacy(ctx, i, 0)
}
