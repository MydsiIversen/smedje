package crypto

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/json"
	"testing"

	"github.com/smedje/smedje/internal/entropy"
)

func TestJWKSFromRSA(t *testing.T) {
	key, _ := rsa.GenerateKey(entropy.Reader, 2048)
	b, err := jwksFromKey("test-kid", &key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	var jwks map[string]interface{}
	if err := json.Unmarshal(b, &jwks); err != nil {
		t.Fatal(err)
	}
	keys := jwks["keys"].([]interface{})
	if len(keys) != 1 {
		t.Fatalf("keys = %d, want 1", len(keys))
	}
	jwk := keys[0].(map[string]interface{})
	if jwk["kty"] != "RSA" {
		t.Fatalf("kty = %q, want RSA", jwk["kty"])
	}
	if jwk["kid"] != "test-kid" {
		t.Fatalf("kid = %q, want test-kid", jwk["kid"])
	}
}

func TestJWKSFromEC(t *testing.T) {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), entropy.Reader)
	b, err := jwksFromKey("ec-kid", &key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	var jwks map[string]interface{}
	json.Unmarshal(b, &jwks)
	keys := jwks["keys"].([]interface{})
	jwk := keys[0].(map[string]interface{})
	if jwk["kty"] != "EC" {
		t.Fatalf("kty = %q, want EC", jwk["kty"])
	}
}

func TestJWKSFromEd25519(t *testing.T) {
	pub, _, _ := ed25519.GenerateKey(entropy.Reader)
	b, err := jwksFromKey("ed-kid", pub)
	if err != nil {
		t.Fatal(err)
	}
	var jwks map[string]interface{}
	json.Unmarshal(b, &jwks)
	keys := jwks["keys"].([]interface{})
	jwk := keys[0].(map[string]interface{})
	if jwk["kty"] != "OKP" {
		t.Fatalf("kty = %q, want OKP", jwk["kty"])
	}
	if jwk["crv"] != "Ed25519" {
		t.Fatalf("crv = %q, want Ed25519", jwk["crv"])
	}
}
