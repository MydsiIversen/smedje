package crypto

import (
	"context"

	"filippo.io/age"

	"github.com/smedje/smedje/internal/bench"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	forge.Register(&AgeKeypair{})
}

// AgeKeypair generates an age X25519 keypair.
type AgeKeypair struct{}

func (a *AgeKeypair) Name() string             { return "x25519" }
func (a *AgeKeypair) Group() string            { return "age" }
func (a *AgeKeypair) Description() string      { return "Generate an age X25519 keypair" }
func (a *AgeKeypair) Category() forge.Category { return forge.CategoryCrypto }

// Generate returns an age X25519 identity. The private key is Bech32-encoded
// with the AGE-SECRET-KEY- prefix; the public key uses the age1 prefix.
func (a *AgeKeypair) Generate(ctx context.Context, opts forge.Options) (*forge.Output, error) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return nil, err
	}

	return forge.SingleArtifact("age",
		forge.Field{Key: "private-key", Value: identity.String(), Sensitive: true},
		forge.Field{Key: "public-key", Value: identity.Recipient().String()},
	), nil
}

func (a *AgeKeypair) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.RunLegacy(ctx, a, 0)
}
