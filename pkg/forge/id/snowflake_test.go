package id

import (
	"context"
	"strconv"
	"testing"

	"github.com/smedje/smedje/pkg/forge"
)

func TestSnowflakeFormat(t *testing.T) {
	g := &Snowflake{}
	out, err := g.Generate(context.Background(), forge.Options{})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	val := out.Fields[0].Value
	n, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		t.Fatalf("output %q is not a valid int64: %v", val, err)
	}
	if n <= 0 {
		t.Errorf("snowflake ID should be positive, got %d", n)
	}
}

func TestSnowflakeUniqueness(t *testing.T) {
	g := &Snowflake{}
	seen := make(map[string]struct{}, 10000)
	for i := 0; i < 10000; i++ {
		out, err := g.Generate(context.Background(), forge.Options{})
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}
		val := out.Fields[0].Value
		if _, exists := seen[val]; exists {
			t.Fatalf("duplicate snowflake at iteration %d: %s", i, val)
		}
		seen[val] = struct{}{}
	}
}

func TestSnowflakeWorkerID(t *testing.T) {
	g := &Snowflake{}
	opts := forge.Options{Params: map[string]string{"worker": "42"}}
	out, err := g.Generate(context.Background(), opts)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	n, _ := strconv.ParseInt(out.Fields[0].Value, 10, 64)
	worker := (n >> 12) & 0x3FF
	if worker != 42 {
		t.Errorf("worker bits = %d, want 42", worker)
	}
}

func TestSnowflakeWorkerValidation(t *testing.T) {
	g := &Snowflake{}
	tests := []string{"-1", "1024", "abc"}
	for _, w := range tests {
		opts := forge.Options{Params: map[string]string{"worker": w}}
		_, err := g.Generate(context.Background(), opts)
		if err == nil {
			t.Errorf("worker=%q should have failed", w)
		}
	}
}

func TestSnowflakeMetadata(t *testing.T) {
	g := &Snowflake{}
	if g.Name() != "snowflake" {
		t.Errorf("Name() = %q, want %q", g.Name(), "snowflake")
	}
	if g.Category() != forge.CategoryID {
		t.Errorf("Category() = %q, want %q", g.Category(), forge.CategoryID)
	}
}
