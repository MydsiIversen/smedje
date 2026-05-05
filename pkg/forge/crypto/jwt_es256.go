package crypto

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"

	"github.com/smedje/smedje/internal/bench"
	"github.com/smedje/smedje/internal/entropy"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	forge.Register(&JWTES256{})
}

// JWTES256 generates an ECDSA P-256 keypair for use with JWT ES256
// (ECDSA using P-256 and SHA-256). Output includes PKCS#8 private key PEM,
// PKIX public key PEM, and a single-entry JWKS document.
type JWTES256 struct{}

func (j *JWTES256) Name() string             { return "es256" }
func (j *JWTES256) Group() string            { return "jwt" }
func (j *JWTES256) Description() string      { return "Generate a JWT ES256 ECDSA P-256 keypair with JWKS" }
func (j *JWTES256) Category() forge.Category { return forge.CategoryCrypto }

func (j *JWTES256) Generate(ctx context.Context, opts forge.Options) (*forge.Output, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), entropy.Reader)
	if err != nil {
		return nil, fmt.Errorf("jwt: es256: keygen: %w", err)
	}

	privPEM, err := marshalPrivateKeyPEM(key)
	if err != nil {
		return nil, err
	}

	pubPEM, err := marshalPublicKeyPEM(&key.PublicKey)
	if err != nil {
		return nil, err
	}

	kid := opts.Params["kid"]
	if kid == "" {
		kid = pubKeyKID(&key.PublicKey)
	}

	jwks, err := jwksFromKey(kid, "ES256", &key.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("jwt: es256: jwks: %w", err)
	}

	return forge.SingleArtifact("jwt-es256",
		forge.Field{Key: "private-key", Value: privPEM, Sensitive: true},
		forge.Field{Key: "public-key", Value: pubPEM},
		forge.Field{Key: "jwks", Value: string(jwks)},
	), nil
}

// Flags implements forge.FlagDescriber.
func (j *JWTES256) Flags() []forge.FlagDef {
	return []forge.FlagDef{
		{Name: "kid", Type: "string", Description: "Key ID for JWKS (auto-generated from public key fingerprint if empty)"},
	}
}

func (j *JWTES256) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.RunLegacy(ctx, j, 0)
}
