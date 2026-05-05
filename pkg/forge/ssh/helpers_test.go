package ssh

import (
	"strings"
	"testing"
)

func TestGenerateSSHKeypairEd25519(t *testing.T) {
	priv, pub, err := generateSSHKeypair("ed25519", 0)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(priv, "OPENSSH PRIVATE KEY") {
		t.Fatal("expected OpenSSH private key format")
	}
	if !strings.HasPrefix(pub, "ssh-ed25519 ") {
		t.Fatal("expected ssh-ed25519 prefix")
	}
}

func TestGenerateSSHKeypairRSA(t *testing.T) {
	priv, pub, err := generateSSHKeypair("rsa", 2048)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(priv, "OPENSSH PRIVATE KEY") {
		t.Fatal("expected OpenSSH private key format")
	}
	if !strings.HasPrefix(pub, "ssh-rsa ") {
		t.Fatal("expected ssh-rsa prefix")
	}
}

func TestGenerateSSHKeypairECDSA(t *testing.T) {
	priv, pub, err := generateSSHKeypair("ecdsa-p256", 0)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(priv, "OPENSSH PRIVATE KEY") {
		t.Fatal("expected OpenSSH private key format")
	}
	if !strings.HasPrefix(pub, "ecdsa-sha2-nistp256 ") {
		t.Fatalf("expected ecdsa prefix, got %q", pub[:30])
	}
}
