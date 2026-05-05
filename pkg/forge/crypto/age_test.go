package crypto

import (
	"context"
	"strings"
	"testing"

	"github.com/smedje/smedje/pkg/forge"
)

func TestAgeKeypair(t *testing.T) {
	g := &AgeKeypair{}
	out, err := g.Generate(context.Background(), forge.Options{})
	if err != nil {
		t.Fatal(err)
	}
	fields := out.PrimaryFields()
	if len(fields) != 2 {
		t.Fatalf("fields = %d, want 2", len(fields))
	}
	if !strings.HasPrefix(fields[0].Value, "AGE-SECRET-KEY-") {
		t.Fatalf("private key should start with AGE-SECRET-KEY-, got %q", fields[0].Value[:20])
	}
	if !strings.HasPrefix(fields[1].Value, "age1") {
		t.Fatalf("public key should start with age1, got %q", fields[1].Value[:10])
	}
	if !fields[0].Sensitive {
		t.Fatal("private key should be sensitive")
	}
}
