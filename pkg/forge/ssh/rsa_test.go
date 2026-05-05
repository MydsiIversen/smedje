package ssh

import (
	"context"
	"strings"
	"testing"

	"github.com/smedje/smedje/pkg/forge"
)

func TestSSHRSA(t *testing.T) {
	g := &RSA{}
	out, err := g.Generate(context.Background(), forge.Options{})
	if err != nil {
		t.Fatal(err)
	}
	fields := out.PrimaryFields()
	if !strings.HasPrefix(fields[1].Value, "ssh-rsa ") {
		t.Fatalf("expected ssh-rsa prefix, got %q", fields[1].Value[:20])
	}
	if !fields[0].Sensitive {
		t.Fatal("private key should be sensitive")
	}
}
