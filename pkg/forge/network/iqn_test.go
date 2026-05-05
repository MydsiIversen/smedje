package network

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/smedje/smedje/pkg/forge"
)

func TestIQNFormat(t *testing.T) {
	g := &IQN{}
	out, err := g.Generate(context.Background(), forge.Options{
		Params: map[string]string{
			"authority": "com.example",
			"target":    "storage.lun0",
			"date":      "2024-01",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	val := out.PrimaryFields()[0].Value
	want := "iqn.2024-01.example.com:storage.lun0"
	if val != want {
		t.Errorf("IQN = %q, want %q", val, want)
	}
}

func TestIQNFormatThreePart(t *testing.T) {
	g := &IQN{}
	out, err := g.Generate(context.Background(), forge.Options{
		Params: map[string]string{
			"authority": "com.example.storage",
			"target":    "lun0",
			"date":      "2025-03",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	val := out.PrimaryFields()[0].Value
	want := "iqn.2025-03.storage.example.com:lun0"
	if val != want {
		t.Errorf("IQN = %q, want %q", val, want)
	}
}

func TestIQNDefaultDate(t *testing.T) {
	g := &IQN{}
	fixedTime := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	out, err := g.Generate(context.Background(), forge.Options{
		Params: map[string]string{
			"authority": "com.example",
			"target":    "disk0",
		},
		Time: func() time.Time { return fixedTime },
	})
	if err != nil {
		t.Fatal(err)
	}
	val := out.PrimaryFields()[0].Value
	if !strings.HasPrefix(val, "iqn.2025-06.") {
		t.Errorf("IQN %q should start with iqn.2025-06.", val)
	}
}

func TestIQNMissingAuthority(t *testing.T) {
	g := &IQN{}
	_, err := g.Generate(context.Background(), forge.Options{
		Params: map[string]string{
			"target": "storage.lun0",
		},
	})
	if err == nil {
		t.Fatal("expected error for missing authority, got nil")
	}
	if !strings.Contains(err.Error(), "--authority") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestIQNMissingTarget(t *testing.T) {
	g := &IQN{}
	_, err := g.Generate(context.Background(), forge.Options{
		Params: map[string]string{
			"authority": "com.example",
		},
	})
	if err == nil {
		t.Fatal("expected error for missing target, got nil")
	}
	if !strings.Contains(err.Error(), "--target") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestIQNMetadata(t *testing.T) {
	g := &IQN{}
	if g.Name() != "iqn" {
		t.Errorf("Name() = %q, want %q", g.Name(), "iqn")
	}
	if g.Group() != "network" {
		t.Errorf("Group() = %q, want %q", g.Group(), "network")
	}
	if g.Category() != forge.CategoryNetwork {
		t.Errorf("Category() = %q, want %q", g.Category(), forge.CategoryNetwork)
	}
}
