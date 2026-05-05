package email

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"

	"github.com/smedje/smedje/internal/bench"
	"github.com/smedje/smedje/internal/entropy"
	"github.com/smedje/smedje/pkg/forge"
)

func init() { forge.Register(&DKIM{}) }

type DKIM struct{}

func (d *DKIM) Name() string             { return "dkim" }
func (d *DKIM) Group() string            { return "email" }
func (d *DKIM) Description() string      { return "Generate a DKIM keypair with DNS TXT record" }
func (d *DKIM) Category() forge.Category { return forge.CategoryCrypto }

func (d *DKIM) Generate(ctx context.Context, opts forge.Options) (*forge.Output, error) {
	selector, ok := opts.Params["selector"]
	if !ok || selector == "" {
		return nil, fmt.Errorf("dkim: --selector is required")
	}
	domain, ok := opts.Params["domain"]
	if !ok || domain == "" {
		return nil, fmt.Errorf("dkim: --domain is required")
	}
	bits := 2048
	if v, ok := opts.Params["bits"]; ok {
		fmt.Sscanf(v, "%d", &bits)
	}
	key, err := rsa.GenerateKey(entropy.Reader, bits)
	if err != nil {
		return nil, fmt.Errorf("dkim: keygen: %w", err)
	}
	privPEM := string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}))
	pubDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("dkim: marshal public key: %w", err)
	}
	pubB64 := base64.StdEncoding.EncodeToString(pubDER)
	dnsRecord := fmt.Sprintf("v=DKIM1; k=rsa; p=%s", pubB64)
	dnsName := fmt.Sprintf("%s._domainkey.%s", selector, domain)

	return &forge.Output{
		Name: "dkim",
		Artifacts: []forge.Artifact{
			{
				Label: "private-key", Filename: "dkim-private.pem",
				Fields: []forge.Field{{Key: "private-key", Value: privPEM, Sensitive: true}},
			},
			{
				Label: "dns-record", Filename: "dkim-dns.txt",
				Fields: []forge.Field{
					{Key: "record-name", Value: dnsName},
					{Key: "record-value", Value: dnsRecord},
				},
			},
		},
	}, nil
}

func (d *DKIM) Flags() []forge.FlagDef {
	return []forge.FlagDef{
		{Name: "selector", Type: "string", Description: "DKIM selector (e.g., mail) [required]"},
		{Name: "domain", Type: "string", Description: "Domain name (e.g., example.com) [required]"},
		{Name: "bits", Type: "int", Default: "2048", Description: "RSA key size"},
	}
}

func (d *DKIM) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.RunLegacy(ctx, d, 0)
}
