package tls

import (
	"context"
	"crypto/x509"
	"fmt"

	"github.com/smedje/smedje/internal/bench"
	"github.com/smedje/smedje/internal/config"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	forge.Register(&RSACert{})
}

// RSACert generates a self-signed RSA TLS certificate.
type RSACert struct{}

func (r *RSACert) Name() string             { return "rsa" }
func (r *RSACert) Group() string            { return "tls" }
func (r *RSACert) Description() string      { return "Generate an RSA self-signed TLS certificate" }
func (r *RSACert) Category() forge.Category { return forge.CategoryCrypto }

func (r *RSACert) Generate(ctx context.Context, opts forge.Options) (*forge.Output, error) {
	cn := "localhost"
	if def := config.GetDefault("tls.cn"); def != "" {
		cn = def
	}
	if v, ok := opts.Params["cn"]; ok {
		cn = v
	}

	days := 825
	if def := config.GetDefault("tls.days"); def != "" {
		fmt.Sscanf(def, "%d", &days)
	}
	if v, ok := opts.Params["days"]; ok {
		fmt.Sscanf(v, "%d", &days)
	}

	bits := "2048"
	if v, ok := opts.Params["bits"]; ok {
		bits = v
	}
	algo := "rsa-" + bits

	var sanList []string
	if v, ok := opts.Params["san"]; ok && v != "" {
		dns, ips := parseSANs(v)
		for _, d := range dns {
			sanList = append(sanList, d)
		}
		for _, ip := range ips {
			sanList = append(sanList, ip.String())
		}
	}
	if len(sanList) == 0 {
		sanList = []string{cn, "127.0.0.1"}
	}

	key, err := generateKey(algo)
	if err != nil {
		return nil, fmt.Errorf("tls: keygen: %w", err)
	}

	tmpl, err := certTemplate(cn, days, sanList)
	if err != nil {
		return nil, fmt.Errorf("tls: template: %w", err)
	}
	tmpl.KeyUsage = x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment
	tmpl.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}

	der, err := signCert(tmpl, nil, key.Public(), key)
	if err != nil {
		return nil, fmt.Errorf("tls: sign: %w", err)
	}

	keyPEM, err := encodeKeyPEM(key)
	if err != nil {
		return nil, err
	}

	return forge.SingleArtifact("tls",
		forge.Field{Key: "certificate", Value: encodeCertPEM(der)},
		forge.Field{Key: "private-key", Value: keyPEM, Sensitive: true},
	), nil
}

// Flags implements forge.FlagDescriber.
func (r *RSACert) Flags() []forge.FlagDef {
	return []forge.FlagDef{
		{Name: "cn", Type: "string", Default: "localhost", Description: "Certificate hostname (e.g. myapp.local)"},
		{Name: "days", Type: "int", Default: "825", Description: "Validity in days (825 = max for public trust stores)"},
		{Name: "san", Type: "string", Description: "Extra hostnames or IPs, comma-separated"},
		{Name: "bits", Type: "int", Default: "2048", Description: "RSA key size (2048 standard, 4096 higher security)", Options: []string{"2048", "4096"}},
	}
}

func (r *RSACert) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.RunLegacy(ctx, r, 0)
}
