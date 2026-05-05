package output

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/smedje/smedje/pkg/forge"
)

func TestTextSingleValue(t *testing.T) {
	o := forge.SingleArtifact("uuidv7", forge.Field{Key: "uuid", Value: "0192af8e-dead-7fff-beef-cafe12345678"})
	var buf bytes.Buffer
	if err := Text(&buf, o); err != nil {
		t.Fatal(err)
	}
	want := "0192af8e-dead-7fff-beef-cafe12345678\n"
	if buf.String() != want {
		t.Errorf("got %q, want %q", buf.String(), want)
	}
}

func TestTextMultiField(t *testing.T) {
	o := forge.SingleArtifact("ed25519-keypair",
		forge.Field{Key: "private-key", Value: "PRIVATE"},
		forge.Field{Key: "public-key", Value: "PUBLIC"},
	)
	var buf bytes.Buffer
	if err := Text(&buf, o); err != nil {
		t.Fatal(err)
	}
	want := "private-key: PRIVATE\npublic-key: PUBLIC\n"
	if buf.String() != want {
		t.Errorf("got %q, want %q", buf.String(), want)
	}
}

func TestTextMultiArtifact(t *testing.T) {
	o := &forge.Output{
		Name: "tls-chain",
		Artifacts: []forge.Artifact{
			{
				Label:  "root-ca",
				Fields: []forge.Field{{Key: "cert", Value: "ROOT_CERT"}},
			},
			{
				Label:  "leaf",
				Fields: []forge.Field{{Key: "cert", Value: "LEAF_CERT"}, {Key: "key", Value: "LEAF_KEY"}},
			},
		},
	}
	var buf bytes.Buffer
	if err := Text(&buf, o); err != nil {
		t.Fatal(err)
	}
	want := "--- root-ca ---\nROOT_CERT\n\n--- leaf ---\ncert: LEAF_CERT\nkey: LEAF_KEY\n"
	if buf.String() != want {
		t.Errorf("got %q, want %q", buf.String(), want)
	}
}

func TestJSONSingleArtifact(t *testing.T) {
	o := forge.SingleArtifact("uuidv7", forge.Field{Key: "uuid", Value: "abc-123"})
	var buf bytes.Buffer
	if err := JSON(&buf, o); err != nil {
		t.Fatal(err)
	}
	var m map[string]string
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatal(err)
	}
	if m["uuid"] != "abc-123" {
		t.Errorf("got uuid=%q, want %q", m["uuid"], "abc-123")
	}
}

