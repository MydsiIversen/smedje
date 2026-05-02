package explain

import (
	"testing"
)

func TestUUIDv7Layout(t *testing.T) {
	input := "01912f2b-95a8-7b4c-b5a1-9e3c6a7d8f0e"
	m, ok := (&uuidDetector{}).Detect(input)
	if !ok {
		t.Fatal("expected UUIDv7 to be detected")
	}
	if m.Layout == nil {
		t.Fatal("expected Layout to be populated")
	}

	// UUIDv7 should have 11 segments: time-high, sep, time-mid, sep,
	// version, rand-a, sep, variant, rand-b, sep, rand-b-low.
	if got := len(m.Layout); got != 11 {
		t.Fatalf("expected 11 layout segments, got %d", got)
	}

	assertSegment(t, m.Layout[0], 0, 8, "time-high", "time")
	assertSegment(t, m.Layout[1], 8, 9, "sep", "meta")
	assertSegment(t, m.Layout[2], 9, 13, "time-mid", "time")
	assertSegment(t, m.Layout[3], 13, 14, "sep", "meta")
	assertSegment(t, m.Layout[4], 14, 15, "version", "version")
	assertSegment(t, m.Layout[5], 15, 18, "rand-a", "random")
	assertSegment(t, m.Layout[6], 18, 19, "sep", "meta")
	assertSegment(t, m.Layout[7], 19, 20, "variant", "version")
	assertSegment(t, m.Layout[8], 20, 23, "rand-b", "random")
	assertSegment(t, m.Layout[9], 23, 24, "sep", "meta")
	assertSegment(t, m.Layout[10], 24, 36, "rand-b-low", "random")

	assertLayoutCoversInput(t, m.Layout, len(input))
	assertLayoutValuesNonEmpty(t, m.Layout)
}

func TestUUIDv4Layout(t *testing.T) {
	input := "550e8400-e29b-41d4-a716-446655440000"
	m, ok := (&uuidDetector{}).Detect(input)
	if !ok {
		t.Fatal("expected UUIDv4 to be detected")
	}
	if m.Layout == nil {
		t.Fatal("expected Layout to be populated")
	}

	// v4 has 11 segments like v7.
	if got := len(m.Layout); got != 11 {
		t.Fatalf("expected 11 layout segments, got %d", got)
	}

	assertSegment(t, m.Layout[0], 0, 8, "rand-a", "random")
	assertSegment(t, m.Layout[4], 14, 15, "version", "version")
	assertLayoutCoversInput(t, m.Layout, len(input))
}

func TestUUIDv1Layout(t *testing.T) {
	input := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	m, ok := (&uuidDetector{}).Detect(input)
	if !ok {
		t.Fatal("expected UUIDv1 to be detected")
	}
	if m.Layout == nil {
		t.Fatal("expected Layout to be populated")
	}

	assertSegment(t, m.Layout[0], 0, 8, "time-low", "time")
	assertSegment(t, m.Layout[4], 14, 15, "version", "version")
	assertSegment(t, m.Layout[5], 15, 18, "time-high", "time")
	assertSegment(t, m.Layout[10], 24, 36, "node", "meta")
	assertLayoutCoversInput(t, m.Layout, len(input))
}

func TestNilUUIDLayout(t *testing.T) {
	input := "00000000-0000-0000-0000-000000000000"
	m, ok := (&uuidDetector{}).Detect(input)
	if !ok {
		t.Fatal("expected nil UUID to be detected")
	}
	if len(m.Layout) != 1 {
		t.Fatalf("expected 1 layout segment for nil UUID, got %d", len(m.Layout))
	}
	assertSegment(t, m.Layout[0], 0, 36, "nil", "meta")
}

func TestMaxUUIDLayout(t *testing.T) {
	input := "ffffffff-ffff-ffff-ffff-ffffffffffff"
	m, ok := (&uuidDetector{}).Detect(input)
	if !ok {
		t.Fatal("expected max UUID to be detected")
	}
	if len(m.Layout) != 1 {
		t.Fatalf("expected 1 layout segment for max UUID, got %d", len(m.Layout))
	}
	assertSegment(t, m.Layout[0], 0, 36, "max", "meta")
}

func TestULIDLayout(t *testing.T) {
	input := "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	m, ok := (&ulidDetector{}).Detect(input)
	if !ok {
		t.Fatal("expected ULID to be detected")
	}
	if m.Layout == nil {
		t.Fatal("expected Layout to be populated")
	}

	if got := len(m.Layout); got != 2 {
		t.Fatalf("expected 2 layout segments, got %d", got)
	}

	assertSegment(t, m.Layout[0], 0, 10, "timestamp", "time")
	assertSegment(t, m.Layout[1], 10, 26, "randomness", "random")
	assertLayoutCoversInput(t, m.Layout, len(input))
	assertLayoutValuesNonEmpty(t, m.Layout)
}

