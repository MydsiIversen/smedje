package crypto

import (
	"context"
	"crypto/ed25519"
	"fmt"

	"github.com/smedje/smedje/internal/bench"
	"github.com/smedje/smedje/internal/entropy"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	forge.Register(&JWTEdDSA{})
}

// JWTEdDSA generates an Ed25519 keypair for use with JWT EdDSA (RFC 8037).
// Output includes PKCS#8 private key PEM, PKIX public key PEM, and a
// single-entry JWKS document.
type JWTEdDSA struct{}

func (j *JWTEdDSA) Name() string             { return "eddsa" }
func (j *JWTEdDSA) Group() string            { return "jwt" }
func (j *JWTEdDSA) Description() string      { return "Generate a JWT EdDSA Ed25519 keypair with JWKS" }
func (j *JWTEdDSA) Category() forge.Category { return forge.CategoryCrypto }

func (j *JWTEdDSA) Generate(ctx context.Context, opts forge.Options) (*forge.Output, error) {
	pub, priv, err := ed25519.GenerateKey(entropy.Reader)
	if err != nil {
		return nil, fmt.Errorf("jwt: eddsa: keygen: %w", err)
	}

	privPEM, err := marshalPrivateKeyPEM(priv)
	if err != nil {
		return nil, err
	}

	pubPEM, err := marshalPublicKeyPEM(pub)
	if err != nil {
		return nil, err
	}

	kid := opts.Params["kid"]
	if kid == "" {
		kid = pubKeyKID(pub)
	}

	jwks, err := jwksFromKey(kid, "EdDSA", pub)
	if err != nil {
		return nil, fmt.Errorf("jwt: eddsa: jwks: %w", err)
	}

	return forge.SingleArtifact("jwt-eddsa",
		forge.Field{Key: "private-key", Value: privPEM, Sensitive: true},
		forge.Field{Key: "public-key", Value: pubPEM},
		forge.Field{Key: "jwks", Value: string(jwks)},
	), nil
}

// Flags implements forge.FlagDescriber.
func (j *JWTEdDSA) Flags() []forge.FlagDef {
	return []forge.FlagDef{
		{Name: "kid", Type: "string", Description: "Key ID for JWKS (auto-generated from public key fingerprint if empty)"},
	}
}

func (j *JWTEdDSA) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.RunLegacy(ctx, j, 0)
}
