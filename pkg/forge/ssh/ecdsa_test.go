package ssh

import (
	"context"
	"strings"
	"testing"

	"github.com/smedje/smedje/pkg/forge"
)

func TestSSHECDSA(t *testing.T) {
	g := &ECDSA{}
	out, err := g.Generate(context.Background(), forge.Options{})
	if err != nil {
		t.Fatal(err)
	}
	fields := out.PrimaryFields()
	if !strings.HasPrefix(fields[1].Value, "ecdsa-sha2-nistp256 ") {
		t.Fatalf("expected ecdsa prefix, got %q", fields[1].Value[:30])
	}
}

func TestSSHECDSAP384(t *testing.T) {
	g := &ECDSA{}
	out, err := g.Generate(context.Background(), forge.Options{
		Params: map[string]string{"curve": "p384"},
	})
	if err != nil {
		t.Fatal(err)
	}
	fields := out.PrimaryFields()
	if !strings.HasPrefix(fields[1].Value, "ecdsa-sha2-nistp384 ") {
		t.Fatalf("expected ecdsa-p384 prefix, got %q", fields[1].Value[:30])
	}
}
