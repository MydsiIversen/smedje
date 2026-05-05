package network

import (
	"context"
	"encoding/hex"
	"testing"

	"github.com/smedje/smedje/pkg/forge"
)

func TestIPsecPSK(t *testing.T) {
	g := &IPsecPSK{}
	out, err := g.Generate(context.Background(), forge.Options{})
	if err != nil {
		t.Fatal(err)
	}
	val := out.PrimaryFields()[0].Value
	b, err := hex.DecodeString(val)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) != 32 {
		t.Fatalf("length = %d, want 32", len(b))
	}
}

func TestIPsecPSKCustomLength(t *testing.T) {
	g := &IPsecPSK{}
	out, err := g.Generate(context.Background(), forge.Options{
		Params: map[string]string{"length": "64"},
	})
	if err != nil {
		t.Fatal(err)
	}
	b, err := hex.DecodeString(out.PrimaryFields()[0].Value)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) != 64 {
		t.Fatalf("length = %d, want 64", len(b))
	}
}

func TestIPsecPSKLengthValidation(t *testing.T) {
	g := &IPsecPSK{}
	_, err := g.Generate(context.Background(), forge.Options{
		Params: map[string]string{"length": "8"},
	})
	if err == nil {
		t.Fatal("expected error for length < 16")
	}
	_, err = g.Generate(context.Background(), forge.Options{
		Params: map[string]string{"length": "200"},
	})
	if err == nil {
		t.Fatal("expected error for length > 128")
	}
}
