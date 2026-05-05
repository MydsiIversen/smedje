package id

import (
	"context"
	"testing"

	"github.com/smedje/smedje/pkg/forge"
)

func TestUUIDv4Format(t *testing.T) {
	g := &UUIDv4{}
	out, err := g.Generate(context.Background(), forge.Options{})
	if err != nil {
		t.Fatal(err)
	}
	val := out.PrimaryFields()[0].Value
	if len(val) != 36 {
		t.Errorf("expected 36 chars, got %d: %s", len(val), val)
	}
	if val[14] != '4' {
		t.Errorf("version nibble = %c, want '4'", val[14])
	}
	v := val[19]
	if v != '8' && v != '9' && v != 'a' && v != 'b' {
		t.Errorf("variant nibble = %c, want 8/9/a/b", v)
	}
}

func TestUUIDv4Uniqueness(t *testing.T) {
	g := &UUIDv4{}
	seen := make(map[string]bool, 1000)
	for range 1000 {
		out, err := g.Generate(context.Background(), forge.Options{})
		if err != nil {
			t.Fatal(err)
		}
		if seen[out.PrimaryFields()[0].Value] {
			t.Fatal("duplicate")
		}
		seen[out.PrimaryFields()[0].Value] = true
	}
}
