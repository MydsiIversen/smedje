package secret

import (
	"context"
	"strings"
	"testing"

	"github.com/smedje/smedje/pkg/forge"
)

func TestPasswordDefaultLength(t *testing.T) {
	g := &Password{}
	out, err := g.Generate(context.Background(), forge.Options{})
	if err != nil {
		t.Fatal(err)
	}
	pw := out.PrimaryFields()[0].Value
	if len(pw) != 24 {
		t.Errorf("default length = %d, want 24", len(pw))
	}
	if !out.PrimaryFields()[0].Sensitive {
		t.Error("password should be marked sensitive")
	}
}

func TestPasswordCustomLength(t *testing.T) {
	g := &Password{}
	opts := forge.Options{Params: map[string]string{"length": "32"}}
	out, err := g.Generate(context.Background(), opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.PrimaryFields()[0].Value) != 32 {
		t.Errorf("length = %d, want 32", len(out.PrimaryFields()[0].Value))
	}
}

func TestPasswordMinLength(t *testing.T) {
	g := &Password{}
	opts := forge.Options{Params: map[string]string{"length": "4"}}
	_, err := g.Generate(context.Background(), opts)
	if err == nil {
		t.Error("expected error for length < 8")
	}
}

func TestPasswordCharsets(t *testing.T) {
	g := &Password{}
	tests := []struct {
		charset string
		allowed string
	}{
		{"alpha", charsetLower + charsetUpper},
		{"alphanum", charsetLower + charsetUpper + charsetDigits},
		{"digits", charsetDigits},
	}
	for _, tc := range tests {
		opts := forge.Options{Params: map[string]string{"charset": tc.charset, "length": "100"}}
		out, err := g.Generate(context.Background(), opts)
		if err != nil {
			t.Fatalf("charset=%q: %v", tc.charset, err)
		}
		for _, c := range out.PrimaryFields()[0].Value {
			if !strings.ContainsRune(tc.allowed, c) {
				t.Errorf("charset=%q: unexpected char %q", tc.charset, c)
				break
			}
		}
	}
}

func TestPasswordUniqueness(t *testing.T) {
	g := &Password{}
	seen := make(map[string]struct{}, 100)
	for i := 0; i < 100; i++ {
		out, err := g.Generate(context.Background(), forge.Options{})
		if err != nil {
			t.Fatal(err)
		}
		pw := out.PrimaryFields()[0].Value
		if _, exists := seen[pw]; exists {
			t.Fatalf("duplicate password at iteration %d", i)
		}
		seen[pw] = struct{}{}
	}
}

func TestPasswordFlags(t *testing.T) {
	g := &Password{}
	fd, ok := (forge.Generator)(g).(forge.FlagDescriber)
	if !ok {
		t.Fatal("Password does not implement FlagDescriber")
	}
	flags := fd.Flags()
	if len(flags) != 2 {
		t.Fatalf("got %d flags, want 2", len(flags))
	}
}

func TestPasswordMetadata(t *testing.T) {
	g := &Password{}
	if g.Name() != "password" {
		t.Errorf("Name() = %q, want %q", g.Name(), "password")
	}
	if g.Category() != forge.CategorySecret {
		t.Errorf("Category() = %q, want %q", g.Category(), forge.CategorySecret)
	}
}
