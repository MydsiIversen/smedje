package network

import (
	"context"
	"testing"

	"github.com/smedje/smedje/pkg/forge"
)

func TestSNMPCommunity(t *testing.T) {
	g := &SNMPCommunity{}
	out, err := g.Generate(context.Background(), forge.Options{})
	if err != nil {
		t.Fatal(err)
	}
	val := out.PrimaryFields()[0].Value
	if len(val) != 16 {
		t.Fatalf("length = %d, want 16", len(val))
	}
	// All characters must be in the allowed charset.
	for _, c := range val {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			t.Fatalf("unexpected character %q in community string", c)
		}
	}
}

func TestSNMPCommunityCustomLength(t *testing.T) {
	g := &SNMPCommunity{}
	out, err := g.Generate(context.Background(), forge.Options{
		Params: map[string]string{"length": "32"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.PrimaryFields()[0].Value) != 32 {
		t.Fatalf("length = %d, want 32", len(out.PrimaryFields()[0].Value))
	}
}
