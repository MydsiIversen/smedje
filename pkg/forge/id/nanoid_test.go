package id

import (
	"context"
	"testing"

	"github.com/smedje/smedje/pkg/forge"
)

func TestNanoIDDefaultLength(t *testing.T) {
	g := &NanoID{}
	out, err := g.Generate(context.Background(), forge.Options{})
	if err != nil {
		t.Fatal(err)
	}
	val := out.PrimaryFields()[0].Value
	if len(val) != 21 {
		t.Errorf("expected 21 chars, got %d: %s", len(val), val)
	}
}

func TestNanoIDCustomLength(t *testing.T) {
	g := &NanoID{}
	out, err := g.Generate(context.Background(), forge.Options{
		Params: map[string]string{"length": "10"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.PrimaryFields()[0].Value) != 10 {
		t.Errorf("expected 10 chars, got %d", len(out.PrimaryFields()[0].Value))
	}
}

func TestNanoIDURLSafe(t *testing.T) {
	g := &NanoID{}
	for range 100 {
		out, err := g.Generate(context.Background(), forge.Options{})
		if err != nil {
			t.Fatal(err)
		}
		for _, c := range out.PrimaryFields()[0].Value {
			found := false
			for _, a := range defaultNanoIDAlphabet {
				if c == a {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("char %c not in default alphabet", c)
			}
		}
	}
}

func TestNanoIDFlags(t *testing.T) {
	g := &NanoID{}
	fd, ok := (forge.Generator)(g).(forge.FlagDescriber)
	if !ok {
		t.Fatal("NanoID does not implement FlagDescriber")
	}
	flags := fd.Flags()
	if len(flags) != 2 {
		t.Fatalf("got %d flags, want 2", len(flags))
	}
}

func TestNanoIDUniqueness(t *testing.T) {
	g := &NanoID{}
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
