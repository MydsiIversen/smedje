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

	return Match{
		Format:     fields["format"],
		Confidence: 0.95,
		Fields:     fields,
	}, true
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

	return Match{
		Format:     "ULID",
		Confidence: 0.85,
		Fields: map[string]string{
			"timestamp": ts.UTC().Format(time.RFC3339Nano),
		},
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
	return Match{
		Format:     "NanoID (probable)",
		Confidence: 0.5,
		Fields: map[string]string{
			"length": fmt.Sprintf("%d", len(input)),
		},
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

	return Match{
		Format:     "Snowflake ID",
		Confidence: 0.7,
		Fields: map[string]string{
			"timestamp": t.UTC().Format(time.RFC3339),
			"worker":    fmt.Sprintf("%d", worker),
			"sequence":  fmt.Sprintf("%d", seq),
		},
	}, true
}
