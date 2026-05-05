package network

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/smedje/smedje/pkg/forge"
)

var ouiMACColonPattern = regexp.MustCompile(`^[0-9a-f]{2}(:[0-9a-f]{2}){5}$`)

func TestOUIMACRandomLength(t *testing.T) {
	g := &OUIMAC{}
	out, err := g.Generate(context.Background(), forge.Options{})
	if err != nil {
		t.Fatal(err)
	}
	val := out.PrimaryFields()[0].Value
	if len(val) != 17 {
		t.Errorf("MAC length = %d, want 17; got %q", len(val), val)
	}
	if !ouiMACColonPattern.MatchString(val) {
		t.Errorf("MAC %q does not match colon pattern", val)
	}
}

func TestOUIMACRandomVendorPopulated(t *testing.T) {
	g := &OUIMAC{}
	out, err := g.Generate(context.Background(), forge.Options{})
	if err != nil {
		t.Fatal(err)
	}
	fields := out.PrimaryFields()
	if len(fields) < 2 {
		t.Fatalf("expected at least 2 fields, got %d", len(fields))
	}
	if fields[1].Key != "vendor" {
		t.Errorf("second field key = %q, want %q", fields[1].Key, "vendor")
	}
	if fields[1].Value == "" {
		t.Error("vendor field is empty")
	}
}

func TestOUIMACSpecificVMware(t *testing.T) {
	g := &OUIMAC{}
	out, err := g.Generate(context.Background(), forge.Options{
		Params: map[string]string{"oui": "00:50:56"},
	})
	if err != nil {
		t.Fatal(err)
	}
	fields := out.PrimaryFields()
	mac := fields[0].Value
	vendor := fields[1].Value

	if !strings.HasPrefix(strings.ToLower(mac), "00:50:56") {
		t.Errorf("MAC %q does not start with VMware OUI 00:50:56", mac)
	}
	if vendor != "VMware" {
		t.Errorf("vendor = %q, want %q", vendor, "VMware")
	}
}

func TestOUIMACUnknownPrefix(t *testing.T) {
	g := &OUIMAC{}
	_, err := g.Generate(context.Background(), forge.Options{
		Params: map[string]string{"oui": "DE:AD:BE"},
	})
	if err == nil {
		t.Fatal("expected error for unknown OUI, got nil")
	}
	if !strings.Contains(err.Error(), "not in table") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestOUIMACFormatDash(t *testing.T) {
	g := &OUIMAC{}
	out, err := g.Generate(context.Background(), forge.Options{
		Params: map[string]string{"style": "dash"},
	})
	if err != nil {
		t.Fatal(err)
	}
	val := out.PrimaryFields()[0].Value
	dashPattern := regexp.MustCompile(`^[0-9a-f]{2}(-[0-9a-f]{2}){5}$`)
	if !dashPattern.MatchString(val) {
		t.Errorf("dash format %q does not match expected pattern", val)
	}
}

func TestOUIMACFormatDot(t *testing.T) {
	g := &OUIMAC{}
	out, err := g.Generate(context.Background(), forge.Options{
		Params: map[string]string{"style": "dot"},
	})
	if err != nil {
		t.Fatal(err)
	}
	val := out.PrimaryFields()[0].Value
	dotPattern := regexp.MustCompile(`^[0-9a-f]{4}\.[0-9a-f]{4}\.[0-9a-f]{4}$`)
	if !dotPattern.MatchString(val) {
		t.Errorf("dot format %q does not match expected pattern", val)
	}
}

func TestOUIMACMetadata(t *testing.T) {
	g := &OUIMAC{}
	if g.Name() != "oui" {
		t.Errorf("Name() = %q, want %q", g.Name(), "oui")
	}
	if g.Group() != "mac" {
		t.Errorf("Group() = %q, want %q", g.Group(), "mac")
	}
	if g.Category() != forge.CategoryNetwork {
		t.Errorf("Category() = %q, want %q", g.Category(), forge.CategoryNetwork)
	}
}
