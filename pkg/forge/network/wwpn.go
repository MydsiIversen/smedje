package network

import (
	"context"
	"fmt"

	"github.com/smedje/smedje/internal/bench"
	"github.com/smedje/smedje/internal/entropy"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	forge.Register(&WWPN{})
}

// WWPN generates a Fibre Channel World Wide Port Name in NAA 5 format.
// The first nibble is forced to 0x5 (NAA identifier = 5, IEEE Registered).
// The remaining 60 bits are random. Output is 8 colon-separated hex octets,
// for example: 50:ab:cd:12:34:56:78:9a.
type WWPN struct{}

func (w *WWPN) Name() string             { return "wwpn" }
func (w *WWPN) Group() string            { return "network" }
func (w *WWPN) Description() string      { return "Generate a Fibre Channel World Wide Port Name (WWPN)" }
func (w *WWPN) Category() forge.Category { return forge.CategoryNetwork }

// Generate returns a random NAA 5 WWPN.
func (w *WWPN) Generate(ctx context.Context, opts forge.Options) (*forge.Output, error) {
	var b [8]byte
	if _, err := entropy.Read(b[:]); err != nil {
		return nil, fmt.Errorf("wwpn: entropy read: %w", err)
	}

	// NAA 5: high nibble of first byte = 0x5.
	b[0] = (b[0] & 0x0F) | 0x50

	wwpn := fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x:%02x:%02x",
		b[0], b[1], b[2], b[3], b[4], b[5], b[6], b[7])

	return forge.SingleArtifact("wwpn",
		forge.Field{Key: "value", Value: wwpn},
	), nil
}

func (w *WWPN) Flags() []forge.FlagDef {
	return nil
}

func (w *WWPN) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.RunLegacy(ctx, w, 0)
}
