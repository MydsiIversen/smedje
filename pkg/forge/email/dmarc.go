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
	record := strings.Join(parts, "; ")
	dnsName := fmt.Sprintf("_dmarc.%s", domain)

	return forge.SingleArtifact("dmarc",
		forge.Field{Key: "record-name", Value: dnsName},
		forge.Field{Key: "record-value", Value: record},
	), nil
}

func (d *DMARC) Flags() []forge.FlagDef {
	return []forge.FlagDef{
		{Name: "domain", Type: "string", Description: "Domain name [required]"},
		{Name: "policy", Type: "string", Default: "none", Description: "DMARC policy", Options: []string{"none", "quarantine", "reject"}},
		{Name: "rua", Type: "string", Description: "Aggregate report email address"},
		{Name: "ruf", Type: "string", Description: "Forensic report email address"},
	}
}

func (d *DMARC) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.RunLegacy(ctx, d, 0)
}
