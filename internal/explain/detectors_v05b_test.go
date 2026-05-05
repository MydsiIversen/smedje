package explain

import (
	"testing"
)

func TestMACDetector(t *testing.T) {
	m, ok := (&macDetector{}).Detect("00:50:56:ab:cd:ef")
	if !ok {
		t.Fatal("should detect MAC")
	}
	if m.Fields["vendor"] != "VMware" {
		t.Fatalf("vendor = %q, want VMware", m.Fields["vendor"])
	}
	if m.Confidence != 0.85 {
		t.Fatalf("confidence = %f, want 0.85", m.Confidence)
	}
}

func TestMACDetectorUnknownOUI(t *testing.T) {
	m, ok := (&macDetector{}).Detect("aa:bb:cc:dd:ee:ff")
	if !ok {
		t.Fatal("should still detect MAC format")
	}
	if m.Fields["vendor"] != "Unknown" {
		t.Fatalf("vendor = %q, want Unknown", m.Fields["vendor"])
	}
}

func TestJWTDetector(t *testing.T) {
	jwt := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIn0.fakesig"
	m, ok := (&jwtDetector{}).Detect(jwt)
	if !ok {
		t.Fatal("should detect JWT")
	}
	if m.Fields["algorithm"] != "HS256" {
		t.Fatalf("algorithm = %q, want HS256", m.Fields["algorithm"])
	}
}

func TestSSHPubKeyDetector(t *testing.T) {
	key := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI testkey"
	m, ok := (&sshPubKeyDetector{}).Detect(key)
	if !ok {
		t.Fatal("should detect SSH public key")
	}
	if m.Fields["algorithm"] != "ssh-ed25519" {
		t.Fatalf("algorithm = %q", m.Fields["algorithm"])
	}
}

func TestPEMDetector(t *testing.T) {
	pem := "-----BEGIN CERTIFICATE-----\nMIIB...\n-----END CERTIFICATE-----"
	m, ok := (&pemDetector{}).Detect(pem)
	if !ok {
		t.Fatal("should detect PEM")
	}
	if m.Fields["type"] != "CERTIFICATE" {
		t.Fatalf("type = %q, want CERTIFICATE", m.Fields["type"])
	}
}

func TestIQNDetector(t *testing.T) {
	m, ok := (&iqnDetector{}).Detect("iqn.2024-01.com.example:storage.lun0")
	if !ok {
		t.Fatal("should detect IQN")
	}
	if m.Fields["authority"] != "com.example" {
		t.Fatalf("authority = %q", m.Fields["authority"])
	}
}

func TestAgeKeyDetector(t *testing.T) {
	m, ok := (&ageKeyDetector{}).Detect("age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p")
	if !ok {
		t.Fatal("should detect age public key")
	}
	if m.Confidence != 0.95 {
		t.Fatalf("confidence = %f, want 0.95", m.Confidence)
	}
}

func TestWireGuardKeyDetector(t *testing.T) {
	m, ok := (&wgKeyDetector{}).Detect("YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXowMTIzNDU=")
	if !ok {
		t.Fatal("should detect possible WG key")
	}
	if m.Confidence != 0.40 {
		t.Fatalf("confidence = %f, want 0.40", m.Confidence)
	}
}
