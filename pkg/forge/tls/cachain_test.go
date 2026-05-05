package tls

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/smedje/smedje/pkg/forge"
)

func TestCAChainDepth3(t *testing.T) {
	g := &CAChain{}
	out, err := g.Generate(context.Background(), forge.Options{
		Params: map[string]string{"cn": "TestCA", "days": "30", "depth": "3"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Artifacts) != 3 {
		t.Fatalf("artifacts = %d, want 3", len(out.Artifacts))
	}
	if out.Artifacts[0].Label != "root-ca" {
		t.Fatalf("artifact[0].Label = %q, want %q", out.Artifacts[0].Label, "root-ca")
	}
	if out.Artifacts[1].Label != "intermediate-ca" {
		t.Fatalf("artifact[1].Label = %q, want %q", out.Artifacts[1].Label, "intermediate-ca")
	}
	if out.Artifacts[2].Label != "leaf" {
		t.Fatalf("artifact[2].Label = %q, want %q", out.Artifacts[2].Label, "leaf")
	}

	for i, a := range out.Artifacts {
		if len(a.Fields) != 2 {
			t.Fatalf("artifact[%d] fields = %d, want 2", i, len(a.Fields))
		}
		if a.Fields[0].Key != "certificate" {
			t.Fatalf("artifact[%d].Fields[0].Key = %q, want %q", i, a.Fields[0].Key, "certificate")
		}
		if a.Fields[1].Key != "private-key" {
			t.Fatalf("artifact[%d].Fields[1].Key = %q, want %q", i, a.Fields[1].Key, "private-key")
		}
		if !a.Fields[1].Sensitive {
			t.Fatalf("artifact[%d] private-key should be sensitive", i)
		}
	}

	rootCert := parseCertPEM(t, out.Artifacts[0].Fields[0].Value)
	intermediateCert := parseCertPEM(t, out.Artifacts[1].Fields[0].Value)
	leafCert := parseCertPEM(t, out.Artifacts[2].Fields[0].Value)

	if !rootCert.IsCA {
		t.Fatal("root should be CA")
	}
	if !intermediateCert.IsCA {
		t.Fatal("intermediate should be CA")
	}
	if leafCert.IsCA {
		t.Fatal("leaf should not be CA")
	}

	if err := intermediateCert.CheckSignatureFrom(rootCert); err != nil {
		t.Fatalf("intermediate not signed by root: %v", err)
	}
	if err := leafCert.CheckSignatureFrom(intermediateCert); err != nil {
		t.Fatalf("leaf not signed by intermediate: %v", err)
	}
}

func TestCAChainDepth2(t *testing.T) {
	g := &CAChain{}
	out, err := g.Generate(context.Background(), forge.Options{
		Params: map[string]string{"depth": "2"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Artifacts) != 2 {
		t.Fatalf("artifacts = %d, want 2", len(out.Artifacts))
	}
	if out.Artifacts[0].Label != "root-ca" {
		t.Fatalf("artifact[0].Label = %q, want %q", out.Artifacts[0].Label, "root-ca")
	}
	if out.Artifacts[1].Label != "leaf" {
		t.Fatalf("artifact[1].Label = %q, want %q", out.Artifacts[1].Label, "leaf")
	}
}

func TestCAChainFlagDescriber(t *testing.T) {
	g := &CAChain{}
	fd, ok := interface{}(g).(forge.FlagDescriber)
	if !ok {
		t.Fatal("CAChain should implement FlagDescriber")
	}
	flags := fd.Flags()
	if len(flags) != 4 {
		t.Fatalf("flags = %d, want 4", len(flags))
	}
}

func TestCAChainFilenames(t *testing.T) {
	g := &CAChain{}
	out, _ := g.Generate(context.Background(), forge.Options{
		Params: map[string]string{"depth": "3"},
	})
	expected := []string{"root-ca.pem", "intermediate-ca.pem", "leaf.pem"}
	for i, a := range out.Artifacts {
		if a.Filename != expected[i] {
			t.Fatalf("artifact[%d].Filename = %q, want %q", i, a.Filename, expected[i])
		}
	}
}

func parseCertPEM(t *testing.T, pemStr string) *x509.Certificate {
	t.Helper()
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		t.Fatal("no PEM block found")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	return cert
}
