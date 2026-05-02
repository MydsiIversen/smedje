package explain

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"
)

func init() {
	Register(&uuidDetector{})
	Register(&ulidDetector{})
	Register(&nanoidDetector{})
	Register(&snowflakeDetector{})
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
