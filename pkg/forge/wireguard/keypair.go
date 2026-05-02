// Package wireguard provides WireGuard key generators.
package wireguard

import (
	"context"
	"encoding/base64"

	"github.com/smedje/smedje/internal/bench"
	"github.com/smedje/smedje/internal/entropy"
	"github.com/smedje/smedje/pkg/forge"

	"golang.org/x/crypto/curve25519"
)

func init() {
	forge.Register(&Keypair{})
}

// Keypair generates a WireGuard Curve25519 keypair.
type Keypair struct{}

func (k *Keypair) Name() string             { return "keypair" }
func (k *Keypair) Description() string      { return "Generate a WireGuard Curve25519 keypair" }
func (k *Keypair) Category() forge.Category { return forge.CategoryCrypto }

func (k *Keypair) Generate(ctx context.Context, opts forge.Options) (*forge.Output, error) {
	var priv [32]byte
	if _, err := entropy.Read(priv[:]); err != nil {
		return nil, err
	}

	// Clamp the private key per Curve25519 convention, matching wg(8).
	priv[0] &= 248
	priv[31] = (priv[31] & 127) | 64

	pub, err := curve25519.X25519(priv[:], curve25519.Basepoint)
	if err != nil {
		return nil, err
	}

	return &forge.Output{
		Name: "wireguard-keypair",
		Fields: []forge.Field{
			{Key: "private-key", Value: base64.StdEncoding.EncodeToString(priv[:]), Sensitive: true},
			{Key: "public-key", Value: base64.StdEncoding.EncodeToString(pub)},
		},
	}, nil
}

func (k *Keypair) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.RunLegacy(ctx, k, 0)
}