func TestSnowflakeLayout(t *testing.T) {
	// Use a Snowflake ID that decodes to a valid timestamp (2024+).
	// Construct one: timestamp=100000000ms from smedje epoch, worker=1, seq=42.
	// n = (100000000 << 22) | (1 << 12) | 42 = 419430400004138 (15 digits).
	input := "419430400004138"
	m, ok := (&snowflakeDetector{}).Detect(input)
	if !ok {
		t.Fatal("expected Snowflake to be detected")
	}
	if m.Layout == nil {
		t.Fatal("expected Layout to be populated")
	}

	if got := len(m.Layout); got != 2 {
		t.Fatalf("expected 2 layout segments, got %d", got)
	}

	// Layout segments should cover the full string.
	assertLayoutCoversInput(t, m.Layout, len(input))
	assertLayoutValuesNonEmpty(t, m.Layout)

	// First segment is timestamp, second is worker+seq.
	if m.Layout[0].Type != "time" {
		t.Errorf("expected first segment type 'time', got %q", m.Layout[0].Type)
	}
	if m.Layout[1].Type != "meta" {
		t.Errorf("expected second segment type 'meta', got %q", m.Layout[1].Type)
	}
}

func TestNanoIDLayout(t *testing.T) {
	input := "V1StGXR8_Z5jdHi6B-myT"
	m, ok := (&nanoidDetector{}).Detect(input)
	if !ok {
		t.Fatal("expected NanoID to be detected")
	}
	if m.Layout == nil {
		t.Fatal("expected Layout to be populated")
	}

	if got := len(m.Layout); got != 1 {
		t.Fatalf("expected 1 layout segment, got %d", got)
	}

	assertSegment(t, m.Layout[0], 0, len(input), "random", "random")
	assertLayoutCoversInput(t, m.Layout, len(input))
	assertLayoutValuesNonEmpty(t, m.Layout)
}

func TestLayoutSegmentValues(t *testing.T) {
	// Verify Value field matches the input substring.
	input := "01912f2b-95a8-7b4c-b5a1-9e3c6a7d8f0e"
	m, ok := (&uuidDetector{}).Detect(input)
	if !ok {
		t.Fatal("expected UUID to be detected")
	}

	for i, seg := range m.Layout {
		expected := input[seg.Start:seg.End]
		if seg.Value != expected {
			t.Errorf("segment %d (%s): Value=%q, expected=%q", i, seg.Label, seg.Value, expected)
		}
	}
}

// assertSegment checks that a layout segment has the expected bounds, label, and type.
func assertSegment(t *testing.T, seg LayoutSegment, start, end int, label, typ string) {
	t.Helper()
	if seg.Start != start || seg.End != end {
		t.Errorf("segment %q: range [%d,%d), expected [%d,%d)", seg.Label, seg.Start, seg.End, start, end)
	}
	if seg.Label != label {
		t.Errorf("segment at [%d,%d): label=%q, expected %q", start, end, seg.Label, label)
	}
	if seg.Type != typ {
		t.Errorf("segment %q: type=%q, expected %q", label, seg.Type, typ)
	}
}

// assertLayoutCoversInput checks that layout segments cover the full input with no gaps.
func assertLayoutCoversInput(t *testing.T, segs []LayoutSegment, inputLen int) {
	t.Helper()
	if len(segs) == 0 {
		t.Error("no layout segments")
		return
	}
	if segs[0].Start != 0 {
		t.Errorf("first segment starts at %d, expected 0", segs[0].Start)
	}
	for i := 1; i < len(segs); i++ {
		if segs[i].Start != segs[i-1].End {
			t.Errorf("gap between segments %d and %d: [%d,%d) -> [%d,%d)",
				i-1, i, segs[i-1].Start, segs[i-1].End, segs[i].Start, segs[i].End)
		}
	}
	last := segs[len(segs)-1]
	if last.End != inputLen {
		t.Errorf("last segment ends at %d, expected %d", last.End, inputLen)
	}
}

// assertLayoutValuesNonEmpty checks that every segment has non-empty Label, Type, and Value.
func assertLayoutValuesNonEmpty(t *testing.T, segs []LayoutSegment) {
	t.Helper()
	for i, seg := range segs {
		if seg.Label == "" {
			t.Errorf("segment %d has empty Label", i)
		}
		if seg.Type == "" {
			t.Errorf("segment %d (%s) has empty Type", i, seg.Label)
		}
		if seg.Value == "" {
			t.Errorf("segment %d (%s) has empty Value", i, seg.Label)
		}
	}
}
