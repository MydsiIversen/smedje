package ssh

const whyEd25519 = `About Ed25519 (RFC 8709, OpenSSH 6.5+):
  Elliptic-curve signature scheme using Curve25519. 256-bit keys with
  128-bit security level. Deterministic signing (no per-signature
  randomness needed). Fast, compact, and constant-time.

Why default:
  Ed25519 is the recommended key type for new deployments. Faster than
  RSA, smaller keys, no parameter choices to get wrong, resistant to
  timing attacks by design.

Alternatives:
  RSA-3072    wider compatibility (older systems), larger keys, slower
              — use only if Ed25519 is unsupported (planned for v0.4)
  ECDSA       curve choice matters (P-256 vs P-384); Ed25519 preferred
              unless FIPS mandates NIST curves

See also: smedje recommend ssh-key
`

func (e *Ed25519) Why() string { return whyEd25519 }
