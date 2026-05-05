package crypto

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/smedje/smedje/pkg/forge"
)

func TestJWTRS256(t *testing.T) {
	g := &JWTRS256{}
	out, err := g.Generate(context.Background(), forge.Options{})
	if err != nil {
		t.Fatal(err)
	}
	fields := out.PrimaryFields()
	if len(fields) != 3 {
		t.Fatalf("field count = %d, want 3", len(fields))
	}
	if fields[0].Key != "private-key" {
		t.Fatalf("fields[0].Key = %q, want private-key", fields[0].Key)
	}
	if !fields[0].Sensitive {
		t.Fatal("private-key must be sensitive")
	}
	if fields[1].Key != "public-key" {
		t.Fatalf("fields[1].Key = %q, want public-key", fields[1].Key)
	}
	if fields[2].Key != "jwks" {
		t.Fatalf("fields[2].Key = %q, want jwks", fields[2].Key)
	}
}

func TestJWTRS256PEM(t *testing.T) {
	g := &JWTRS256{}
	out, _ := g.Generate(context.Background(), forge.Options{})
	fields := out.PrimaryFields()
	if !strings.Contains(fields[0].Value, "-----BEGIN PRIVATE KEY-----") {
		t.Fatal("private-key is not a PKCS#8 PEM block")
	}
	if !strings.Contains(fields[1].Value, "-----BEGIN PUBLIC KEY-----") {
		t.Fatal("public-key is not a PKIX PEM block")
	}
}

func TestJWTRS256JWKS(t *testing.T) {
	g := &JWTRS256{}
	out, _ := g.Generate(context.Background(), forge.Options{})
	fields := out.PrimaryFields()
	var jwks map[string]interface{}
	if err := json.Unmarshal([]byte(fields[2].Value), &jwks); err != nil {
		t.Fatalf("jwks not valid JSON: %v", err)
	}
	keys := jwks["keys"].([]interface{})
	jwk := keys[0].(map[string]interface{})
	if jwk["kty"] != "RSA" {
		t.Fatalf("kty = %q, want RSA", jwk["kty"])
	}
}

func TestJWTRS256KidFlag(t *testing.T) {
	g := &JWTRS256{}
	out, err := g.Generate(context.Background(), forge.Options{Params: map[string]string{"kid": "my-key"}})
	if err != nil {
		t.Fatal(err)
	}
	fields := out.PrimaryFields()
	var jwks map[string]interface{}
	json.Unmarshal([]byte(fields[2].Value), &jwks)
	keys := jwks["keys"].([]interface{})
	jwk := keys[0].(map[string]interface{})
	if jwk["kid"] != "my-key" {
		t.Fatalf("kid = %q, want my-key", jwk["kid"])
	}
}

func TestJWTRS256Flags(t *testing.T) {
	g := &JWTRS256{}
	flags := g.Flags()
	findFlag(t, flags, "kid")
}
