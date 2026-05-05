package output

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/smedje/smedje/pkg/forge"
)

func TestWriteDir(t *testing.T) {
	dir := t.TempDir()
	out := forge.SingleArtifact("test", forge.Field{Key: "value", Value: "hello"})

	paths, err := WriteDir(dir, out, "text")
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 1 {
		t.Fatalf("got %d paths, want 1", len(paths))
	}

	data, err := os.ReadFile(paths[0])
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello\n" {
		t.Errorf("file content = %q, want %q", string(data), "hello\n")
	}
}

func TestWriteDirMultiArtifact(t *testing.T) {
	dir := t.TempDir()
	out := &forge.Output{
		Name: "ca-chain",
		Artifacts: []forge.Artifact{
			{Label: "root-ca", Fields: []forge.Field{{Key: "cert", Value: "ROOT"}}},
			{Label: "leaf", Fields: []forge.Field{{Key: "cert", Value: "LEAF"}}},
		},
	}

	paths, err := WriteDir(dir, out, "text")
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 2 {
		t.Fatalf("got %d paths, want 2", len(paths))
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("file %s does not exist", p)
		}
	}
}

func TestWriteDirWithFilename(t *testing.T) {
	dir := t.TempDir()
	out := &forge.Output{
		Name: "tls",
		Artifacts: []forge.Artifact{
			{Label: "cert", Filename: "server.pem", Fields: []forge.Field{{Key: "cert", Value: "CERT"}}},
		},
	}

	paths, err := WriteDir(dir, out, "text")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(paths[0]) != "server.pem" {
		t.Errorf("filename = %q, want server.pem", filepath.Base(paths[0]))
	}
}
