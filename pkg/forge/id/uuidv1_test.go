package id

import (
	"context"
	"testing"

	"github.com/smedje/smedje/pkg/forge"
)

func TestUUIDv1Format(t *testing.T) {
	g := &UUIDv1{}
	out, err := g.Generate(context.Background(), forge.Options{})
	if err != nil {
		t.Fatal(err)
	}
	val := out.Fields[0].Value
	if len(val) != 36 {
		t.Errorf("expected 36 chars, got %d: %s", len(val), val)
	}
	// Version nibble should be '1'
	if val[14] != '1' {
		t.Errorf("version nibble = %c, want '1'", val[14])
	}
	// Variant should be 8, 9, a, or b
	v := val[19]
	if v != '8' && v != '9' && v != 'a' && v != 'b' {
		t.Errorf("variant nibble = %c, want 8/9/a/b", v)
	}
}

func TestUUIDv1Uniqueness(t *testing.T) {
	g := &UUIDv1{}
	seen := make(map[string]bool, 1000)
	for range 1000 {
		out, err := g.Generate(context.Background(), forge.Options{})
		if err != nil {
			t.Fatal(err)
		}
		val := out.Fields[0].Value
		if seen[val] {
			t.Fatalf("duplicate UUID: %s", val)
		}
		seen[val] = true
	}
}
