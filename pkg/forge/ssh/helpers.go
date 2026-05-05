package ssh

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/pem"
	"fmt"

	"github.com/smedje/smedje/internal/entropy"

	gossh "golang.org/x/crypto/ssh"
)

// generateSSHKeypair generates an SSH keypair for the given algorithm.
// For RSA, bits specifies the key size; pass 0 to use the default (4096).
// For ECDSA variants, bits is ignored.
// Returns the private key in OpenSSH PEM format and the public key in
// authorized_keys format.
func generateSSHKeypair(algo string, bits int, comment string) (privPEM, pubAuthorized string, err error) {
	var signer interface{}

	switch algo {
	case "ed25519":
		_, priv, err := ed25519.GenerateKey(entropy.Reader)
		if err != nil {
			return "", "", fmt.Errorf("ssh: ed25519 keygen: %w", err)
		}
		signer = priv
	case "rsa":
		if bits == 0 {
			bits = 4096
		}
		priv, err := rsa.GenerateKey(entropy.Reader, bits)
		if err != nil {
			return "", "", fmt.Errorf("ssh: rsa keygen: %w", err)
		}
		signer = priv
	case "ecdsa-p256":
		priv, err := ecdsa.GenerateKey(elliptic.P256(), entropy.Reader)
		if err != nil {
			return "", "", fmt.Errorf("ssh: ecdsa keygen: %w", err)
		}
		signer = priv
	case "ecdsa-p384":
		priv, err := ecdsa.GenerateKey(elliptic.P384(), entropy.Reader)
		if err != nil {
			return "", "", fmt.Errorf("ssh: ecdsa keygen: %w", err)
		}
		signer = priv
	default:
		return "", "", fmt.Errorf("ssh: unsupported algo %q", algo)
	}

	privBlock, err := gossh.MarshalPrivateKey(signer, comment)
	if err != nil {
		return "", "", fmt.Errorf("ssh: marshal private key: %w", err)
	}
	privStr := string(pem.EncodeToMemory(privBlock))

	pub, err := gossh.NewPublicKey(publicKeyFrom(signer))
	if err != nil {
		return "", "", fmt.Errorf("ssh: marshal public key: %w", err)
	}
	pubStr := string(gossh.MarshalAuthorizedKey(pub))

	return privStr, pubStr, nil
}

// publicKeyFrom extracts the public key from a private key of any supported type.
func publicKeyFrom(key interface{}) interface{} {
	switch k := key.(type) {
	case ed25519.PrivateKey:
		return k.Public()
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		panic(fmt.Sprintf("ssh: unknown key type %T", key))
	}
}
