package tls

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/smedje/smedje/pkg/forge"
)

func TestRSACert(t *testing.T) {
	g := &RSACert{}
	out, err := g.Generate(context.Background(), forge.Options{
		Params: map[string]string{"cn": "test.local", "days": "30", "bits": "2048"},
	})
	if err != nil {
		t.Fatal(err)
	}
	fields := out.PrimaryFields()
	if len(fields) != 2 {
		t.Fatalf("fields = %d, want 2", len(fields))
	}

	block, _ := pem.Decode([]byte(fields[1].Value))
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := key.(*rsa.PrivateKey); !ok {
		t.Fatalf("expected *rsa.PrivateKey, got %T", key)
	}
}

func TestRSACertFlagDescriber(t *testing.T) {
	g := &RSACert{}
	fd, ok := interface{}(g).(forge.FlagDescriber)
	if !ok {
		t.Fatal("RSACert should implement FlagDescriber")
	}
	flags := fd.Flags()
	found := false
	for _, f := range flags {
		if f.Name == "bits" {
			found = true
		}
	}
	if !found {
		t.Fatal("missing bits flag")
	}
}
