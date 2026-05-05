// Package ssh provides SSH key generators.
package ssh

import (
	"context"
	"crypto/ed25519"
	"encoding/pem"

	"github.com/smedje/smedje/internal/bench"
	"github.com/smedje/smedje/internal/entropy"
	"github.com/smedje/smedje/pkg/forge"

	gossh "golang.org/x/crypto/ssh"
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
	pub, priv, err := ed25519.GenerateKey(entropy.Reader)
	if err != nil {
		return nil, err
	}

	privPEM, err := gossh.MarshalPrivateKey(priv, "")
	if err != nil {
		return nil, err
	}
	privStr := string(pem.EncodeToMemory(privPEM))

	sshPub, err := gossh.NewPublicKey(pub)
	if err != nil {
		return nil, err
	}
	pubStr := string(gossh.MarshalAuthorizedKey(sshPub))

	return forge.SingleArtifact("ssh",
		forge.Field{Key: "private-key", Value: privStr, Sensitive: true},
		forge.Field{Key: "public-key", Value: pubStr},
	), nil
}

func (e *Ed25519) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.RunLegacy(ctx, e, 0)
}
