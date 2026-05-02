package ssh

import (
	"context"
	"strings"
	"testing"

	"github.com/smedje/smedje/pkg/forge"

	gossh "golang.org/x/crypto/ssh"
)

func TestEd25519Generate(t *testing.T) {
	g := &Ed25519{}
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

	if !strings.Contains(privField.Value, "BEGIN OPENSSH PRIVATE KEY") {
		t.Error("private key missing PEM header")
	}

	if !strings.HasPrefix(pubField.Value, "ssh-ed25519 ") {
		t.Error("public key should start with ssh-ed25519")
	}
}

func TestEd25519KeyParseable(t *testing.T) {
	g := &Ed25519{}
	out, err := g.Generate(context.Background(), forge.Options{})
	if err != nil {
		t.Fatal(err)
	}

	_, err = gossh.ParsePublicKey([]byte(out.Fields[1].Value))
	if err != nil {
		// ParsePublicKey wants wire format; use ParseAuthorizedKey instead.
		_, _, _, _, err = gossh.ParseAuthorizedKey([]byte(out.Fields[1].Value))
		if err != nil {
			t.Fatalf("public key not parseable: %v", err)
		}
	}
}

func TestEd25519Uniqueness(t *testing.T) {
	g := &Ed25519{}
	keys := make(map[string]struct{}, 10)
	for i := 0; i < 10; i++ {
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

func TestEd25519Metadata(t *testing.T) {
	g := &Ed25519{}
	if g.Name() != "ed25519" {
		t.Errorf("Name() = %q, want %q", g.Name(), "ed25519")
	}
	if g.Category() != forge.CategoryCrypto {
		t.Errorf("Category() = %q, want %q", g.Category(), forge.CategoryCrypto)
	}
}
