package tls

import (
	"context"
	"crypto"
	"crypto/x509"
	"fmt"

	"github.com/smedje/smedje/internal/bench"
	"github.com/smedje/smedje/internal/config"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	forge.Register(&CAChain{})
}

// CAChain generates a TLS CA chain of configurable depth. Depth 3 produces
// root CA → intermediate CA → leaf cert; depth 2 skips the intermediate.
type CAChain struct{}

func (c *CAChain) Name() string  { return "ca-chain" }
func (c *CAChain) Group() string { return "tls" }
func (c *CAChain) Description() string {
	return "Generate a TLS CA chain (root → intermediate → leaf)"
}
func (c *CAChain) Category() forge.Category { return forge.CategoryCrypto }

func (c *CAChain) Generate(ctx context.Context, opts forge.Options) (*forge.Output, error) {
	cn := "My CA"
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

	depth := 3
	if v, ok := opts.Params["depth"]; ok {
		fmt.Sscanf(v, "%d", &depth)
	}
	if depth < 2 || depth > 3 {
		return nil, fmt.Errorf("tls: depth must be 2 or 3, got %d", depth)
	}

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
		sanList = []string{"localhost", "127.0.0.1"}
	}

	rootKey, err := generateKey("ed25519")
	if err != nil {
		return nil, fmt.Errorf("tls: root keygen: %w", err)
	}
	rootTmpl, err := certTemplate(cn+" Root CA", days, nil)
	if err != nil {
		return nil, fmt.Errorf("tls: root template: %w", err)
	}
	rootTmpl.IsCA = true
	rootTmpl.BasicConstraintsValid = true
	rootTmpl.KeyUsage = x509.KeyUsageCertSign | x509.KeyUsageCRLSign
	rootTmpl.MaxPathLen = 1
	rootTmpl.MaxPathLenZero = false

	rootDER, err := signCert(rootTmpl, nil, rootKey.Public(), rootKey)
	if err != nil {
		return nil, fmt.Errorf("tls: root sign: %w", err)
	}
	rootCert, err := x509.ParseCertificate(rootDER)
	if err != nil {
		return nil, fmt.Errorf("tls: root parse: %w", err)
	}

	rootKeyPEM, err := encodeKeyPEM(rootKey)
	if err != nil {
		return nil, err
	}

	artifacts := []forge.Artifact{
		{
			Label:    "root-ca",
			Filename: "root-ca.pem",
			Fields: []forge.Field{
				{Key: "certificate", Value: encodeCertPEM(rootDER)},
				{Key: "private-key", Value: rootKeyPEM, Sensitive: true},
			},
		},
	}

	var leafParentCert *x509.Certificate
	var leafParentKey crypto.Signer

	if depth == 3 {
		intKey, err := generateKey("ed25519")
		if err != nil {
			return nil, fmt.Errorf("tls: intermediate keygen: %w", err)
		}
		intTmpl, err := certTemplate(cn+" Intermediate CA", days, nil)
		if err != nil {
			return nil, fmt.Errorf("tls: intermediate template: %w", err)
		}
		intTmpl.IsCA = true
		intTmpl.BasicConstraintsValid = true
		intTmpl.KeyUsage = x509.KeyUsageCertSign | x509.KeyUsageCRLSign
		intTmpl.MaxPathLen = 0
		intTmpl.MaxPathLenZero = true

		intDER, err := signCert(intTmpl, rootCert, intKey.Public(), rootKey)
		if err != nil {
			return nil, fmt.Errorf("tls: intermediate sign: %w", err)
		}
		intCert, err := x509.ParseCertificate(intDER)
		if err != nil {
			return nil, fmt.Errorf("tls: intermediate parse: %w", err)
		}

		intKeyPEM, err := encodeKeyPEM(intKey)
		if err != nil {
			return nil, err
		}

		artifacts = append(artifacts, forge.Artifact{
			Label:    "intermediate-ca",
			Filename: "intermediate-ca.pem",
			Fields: []forge.Field{
				{Key: "certificate", Value: encodeCertPEM(intDER)},
				{Key: "private-key", Value: intKeyPEM, Sensitive: true},
			},
		})

		leafParentCert = intCert
		leafParentKey = intKey
	} else {
		leafParentCert = rootCert
		leafParentKey = rootKey
	}

	leafKey, err := generateKey("ed25519")
	if err != nil {
		return nil, fmt.Errorf("tls: leaf keygen: %w", err)
	}
	leafTmpl, err := certTemplate(cn, days, sanList)
	if err != nil {
		return nil, fmt.Errorf("tls: leaf template: %w", err)
	}
	leafTmpl.KeyUsage = x509.KeyUsageDigitalSignature
	leafTmpl.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}

	leafDER, err := signCert(leafTmpl, leafParentCert, leafKey.Public(), leafParentKey)
	if err != nil {
		return nil, fmt.Errorf("tls: leaf sign: %w", err)
	}

	leafKeyPEM, err := encodeKeyPEM(leafKey)
	if err != nil {
		return nil, err
	}

	artifacts = append(artifacts, forge.Artifact{
		Label:    "leaf",
		Filename: "leaf.pem",
		Fields: []forge.Field{
			{Key: "certificate", Value: encodeCertPEM(leafDER)},
			{Key: "private-key", Value: leafKeyPEM, Sensitive: true},
		},
	})

	return &forge.Output{Name: "ca-chain", Artifacts: artifacts}, nil
}

// Flags returns the generator-specific flags for CLI wiring.
func (c *CAChain) Flags() []forge.FlagDef {
	return []forge.FlagDef{
		{Name: "cn", Type: "string", Default: "My CA", Description: "Base name for CA chain (produces 'My CA Root CA', 'My CA Intermediate CA', etc.)"},
		{Name: "days", Type: "int", Default: "825", Description: "Validity in days for all certs in the chain"},
		{Name: "depth", Type: "int", Default: "3", Description: "Chain depth", Options: []string{"2", "3"}},
		{Name: "san", Type: "string", Description: "Leaf certificate SANs, comma-separated. Defaults to localhost,127.0.0.1"},
	}
}

// Bench runs a self-benchmark for the ca-chain generator.
func (c *CAChain) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.RunLegacy(ctx, c, 0)
}
