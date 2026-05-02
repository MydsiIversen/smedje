package tls

const whySelfSigned = `About self-signed TLS (Ed25519):
  A self-signed X.509 certificate with Ed25519 key. Not trusted by
  browsers or system trust stores without explicit installation.
  Default validity is 825 days (Apple's maximum for local trust).

Why use it:
  Local development, internal services, mTLS between services you
  control. Avoids the overhead of a full PKI for environments where
  trust is established out of band.

Alternatives:
  Let's Encrypt     free, publicly trusted — use for public-facing TLS
  Internal CA       sign many leaves from one root (planned for v0.4)
  mkcaol/step-ca   tools that install a local CA into system trust

Note:
  Self-signed certs require adding to the trust store on every client.
  For multi-service environments, an internal CA is usually simpler.

See also: smedje recommend tls-cert
`

func (s *SelfSigned) Why() string { return whySelfSigned }
