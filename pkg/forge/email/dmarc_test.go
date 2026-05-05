package email

import (
	"context"
	"testing"

	"github.com/smedje/smedje/pkg/forge"
)

func TestDMARC(t *testing.T) {
	g := &DMARC{}
	out, err := g.Generate(context.Background(), forge.Options{
		Params: map[string]string{
			"domain": "example.com",
			"policy": "reject",
			"rua":    "dmarc@example.com",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	fields := out.PrimaryFields()
	if fields[0].Value != "_dmarc.example.com" {
		t.Fatalf("record-name = %q", fields[0].Value)
	}
	expected := "v=DMARC1; p=reject; rua=mailto:dmarc@example.com"
	if fields[1].Value != expected {
		t.Fatalf("record-value = %q, want %q", fields[1].Value, expected)
	}
}

func TestDMARCMissingDomain(t *testing.T) {
	g := &DMARC{}
	_, err := g.Generate(context.Background(), forge.Options{})
	if err == nil {
		t.Fatal("expected error for missing domain")
	}
}
