package tls

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net"
	"testing"
)

func TestGenerateKey(t *testing.T) {
	tests := []struct {
		algo    string
		keyType string
	}{
		{"ed25519", "ed25519"},
		{"rsa-2048", "rsa"},
		{"rsa-4096", "rsa"},
		{"ecdsa-p256", "ecdsa"},
		{"ecdsa-p384", "ecdsa"},
	}
	for _, tt := range tests {
		t.Run(tt.algo, func(t *testing.T) {
			key, err := generateKey(tt.algo)
			if err != nil {
				t.Fatalf("generateKey(%q): %v", tt.algo, err)
			}
			switch tt.keyType {
			case "ed25519":
				if _, ok := key.(ed25519.PrivateKey); !ok {
					t.Fatalf("expected ed25519.PrivateKey, got %T", key)
				}
			case "rsa":
				k, ok := key.(*rsa.PrivateKey)
				if !ok {
					t.Fatalf("expected *rsa.PrivateKey, got %T", key)
				}
				if tt.algo == "rsa-4096" && k.N.BitLen() != 4096 {
					t.Fatalf("expected 4096-bit key, got %d", k.N.BitLen())
				}
			case "ecdsa":
				k, ok := key.(*ecdsa.PrivateKey)
				if !ok {
					t.Fatalf("expected *ecdsa.PrivateKey, got %T", key)
				}
				if tt.algo == "ecdsa-p384" && k.Curve != elliptic.P384() {
					t.Fatalf("expected P-384 curve")
				}
			}
		})
	}
}

func TestGenerateKeyInvalid(t *testing.T) {
	_, err := generateKey("invalid")
	if err == nil {
		t.Fatal("expected error for invalid algo")
	}
}

func TestCertTemplate(t *testing.T) {
	tmpl, err := certTemplate("test.local", 30, []string{"test.local", "127.0.0.1"})
	if err != nil {
		t.Fatal(err)
	}
	if tmpl.Subject.CommonName != "test.local" {
		t.Fatalf("CN = %q, want %q", tmpl.Subject.CommonName, "test.local")
	}
	if len(tmpl.DNSNames) != 1 || tmpl.DNSNames[0] != "test.local" {
		t.Fatalf("DNSNames = %v, want [test.local]", tmpl.DNSNames)
	}
	if len(tmpl.IPAddresses) != 1 || !tmpl.IPAddresses[0].Equal(net.ParseIP("127.0.0.1")) {
		t.Fatalf("IPAddresses = %v, want [127.0.0.1]", tmpl.IPAddresses)
	}
}

func TestSignCertSelfSigned(t *testing.T) {
	key, _ := generateKey("ed25519")
	tmpl, _ := certTemplate("test.local", 30, nil)
	der, err := signCert(tmpl, nil, key.Public(), key)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	if cert.Subject.CommonName != "test.local" {
		t.Fatalf("CN = %q, want %q", cert.Subject.CommonName, "test.local")
	}
}

func TestEncodeCertPEM(t *testing.T) {
	key, _ := generateKey("ed25519")
	tmpl, _ := certTemplate("test", 1, nil)
	der, _ := signCert(tmpl, nil, key.Public(), key)
	pemStr := encodeCertPEM(der)
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil || block.Type != "CERTIFICATE" {
		t.Fatal("expected CERTIFICATE PEM block")
	}
}

func TestEncodeKeyPEM(t *testing.T) {
	key, _ := generateKey("ed25519")
	pemStr, err := encodeKeyPEM(key)
	if err != nil {
		t.Fatal(err)
	}
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil || block.Type != "PRIVATE KEY" {
		t.Fatal("expected PRIVATE KEY PEM block")
	}
}

func TestParseSANs(t *testing.T) {
	dns, ips := parseSANs("example.com,192.168.1.1,test.local")
	if len(dns) != 2 {
		t.Fatalf("dns = %v, want 2 entries", dns)
	}
	if len(ips) != 1 {
		t.Fatalf("ips = %v, want 1 entry", ips)
	}
}

func TestRandSerial(t *testing.T) {
	s1, err := randSerial()
	if err != nil {
		t.Fatal(err)
	}
	s2, _ := randSerial()
	if s1.Cmp(s2) == 0 {
		t.Fatal("two serials should not be equal")
	}
}

func TestPublicKey(t *testing.T) {
	key, _ := generateKey("rsa-2048")
	pub := publicKey(key)
	if _, ok := pub.(*rsa.PublicKey); !ok {
		t.Fatalf("expected *rsa.PublicKey, got %T", pub)
	}
}
