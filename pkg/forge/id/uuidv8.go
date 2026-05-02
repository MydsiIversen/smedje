package id

import (
	"context"
	"fmt"

	"github.com/smedje/smedje/internal/bench"
	"github.com/smedje/smedje/internal/entropy"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	forge.Register(&UUIDv8{})
}

// UUIDv8 generates RFC 9562 UUIDv8 identifiers with custom payload. When no
// payload is specified, all custom bits are filled from crypto/rand.
type UUIDv8 struct{}

func (u *UUIDv8) Name() string             { return "v8" }
func (u *UUIDv8) Group() string            { return "uuid" }
func (u *UUIDv8) Description() string      { return "Generate a UUIDv8 (RFC 9562, custom)" }
func (u *UUIDv8) Category() forge.Category { return forge.CategoryID }

func (u *UUIDv8) Generate(ctx context.Context, opts forge.Options) (*forge.Output, error) {
	var uuid [16]byte

	// Fill with random data.
	if _, err := entropy.Read(uuid[:]); err != nil {
		return nil, fmt.Errorf("uuidv8: entropy read: %w", err)
	}

	// Version 8
	uuid[6] = (uuid[6] & 0x0F) | 0x80
	// Variant 10
	uuid[8] = (uuid[8] & 0x3F) | 0x80

	return &forge.Output{
		Name: "uuid",
		Fields: []forge.Field{
			{Key: "value", Value: formatUUID(uuid)},
		},
	}, nil
}

func (u *UUIDv8) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.RunLegacy(ctx, u, 0)
}
