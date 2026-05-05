package network

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/smedje/smedje/internal/bench"
	"github.com/smedje/smedje/internal/entropy"
	"github.com/smedje/smedje/pkg/forge"
)

func init() { forge.Register(&SNMPCommunity{}) }

// SNMPCommunity generates a random alphanumeric SNMP community string.
// The default length is 16 characters. All characters are drawn from
// [a-zA-Z0-9] to stay compatible with the widest range of SNMP agents.
type SNMPCommunity struct{}

func (s *SNMPCommunity) Name() string             { return "snmp-community" }
func (s *SNMPCommunity) Group() string            { return "network" }
func (s *SNMPCommunity) Description() string      { return "Generate an SNMP community string" }
func (s *SNMPCommunity) Category() forge.Category { return forge.CategorySecret }

const snmpCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// Generate returns a random alphanumeric string of the requested length.
func (s *SNMPCommunity) Generate(ctx context.Context, opts forge.Options) (*forge.Output, error) {
	length := 16
	if v, ok := opts.Params["length"]; ok {
		fmt.Sscanf(v, "%d", &length)
	}

	charsetLen := big.NewInt(int64(len(snmpCharset)))
	result := make([]byte, length)
	for i := range result {
		idx, err := rand.Int(entropy.Reader, charsetLen)
		if err != nil {
			return nil, fmt.Errorf("snmp: %w", err)
		}
		result[i] = snmpCharset[idx.Int64()]
	}

	return forge.SingleArtifact("snmp-community",
		forge.Field{Key: "value", Value: string(result), Sensitive: true},
	), nil
}

func (s *SNMPCommunity) Flags() []forge.FlagDef {
	return []forge.FlagDef{
		{Name: "length", Type: "int", Default: "16", Description: "String length in characters (alphanumeric a-zA-Z0-9)"},
	}
}

func (s *SNMPCommunity) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.RunLegacy(ctx, s, 0)
}
