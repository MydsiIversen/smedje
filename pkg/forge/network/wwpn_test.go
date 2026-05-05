package network

import (
	"context"
	"strings"
	"testing"

	"github.com/smedje/smedje/pkg/forge"
)

func TestWWPNFormat(t *testing.T) {
	g := &WWPN{}
	out, err := g.Generate(context.Background(), forge.Options{})
	if err != nil {
		t.Fatal(err)
	}
	val := out.PrimaryFields()[0].Value
	parts := strings.Split(val, ":")
	if len(parts) != 8 {
		t.Fatalf("WWPN %q: got %d colon-separated parts, want 8", val, len(parts))
	}
	for i, p := range parts {
		if len(p) != 2 {
			t.Errorf("WWPN part %d %q: length = %d, want 2", i, p, len(p))
		}
	}
}

func TestWWPNNAA5(t *testing.T) {
	g := &WWPN{}
	for i := 0; i < 100; i++ {
		out, err := g.Generate(context.Background(), forge.Options{})
		if err != nil {
			t.Fatal(err)
		}
		val := out.PrimaryFields()[0].Value
		if !strings.HasPrefix(val, "5") {
			t.Errorf("iteration %d: WWPN %q does not start with '5' (NAA 5)", i, val)
		}
	}
}

func TestWWPNUniqueness(t *testing.T) {
	g := &WWPN{}
	seen := make(map[string]struct{}, 1000)
	for i := 0; i < 1000; i++ {
		out, err := g.Generate(context.Background(), forge.Options{})
		if err != nil {
			t.Fatal(err)
		}
		val := out.PrimaryFields()[0].Value
		if _, exists := seen[val]; exists {
			t.Fatalf("duplicate WWPN at iteration %d: %s", i, val)
		}
		seen[val] = struct{}{}
	}
}

func TestWWPNMetadata(t *testing.T) {
	g := &WWPN{}
	if g.Name() != "wwpn" {
		t.Errorf("Name() = %q, want %q", g.Name(), "wwpn")
	}
	if g.Group() != "network" {
		t.Errorf("Group() = %q, want %q", g.Group(), "network")
	}
	if g.Category() != forge.CategoryNetwork {
		t.Errorf("Category() = %q, want %q", g.Category(), forge.CategoryNetwork)
	}
}
