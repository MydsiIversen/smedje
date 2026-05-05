package network

import (
	"context"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/smedje/smedje/internal/bench"
	"github.com/smedje/smedje/internal/entropy"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	forge.Register(&OUIMAC{})
}

// OUIMAC generates a unicast MAC address whose first three octets come from a
// real vendor OUI in the built-in table. The remaining three octets are random.
// Use --oui to pin a specific prefix; omit it to pick one at random.
type OUIMAC struct{}

func (o *OUIMAC) Name() string             { return "oui" }
func (o *OUIMAC) Group() string            { return "mac" }
func (o *OUIMAC) Description() string      { return "Generate a MAC address with a real vendor OUI prefix" }
func (o *OUIMAC) Category() forge.Category { return forge.CategoryNetwork }

// Generate returns a MAC address prefixed with a vendor OUI.
func (o *OUIMAC) Generate(ctx context.Context, opts forge.Options) (*forge.Output, error) {
	format := "colon"
	if v, ok := opts.Params["format"]; ok && v != "" {
		format = v
	}
	if format != "colon" && format != "dash" && format != "dot" {
		return nil, fmt.Errorf("oui-mac: unknown format %q; choose colon, dash, or dot", format)
	}

	var prefix string
	var vendor string

	if v, ok := opts.Params["oui"]; ok && v != "" {
		prefix = normalizeOUI(v)
		name, found := ouiTable[prefix]
		if !found {
			return nil, fmt.Errorf("oui-mac: OUI %q not in table", prefix)
		}
		vendor = name
	} else {
		keys := ouiKeys()
		idx, err := randIndex(len(keys))
		if err != nil {
			return nil, fmt.Errorf("oui-mac: entropy: %w", err)
		}
		prefix = keys[idx]
		vendor = ouiTable[prefix]
	}

	// Parse the 3-byte prefix.
	var addr [6]byte
	parts := strings.Split(prefix, ":")
	if len(parts) != 3 {
		return nil, fmt.Errorf("oui-mac: malformed OUI %q in table", prefix)
	}
	for i, p := range parts {
		var b byte
		fmt.Sscanf(p, "%02X", &b)
		addr[i] = b
	}

	// Fill the remaining 3 bytes randomly.
	if _, err := entropy.Read(addr[3:]); err != nil {
		return nil, fmt.Errorf("oui-mac: entropy read: %w", err)
	}

	mac := formatMAC(addr, format)
	return forge.SingleArtifact("oui-mac",
		forge.Field{Key: "value", Value: mac},
		forge.Field{Key: "vendor", Value: vendor},
	), nil
}

func (o *OUIMAC) Flags() []forge.FlagDef {
	return []forge.FlagDef{
		{Name: "oui", Type: "string", Description: "Vendor OUI prefix (e.g. 00:50:56 for VMware). Random vendor if empty"},
		{Name: "format", Type: "string", Default: "colon", Description: "Output style: colon (aa:bb:cc), dash (aa-bb-cc), dot (aabb.ccdd Cisco)", Options: []string{"colon", "dash", "dot"}},
	}
}

func (o *OUIMAC) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.RunLegacy(ctx, o, 0)
}

// formatMAC formats a 6-byte address as a MAC string in the requested style.
// Styles: "colon" → aa:bb:cc:dd:ee:ff, "dash" → aa-bb-cc-dd-ee-ff,
// "dot" → aabb.ccdd.eeff.
func formatMAC(addr [6]byte, style string) string {
	switch style {
	case "dash":
		return fmt.Sprintf("%02x-%02x-%02x-%02x-%02x-%02x",
			addr[0], addr[1], addr[2], addr[3], addr[4], addr[5])
	case "dot":
		// Cisco dot notation: four-nibble groups.
		w0 := binary.BigEndian.Uint16(addr[0:2])
		w1 := binary.BigEndian.Uint16(addr[2:4])
		w2 := binary.BigEndian.Uint16(addr[4:6])
		return fmt.Sprintf("%04x.%04x.%04x", w0, w1, w2)
	default: // colon
		return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
			addr[0], addr[1], addr[2], addr[3], addr[4], addr[5])
	}
}

// normalizeOUI converts a raw OUI string to the canonical uppercase colon form
// used as keys in ouiTable. It accepts colons, dashes, and no separators.
func normalizeOUI(s string) string {
	s = strings.ToUpper(s)
	// Strip separators, then re-insert colons.
	clean := strings.ReplaceAll(strings.ReplaceAll(s, ":", ""), "-", "")
	if len(clean) != 6 {
		return s // return as-is; caller will get a "not in table" error
	}
	return clean[0:2] + ":" + clean[2:4] + ":" + clean[4:6]
}

// randIndex returns a cryptographically random index in [0, max).
func randIndex(max int) (int, error) {
	var b [4]byte
	if _, err := entropy.Read(b[:]); err != nil {
		return 0, err
	}
	v := binary.BigEndian.Uint32(b[:])
	return int(v) % max, nil
}
