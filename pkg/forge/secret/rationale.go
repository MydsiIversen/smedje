package secret

const whyPassword = `About password generation:
  Random password from crypto/rand using math/big for uniform
  distribution across the charset. Default: 24 characters from
  the full printable ASCII set (uppercase, lowercase, digits,
  symbols). ~157 bits of entropy at default length.

Why these defaults:
  24 characters provides excellent security against brute force while
  staying within most systems' maximum password length. Full charset
  maximizes entropy per character.

Alternatives:
  --charset alpha      letters only (weaker but no symbol issues)
  --charset alphanum   letters + digits (safe for systems that reject symbols)
  --length 32          for service accounts or high-security contexts
  diceware             passphrase-style (planned for v0.4)

See also: smedje recommend password
`

const whyTOTP = `About TOTP (RFC 6238):
  Time-based One-Time Password secret. Generates a 20-byte (160-bit)
  base32-encoded secret matching the SHA-1 HMAC block size used by
  most authenticator apps. Includes an otpauth:// URI for QR code
  generation.

Why these defaults:
  20 bytes matches the SHA-1 HMAC key size — shorter wastes HMAC
  capacity, longer provides no additional security. 6 digits and
  30-second period are the de facto standard supported by all major
  authenticator apps.

Alternatives:
  --digits 8      8-digit codes (some enterprise systems require this)
  --period 60     longer validity window (reduces user pressure)

See also: smedje recommend secret
`

func (p *Password) Why() string { return whyPassword }
func (t *TOTP) Why() string    { return whyTOTP }
