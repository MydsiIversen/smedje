// Package tls provides TLS certificate generators.
package tls

import (
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"strings"
	"time"

	"github.com/smedje/smedje/internal/bench"
	"github.com/smedje/smedje/internal/config"
	"github.com/smedje/smedje/internal/entropy"
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

	days := 365
	if def := config.GetDefault("tls.days"); def != "" {
		fmt.Sscanf(def, "%d", &days)
	}
	if v, ok := opts.Params["days"]; ok {
		fmt.Sscanf(v, "%d", &days)
	}

	var dnsNames []string
	var ipAddrs []net.IP
	if sans, ok := opts.Params["san"]; ok {
		for _, s := range strings.Split(sans, ",") {
			s = strings.TrimSpace(s)
			if ip := net.ParseIP(s); ip != nil {
				ipAddrs = append(ipAddrs, ip)
			} else {
				dnsNames = append(dnsNames, s)
			}
		}
	}
	if len(dnsNames) == 0 && len(ipAddrs) == 0 {
		dnsNames = []string{cn}
		ipAddrs = []net.IP{net.IPv4(127, 0, 0, 1)}
	}

	pub, priv, err := ed25519.GenerateKey(entropy.Reader)
	if err != nil {
		return nil, fmt.Errorf("tls: keygen: %w", err)
	}

	serialMax := new(big.Int).Lsh(big.NewInt(1), 128)
	serial, err := randBigInt(serialMax)
	if err != nil {
		return nil, fmt.Errorf("tls: serial: %w", err)
	}

	now := time.Now()
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: cn},
		NotBefore:    now,
		NotAfter:     now.AddDate(0, 0, days),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     dnsNames,
		IPAddresses:  ipAddrs,
	}

	certDER, err := x509.CreateCertificate(entropy.Reader, tmpl, tmpl, pub, priv)
	if err != nil {
		return nil, fmt.Errorf("tls: create cert: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, fmt.Errorf("tls: marshal key: %w", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})

	return forge.SingleArtifact("tls",
		forge.Field{Key: "certificate", Value: string(certPEM)},
		forge.Field{Key: "private-key", Value: string(keyPEM), Sensitive: true},
	), nil
}

func (s *SelfSigned) Flags() []forge.FlagDef {
	return []forge.FlagDef{
		{Name: "cn", Type: "string", Default: "localhost", Description: "Common name"},
		{Name: "days", Type: "int", Default: "825", Description: "Validity in days"},
		{Name: "san", Type: "string", Description: "Subject alternative names (comma-separated)"},
	}
}

func (s *SelfSigned) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.RunLegacy(ctx, s, 0)
}

func randBigInt(max *big.Int) (*big.Int, error) {
	b := make([]byte, max.BitLen()/8+1)
	if _, err := entropy.Read(b); err != nil {
		return nil, err
	}
	n := new(big.Int).SetBytes(b)
	n.Mod(n, max)
	return n, nil
}
