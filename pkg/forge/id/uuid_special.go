package id

import (
	"context"

	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	forge.Register(&UUIDNil{})
	forge.Register(&UUIDMax{})
}

// UUIDNil outputs the nil UUID (00000000-0000-0000-0000-000000000000).
type UUIDNil struct{}

func (u *UUIDNil) Name() string             { return "nil" }
func (u *UUIDNil) Description() string      { return "Output the nil UUID (all zeros)" }
func (u *UUIDNil) Category() forge.Category { return forge.CategoryID }

func (u *UUIDNil) Generate(ctx context.Context, opts forge.Options) (*forge.Output, error) {
	return &forge.Output{
		Name: "uuid-nil",
		Fields: []forge.Field{
			{Key: "value", Value: "00000000-0000-0000-0000-000000000000"},
		},
	}, nil
}

func (u *UUIDNil) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return &forge.BenchResult{Generator: "nil", Iterations: 1, OpsPerSec: 0}, nil
}

// UUIDMax outputs the max UUID (ffffffff-ffff-ffff-ffff-ffffffffffff).
type UUIDMax struct{}

func (u *UUIDMax) Name() string             { return "max" }
func (u *UUIDMax) Description() string      { return "Output the max UUID (all ones)" }
func (u *UUIDMax) Category() forge.Category { return forge.CategoryID }

func (u *UUIDMax) Generate(ctx context.Context, opts forge.Options) (*forge.Output, error) {
	return &forge.Output{
		Name: "uuid-max",
		Fields: []forge.Field{
			{Key: "value", Value: "ffffffff-ffff-ffff-ffff-ffffffffffff"},
		},
	}, nil
}

func (u *UUIDMax) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return &forge.BenchResult{Generator: "max", Iterations: 1, OpsPerSec: 0}, nil
}
