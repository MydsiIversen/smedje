package wireguard

const whyKeypair = `About WireGuard keypair (Curve25519):
  X25519 key agreement keypair with proper clamping per wg(8). The
  private key is 32 bytes from crypto/rand with bits 0, 1, 2, 255
  cleared/set per the Curve25519 spec. Output matches wg genkey/pubkey
  base64 format.

Why use it:
  WireGuard requires exactly this key format. There are no alternative
  curves or parameters — Curve25519 is the only option.

Note:
  Smedje generates keys but does not manage them. For production
  tunnels, integrate with your secrets manager or use wg genkey
  directly on the target host to avoid key transport.

See also: smedje recommend vpn-key
`

func (k *Keypair) Why() string { return whyKeypair }
