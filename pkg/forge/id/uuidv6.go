package id

import (
	"context"
	"encoding/binary"
	"fmt"

	"github.com/smedje/smedje/internal/bench"
	"github.com/smedje/smedje/internal/entropy"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	forge.Register(&UUIDv6{})
}

// UUIDv6 generates RFC 9562 UUIDv6 identifiers. UUIDv6 is a reordering of
// UUIDv1 fields so that the timestamp sorts naturally in byte order.
type UUIDv6 struct{}

func (u *UUIDv6) Name() string             { return "v6" }
func (u *UUIDv6) Group() string            { return "uuid" }
func (u *UUIDv6) Description() string      { return "Generate a UUIDv6 (RFC 9562, reordered time)" }
func (u *UUIDv6) Category() forge.Category { return forge.CategoryID }

func (u *UUIDv6) Generate(ctx context.Context, opts forge.Options) (*forge.Output, error) {
	var uuid [16]byte

	// 60-bit Gregorian timestamp (same as v1).
	const uuidEpoch = 122192928000000000
	now := optTime(opts)
	t := uint64(now.UnixNano()/100) + uuidEpoch

	// UUIDv6 layout: time_high (32) | time_mid (16) | ver(4) + time_low (12) | var(2) + clock_seq (14) | node (48)
	// Top 32 bits of timestamp
	binary.BigEndian.PutUint32(uuid[0:4], uint32(t>>28))
	// Next 16 bits
	binary.BigEndian.PutUint16(uuid[4:6], uint16(t>>12))
	// Low 12 bits + version 6
	binary.BigEndian.PutUint16(uuid[6:8], uint16(t&0x0FFF)|0x6000)

	// Clock sequence: 14 random bits.
	var clk [2]byte
	if _, err := entropy.Read(clk[:]); err != nil {
		return nil, fmt.Errorf("uuidv6: clock seq: %w", err)
	}
	uuid[8] = (clk[0] & 0x3F) | 0x80
	uuid[9] = clk[1]

	// Node: 6 random bytes with multicast bit set.
	var node [6]byte
	if _, err := entropy.Read(node[:]); err != nil {
		return nil, fmt.Errorf("uuidv6: node: %w", err)
	}
	node[0] |= 0x01
	copy(uuid[10:16], node[:])

	return &forge.Output{
		Name: "uuid",
		Fields: []forge.Field{
			{Key: "value", Value: formatUUID(uuid)},
		},
	}, nil
}

func (u *UUIDv6) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.RunLegacy(ctx, u, 0)
}