func TestJSONMultiArtifact(t *testing.T) {
	o := &forge.Output{
		Name: "tls-chain",
		Artifacts: []forge.Artifact{
			{
				Label:  "root-ca",
				Fields: []forge.Field{{Key: "cert", Value: "ROOT"}},
			},
			{
				Label:  "leaf",
				Fields: []forge.Field{{Key: "cert", Value: "LEAF"}},
			},
		},
	}
	var buf bytes.Buffer
	if err := JSON(&buf, o); err != nil {
		t.Fatal(err)
	}
	var items []struct {
		Label  string            `json:"label"`
		Fields map[string]string `json:"fields"`
	}
	if err := json.Unmarshal(buf.Bytes(), &items); err != nil {
		t.Fatalf("expected array, got error: %v\nraw: %s", err, buf.String())
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Label != "root-ca" {
		t.Errorf("item[0].label = %q, want %q", items[0].Label, "root-ca")
	}
	if items[0].Fields["cert"] != "ROOT" {
		t.Errorf("item[0].fields.cert = %q, want %q", items[0].Fields["cert"], "ROOT")
	}
	if items[1].Label != "leaf" {
		t.Errorf("item[1].label = %q, want %q", items[1].Label, "leaf")
	}
}

func TestQuietMultiArtifact(t *testing.T) {
	o := &forge.Output{
		Name: "tls-chain",
		Artifacts: []forge.Artifact{
			{
				Label:  "root-ca",
				Fields: []forge.Field{{Key: "cert", Value: "ROOT_CERT"}},
			},
			{
				Label:  "leaf",
				Fields: []forge.Field{{Key: "cert", Value: "LEAF_CERT"}, {Key: "key", Value: "LEAF_KEY"}},
			},
		},
	}
	var buf bytes.Buffer
	if err := Quiet(&buf, o); err != nil {
		t.Fatal(err)
	}
	want := "ROOT_CERT\n\nLEAF_CERT\nLEAF_KEY\n"
	if buf.String() != want {
		t.Errorf("got %q, want %q", buf.String(), want)
	}
}

func TestEnvSingleArtifact(t *testing.T) {
	o := forge.SingleArtifact("ssh",
		forge.Field{Key: "private-key", Value: "PRIV"},
		forge.Field{Key: "public-key", Value: "PUB"},
	)
	var buf bytes.Buffer
	if err := Env(&buf, o, "ssh"); err != nil {
		t.Fatal(err)
	}
	want := "SSH_PRIVATE_KEY=PRIV\nSSH_PUBLIC_KEY=PUB\n"
	if buf.String() != want {
		t.Errorf("got %q, want %q", buf.String(), want)
	}
}

func TestEnvMultiArtifact(t *testing.T) {
	o := &forge.Output{
		Name: "tls-chain",
		Artifacts: []forge.Artifact{
			{
				Label:  "root-ca",
				Fields: []forge.Field{{Key: "cert", Value: "ROOT"}},
			},
			{
				Label:  "leaf",
				Fields: []forge.Field{{Key: "cert", Value: "LEAF"}, {Key: "key", Value: "KEY"}},
			},
		},
	}
	var buf bytes.Buffer
	if err := Env(&buf, o, "tls"); err != nil {
		t.Fatal(err)
	}
	want := "ROOT_CA_CERT=ROOT\nLEAF_CERT=LEAF\nLEAF_KEY=KEY\n"
	if buf.String() != want {
		t.Errorf("got %q, want %q", buf.String(), want)
	}
}

func TestPEMMultiArtifact(t *testing.T) {
	o := &forge.Output{
		Name: "tls-chain",
		Artifacts: []forge.Artifact{
			{
				Label:  "root-ca",
				Fields: []forge.Field{{Key: "cert", Value: "-----BEGIN CERTIFICATE-----\nROOT\n-----END CERTIFICATE-----\n"}},
			},
			{
				Label:  "leaf",
				Fields: []forge.Field{{Key: "cert", Value: "-----BEGIN CERTIFICATE-----\nLEAF\n-----END CERTIFICATE-----\n"}},
			},
		},
	}
	var buf bytes.Buffer
	if err := PEM(&buf, o); err != nil {
		t.Fatal(err)
	}
	want := "-----BEGIN CERTIFICATE-----\nROOT\n-----END CERTIFICATE-----\n-----BEGIN CERTIFICATE-----\nLEAF\n-----END CERTIFICATE-----\n"
	if buf.String() != want {
		t.Errorf("got %q, want %q", buf.String(), want)
	}
}

func TestRenderDispatch(t *testing.T) {
	o := forge.SingleArtifact("uuidv7", forge.Field{Key: "uuid", Value: "test-uuid"})

	tests := []struct {
		format string
		want   string
	}{
		{"text", "test-uuid\n"},
		{"quiet", "test-uuid\n"},
	}
	for _, tt := range tests {
		var buf bytes.Buffer
		if err := Render(&buf, o, tt.format); err != nil {
			t.Fatalf("Render(%q): %v", tt.format, err)
		}
		if buf.String() != tt.want {
			t.Errorf("Render(%q) = %q, want %q", tt.format, buf.String(), tt.want)
		}
	}
}

func TestBatchJSON(t *testing.T) {
	outputs := []*forge.Output{
		forge.SingleArtifact("uuidv7", forge.Field{Key: "uuid", Value: "aaa"}),
		forge.SingleArtifact("uuidv7", forge.Field{Key: "uuid", Value: "bbb"}),
	}
	var buf bytes.Buffer
	if err := RenderBatch(&buf, outputs, "json", BatchOptions{}); err != nil {
		t.Fatal(err)
	}
	var items []map[string]string
	if err := json.Unmarshal(buf.Bytes(), &items); err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	if items[0]["uuid"] != "aaa" {
		t.Errorf("items[0][uuid] = %q, want %q", items[0]["uuid"], "aaa")
	}
}

func TestBatchCSV(t *testing.T) {
	outputs := []*forge.Output{
		forge.SingleArtifact("uuidv7", forge.Field{Key: "uuid", Value: "aaa"}),
		forge.SingleArtifact("uuidv7", forge.Field{Key: "uuid", Value: "bbb"}),
	}
	var buf bytes.Buffer
	if err := RenderBatch(&buf, outputs, "csv", BatchOptions{}); err != nil {
		t.Fatal(err)
	}
	want := "uuidv7\naaa\nbbb\n"
	if buf.String() != want {
		t.Errorf("got %q, want %q", buf.String(), want)
	}
}

func TestBatchSQL(t *testing.T) {
	outputs := []*forge.Output{
		forge.SingleArtifact("uuidv7", forge.Field{Key: "uuid", Value: "aaa"}),
		forge.SingleArtifact("uuidv7", forge.Field{Key: "uuid", Value: "bbb"}),
	}
	var buf bytes.Buffer
	if err := RenderBatch(&buf, outputs, "sql", BatchOptions{SQLTable: "ids"}); err != nil {
		t.Fatal(err)
	}
	want := "INSERT INTO ids (uuid) VALUES\n  ('aaa'),\n  ('bbb');\n"
	if buf.String() != want {
		t.Errorf("got %q, want %q", buf.String(), want)
	}
}

func TestNilArtifacts(t *testing.T) {
	o := &forge.Output{Name: "empty"}
	var buf bytes.Buffer
	if err := Text(&buf, o); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "" {
		t.Errorf("expected empty output, got %q", buf.String())
	}
}
