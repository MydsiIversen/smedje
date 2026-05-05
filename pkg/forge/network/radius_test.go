package network

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/smedje/smedje/pkg/forge"
)

func TestRADIUSSecret(t *testing.T) {
	g := &RADIUSSecret{}
	out, err := g.Generate(context.Background(), forge.Options{})
	if err != nil {
		t.Fatal(err)
	}
	val := out.PrimaryFields()[0].Value
	b, err := base64.StdEncoding.DecodeString(val)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) != 24 {
		t.Fatalf("length = %d, want 24", len(b))
	}
}

func TestRADIUSSecretCustomLength(t *testing.T) {
	g := &RADIUSSecret{}
	out, err := g.Generate(context.Background(), forge.Options{
		Params: map[string]string{"length": "48"},
	})
	if err != nil {
		t.Fatal(err)
	}
	b, err := base64.StdEncoding.DecodeString(out.PrimaryFields()[0].Value)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) != 48 {
		t.Fatalf("length = %d, want 48", len(b))
	}
}
