package email

import (
	"context"
	"strings"
	"testing"

	"github.com/smedje/smedje/pkg/forge"
)

func TestDKIM(t *testing.T) {
	g := &DKIM{}
	out, err := g.Generate(context.Background(), forge.Options{
		Params: map[string]string{"selector": "mail", "domain": "example.com"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Artifacts) != 2 {
		t.Fatalf("artifacts = %d, want 2", len(out.Artifacts))
	}
	if out.Artifacts[0].Label != "private-key" {
		t.Fatalf("artifact[0].Label = %q, want private-key", out.Artifacts[0].Label)
	}
	if !strings.Contains(out.Artifacts[0].Fields[0].Value, "RSA PRIVATE KEY") {
		t.Fatal("expected RSA PRIVATE KEY PEM")
	}
	dnsName := out.Artifacts[1].Fields[0].Value
	if dnsName != "mail._domainkey.example.com" {
		t.Fatalf("dns name = %q, want %q", dnsName, "mail._domainkey.example.com")
	}
	dnsValue := out.Artifacts[1].Fields[1].Value
	if !strings.HasPrefix(dnsValue, "v=DKIM1; k=rsa; p=") {
		t.Fatalf("dns value should start with DKIM1 header, got %q", dnsValue[:30])
	}
}

func TestDKIMMissingSelectorDomain(t *testing.T) {
	g := &DKIM{}
	_, err := g.Generate(context.Background(), forge.Options{})
	if err == nil {
		t.Fatal("expected error for missing selector/domain")
	}
}
