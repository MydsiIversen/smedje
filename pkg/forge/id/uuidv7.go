// Package id provides identifier generators (UUIDv7, Snowflake).
package id

import (
	"context"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/smedje/smedje/internal/bench"
	"github.com/smedje/smedje/internal/entropy"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	forge.Register(&UUIDv7{})
}

// UUIDv7 generates RFC 9562 UUIDv7 identifiers with millisecond timestamp
// precision and 74 bits of crypto/rand randomness.
type UUIDv7 struct{}

func (u *UUIDv7) Name() string             { return "v7" }
func (u *UUIDv7) Description() string      { return "Generate a UUIDv7 (RFC 9562, time-ordered)" }
func (u *UUIDv7) Category() forge.Category { return forge.CategoryID }

func (u *UUIDv7) Generate(ctx context.Context, opts forge.Options) (*forge.Output, error) {
	id, err := newUUIDv7()
	if err != nil {
		return nil, err
	}
	return &forge.Output{
		Name: "uuidv7",
		Fields: []forge.Field{
			{Key: "value", Value: formatUUID(id)},
		},
	}, nil
}

func (u *UUIDv7) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.RunLegacy(ctx, u, 0)
}

// newUUIDv7 builds a 128-bit UUIDv7 per RFC 9562 section 5.7:
//
//	0                   1                   2                   3
//	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	|                         unix_ts_ms (48 bits)                  |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	|          unix_ts_ms           | ver(4) |  rand_a (12 bits)    |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	| var(2)|              rand_b (62 bits)                         |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	|                         rand_b                                |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
func newUUIDv7() ([16]byte, error) {
	var uuid [16]byte

	// 48-bit millisecond Unix timestamp.
	ms := uint64(time.Now().UnixMilli())
	uuid[0] = byte(ms >> 40)
	uuid[1] = byte(ms >> 32)
	uuid[2] = byte(ms >> 24)
	uuid[3] = byte(ms >> 16)
	uuid[4] = byte(ms >> 8)
	uuid[5] = byte(ms)

	// 74 bits of randomness from crypto/rand.
	var rnd [10]byte
	if _, err := entropy.Read(rnd[:]); err != nil {
		return uuid, fmt.Errorf("uuidv7: entropy read: %w", err)
	}

	// rand_a: 12 bits in uuid[6..7], with version nibble 0b0111.
	uuid[6] = 0x70 | (rnd[0] & 0x0F)
	uuid[7] = rnd[1]

	// rand_b: 62 bits in uuid[8..15], with variant bits 0b10.
	copy(uuid[8:], rnd[2:])
	uuid[8] = (uuid[8] & 0x3F) | 0x80

	return uuid, nil
}

func formatUUID(u [16]byte) string {
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		binary.BigEndian.Uint32(u[0:4]),
		binary.BigEndian.Uint16(u[4:6]),
		binary.BigEndian.Uint16(u[6:8]),
		binary.BigEndian.Uint16(u[8:10]),
		u[10:16],
	)
}
