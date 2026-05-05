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
	forge.Register(&MTLS{})
}

// MTLS generates a mutual TLS bundle: a self-signed CA, a server cert signed
// by that CA, and a client cert signed by that CA.
type MTLS struct{}

func (m *MTLS) Name() string             { return "mtls" }
func (m *MTLS) Group() string            { return "tls" }
func (m *MTLS) Description() string      { return "Generate a mutual TLS bundle (CA + server + client)" }
func (m *MTLS) Category() forge.Category { return forge.CategoryCrypto }

func (m *MTLS) Generate(ctx context.Context, opts forge.Options) (*forge.Output, error) {
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

	// CA
	caKey, err := generateKey("ed25519")
	if err != nil {
		return nil, fmt.Errorf("tls: ca keygen: %w", err)
	}
	caTmpl, err := certTemplate(cn+" CA", days, nil)
	if err != nil {
		return nil, fmt.Errorf("tls: ca template: %w", err)
	}
	caTmpl.IsCA = true
	caTmpl.BasicConstraintsValid = true
	caTmpl.KeyUsage = x509.KeyUsageCertSign | x509.KeyUsageCRLSign

	caDER, err := signCert(caTmpl, nil, caKey.Public(), caKey)
	if err != nil {
		return nil, fmt.Errorf("tls: ca sign: %w", err)
	}
	caCert, err := x509.ParseCertificate(caDER)
	if err != nil {
		return nil, fmt.Errorf("tls: ca parse: %w", err)
	}
	caKeyPEM, err := encodeKeyPEM(caKey)
	if err != nil {
		return nil, err
	}

	// Server cert
	serverKey, err := generateKey("ed25519")
	if err != nil {
		return nil, fmt.Errorf("tls: server keygen: %w", err)
	}
	serverTmpl, err := certTemplate(cn, days, sanList)
	if err != nil {
		return nil, fmt.Errorf("tls: server template: %w", err)
	}
	serverTmpl.KeyUsage = x509.KeyUsageDigitalSignature
	serverTmpl.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}

	serverDER, err := signCert(serverTmpl, caCert, serverKey.Public(), caKey)
	if err != nil {
		return nil, fmt.Errorf("tls: server sign: %w", err)
	}
	serverKeyPEM, err := encodeKeyPEM(serverKey)
	if err != nil {
		return nil, err
	}

	// Client cert
	clientKey, err := generateKey("ed25519")
	if err != nil {
		return nil, fmt.Errorf("tls: client keygen: %w", err)
	}
	clientTmpl, err := certTemplate("client", days, nil)
	if err != nil {
		return nil, fmt.Errorf("tls: client template: %w", err)
	}
	clientTmpl.KeyUsage = x509.KeyUsageDigitalSignature
	clientTmpl.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}

	clientDER, err := signCert(clientTmpl, caCert, clientKey.Public(), caKey)
	if err != nil {
		return nil, fmt.Errorf("tls: client sign: %w", err)
	}
	clientKeyPEM, err := encodeKeyPEM(clientKey)
	if err != nil {
		return nil, err
	}

	return &forge.Output{
		Name: "mtls",
		Artifacts: []forge.Artifact{
			{
				Label: "ca", Filename: "ca.pem",
				Fields: []forge.Field{
					{Key: "certificate", Value: encodeCertPEM(caDER)},
					{Key: "private-key", Value: caKeyPEM, Sensitive: true},
				},
			},
			{
				Label: "server", Filename: "server.pem",
				Fields: []forge.Field{
					{Key: "certificate", Value: encodeCertPEM(serverDER)},
					{Key: "private-key", Value: serverKeyPEM, Sensitive: true},
				},
			},
			{
				Label: "client", Filename: "client.pem",
				Fields: []forge.Field{
					{Key: "certificate", Value: encodeCertPEM(clientDER)},
					{Key: "private-key", Value: clientKeyPEM, Sensitive: true},
				},
			},
		},
	}, nil
}

// Flags returns the generator-specific flags for CLI wiring.
func (m *MTLS) Flags() []forge.FlagDef {
	return []forge.FlagDef{
		{Name: "cn", Type: "string", Default: "localhost", Description: "Common name for server cert"},
		{Name: "days", Type: "int", Default: "825", Description: "Validity in days"},
		{Name: "san", Type: "string", Description: "Server SANs (comma-separated)"},
	}
}

// Bench runs a self-benchmark for the mtls generator.
func (m *MTLS) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.RunLegacy(ctx, m, 0)
}
