package crypto

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/smedje/smedje/pkg/forge"
)

func TestJWTHS256(t *testing.T) {
	g := &JWTHS256{}
	out, err := g.Generate(context.Background(), forge.Options{})
	if err != nil {
		t.Fatal(err)
	}
	fields := out.PrimaryFields()
	if fields[0].Key != "secret" {
		t.Fatalf("key = %q, want %q", fields[0].Key, "secret")
	}
	b, err := base64.StdEncoding.DecodeString(fields[0].Value)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) != 32 {
		t.Fatalf("secret length = %d, want 32", len(b))
	}
}

func TestJWTHS256Sensitive(t *testing.T) {
	g := &JWTHS256{}
	out, err := g.Generate(context.Background(), forge.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !out.PrimaryFields()[0].Sensitive {
		t.Fatal("secret field must be sensitive")
	}
}

func TestJWTHS256TooShort(t *testing.T) {
	g := &JWTHS256{}
	_, err := g.Generate(context.Background(), forge.Options{Params: map[string]string{"length": "8"}})
	if err == nil {
		t.Fatal("expected error for length < 16")
	}
}

func TestJWTHS256Flags(t *testing.T) {
	g := &JWTHS256{}
	flags := g.Flags()
	if len(flags) == 0 {
		t.Fatal("expected at least one flag")
	}
	found := false
	for _, f := range flags {
		if f.Name == "length" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected length flag")
	}
}
