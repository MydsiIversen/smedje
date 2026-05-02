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
	forge.Register(&UUIDv1{})
}

// UUIDv1 generates RFC 9562 UUIDv1 identifiers with a random node ID
// (locally-administered MAC) by default, avoiding host identity leakage.
type UUIDv1 struct{}

func (u *UUIDv1) Name() string  { return "v1" }
func (u *UUIDv1) Group() string { return "uuid" }
func (u *UUIDv1) Description() string {
	return "Generate a UUIDv1 (RFC 9562, time-based with random node)"
}
func (u *UUIDv1) Category() forge.Category { return forge.CategoryID }

func (u *UUIDv1) Generate(ctx context.Context, opts forge.Options) (*forge.Output, error) {
	id, err := newUUIDv1()
	if err != nil {
		return nil, err
	}
	return &forge.Output{
		Name: "uuidv1",
		Fields: []forge.Field{
			{Key: "value", Value: formatUUID(id)},
		},
	}, nil
}

func (u *UUIDv1) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.RunLegacy(ctx, u, 0)
}

// newUUIDv1 builds a 128-bit UUIDv1 per RFC 9562 section 5.1:
//
//	time_low (32) | time_mid (16) | ver(4) + time_hi (12) | var(2) + clock_seq (14) | node (48)
//
// Uses a random node ID with the multicast bit set (locally administered).
func newUUIDv1() ([16]byte, error) {
	var uuid [16]byte

	// 60-bit timestamp: 100-nanosecond intervals since 1582-10-15.
	const uuidEpoch = 122192928000000000 // 100ns intervals from 1582 to 1970
	now := time.Now()
	t := uint64(now.UnixNano()/100) + uuidEpoch

	// time_low
	binary.BigEndian.PutUint32(uuid[0:4], uint32(t))
	// time_mid
	binary.BigEndian.PutUint16(uuid[4:6], uint16(t>>32))
	// time_hi + version 1
	binary.BigEndian.PutUint16(uuid[6:8], uint16(t>>48)&0x0FFF|0x1000)

	// Clock sequence: 14 random bits.
	var clk [2]byte
	if _, err := entropy.Read(clk[:]); err != nil {
		return uuid, fmt.Errorf("uuidv1: clock seq: %w", err)
	}
	uuid[8] = (clk[0] & 0x3F) | 0x80 // variant bits
	uuid[9] = clk[1]

	// Node: 6 random bytes with multicast bit set.
	var node [6]byte
	if _, err := entropy.Read(node[:]); err != nil {
		return uuid, fmt.Errorf("uuidv1: node: %w", err)
	}
	node[0] |= 0x01 // multicast bit = locally administered
	copy(uuid[10:16], node[:])

	return uuid, nil
}
