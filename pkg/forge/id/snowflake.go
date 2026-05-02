package id

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/smedje/smedje/internal/bench"
	"github.com/smedje/smedje/internal/config"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	forge.Register(&Snowflake{})
}

// Snowflake generates Twitter-style 64-bit Snowflake IDs.
//
// Layout (MSB to LSB):
//
//	1 bit unused | 41 bits timestamp (ms since epoch) | 10 bits worker | 12 bits sequence
//
// The epoch is 2024-01-01T00:00:00Z. Worker ID defaults to 0 and can be
// set via the "worker" option (0-1023).
type Snowflake struct {
	mu       sync.Mutex
	lastMS   int64
	sequence int64
}

const snowflakeEpoch = 1704067200000 // 2024-01-01T00:00:00Z in ms

func (s *Snowflake) Name() string             { return "snowflake" }
func (s *Snowflake) Description() string      { return "Generate a Twitter-style Snowflake ID (64-bit)" }
func (s *Snowflake) Category() forge.Category { return forge.CategoryID }

func (s *Snowflake) Generate(ctx context.Context, opts forge.Options) (*forge.Output, error) {
	workerID := int64(0)
	if def := config.GetDefault("snowflake.worker"); def != "" {
		if n, err := strconv.ParseInt(def, 10, 64); err == nil && n >= 0 && n <= 1023 {
			workerID = n
		}
	}
	if w, ok := opts.Params["worker"]; ok {
		n, err := strconv.ParseInt(w, 10, 64)
		if err != nil || n < 0 || n > 1023 {
			return nil, fmt.Errorf("snowflake: worker must be 0-1023, got %q", w)
		}
		workerID = n
	}

	s.mu.Lock()
	now := time.Now().UnixMilli() - snowflakeEpoch
	if now == s.lastMS {
		s.sequence++
		if s.sequence > 4095 {
			// Spin until next millisecond.
			for now <= s.lastMS {
				now = time.Now().UnixMilli() - snowflakeEpoch
			}
			s.sequence = 0
		}
	} else {
		s.sequence = 0
	}
	s.lastMS = now
	seq := s.sequence
	s.mu.Unlock()

	id := (now << 22) | (workerID << 12) | seq

	return &forge.Output{
		Name: "snowflake",
		Fields: []forge.Field{
			{Key: "value", Value: strconv.FormatInt(id, 10)},
		},
	}, nil
}

func (s *Snowflake) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.Run(ctx, s, 0)
}
