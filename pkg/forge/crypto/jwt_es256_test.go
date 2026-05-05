package crypto

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/smedje/smedje/pkg/forge"
)

func TestJWTES256(t *testing.T) {
	g := &JWTES256{}
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

func TestJWTES256PEM(t *testing.T) {
	g := &JWTES256{}
	out, _ := g.Generate(context.Background(), forge.Options{})
	fields := out.PrimaryFields()
	if !strings.Contains(fields[0].Value, "-----BEGIN PRIVATE KEY-----") {
		t.Fatal("private-key is not a PKCS#8 PEM block")
	}
	if !strings.Contains(fields[1].Value, "-----BEGIN PUBLIC KEY-----") {
		t.Fatal("public-key is not a PKIX PEM block")
	}
}

func TestJWTES256JWKS(t *testing.T) {
	g := &JWTES256{}
	out, _ := g.Generate(context.Background(), forge.Options{})
	fields := out.PrimaryFields()
	var jwks map[string]interface{}
	if err := json.Unmarshal([]byte(fields[2].Value), &jwks); err != nil {
		t.Fatalf("jwks not valid JSON: %v", err)
	}
	keys := jwks["keys"].([]interface{})
	jwk := keys[0].(map[string]interface{})
	if jwk["kty"] != "EC" {
		t.Fatalf("kty = %q, want EC", jwk["kty"])
	}
	if jwk["crv"] != "P-256" {
		t.Fatalf("crv = %q, want P-256", jwk["crv"])
	}
}

func TestJWTES256Flags(t *testing.T) {
	g := &JWTES256{}
	flags := g.Flags()
	findFlag(t, flags, "kid")
}
