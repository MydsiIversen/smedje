package id

import (
	"context"
	"fmt"

	"github.com/smedje/smedje/internal/bench"
	"github.com/smedje/smedje/internal/entropy"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	forge.Register(&UUIDv4{})
}

// UUIDv4 generates RFC 9562 UUIDv4 identifiers with 122 bits of randomness.
type UUIDv4 struct{}

func (u *UUIDv4) Name() string             { return "v4" }
func (u *UUIDv4) Description() string      { return "Generate a UUIDv4 (RFC 9562, random)" }
func (u *UUIDv4) Category() forge.Category { return forge.CategoryID }

func (u *UUIDv4) Generate(ctx context.Context, opts forge.Options) (*forge.Output, error) {
	var uuid [16]byte
	if _, err := entropy.Read(uuid[:]); err != nil {
		return nil, fmt.Errorf("uuidv4: entropy read: %w", err)
	}

	// Version 4
	uuid[6] = (uuid[6] & 0x0F) | 0x40
	// Variant 10
	uuid[8] = (uuid[8] & 0x3F) | 0x80

	return &forge.Output{
		Name: "uuidv4",
		Fields: []forge.Field{
			{Key: "value", Value: formatUUID(uuid)},
		},
	}, nil
}

func (u *UUIDv4) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.RunLegacy(ctx, u, 0)
}
