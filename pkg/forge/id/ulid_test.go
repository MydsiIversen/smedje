package id

import (
	"context"
	"testing"

	"github.com/smedje/smedje/pkg/forge"
)

func TestULIDFormat(t *testing.T) {
	g := &ULID{}
	out, err := g.Generate(context.Background(), forge.Options{})
	if err != nil {
		t.Fatal(err)
	}
	val := out.PrimaryFields()[0].Value
	if len(val) != 26 {
		t.Errorf("expected 26 chars, got %d: %s", len(val), val)
	}
	for _, c := range val {
		if !isCrockford(byte(c)) {
			t.Errorf("invalid Crockford char: %c in %s", c, val)
		}
	}
}

func TestULIDTimestampSortable(t *testing.T) {
	g := &ULID{}
	// ULIDs generated across different milliseconds must have ascending
	// timestamp prefixes (first 10 chars).
	prev := ""
	for range 100 {
		out, err := g.Generate(context.Background(), forge.Options{})
		if err != nil {
			t.Fatal(err)
		}
		val := out.PrimaryFields()[0].Value
		ts := val[:10]
		if prev != "" && ts < prev {
			t.Errorf("timestamp prefix not sorted: %s < %s", ts, prev)
		}
		prev = ts
	}
}

func isCrockford(b byte) bool {
	for _, c := range []byte(crockford) {
		if b == c {
			return true
		}
	}
	return false
}
