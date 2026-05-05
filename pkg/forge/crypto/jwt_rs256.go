package crypto

import (
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/smedje/smedje/internal/bench"
	"github.com/smedje/smedje/internal/entropy"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	forge.Register(&JWTRS256{})
}

// JWTRS256 generates an RSA keypair for use with JWT RS256 (RSASSA-PKCS1-v1_5
// using SHA-256). Output includes PKCS#8 private key PEM, PKIX public key PEM,
// and a single-entry JWKS document.
type JWTRS256 struct{}

func (j *JWTRS256) Name() string             { return "rs256" }
func (j *JWTRS256) Group() string            { return "jwt" }
func (j *JWTRS256) Description() string      { return "Generate a JWT RS256 RSA keypair with JWKS" }
func (j *JWTRS256) Category() forge.Category { return forge.CategoryCrypto }

func (j *JWTRS256) Generate(ctx context.Context, opts forge.Options) (*forge.Output, error) {
	bits := 2048
	if v, ok := opts.Params["bits"]; ok {
		fmt.Sscanf(v, "%d", &bits)
	}

	key, err := rsa.GenerateKey(entropy.Reader, bits)
	if err != nil {
		return nil, fmt.Errorf("jwt: rs256: keygen: %w", err)
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

	jwks, err := jwksFromKey(kid, &key.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("jwt: rs256: jwks: %w", err)
	}

	return forge.SingleArtifact("jwt-rs256",
		forge.Field{Key: "private-key", Value: privPEM, Sensitive: true},
		forge.Field{Key: "public-key", Value: pubPEM},
		forge.Field{Key: "jwks", Value: string(jwks)},
	), nil
}

// Flags implements forge.FlagDescriber.
func (j *JWTRS256) Flags() []forge.FlagDef {
	return []forge.FlagDef{
		{Name: "bits", Type: "int", Default: "2048", Description: "RSA key size in bits", Options: []string{"2048", "4096"}},
		{Name: "kid", Type: "string", Default: "", Description: "Key ID; auto-generated from public key SHA-256 if omitted"},
	}
}

func (j *JWTRS256) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.RunLegacy(ctx, j, 0)
}

// marshalPrivateKeyPEM encodes a private key as a PKCS#8 PEM PRIVATE KEY block.
func marshalPrivateKeyPEM(key interface{}) (string, error) {
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return "", fmt.Errorf("jwt: marshal private key: %w", err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})), nil
}

// marshalPublicKeyPEM encodes a public key as a PKIX PEM PUBLIC KEY block.
func marshalPublicKeyPEM(pub interface{}) (string, error) {
	der, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return "", fmt.Errorf("jwt: marshal public key: %w", err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})), nil
}

// pubKeyKID returns a short hex string derived from the SHA-256 of the PKIX
// encoding of the public key. This provides a stable, unique identifier.
func pubKeyKID(pub interface{}) string {
	der, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return "unknown"
	}
	sum := sha256.Sum256(der)
	return fmt.Sprintf("%x", sum[:8])
}
