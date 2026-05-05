package wireguard

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/smedje/smedje/pkg/forge"
)

func TestMesh3Peers(t *testing.T) {
	g := &Mesh{}
	out, err := g.Generate(context.Background(), forge.Options{
		Params: map[string]string{"peers": "3"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Artifacts) != 3 {
		t.Fatalf("artifacts = %d, want 3", len(out.Artifacts))
	}
	for i, a := range out.Artifacts {
		if a.Label != fmt.Sprintf("peer-%d", i+1) {
			t.Fatalf("artifact[%d].Label = %q", i, a.Label)
		}
		if a.Filename != fmt.Sprintf("peer-%d.conf", i+1) {
			t.Fatalf("artifact[%d].Filename = %q", i, a.Filename)
		}

		var hasPrivKey, hasPubKey, hasConfig bool
		for _, f := range a.Fields {
			switch f.Key {
			case "private-key":
				hasPrivKey = true
				if !f.Sensitive {
					t.Fatal("private-key should be sensitive")
				}
			case "public-key":
				hasPubKey = true
			case "config":
				hasConfig = true
				if !strings.Contains(f.Value, "[Interface]") {
					t.Fatal("config missing [Interface]")
				}
				if !strings.Contains(f.Value, "[Peer]") {
					t.Fatal("config missing [Peer]")
				}
			}
		}
		if !hasPrivKey || !hasPubKey || !hasConfig {
			t.Fatalf("artifact[%d] missing fields: priv=%v pub=%v cfg=%v", i, hasPrivKey, hasPubKey, hasConfig)
		}
	}
}

func TestMeshPeerCrossReference(t *testing.T) {
	g := &Mesh{}
	out, err := g.Generate(context.Background(), forge.Options{
		Params: map[string]string{"peers": "2"},
	})
	if err != nil {
		t.Fatal(err)
	}

	pub2 := fieldValue(out.Artifacts[1], "public-key")
	cfg1 := fieldValue(out.Artifacts[0], "config")
	if !strings.Contains(cfg1, pub2) {
		t.Fatal("peer-1 config should reference peer-2 public key")
	}

	pub1 := fieldValue(out.Artifacts[0], "public-key")
	cfg2 := fieldValue(out.Artifacts[1], "config")
	if !strings.Contains(cfg2, pub1) {
		t.Fatal("peer-2 config should reference peer-1 public key")
	}
}

func TestMeshWithEndpoints(t *testing.T) {
	g := &Mesh{}
	out, err := g.Generate(context.Background(), forge.Options{
		Params: map[string]string{
			"peers":    "2",
			"endpoint": "10.0.0.1:51820,10.0.0.2:51820",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	cfg1 := fieldValue(out.Artifacts[0], "config")
	if !strings.Contains(cfg1, "10.0.0.2:51820") {
		t.Fatal("peer-1 config should reference peer-2 endpoint")
	}
}

func fieldValue(a forge.Artifact, key string) string {
	for _, f := range a.Fields {
		if f.Key == key {
			return f.Value
		}
	}
	return ""
}
