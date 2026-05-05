package network

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/smedje/smedje/internal/bench"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	forge.Register(&IQN{})
}

// IQN generates an iSCSI Qualified Name as defined in RFC 3720.
// Format: iqn.YYYY-MM.<reversed-authority>:<target>
// Example: iqn.2024-01.com.example:storage.lun0
type IQN struct{}

func (i *IQN) Name() string             { return "iqn" }
func (i *IQN) Group() string            { return "network" }
func (i *IQN) Description() string      { return "Generate an iSCSI Qualified Name (IQN)" }
func (i *IQN) Category() forge.Category { return forge.CategoryNetwork }

// Generate returns a formatted IQN string.
// Required params: "authority" (e.g. com.example), "target" (e.g. storage.lun0).
// Optional: "date" in YYYY-MM format; defaults to the current month.
func (i *IQN) Generate(ctx context.Context, opts forge.Options) (*forge.Output, error) {
	authority := opts.Params["authority"]
	if authority == "" {
		return nil, fmt.Errorf("iqn: --authority is required (e.g. com.example)")
	}
	target := opts.Params["target"]
	if target == "" {
		return nil, fmt.Errorf("iqn: --target is required (e.g. storage.lun0)")
	}

	var date string
	if v, ok := opts.Params["date"]; ok && v != "" {
		date = v
	} else {
		now := time.Now()
		if opts.Time != nil {
			now = opts.Time()
		}
		date = now.Format("2006-01")
	}

	reversed := reverseAuthority(authority)
	iqn := fmt.Sprintf("iqn.%s.%s:%s", date, reversed, target)

	return forge.SingleArtifact("iqn",
		forge.Field{Key: "value", Value: iqn},
	), nil
}

func (i *IQN) Flags() []forge.FlagDef {
	return []forge.FlagDef{
		{Name: "authority", Type: "string", Default: "", Description: "DNS authority in forward order (e.g. com.example)"},
		{Name: "target", Type: "string", Default: "", Description: "Target name (e.g. storage.lun0)"},
		{Name: "date", Type: "string", Default: "", Description: "Date in YYYY-MM format; defaults to current month"},
	}
}

func (i *IQN) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.RunLegacy(ctx, i, 0)
}

// reverseAuthority reverses a dot-separated authority string.
// "com.example.storage" becomes "storage.example.com".
func reverseAuthority(authority string) string {
	parts := strings.Split(authority, ".")
	for l, r := 0, len(parts)-1; l < r; l, r = l+1, r-1 {
		parts[l], parts[r] = parts[r], parts[l]
	}
	return strings.Join(parts, ".")
}
