// Package ssh provides SSH key generators.
package ssh

import (
	"context"

	"github.com/smedje/smedje/internal/bench"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	forge.Register(&Ed25519{})
}

// Ed25519 generates an OpenSSH ed25519 keypair.
type Ed25519 struct{}

func (e *Ed25519) Name() string             { return "ed25519" }
func (e *Ed25519) Group() string            { return "ssh" }
func (e *Ed25519) Description() string      { return "Generate an Ed25519 OpenSSH keypair" }
func (e *Ed25519) Category() forge.Category { return forge.CategoryCrypto }

func (e *Ed25519) Generate(ctx context.Context, opts forge.Options) (*forge.Output, error) {
	priv, pub, err := generateSSHKeypair("ed25519", 0)
	if err != nil {
		return nil, err
	}

	return forge.SingleArtifact("ssh",
		forge.Field{Key: "private-key", Value: priv, Sensitive: true},
		forge.Field{Key: "public-key", Value: pub},
	), nil
}

func (e *Ed25519) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.RunLegacy(ctx, e, 0)
}
