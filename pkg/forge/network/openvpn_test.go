package network

import (
	"context"
	"strings"
	"testing"

	"github.com/smedje/smedje/pkg/forge"
)

func TestOpenVPNTLSAuth(t *testing.T) {
	g := &OpenVPNTLSAuth{}
	out, err := g.Generate(context.Background(), forge.Options{})
	if err != nil {
		t.Fatal(err)
	}
	val := out.PrimaryFields()[0].Value
	if !strings.Contains(val, "-----BEGIN OpenVPN Static key V1-----") {
		t.Fatal("missing OpenVPN key header")
	}
	if !strings.Contains(val, "-----END OpenVPN Static key V1-----") {
		t.Fatal("missing OpenVPN key footer")
	}
}

func TestOpenVPNTLSAuthLineFormat(t *testing.T) {
	g := &OpenVPNTLSAuth{}
	out, err := g.Generate(context.Background(), forge.Options{})
	if err != nil {
		t.Fatal(err)
	}
	val := out.PrimaryFields()[0].Value
	lines := strings.Split(val, "\n")

	// Find hex body lines (between header and footer).
	inBody := false
	hexLineCount := 0
	for _, line := range lines {
		if line == "-----BEGIN OpenVPN Static key V1-----" {
			inBody = true
			continue
		}
		if line == "-----END OpenVPN Static key V1-----" {
			inBody = false
			continue
		}
		if inBody && len(line) > 0 {
			hexLineCount++
			if len(line) > 32 {
				t.Fatalf("hex line too long: %d chars (max 32)", len(line))
			}
		}
	}

	// 2048 bits = 256 bytes = 512 hex chars; at 32 chars/line = 16 lines.
	if hexLineCount != 16 {
		t.Fatalf("hex line count = %d, want 16", hexLineCount)
	}
}
