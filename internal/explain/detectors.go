package explain

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/smedje/smedje/pkg/forge/network"
)

func init() {
	Register(&uuidDetector{})
	Register(&ulidDetector{})
	Register(&nanoidDetector{})
	Register(&snowflakeDetector{})
	Register(&macDetector{})
	Register(&jwtDetector{})
	Register(&sshPubKeyDetector{})
	Register(&pemDetector{})
	Register(&iqnDetector{})
	Register(&ageKeyDetector{})
	Register(&wgKeyDetector{})
}

var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-([0-9a-f])[0-9a-f]{3}-([0-9a-f])[0-9a-f]{3}-[0-9a-f]{12}$`)

type uuidDetector struct{}

func (d *uuidDetector) Name() string { return "UUID" }

func (d *uuidDetector) Detect(input string) (Match, bool) {
	input = strings.ToLower(strings.TrimSpace(input))
	m := uuidRegex.FindStringSubmatch(input)
	if m == nil {
		return Match{}, false
	}

	version := m[1]
	variant := m[2]

	fields := map[string]string{
		"version": "v" + version,
		"variant": variantName(variant),
	}

	// Decode timestamp for time-based versions.
	raw := strings.ReplaceAll(input, "-", "")
	bytes, _ := hex.DecodeString(raw)

	switch version {
	case "1":
		fields["format"] = "UUIDv1 (time-based)"
		if ts := decodeV1Timestamp(bytes); !ts.IsZero() {
			fields["timestamp"] = ts.UTC().Format(time.RFC3339Nano)
		}
	case "4":
		fields["format"] = "UUIDv4 (random)"
	case "6":
		fields["format"] = "UUIDv6 (reordered time)"
		if ts := decodeV6Timestamp(bytes); !ts.IsZero() {
			fields["timestamp"] = ts.UTC().Format(time.RFC3339Nano)
		}
	case "7":
		fields["format"] = "UUIDv7 (Unix time-ordered)"
		if ts := decodeV7Timestamp(bytes); !ts.IsZero() {
			fields["timestamp"] = ts.UTC().Format(time.RFC3339Nano)
		}
	case "8":
		fields["format"] = "UUIDv8 (custom)"
	default:
		fields["format"] = fmt.Sprintf("UUID (version %s)", version)
	}

	// Nil and max detection.
	if input == "00000000-0000-0000-0000-000000000000" {
		fields["format"] = "Nil UUID"
		fields["version"] = "nil"
	} else if input == "ffffffff-ffff-ffff-ffff-ffffffffffff" {
		fields["format"] = "Max UUID"
		fields["version"] = "max"
	}

	layout := uuidLayout(input, version, fields["version"])

	return Match{
		Format:     fields["format"],
		Confidence: 0.95,
		Fields:     fields,
		Layout:     layout,
	}, true
}

// uuidLayout returns the layout segments for a UUID string based on version.
func uuidLayout(input, version, resolvedVersion string) []LayoutSegment {
	// Nil and Max UUIDs: single meta segment.
	if resolvedVersion == "nil" || resolvedVersion == "max" {
		label := "nil"
		desc := "Nil UUID (all zeros)"
		if resolvedVersion == "max" {
			label = "max"
			desc = "Max UUID (all ones)"
		}
		return []LayoutSegment{
			{Start: 0, End: 36, Label: label, Type: "meta", Value: input, Description: desc},
		}
	}

	sep := func(pos int) LayoutSegment {
		return LayoutSegment{Start: pos, End: pos + 1, Label: "sep", Type: "meta", Value: "-", Description: "Separator"}
	}

	switch version {
	case "1":
		return []LayoutSegment{
			{Start: 0, End: 8, Label: "time-low", Type: "time", Value: input[0:8], Description: "Timestamp (low 32 bits)"},
			sep(8),
			{Start: 9, End: 13, Label: "time-mid", Type: "time", Value: input[9:13], Description: "Timestamp (mid 16 bits)"},
			sep(13),
			{Start: 14, End: 15, Label: "version", Type: "version", Value: input[14:15], Description: "Version (1)"},
			{Start: 15, End: 18, Label: "time-high", Type: "time", Value: input[15:18], Description: "Timestamp (high 12 bits)"},
			sep(18),
			{Start: 19, End: 20, Label: "variant", Type: "version", Value: input[19:20], Description: "Variant (RFC 9562)"},
			{Start: 20, End: 23, Label: "clock-seq", Type: "counter", Value: input[20:23], Description: "Clock sequence"},
			sep(23),
			{Start: 24, End: 36, Label: "node", Type: "meta", Value: input[24:36], Description: "Node (MAC address)"},
		}
	case "4":
		return []LayoutSegment{
			{Start: 0, End: 8, Label: "rand-a", Type: "random", Value: input[0:8], Description: "Random (32 bits)"},
			sep(8),
			{Start: 9, End: 13, Label: "rand-b", Type: "random", Value: input[9:13], Description: "Random (16 bits)"},
			sep(13),
			{Start: 14, End: 15, Label: "version", Type: "version", Value: input[14:15], Description: "Version (4)"},
			{Start: 15, End: 18, Label: "rand-c", Type: "random", Value: input[15:18], Description: "Random (12 bits)"},
			sep(18),
			{Start: 19, End: 20, Label: "variant", Type: "version", Value: input[19:20], Description: "Variant (RFC 9562)"},
			{Start: 20, End: 23, Label: "rand-d", Type: "random", Value: input[20:23], Description: "Random (high)"},
			sep(23),
			{Start: 24, End: 36, Label: "rand-e", Type: "random", Value: input[24:36], Description: "Random (low 48 bits)"},
		}
	case "6":
		return []LayoutSegment{
			{Start: 0, End: 8, Label: "time-high", Type: "time", Value: input[0:8], Description: "Timestamp (high 32 bits)"},
			sep(8),
			{Start: 9, End: 13, Label: "time-mid", Type: "time", Value: input[9:13], Description: "Timestamp (mid 16 bits)"},
			sep(13),
			{Start: 14, End: 15, Label: "version", Type: "version", Value: input[14:15], Description: "Version (6)"},
			{Start: 15, End: 18, Label: "time-low", Type: "time", Value: input[15:18], Description: "Timestamp (low 12 bits)"},
			sep(18),
			{Start: 19, End: 20, Label: "variant", Type: "version", Value: input[19:20], Description: "Variant (RFC 9562)"},
			{Start: 20, End: 23, Label: "clock-seq", Type: "counter", Value: input[20:23], Description: "Clock sequence"},
			sep(23),
			{Start: 24, End: 36, Label: "node", Type: "meta", Value: input[24:36], Description: "Node (48 bits)"},
		}
	case "7":
		return []LayoutSegment{
			{Start: 0, End: 8, Label: "time-high", Type: "time", Value: input[0:8], Description: "Timestamp (high 32 bits)"},
			sep(8),
			{Start: 9, End: 13, Label: "time-mid", Type: "time", Value: input[9:13], Description: "Timestamp (mid 16 bits)"},
			sep(13),
			{Start: 14, End: 15, Label: "version", Type: "version", Value: input[14:15], Description: "Version (7)"},
			{Start: 15, End: 18, Label: "rand-a", Type: "random", Value: input[15:18], Description: "Random A (12 bits)"},
			sep(18),
			{Start: 19, End: 20, Label: "variant", Type: "version", Value: input[19:20], Description: "Variant (RFC 9562)"},
			{Start: 20, End: 23, Label: "rand-b", Type: "random", Value: input[20:23], Description: "Random B (high)"},
			sep(23),
			{Start: 24, End: 36, Label: "rand-b-low", Type: "random", Value: input[24:36], Description: "Random B (low 48 bits)"},
		}
	case "8":
		return []LayoutSegment{
			{Start: 0, End: 8, Label: "custom-a", Type: "meta", Value: input[0:8], Description: "Custom data A (32 bits)"},
			sep(8),
			{Start: 9, End: 13, Label: "custom-b", Type: "meta", Value: input[9:13], Description: "Custom data B (16 bits)"},
			sep(13),
			{Start: 14, End: 15, Label: "version", Type: "version", Value: input[14:15], Description: "Version (8)"},
			{Start: 15, End: 18, Label: "custom-c", Type: "meta", Value: input[15:18], Description: "Custom data C (12 bits)"},
			sep(18),
			{Start: 19, End: 20, Label: "variant", Type: "version", Value: input[19:20], Description: "Variant (RFC 9562)"},
			{Start: 20, End: 23, Label: "custom-d", Type: "meta", Value: input[20:23], Description: "Custom data D (high)"},
			sep(23),
			{Start: 24, End: 36, Label: "custom-e", Type: "meta", Value: input[24:36], Description: "Custom data E (low 48 bits)"},
		}
	default:
		return []LayoutSegment{
			{Start: 0, End: 36, Label: "uuid", Type: "meta", Value: input, Description: fmt.Sprintf("UUID version %s", version)},
		}
	}
}

func variantName(nibble string) string {
	switch nibble {
	case "8", "9", "a", "b":
		return "RFC 9562"
	case "c", "d":
		return "Microsoft"
	case "e":
		return "Future"
	default:
		return "NCS"
	}
}

func decodeV1Timestamp(b []byte) time.Time {
	if len(b) < 8 {
		return time.Time{}
	}
	timeLow := uint64(binary.BigEndian.Uint32(b[0:4]))
	timeMid := uint64(binary.BigEndian.Uint16(b[4:6]))
	timeHi := uint64(binary.BigEndian.Uint16(b[6:8])) & 0x0FFF
	t := (timeHi << 48) | (timeMid << 32) | timeLow

	const uuidEpoch = 122192928000000000
	if t < uuidEpoch {
		return time.Time{}
	}
	unixNano := int64(t-uuidEpoch) * 100
	return time.Unix(0, unixNano)
}

func decodeV6Timestamp(b []byte) time.Time {
	if len(b) < 8 {
		return time.Time{}
	}
	timeHigh := uint64(binary.BigEndian.Uint32(b[0:4]))
	timeMid := uint64(binary.BigEndian.Uint16(b[4:6]))
	timeLow := uint64(binary.BigEndian.Uint16(b[6:8])) & 0x0FFF
	t := (timeHigh << 28) | (timeMid << 12) | timeLow

	const uuidEpoch = 122192928000000000
	if t < uuidEpoch {
		return time.Time{}
	}
	unixNano := int64(t-uuidEpoch) * 100
	return time.Unix(0, unixNano)
}

func decodeV7Timestamp(b []byte) time.Time {
	if len(b) < 6 {
		return time.Time{}
	}
	ms := uint64(b[0])<<40 | uint64(b[1])<<32 | uint64(b[2])<<24 |
		uint64(b[3])<<16 | uint64(b[4])<<8 | uint64(b[5])
	return time.UnixMilli(int64(ms))
}

type ulidDetector struct{}

func (d *ulidDetector) Name() string { return "ULID" }

func (d *ulidDetector) Detect(input string) (Match, bool) {
	input = strings.TrimSpace(input)
	if len(input) != 26 {
		return Match{}, false
	}
	for _, c := range input {
		if !isCrockfordChar(byte(c)) {
			return Match{}, false
		}
	}

	// Decode timestamp from first 10 chars.
	ms := decodeCrockford(input[:10])
	ts := time.UnixMilli(int64(ms))

	layout := []LayoutSegment{
		{Start: 0, End: 10, Label: "timestamp", Type: "time", Value: input[0:10], Description: "Timestamp (ms, 48 bits)"},
		{Start: 10, End: 26, Label: "randomness", Type: "random", Value: input[10:26], Description: "Randomness (80 bits)"},
	}

	return Match{
		Format:     "ULID",
		Confidence: 0.85,
		Fields: map[string]string{
			"timestamp": ts.UTC().Format(time.RFC3339Nano),
		},
		Layout: layout,
	}, true
}

const crockfordAlphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

func isCrockfordChar(b byte) bool {
	if b >= 'a' && b <= 'z' {
		b -= 32
	}
	for _, c := range []byte(crockfordAlphabet) {
		if b == c {
			return true
		}
	}
	return false
}

func decodeCrockford(s string) uint64 {
	var val uint64
	for _, c := range []byte(s) {
		if c >= 'a' && c <= 'z' {
			c -= 32
		}
		idx := strings.IndexByte(crockfordAlphabet, c)
		if idx < 0 {
			return 0
		}
		val = (val << 5) | uint64(idx)
	}
	return val
}

type nanoidDetector struct{}

func (d *nanoidDetector) Name() string { return "NanoID" }

func (d *nanoidDetector) Detect(input string) (Match, bool) {
	input = strings.TrimSpace(input)
	if len(input) < 10 || len(input) > 256 {
		return Match{}, false
	}
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_-"
	for _, c := range input {
		if !strings.ContainsRune(alphabet, c) {
			return Match{}, false
		}
	}
	// NanoID has lower confidence than UUID/ULID since many formats use this charset.
	layout := []LayoutSegment{
		{Start: 0, End: len(input), Label: "random", Type: "random", Value: input, Description: "Random bytes (URL-safe alphabet)"},
	}

	return Match{
		Format:     "NanoID (probable)",
		Confidence: 0.5,
		Fields: map[string]string{
			"length": fmt.Sprintf("%d", len(input)),
		},
		Layout: layout,
	}, true
}

type snowflakeDetector struct{}

func (d *snowflakeDetector) Name() string { return "Snowflake" }

func (d *snowflakeDetector) Detect(input string) (Match, bool) {
	input = strings.TrimSpace(input)
	if len(input) < 15 || len(input) > 20 {
		return Match{}, false
	}
	for _, c := range input {
		if c < '0' || c > '9' {
			return Match{}, false
		}
	}

	var n uint64
	for _, c := range input {
		n = n*10 + uint64(c-'0')
	}

	// Snowflake with smedje epoch (2024-01-01)
	const smedjeEpoch = 1704067200000
	ts := (n >> 22) + smedjeEpoch
	worker := (n >> 12) & 0x3FF
	seq := n & 0xFFF

	t := time.UnixMilli(int64(ts))
	if t.Year() < 2024 || t.Year() > 2100 {
		return Match{}, false
	}

	// Snowflake is a 64-bit integer; bit boundaries don't map cleanly to
	// decimal digits. Use approximate character ranges based on typical
	// 18-19 digit IDs.
	tsBound := len(input) - 5 // last ~5 digits encode worker+sequence
	if tsBound < 1 {
		tsBound = 1
	}
	layout := []LayoutSegment{
		{Start: 0, End: tsBound, Label: "timestamp-bits", Type: "time", Value: input[:tsBound], Description: "Encodes timestamp (41 bits, approximate digit range)"},
		{Start: tsBound, End: len(input), Label: "worker-seq", Type: "meta", Value: input[tsBound:], Description: "Encodes worker ID (10 bits) and sequence (12 bits)"},
	}

	return Match{
		Format:     "Snowflake ID",
		Confidence: 0.7,
		Fields: map[string]string{
			"timestamp": t.UTC().Format(time.RFC3339),
			"worker":    fmt.Sprintf("%d", worker),
			"sequence":  fmt.Sprintf("%d", seq),
		},
		Layout: layout,
	}, true
}

// --- MAC address detector ---

var macRegex = regexp.MustCompile(`^([0-9a-fA-F]{2}[:\-]){5}[0-9a-fA-F]{2}$`)

type macDetector struct{}

func (d *macDetector) Name() string { return "MAC" }

func (d *macDetector) Detect(input string) (Match, bool) {
	input = strings.TrimSpace(input)
	if !macRegex.MatchString(input) {
		return Match{}, false
	}

	// Normalise to colon-separated uppercase for OUI lookup.
	normalised := strings.ToUpper(strings.ReplaceAll(input, "-", ":"))
	prefix := normalised[:8] // first 3 octets, e.g. "00:50:56"

	vendor := network.LookupVendor(prefix)
	if vendor == "" {
		vendor = "Unknown"
	}

	return Match{
		Format:     "MAC Address",
		Confidence: 0.85,
		Fields: map[string]string{
			"vendor": vendor,
			"oui":    prefix,
		},
	}, true
}

// --- JWT detector ---

type jwtDetector struct{}

func (d *jwtDetector) Name() string { return "JWT" }

func (d *jwtDetector) Detect(input string) (Match, bool) {
	input = strings.TrimSpace(input)
	parts := strings.Split(input, ".")
	if len(parts) != 3 {
		return Match{}, false
	}

	// Decode header (first segment, base64url without padding).
	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return Match{}, false
	}

	var header struct {
		Alg string `json:"alg"`
		Typ string `json:"typ"`
		Kid string `json:"kid"`
	}
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return Match{}, false
	}

	if header.Alg == "" {
		return Match{}, false
	}

	fields := map[string]string{
		"algorithm": header.Alg,
	}
	if header.Typ != "" {
		fields["type"] = header.Typ
	}
	if header.Kid != "" {
		fields["kid"] = header.Kid
	}

	return Match{
		Format:     "JWT",
		Confidence: 0.90,
		Fields:     fields,
	}, true
}

// --- SSH public key detector ---

type sshPubKeyDetector struct{}

func (d *sshPubKeyDetector) Name() string { return "SSH Public Key" }

func (d *sshPubKeyDetector) Detect(input string) (Match, bool) {
	input = strings.TrimSpace(input)
	parts := strings.SplitN(input, " ", 3)
	if len(parts) < 2 {
		return Match{}, false
	}

	algo := parts[0]
	if !strings.HasPrefix(algo, "ssh-") && !strings.HasPrefix(algo, "ecdsa-") {
		return Match{}, false
	}

	fields := map[string]string{
		"algorithm": algo,
	}
	if len(parts) >= 3 {
		fields["comment"] = parts[2]
	}

	return Match{
		Format:     "SSH Public Key",
		Confidence: 0.90,
		Fields:     fields,
	}, true
}

// --- PEM detector ---

var pemTypeRegex = regexp.MustCompile(`^-----BEGIN ([A-Z][A-Z0-9 ]+)-----`)

type pemDetector struct{}

func (d *pemDetector) Name() string { return "PEM" }

func (d *pemDetector) Detect(input string) (Match, bool) {
	input = strings.TrimSpace(input)

	// Try strict decode first.
	if block, _ := pem.Decode([]byte(input)); block != nil {
		return Match{
			Format:     "PEM",
			Confidence: 0.95,
			Fields: map[string]string{
				"type": block.Type,
			},
		}, true
	}

	// Fallback: match the BEGIN marker even if the body is malformed/truncated.
	if m := pemTypeRegex.FindStringSubmatch(input); m != nil {
		return Match{
			Format:     "PEM",
			Confidence: 0.95,
			Fields: map[string]string{
				"type": m[1],
			},
		}, true
	}

	return Match{}, false
}

// --- iSCSI Qualified Name (IQN) detector ---

var iqnRegex = regexp.MustCompile(`^iqn\.(\d{4}-\d{2})\.([^:]+):(.+)$`)

type iqnDetector struct{}

func (d *iqnDetector) Name() string { return "IQN" }

func (d *iqnDetector) Detect(input string) (Match, bool) {
	input = strings.TrimSpace(input)
	m := iqnRegex.FindStringSubmatch(input)
	if m == nil {
		return Match{}, false
	}

	return Match{
		Format:     "iSCSI Qualified Name",
		Confidence: 0.95,
		Fields: map[string]string{
			"date":      m[1],
			"authority": m[2],
			"target":    m[3],
		},
	}, true
}

// --- age public key detector ---

type ageKeyDetector struct{}

func (d *ageKeyDetector) Name() string { return "age Key" }

func (d *ageKeyDetector) Detect(input string) (Match, bool) {
	input = strings.TrimSpace(input)
	if !strings.HasPrefix(input, "age1") || len(input) != 62 {
		return Match{}, false
	}

	return Match{
		Format:     "age Public Key",
		Confidence: 0.95,
		Fields: map[string]string{
			"type": "public key",
		},
	}, true
}

// --- WireGuard key detector ---

type wgKeyDetector struct{}

func (d *wgKeyDetector) Name() string { return "WireGuard Key" }

func (d *wgKeyDetector) Detect(input string) (Match, bool) {
	input = strings.TrimSpace(input)
	if len(input) != 44 {
		return Match{}, false
	}

	decoded, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return Match{}, false
	}
	if len(decoded) != 32 {
		return Match{}, false
	}

	return Match{
		Format:     "WireGuard Key (possible)",
		Confidence: 0.40,
		Fields: map[string]string{
			"decoded_length": "32 bytes",
		},
	}, true
}
