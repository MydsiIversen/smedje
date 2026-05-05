package id

import (
	"context"
	"testing"

	"github.com/smedje/smedje/pkg/forge"
)

func TestUUIDv6Format(t *testing.T) {
	g := &UUIDv6{}
	out, err := g.Generate(context.Background(), forge.Options{})
	if err != nil {
		t.Fatal(err)
	}
	val := out.PrimaryFields()[0].Value
	if len(val) != 36 {
		t.Errorf("expected 36 chars, got %d: %s", len(val), val)
	}
	if val[14] != '6' {
		t.Errorf("version nibble = %c, want '6'", val[14])
	}
	v := val[19]
	if v != '8' && v != '9' && v != 'a' && v != 'b' {
		t.Errorf("variant nibble = %c, want 8/9/a/b", v)
	}
}

func TestUUIDv6Sortable(t *testing.T) {
	g := &UUIDv6{}
	prev := ""
	for range 100 {
		out, err := g.Generate(context.Background(), forge.Options{})
		if err != nil {
			t.Fatal(err)
		}
		val := out.PrimaryFields()[0].Value
		if prev != "" && val < prev {
			t.Errorf("not sorted: %s < %s", val, prev)
		}
		prev = val
	}
}
