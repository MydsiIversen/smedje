package id

import (
	"context"
	"fmt"
	"time"

	"github.com/smedje/smedje/internal/bench"
	"github.com/smedje/smedje/internal/entropy"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	forge.Register(&ULID{})
}

// ULID generates Universally Unique Lexicographically Sortable Identifiers.
// Format: 10-char timestamp (ms) + 16-char randomness, Crockford Base32.
type ULID struct{}

func (u *ULID) Name() string             { return "ulid" }
func (u *ULID) Group() string            { return "ulid" }
func (u *ULID) Description() string      { return "Generate a ULID (Crockford Base32, time-sortable)" }
func (u *ULID) Category() forge.Category { return forge.CategoryID }

func (u *ULID) Generate(ctx context.Context, opts forge.Options) (*forge.Output, error) {
	now := optTime(opts)
	id, err := newULID(now)
	if err != nil {
		return nil, err
	}
	return &forge.Output{
		Name: "ulid",
		Fields: []forge.Field{
			{Key: "value", Value: id},
		},
	}, nil
}

func (u *ULID) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.RunLegacy(ctx, u, 0)
}

// Crockford Base32 alphabet (excludes I, L, O, U to avoid ambiguity).
const crockford = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

func newULID(now time.Time) (string, error) {
	var buf [26]byte

	// Timestamp: 48-bit millisecond Unix time, encoded as 10 base32 chars.
	ms := uint64(now.UnixMilli())
	for i := 9; i >= 0; i-- {
		buf[i] = crockford[ms&0x1F]
		ms >>= 5
	}

	// Randomness: 80 bits (10 bytes), encoded as 16 base32 chars.
	var rnd [10]byte
	if _, err := entropy.Read(rnd[:]); err != nil {
		return "", fmt.Errorf("ulid: entropy read: %w", err)
	}

	// Encode 10 bytes (80 bits) into 16 base32 characters.
	// Process 5 bits at a time from the byte array.
	bits := uint64(0)
	nbits := 0
	pos := 10
	for i := 0; i < len(rnd); i++ {
		bits = (bits << 8) | uint64(rnd[i])
		nbits += 8
		for nbits >= 5 {
			nbits -= 5
			buf[pos] = crockford[(bits>>nbits)&0x1F]
			pos++
		}
	}

	return string(buf[:]), nil
}
