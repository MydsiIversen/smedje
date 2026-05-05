// Package tls provides TLS certificate generators.
package tls

import (
	"context"
	"crypto/x509"
	"fmt"
	"net"

	"github.com/smedje/smedje/internal/bench"
	"github.com/smedje/smedje/internal/config"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	forge.Register(&SelfSigned{})
}

// SelfSigned generates a self-signed TLS leaf certificate with SANs.
// The default key type is Ed25519.
type SelfSigned struct{}

func (s *SelfSigned) Name() string             { return "self-signed" }
func (s *SelfSigned) Group() string            { return "tls" }
func (s *SelfSigned) Description() string      { return "Generate a self-signed TLS certificate" }
func (s *SelfSigned) Category() forge.Category { return forge.CategoryCrypto }

func (s *SelfSigned) Generate(ctx context.Context, opts forge.Options) (*forge.Output, error) {
	cn := config.GetDefault("tls.cn")
	if v, ok := opts.Params["cn"]; ok {
		cn = v
	}
	if cn == "" {
		cn = "localhost"
	}

	days := 825
	if def := config.GetDefault("tls.days"); def != "" {
		fmt.Sscanf(def, "%d", &days)
	}
	if v, ok := opts.Params["days"]; ok {
		fmt.Sscanf(v, "%d", &days)
	}

	var dnsNames []string
	var ipAddrs []net.IP
	if sans, ok := opts.Params["san"]; ok && sans != "" {
		dnsNames, ipAddrs = parseSANs(sans)
	}
	if len(dnsNames) == 0 && len(ipAddrs) == 0 {
		dnsNames = []string{cn}
		ipAddrs = []net.IP{net.IPv4(127, 0, 0, 1)}
	}

	key, err := generateKey("ed25519")
	if err != nil {
		return nil, fmt.Errorf("tls: keygen: %w", err)
	}

	// Build the SAN list for certTemplate from already-parsed names and IPs.
	var sanStrs []string
	for _, d := range dnsNames {
		sanStrs = append(sanStrs, d)
	}
	for _, ip := range ipAddrs {
		sanStrs = append(sanStrs, ip.String())
	}

	tmpl, err := certTemplate(cn, days, sanStrs)
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

func (s *SelfSigned) Flags() []forge.FlagDef {
	return []forge.FlagDef{
		{Name: "cn", Type: "string", Default: "localhost", Description: "Certificate hostname (e.g. myapp.local)"},
		{Name: "days", Type: "int", Default: "825", Description: "Validity in days (825 = max for public trust stores)"},
		{Name: "san", Type: "string", Description: "Extra hostnames or IPs, comma-separated (e.g. api.local,10.0.0.1)"},
	}
}

func (s *SelfSigned) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.RunLegacy(ctx, s, 0)
}
