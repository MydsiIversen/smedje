package tls

import (
	"context"
	"crypto/x509"
	"testing"

	"github.com/smedje/smedje/pkg/forge"
)

func TestMTLS(t *testing.T) {
	g := &MTLS{}
	out, err := g.Generate(context.Background(), forge.Options{
		Params: map[string]string{"cn": "test.local", "days": "30"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Artifacts) != 3 {
		t.Fatalf("artifacts = %d, want 3", len(out.Artifacts))
	}
	if out.Artifacts[0].Label != "ca" {
		t.Fatalf("artifact[0].Label = %q, want %q", out.Artifacts[0].Label, "ca")
	}
	if out.Artifacts[1].Label != "server" {
		t.Fatalf("artifact[1].Label = %q, want %q", out.Artifacts[1].Label, "server")
	}
	if out.Artifacts[2].Label != "client" {
		t.Fatalf("artifact[2].Label = %q, want %q", out.Artifacts[2].Label, "client")
	}

	caCert := parseCertPEM(t, out.Artifacts[0].Fields[0].Value)
	serverCert := parseCertPEM(t, out.Artifacts[1].Fields[0].Value)
	clientCert := parseCertPEM(t, out.Artifacts[2].Fields[0].Value)

	if !caCert.IsCA {
		t.Fatal("CA should be CA")
	}
	if err := serverCert.CheckSignatureFrom(caCert); err != nil {
		t.Fatalf("server not signed by CA: %v", err)
	}
	if err := clientCert.CheckSignatureFrom(caCert); err != nil {
		t.Fatalf("client not signed by CA: %v", err)
	}

	hasServerAuth := false
	for _, u := range serverCert.ExtKeyUsage {
		if u == x509.ExtKeyUsageServerAuth {
			hasServerAuth = true
		}
	}
	if !hasServerAuth {
		t.Fatal("server cert missing ServerAuth")
	}

	hasClientAuth := false
	for _, u := range clientCert.ExtKeyUsage {
		if u == x509.ExtKeyUsageClientAuth {
			hasClientAuth = true
		}
	}
	if !hasClientAuth {
		t.Fatal("client cert missing ClientAuth")
	}
}

func TestMTLSFilenames(t *testing.T) {
	g := &MTLS{}
	out, _ := g.Generate(context.Background(), forge.Options{})
	expected := []string{"ca.pem", "server.pem", "client.pem"}
	for i, a := range out.Artifacts {
		if a.Filename != expected[i] {
			t.Fatalf("artifact[%d].Filename = %q, want %q", i, a.Filename, expected[i])
		}
	}
}
