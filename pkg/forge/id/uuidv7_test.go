package id

import (
	"context"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/smedje/smedje/pkg/forge"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestUUIDv7Format(t *testing.T) {
	g := &UUIDv7{}
	out, err := g.Generate(context.Background(), forge.Options{})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	val := out.Fields[0].Value
	if !uuidPattern.MatchString(val) {
		t.Errorf("output %q does not match UUIDv7 pattern", val)
	}
}

func TestUUIDv7VersionAndVariant(t *testing.T) {
	g := &UUIDv7{}
	for i := 0; i < 100; i++ {
		out, err := g.Generate(context.Background(), forge.Options{})
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}
		val := out.Fields[0].Value

		// Version nibble is at position 14 (0-indexed in the hex string without dashes).
		noDash := strings.ReplaceAll(val, "-", "")
		if noDash[12] != '7' {
			t.Errorf("version nibble = %c, want '7'", noDash[12])
		}

		// Variant bits: character at position 16 must be 8, 9, a, or b.
		v := noDash[16]
		if v != '8' && v != '9' && v != 'a' && v != 'b' {
			t.Errorf("variant nibble = %c, want 8/9/a/b", v)
		}
	}
}

func TestUUIDv7Uniqueness(t *testing.T) {
	g := &UUIDv7{}
	seen := make(map[string]struct{}, 1000)
	for i := 0; i < 1000; i++ {
		out, err := g.Generate(context.Background(), forge.Options{})
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}
		val := out.Fields[0].Value
		if _, exists := seen[val]; exists {
			t.Fatalf("duplicate UUID at iteration %d: %s", i, val)
		}
		seen[val] = struct{}{}
	}
}

func TestUUIDv7TimestampMonotonic(t *testing.T) {
	g := &UUIDv7{}

	out1, err := g.Generate(context.Background(), forge.Options{})
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(2 * time.Millisecond)
	out2, err := g.Generate(context.Background(), forge.Options{})
	if err != nil {
		t.Fatal(err)
	}

	if out1.Fields[0].Value >= out2.Fields[0].Value {
		t.Errorf("UUIDs not time-ordered: %s >= %s", out1.Fields[0].Value, out2.Fields[0].Value)
	}
}

func TestUUIDv7Metadata(t *testing.T) {
	g := &UUIDv7{}
	if g.Name() != "v7" {
		t.Errorf("Name() = %q, want %q", g.Name(), "v7")
	}
	if g.Category() != forge.CategoryID {
		t.Errorf("Category() = %q, want %q", g.Category(), forge.CategoryID)
	}
}
