package tls

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"strings"
	"testing"

	"github.com/smedje/smedje/pkg/forge"
)

func TestSelfSignedGenerate(t *testing.T) {
	g := &SelfSigned{}
	out, err := g.Generate(context.Background(), forge.Options{})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(out.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(out.Fields))
	}

	certField := out.Fields[0]
	keyField := out.Fields[1]

	if !strings.Contains(certField.Value, "BEGIN CERTIFICATE") {
		t.Error("certificate missing PEM header")
	}
	if !strings.Contains(keyField.Value, "BEGIN PRIVATE KEY") {
		t.Error("private key missing PEM header")
	}
	if !keyField.Sensitive {
		t.Error("private key should be marked sensitive")
	}
}

func TestSelfSignedParseable(t *testing.T) {
	g := &SelfSigned{}
	out, err := g.Generate(context.Background(), forge.Options{})
	if err != nil {
		t.Fatal(err)
	}

	block, _ := pem.Decode([]byte(out.Fields[0].Value))
	if block == nil {
		t.Fatal("failed to decode certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("failed to parse certificate: %v", err)
	}

	if cert.Subject.CommonName != "localhost" {
		t.Errorf("CN = %q, want %q", cert.Subject.CommonName, "localhost")
	}
}

func TestSelfSignedWithSANs(t *testing.T) {
	g := &SelfSigned{}
	opts := forge.Options{
		Params: map[string]string{
			"cn":  "example.com",
			"san": "example.com,*.example.com,10.0.0.1",
		},
	}
	out, err := g.Generate(context.Background(), opts)
	if err != nil {
		t.Fatal(err)
	}

	block, _ := pem.Decode([]byte(out.Fields[0].Value))
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}

	if len(cert.DNSNames) != 2 {
		t.Errorf("expected 2 DNS SANs, got %d: %v", len(cert.DNSNames), cert.DNSNames)
	}
	if len(cert.IPAddresses) != 1 {
		t.Errorf("expected 1 IP SAN, got %d", len(cert.IPAddresses))
	}
}

func TestSelfSignedMetadata(t *testing.T) {
	g := &SelfSigned{}
	if g.Name() != "self-signed" {
		t.Errorf("Name() = %q, want %q", g.Name(), "self-signed")
	}
	if g.Category() != forge.CategoryCrypto {
		t.Errorf("Category() = %q, want %q", g.Category(), forge.CategoryCrypto)
	}
}
