package wireguard

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/smedje/smedje/pkg/forge"
)

func TestKeypairGenerate(t *testing.T) {
	g := &Keypair{}
	out, err := g.Generate(context.Background(), forge.Options{})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(out.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(out.Fields))
	}

	privField := out.Fields[0]
	pubField := out.Fields[1]

	if !privField.Sensitive {
		t.Error("private key should be marked sensitive")
	}

	privBytes, err := base64.StdEncoding.DecodeString(privField.Value)
	if err != nil {
		t.Fatalf("private key not valid base64: %v", err)
	}
	if len(privBytes) != 32 {
		t.Errorf("private key length = %d, want 32", len(privBytes))
	}

	pubBytes, err := base64.StdEncoding.DecodeString(pubField.Value)
	if err != nil {
		t.Fatalf("public key not valid base64: %v", err)
	}
	if len(pubBytes) != 32 {
		t.Errorf("public key length = %d, want 32", len(pubBytes))
	}
}

func TestKeypairClamping(t *testing.T) {
	g := &Keypair{}
	for i := 0; i < 50; i++ {
		out, err := g.Generate(context.Background(), forge.Options{})
		if err != nil {
			t.Fatal(err)
		}
		privBytes, _ := base64.StdEncoding.DecodeString(out.Fields[0].Value)
		if privBytes[0]&7 != 0 {
			t.Errorf("iteration %d: low 3 bits of priv[0] not cleared", i)
		}
		if privBytes[31]&128 != 0 {
			t.Errorf("iteration %d: high bit of priv[31] not cleared", i)
		}
		if privBytes[31]&64 == 0 {
			t.Errorf("iteration %d: bit 6 of priv[31] not set", i)
		}
	}
}

func TestKeypairUniqueness(t *testing.T) {
	g := &Keypair{}
	keys := make(map[string]struct{}, 50)
	for i := 0; i < 50; i++ {
		out, err := g.Generate(context.Background(), forge.Options{})
		if err != nil {
			t.Fatal(err)
		}
		pub := out.Fields[1].Value
		if _, exists := keys[pub]; exists {
			t.Fatalf("duplicate key at iteration %d", i)
		}
		keys[pub] = struct{}{}
	}
}

func TestKeypairMetadata(t *testing.T) {
	g := &Keypair{}
	if g.Name() != "keypair" {
		t.Errorf("Name() = %q, want %q", g.Name(), "keypair")
	}
	if g.Category() != forge.CategoryCrypto {
		t.Errorf("Category() = %q, want %q", g.Category(), forge.CategoryCrypto)
	}
}
