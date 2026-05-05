package tls

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"strings"
	"time"

	"github.com/smedje/smedje/internal/entropy"
)

// generateKey returns a crypto.Signer for the named algorithm.
// Supported values: "ed25519", "rsa-2048", "rsa-4096", "ecdsa-p256", "ecdsa-p384".
func generateKey(algo string) (crypto.Signer, error) {
	switch algo {
	case "ed25519":
		_, priv, err := ed25519.GenerateKey(entropy.Reader)
		return priv, err
	case "rsa-2048":
		return rsa.GenerateKey(entropy.Reader, 2048)
	case "rsa-4096":
		return rsa.GenerateKey(entropy.Reader, 4096)
	case "ecdsa-p256":
		return ecdsa.GenerateKey(elliptic.P256(), entropy.Reader)
	case "ecdsa-p384":
		return ecdsa.GenerateKey(elliptic.P384(), entropy.Reader)
	default:
		return nil, fmt.Errorf("tls: unsupported key algorithm %q", algo)
	}
}

// publicKey returns the public key from a crypto.Signer.
func publicKey(key crypto.Signer) crypto.PublicKey {
	return key.Public()
}

// certTemplate builds an x509.Certificate template with the given CN, validity
// duration, and SAN list. Each entry in sans is classified as an IP or DNS name.
func certTemplate(cn string, days int, sans []string) (*x509.Certificate, error) {
	serial, err := randSerial()
	if err != nil {
		return nil, err
	}

	var dnsNames []string
	var ips []net.IP
	for _, s := range sans {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if ip := net.ParseIP(s); ip != nil {
			ips = append(ips, ip)
		} else {
			dnsNames = append(dnsNames, s)
		}
	}

	now := time.Now()
	return &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: cn},
		NotBefore:    now,
		NotAfter:     now.AddDate(0, 0, days),
		DNSNames:     dnsNames,
		IPAddresses:  ips,
	}, nil
}

// signCert calls x509.CreateCertificate. If parent is nil the certificate is
// self-signed (template is used as both subject and issuer).
func signCert(template, parent *x509.Certificate, pub crypto.PublicKey, signer crypto.Signer) ([]byte, error) {
	if parent == nil {
		parent = template
	}
	return x509.CreateCertificate(entropy.Reader, template, parent, pub, signer)
}

// encodeCertPEM wraps a DER-encoded certificate in a PEM CERTIFICATE block.
func encodeCertPEM(der []byte) string {
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
}

// encodeKeyPEM encodes a private key as a PKCS#8 PEM PRIVATE KEY block.
func encodeKeyPEM(key crypto.Signer) (string, error) {
	b, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return "", fmt.Errorf("tls: marshal key: %w", err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: b})), nil
}

// randSerial returns a random 128-bit serial number suitable for X.509 certificates.
func randSerial() (*big.Int, error) {
	max := new(big.Int).Lsh(big.NewInt(1), 128)
	b := make([]byte, max.BitLen()/8+1)
	if _, err := entropy.Read(b); err != nil {
		return nil, err
	}
	n := new(big.Int).SetBytes(b)
	n.Mod(n, max)
	return n, nil
}

// parseSANs splits a comma-separated SAN string into DNS names and IP addresses.
func parseSANs(sans string) (dnsNames []string, ips []net.IP) {
	for _, s := range strings.Split(sans, ",") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if ip := net.ParseIP(s); ip != nil {
			ips = append(ips, ip)
		} else {
			dnsNames = append(dnsNames, s)
		}
	}
	return
}
