package tls

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/smedje/smedje/pkg/forge"
)

func TestCSR(t *testing.T) {
	g := &CSR{}
	out, err := g.Generate(context.Background(), forge.Options{
		Params: map[string]string{"cn": "test.local", "san": "test.local,192.168.1.1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	fields := out.PrimaryFields()
	if len(fields) != 2 {
		t.Fatalf("fields = %d, want 2", len(fields))
	}
	if fields[0].Key != "csr" {
		t.Fatalf("fields[0].Key = %q, want %q", fields[0].Key, "csr")
	}

	block, _ := pem.Decode([]byte(fields[0].Value))
	if block == nil || block.Type != "CERTIFICATE REQUEST" {
		t.Fatal("expected CERTIFICATE REQUEST PEM block")
	}
	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if csr.Subject.CommonName != "test.local" {
		t.Fatalf("CN = %q, want %q", csr.Subject.CommonName, "test.local")
	}
}

func TestCSRAlgoFlag(t *testing.T) {
	g := &CSR{}
	for _, algo := range []string{"ed25519", "rsa", "ecdsa"} {
		t.Run(algo, func(t *testing.T) {
			_, err := g.Generate(context.Background(), forge.Options{
				Params: map[string]string{"cn": "test", "algo": algo},
			})
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
