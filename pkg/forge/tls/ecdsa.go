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
	forge.Register(&ECDSACert{})
}

// ECDSACert generates a self-signed ECDSA TLS certificate.
type ECDSACert struct{}

func (e *ECDSACert) Name() string             { return "ecdsa" }
func (e *ECDSACert) Group() string            { return "tls" }
func (e *ECDSACert) Description() string      { return "Generate an ECDSA self-signed TLS certificate" }
func (e *ECDSACert) Category() forge.Category { return forge.CategoryCrypto }

func (e *ECDSACert) Generate(ctx context.Context, opts forge.Options) (*forge.Output, error) {
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

	curve := "p256"
	if v, ok := opts.Params["curve"]; ok {
		curve = v
	}
	algo := "ecdsa-" + curve

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
	tmpl.KeyUsage = x509.KeyUsageDigitalSignature
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
func (e *ECDSACert) Flags() []forge.FlagDef {
	return []forge.FlagDef{
		{Name: "cn", Type: "string", Default: "localhost", Description: "Common name"},
		{Name: "days", Type: "int", Default: "825", Description: "Validity in days"},
		{Name: "san", Type: "string", Description: "Subject alternative names (comma-separated)"},
		{Name: "curve", Type: "string", Default: "p256", Description: "ECDSA curve", Options: []string{"p256", "p384"}},
	}
}

func (e *ECDSACert) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.RunLegacy(ctx, e, 0)
}
