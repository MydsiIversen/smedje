package network

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/smedje/smedje/pkg/forge"
)

var macPattern = regexp.MustCompile(`^[0-9a-f]{2}(:[0-9a-f]{2}){5}$`)

func TestMACFormat(t *testing.T) {
	g := &MAC{}
	out, err := g.Generate(context.Background(), forge.Options{})
	if err != nil {
		t.Fatal(err)
	}

	val := out.PrimaryFields()[0].Value
	if !macPattern.MatchString(val) {
		t.Errorf("output %q does not match MAC pattern", val)
	}
}

func TestMACLocallyAdministered(t *testing.T) {
	g := &MAC{}
	for i := 0; i < 100; i++ {
		out, err := g.Generate(context.Background(), forge.Options{})
		if err != nil {
			t.Fatal(err)
		}
		val := out.PrimaryFields()[0].Value
		var first byte
		fmt.Sscanf(val[:2], "%02x", &first)

		if first&0x02 == 0 {
			t.Errorf("iteration %d: locally-administered bit not set: %s", i, val)
		}
		if first&0x01 != 0 {
			t.Errorf("iteration %d: unicast bit not cleared: %s", i, val)
		}
	}
}

func TestMACUniqueness(t *testing.T) {
	g := &MAC{}
	seen := make(map[string]struct{}, 1000)
	for i := 0; i < 1000; i++ {
		out, err := g.Generate(context.Background(), forge.Options{})
		if err != nil {
			t.Fatal(err)
		}
		val := out.PrimaryFields()[0].Value
		if _, exists := seen[val]; exists {
			t.Fatalf("duplicate MAC at iteration %d: %s", i, val)
		}
		seen[val] = struct{}{}
	}
}

func TestMACFlags(t *testing.T) {
	g := &MAC{}
	fd, ok := (forge.Generator)(g).(forge.FlagDescriber)
	if !ok {
		t.Fatal("MAC does not implement FlagDescriber")
	}
	flags := fd.Flags()
	if len(flags) != 1 {
		t.Fatalf("got %d flags, want 1", len(flags))
	}
}

func TestMACMetadata(t *testing.T) {
	g := &MAC{}
	if g.Name() != "mac" {
		t.Errorf("Name() = %q, want %q", g.Name(), "mac")
	}
	if g.Category() != forge.CategoryNetwork {
		t.Errorf("Category() = %q, want %q", g.Category(), forge.CategoryNetwork)
	}
}
