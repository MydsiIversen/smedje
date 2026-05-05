// Package network provides network-related generators.
package network

import (
	"context"
	"fmt"

	"github.com/smedje/smedje/internal/bench"
	"github.com/smedje/smedje/internal/entropy"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	forge.Register(&MAC{})
}

// MAC generates a random locally-administered unicast MAC address.
// The second-least-significant bit of the first octet is set (locally
// administered) and the least-significant bit is cleared (unicast).
type MAC struct{}

func (m *MAC) Name() string             { return "mac" }
func (m *MAC) Group() string            { return "mac" }
func (m *MAC) Description() string      { return "Generate a random locally-administered MAC address" }
func (m *MAC) Category() forge.Category { return forge.CategoryNetwork }

func (m *MAC) Generate(ctx context.Context, opts forge.Options) (*forge.Output, error) {
	var addr [6]byte
	if _, err := entropy.Read(addr[:]); err != nil {
		return nil, fmt.Errorf("mac: entropy read: %w", err)
	}

	addr[0] = (addr[0] | 0x02) & 0xFE

	return forge.SingleArtifact("mac", forge.Field{Key: "value", Value: fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
		addr[0], addr[1], addr[2], addr[3], addr[4], addr[5])}), nil
}

func (m *MAC) Flags() []forge.FlagDef {
	return []forge.FlagDef{
		{Name: "format", Type: "string", Default: "colon", Description: "MAC address format",
			Options: []string{"colon", "dash", "dot"}},
	}
}

func (m *MAC) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.RunLegacy(ctx, m, 0)
}
