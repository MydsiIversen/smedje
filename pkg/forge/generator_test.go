package forge

import (
	"context"
	"testing"
)

type mockFlagGenerator struct{}

func (m *mockFlagGenerator) Name() string        { return "mock" }
func (m *mockFlagGenerator) Group() string       { return "mock" }
func (m *mockFlagGenerator) Description() string { return "mock generator" }
func (m *mockFlagGenerator) Category() Category  { return CategoryID }
func (m *mockFlagGenerator) Generate(ctx context.Context, opts Options) (*Output, error) {
	return SingleArtifact("mock", Field{Key: "value", Value: "test"}), nil
}
func (m *mockFlagGenerator) Bench(ctx context.Context) (*BenchResult, error) {
	return &BenchResult{}, nil
}
func (m *mockFlagGenerator) Flags() []FlagDef {
	return []FlagDef{
		{Name: "length", Type: "int", Default: "21", Description: "ID length"},
		{Name: "style", Type: "string", Default: "hex", Description: "Output style", Options: []string{"hex", "base64"}},
	}
}

func TestFlagDescriber(t *testing.T) {
	var g Generator = &mockFlagGenerator{}
	fd, ok := g.(FlagDescriber)
	if !ok {
		t.Fatal("mockFlagGenerator does not implement FlagDescriber")
	}
	flags := fd.Flags()
	if len(flags) != 2 {
		t.Fatalf("got %d flags, want 2", len(flags))
	}
	if flags[0].Name != "length" {
		t.Errorf("flags[0].Name = %q, want %q", flags[0].Name, "length")
	}
	if len(flags[1].Options) != 2 {
		t.Errorf("flags[1].Options = %v, want 2 items", flags[1].Options)
	}
}

func TestSingleArtifact(t *testing.T) {
	out := SingleArtifact("test",
		Field{Key: "a", Value: "1"},
		Field{Key: "b", Value: "2", Sensitive: true},
	)
	if out.Name != "test" {
		t.Errorf("Name = %q, want %q", out.Name, "test")
	}
	if len(out.Artifacts) != 1 {
		t.Fatalf("got %d artifacts, want 1", len(out.Artifacts))
	}
	if len(out.Artifacts[0].Fields) != 2 {
		t.Fatalf("got %d fields, want 2", len(out.Artifacts[0].Fields))
	}
	if out.Artifacts[0].Fields[1].Sensitive != true {
		t.Error("field b should be sensitive")
	}
}

func TestPrimaryFields(t *testing.T) {
	out := SingleArtifact("test", Field{Key: "value", Value: "x"})
	fields := out.PrimaryFields()
	if len(fields) != 1 {
		t.Fatalf("got %d fields, want 1", len(fields))
	}
	if fields[0].Value != "x" {
		t.Errorf("value = %q, want %q", fields[0].Value, "x")
	}

	empty := &Output{Name: "empty"}
	if empty.PrimaryFields() != nil {
		t.Error("PrimaryFields on empty output should return nil")
	}
}

func TestMultiArtifactOutput(t *testing.T) {
	out := &Output{
		Name: "ca-chain",
		Artifacts: []Artifact{
			{Label: "root-ca", Fields: []Field{{Key: "cert", Value: "ROOT"}}},
			{Label: "leaf", Fields: []Field{{Key: "cert", Value: "LEAF"}}},
		},
	}
	if len(out.Artifacts) != 2 {
		t.Fatalf("got %d artifacts, want 2", len(out.Artifacts))
	}
	if out.PrimaryFields()[0].Value != "ROOT" {
		t.Error("PrimaryFields should return first artifact")
	}
}
