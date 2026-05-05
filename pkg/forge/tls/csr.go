package tls

import (
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"

	"github.com/smedje/smedje/internal/bench"
	"github.com/smedje/smedje/internal/config"
	"github.com/smedje/smedje/internal/entropy"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	forge.Register(&CSR{})
}

// CSR generates a TLS certificate signing request with an accompanying private key.
type CSR struct{}

func (c *CSR) Name() string             { return "csr" }
func (c *CSR) Group() string            { return "tls" }
func (c *CSR) Description() string      { return "Generate a TLS certificate signing request" }
func (c *CSR) Category() forge.Category { return forge.CategoryCrypto }

func (c *CSR) Generate(ctx context.Context, opts forge.Options) (*forge.Output, error) {
	cn := "localhost"
	if def := config.GetDefault("tls.cn"); def != "" {
		cn = def
	}
	if v, ok := opts.Params["cn"]; ok {
		cn = v
	}

	algo := "ed25519"
	if v, ok := opts.Params["algo"]; ok {
		switch v {
		case "ed25519":
			algo = "ed25519"
		case "rsa":
			algo = "rsa-2048"
		case "ecdsa":
			algo = "ecdsa-p256"
		default:
			return nil, fmt.Errorf("tls: unsupported CSR algo %q", v)
		}
	}

	key, err := generateKey(algo)
	if err != nil {
		return nil, fmt.Errorf("tls: keygen: %w", err)
	}

	tmpl := &x509.CertificateRequest{
		Subject: pkix.Name{CommonName: cn},
	}

	if org, ok := opts.Params["org"]; ok && org != "" {
		tmpl.Subject.Organization = []string{org}
	}

	if v, ok := opts.Params["san"]; ok && v != "" {
		dns, ips := parseSANs(v)
		tmpl.DNSNames = dns
		tmpl.IPAddresses = ips
	}

	csrDER, err := x509.CreateCertificateRequest(entropy.Reader, tmpl, key)
	if err != nil {
		return nil, fmt.Errorf("tls: create csr: %w", err)
	}

	csrPEM := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrDER}))

	keyPEM, err := encodeKeyPEM(key)
	if err != nil {
		return nil, err
	}

	return forge.SingleArtifact("csr",
		forge.Field{Key: "csr", Value: csrPEM},
		forge.Field{Key: "private-key", Value: keyPEM, Sensitive: true},
	), nil
}

// Flags implements forge.FlagDescriber.
func (c *CSR) Flags() []forge.FlagDef {
	return []forge.FlagDef{
		{Name: "cn", Type: "string", Default: "localhost", Description: "Domain to secure (e.g. example.com)"},
		{Name: "san", Type: "string", Description: "Alternative names, comma-separated (e.g. www.example.com,example.com)"},
		{Name: "algo", Type: "string", Default: "ed25519", Description: "Key algorithm", Options: []string{"ed25519", "rsa", "ecdsa"}},
		{Name: "org", Type: "string", Description: "Organization name for the CSR (e.g. My Company Inc.)"},
	}
}

func (c *CSR) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.RunLegacy(ctx, c, 0)
}
