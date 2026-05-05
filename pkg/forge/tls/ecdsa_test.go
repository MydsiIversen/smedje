package tls

import (
	"context"
	"testing"

	"github.com/smedje/smedje/pkg/forge"
)

func TestECDSACert(t *testing.T) {
	g := &ECDSACert{}
	out, err := g.Generate(context.Background(), forge.Options{
		Params: map[string]string{"cn": "test.local", "days": "30"},
	})
	if err != nil {
		t.Fatal(err)
	}
	fields := out.PrimaryFields()
	if len(fields) != 2 {
		t.Fatalf("fields = %d, want 2", len(fields))
	}
	if fields[0].Key != "certificate" {
		t.Fatalf("fields[0].Key = %q, want %q", fields[0].Key, "certificate")
	}
}

func TestECDSACertP384(t *testing.T) {
	g := &ECDSACert{}
	_, err := g.Generate(context.Background(), forge.Options{
		Params: map[string]string{"curve": "p384"},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestECDSACertFlagDescriber(t *testing.T) {
	g := &ECDSACert{}
	fd, ok := interface{}(g).(forge.FlagDescriber)
	if !ok {
		t.Fatal("ECDSACert should implement FlagDescriber")
	}
	found := false
	for _, f := range fd.Flags() {
		if f.Name == "curve" {
			found = true
		}
	}
	if !found {
		t.Fatal("missing curve flag")
	}
}
