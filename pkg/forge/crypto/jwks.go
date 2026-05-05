// Package crypto contains JWT-related forge generators.
package crypto

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
)

// jwksFromKey encodes a public key as a single-entry JWKS document.
// kid is included verbatim; the caller is responsible for generating it.
func jwksFromKey(kid string, pub interface{}) ([]byte, error) {
	var jwk map[string]string

	switch k := pub.(type) {
	case *rsa.PublicKey:
		jwk = map[string]string{
			"kty": "RSA",
			"kid": kid,
			"use": "sig",
			"n":   base64URLEncode(k.N.Bytes()),
			"e":   base64URLEncode(big.NewInt(int64(k.E)).Bytes()),
		}
	case *ecdsa.PublicKey:
		crv := "P-256"
		size := 32
		if k.Curve == elliptic.P384() {
			crv = "P-384"
			size = 48
		}
		jwk = map[string]string{
			"kty": "EC",
			"kid": kid,
			"use": "sig",
			"crv": crv,
			"x":   base64URLEncode(padLeft(k.X.Bytes(), size)),
			"y":   base64URLEncode(padLeft(k.Y.Bytes(), size)),
		}
	case ed25519.PublicKey:
		jwk = map[string]string{
			"kty": "OKP",
			"kid": kid,
			"use": "sig",
			"crv": "Ed25519",
			"x":   base64URLEncode([]byte(k)),
		}
	default:
		return nil, fmt.Errorf("jwks: unsupported key type %T", pub)
	}

	doc := map[string]interface{}{
		"keys": []interface{}{jwk},
	}
	return json.MarshalIndent(doc, "", "  ")
}

// base64URLEncode encodes b using unpadded base64url encoding (RFC 4648 §5).
func base64URLEncode(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}

// padLeft left-pads b with zero bytes until it is exactly size bytes long.
// Returns b unchanged if len(b) >= size.
func padLeft(b []byte, size int) []byte {
	if len(b) >= size {
		return b
	}
	padded := make([]byte, size)
	copy(padded[size-len(b):], b)
	return padded
}
