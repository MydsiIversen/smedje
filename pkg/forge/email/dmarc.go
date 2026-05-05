package email

import (
	"context"
	"fmt"
	"strings"

	"github.com/smedje/smedje/internal/bench"
	"github.com/smedje/smedje/pkg/forge"
)

func init() { forge.Register(&DMARC{}) }

type DMARC struct{}

func (d *DMARC) Name() string             { return "dmarc" }
func (d *DMARC) Group() string            { return "email" }
func (d *DMARC) Description() string      { return "Generate a DMARC DNS TXT record" }
func (d *DMARC) Category() forge.Category { return forge.CategoryNetwork }

func (d *DMARC) Generate(ctx context.Context, opts forge.Options) (*forge.Output, error) {
	domain, ok := opts.Params["domain"]
	if !ok || domain == "" {
		return nil, fmt.Errorf("dmarc: --domain is required")
	}
	policy := "none"
	if v, ok := opts.Params["policy"]; ok {
		policy = v
	}
	var parts []string
	parts = append(parts, "v=DMARC1")
	parts = append(parts, fmt.Sprintf("p=%s", policy))
	if rua, ok := opts.Params["rua"]; ok && rua != "" {
		parts = append(parts, fmt.Sprintf("rua=mailto:%s", rua))
	}
	if ruf, ok := opts.Params["ruf"]; ok && ruf != "" {
		parts = append(parts, fmt.Sprintf("ruf=mailto:%s", ruf))
	}
	if pct, ok := opts.Params["pct"]; ok && pct != "" && pct != "100" {
		parts = append(parts, fmt.Sprintf("pct=%s", pct))
	}
	record := strings.Join(parts, "; ")
	dnsName := fmt.Sprintf("_dmarc.%s", domain)

	return forge.SingleArtifact("dmarc",
		forge.Field{Key: "record-name", Value: dnsName},
		forge.Field{Key: "record-value", Value: record},
	), nil
}

func (d *DMARC) Flags() []forge.FlagDef {
	return []forge.FlagDef{
		{Name: "domain", Type: "string", Description: "Email domain (e.g. example.com) [required]"},
		{Name: "policy", Type: "string", Default: "none", Description: "Action on auth failure: none (monitor), quarantine (spam folder), reject (block)", Options: []string{"none", "quarantine", "reject"}},
		{Name: "rua", Type: "string", Description: "Email for daily aggregate reports (e.g. dmarc@example.com). mailto: added automatically"},
		{Name: "ruf", Type: "string", Description: "Email for per-failure forensic reports. Note: Gmail/Outlook don't send these"},
		{Name: "pct", Type: "int", Default: "100", Description: "Percentage of messages to apply policy to (1-100). Useful for gradual rollout"},
	}
}

func (d *DMARC) BenchOptions() forge.Options {
	return forge.Options{Params: map[string]string{"domain": "example.com"}}
}

func (d *DMARC) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.RunLegacy(ctx, d, 0)
}
