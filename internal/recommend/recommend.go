// Package recommend provides opinionated recommendations for common
// cryptographic and identity use cases. The data is shared between the
// CLI and the web API.
package recommend

import (
	"sort"
	"strings"
)

// Recommendation describes a single opinionated recommendation for a use case.
type Recommendation struct {
	UseCase      string        `json:"use_case"`
	Primary      string        `json:"primary"`
	Why          string        `json:"why"`
	Command      string        `json:"command"`
	Alternatives []Alternative `json:"alternatives,omitempty"`
	Avoid        []string      `json:"avoid,omitempty"`
}

// Alternative describes a secondary option and when to prefer it.
type Alternative struct {
	Name string `json:"name"`
	When string `json:"when"`
}

// FilterByUseCase returns recommendations whose UseCase contains the query
// (case-insensitive substring match).
func FilterByUseCase(recs []Recommendation, query string) []Recommendation {
	query = strings.ToLower(query)
	var out []Recommendation
	for _, r := range recs {
		if strings.Contains(strings.ToLower(r.UseCase), query) {
			out = append(out, r)
		}
	}
	return out
}

// Topics returns the sorted list of valid recommendation topics.
func Topics() []string {
	topics := make([]string, 0, len(Recommendations))
	for k := range Recommendations {
		topics = append(topics, k)
	}
	sort.Strings(topics)
	return topics
}

// Recommendations maps topic names to their recommendation sets.
var Recommendations = map[string][]Recommendation{
	"id": {
		{
			UseCase: "user-facing API",
			Primary: "Stripe-style prefixed UUIDv7",
			Why:     "time-ordered for DB performance, prefix identifies resource type",
			Command: "smedje uuid v7 (typeid prefix planned for v0.4)",
			Alternatives: []Alternative{
				{Name: "nanoid", When: "shorter URL-safe slugs without time ordering"},
			},
			Avoid: []string{"uuid.v4 for new APIs (no ordering, worse index perf)"},
		},
		{
			UseCase: "internal database primary key",
			Primary: "UUIDv7",
			Why:     "time-ordered, B-tree friendly, 128-bit, standard format",
			Command: "smedje uuid v7",
			Alternatives: []Alternative{
				{Name: "snowflake", When: "need 64-bit integer keys"},
				{Name: "ulid", When: "prefer shorter Crockford representation"},
			},
			Avoid: []string{"auto-increment (leaks volume)", "uuid.v4 (index fragmentation)"},
		},
		{
			UseCase: "distributed system, compact ID",
			Primary: "Snowflake or UUIDv7",
			Why:     "snowflake fits int64 columns; v7 is standard 128-bit",
			Command: "smedje snowflake --worker 1",
			Alternatives: []Alternative{
				{Name: "uuid.v7", When: "no worker coordination needed"},
			},
		},
		{
			UseCase: "URL slug, short",
			Primary: "NanoID with custom alphabet",
			Why:     "configurable length, URL-safe, no padding characters",
			Command: "smedje nanoid --length 12",
			Alternatives: []Alternative{
				{Name: "ulid", When: "need time-sorting in the slug"},
			},
		},
		{
			UseCase: "session token",
			Primary: "NanoID with high entropy",
			Why:     "URL-safe, no encoding needed, configurable length",
			Command: "smedje nanoid --length 32",
			Alternatives: []Alternative{
				{Name: "uuid.v4", When: "standard format preferred"},
			},
			Avoid: []string{"short IDs (< 128 bits entropy for security tokens)"},
		},
		{
			UseCase: "log correlation",
			Primary: "ULID",
			Why:     "time-sortable, compact, grep-friendly (no hyphens)",
			Command: "smedje ulid",
			Alternatives: []Alternative{
				{Name: "uuid.v7", When: "existing UUID infrastructure"},
			},
		},
	},
	"ssh-key": {
		{
			UseCase: "personal key",
			Primary: "Ed25519",
			Why:     "fastest, smallest, most secure modern option",
			Command: "smedje ssh ed25519",
			Alternatives: []Alternative{
				{Name: "RSA-4096", When: "legacy systems without Ed25519 support"},
			},
			Avoid: []string{"RSA-1024 (broken)", "DSA (deprecated)", "ECDSA with P-192"},
		},
		{
			UseCase: "legacy compatibility",
			Primary: "RSA-4096",
			Why:     "widest compatibility across old SSH servers",
			Command: "smedje ssh rsa --bits 4096",
			Alternatives: []Alternative{
				{Name: "Ed25519", When: "server supports it (OpenSSH 6.5+)"},
			},
		},
		{
			UseCase: "code signing",
			Primary: "Ed25519",
			Why:     "deterministic signatures, fast verification, compact",
			Command: "smedje ssh ed25519",
		},
		{
			UseCase: "constrained environments",
			Primary: "ECDSA P-256",
			Why:     "smaller keys than RSA, wider support than Ed25519 on embedded devices",
			Command: "smedje ssh ecdsa",
			Avoid:   []string{"RSA-2048 for new deployments"},
		},
	},
	"tls-cert": {
		{
			UseCase: "local development",
			Primary: "Self-signed Ed25519, 825-day validity",
			Why:     "fast generation, no external deps, Apple trust-store compatible",
			Command: "smedje tls self-signed --days 825",
			Alternatives: []Alternative{
				{Name: "mkcert", When: "want automatic trust-store installation"},
			},
		},
		{
			UseCase: "internal mTLS",
			Primary: "mTLS bundle (CA + server + client)",
			Why:     "mutual authentication for zero-trust service communication",
			Command: "smedje tls mtls",
		},
		{
			UseCase: "publicly-trusted",
			Primary: "Let's Encrypt or your organization's CA",
			Why:     "Smedje does not issue publicly-trusted certs",
			Command: "smedje tls csr --cn example.com",
		},
		{
			UseCase: "internal PKI",
			Primary: "CA chain with Ed25519",
			Why:     "one root in each service's trust store, sign many leaves",
			Command: "smedje tls ca-chain",
			Alternatives: []Alternative{
				{Name: "step-ca", When: "need automated certificate lifecycle management"},
			},
		},
		{
			UseCase: "service mesh mTLS",
			Primary: "mTLS bundle (CA + server + client)",
			Why:     "mutual authentication for zero-trust service communication",
			Command: "smedje tls mtls --cn myservice.local",
		},
		{
			UseCase: "public CA submission",
			Primary: "CSR (Certificate Signing Request)",
			Why:     "generate a private key and CSR for submission to a public CA",
			Command: "smedje tls csr --cn example.com --san example.com,www.example.com",
		},
	},
	"password": {
		{
			UseCase: "user account",
			Primary: "16-character full charset",
			Why:     "~105 bits entropy, passes most complexity requirements",
			Command: "smedje password --length 16",
		},
		{
			UseCase: "service account",
			Primary: "32-character alphanumeric",
			Why:     "high entropy without special-char escaping issues",
			Command: "smedje password --length 32 --charset alphanum",
		},
		{
			UseCase: "automation token",
			Primary: "NanoID with long length",
			Why:     "URL-safe, no quoting needed in configs and env vars",
			Command: "smedje nanoid --length 48",
		},
		{
			UseCase: "passphrase",
			Primary: "Diceware (planned for v0.4)",
			Why:     "memorable, high entropy from word combinations",
			Command: "smedje password --length 24 (interim)",
		},
	},
	"hash": {
		{
			UseCase: "password storage",
			Primary: "argon2id (planned for v0.4)",
			Why:     "memory-hard, resistant to GPU/ASIC attacks",
			Command: "(planned for v0.4; use bcrypt as interim)",
			Alternatives: []Alternative{
				{Name: "bcrypt", When: "acceptable but no longer state-of-the-art"},
			},
			Avoid: []string{"MD5", "SHA-1", "plain SHA-256 (no key stretching)"},
		},
		{
			UseCase: "general-purpose integrity",
			Primary: "BLAKE3 or SHA-256 (planned for v0.4)",
			Why:     "BLAKE3 is fastest; SHA-256 is most widely verified",
			Command: "(planned for v0.4)",
		},
		{
			UseCase: "FIPS compliance",
			Primary: "PBKDF2-SHA256 (planned for v0.4)",
			Why:     "NIST SP 800-132 approved for key derivation",
			Command: "(planned for v0.4)",
			Avoid:   []string{"bcrypt (not FIPS-approved)", "scrypt (not FIPS-approved)"},
		},
	},
	"jwt": {
		{
			UseCase: "OIDC integration",
			Primary: "ES256 / P-256",
			Why:     "required by many OIDC providers, compact signatures",
			Command: "smedje jwt es256",
			Alternatives: []Alternative{
				{Name: "RS256", When: "legacy systems requiring RSA"},
			},
		},
		{
			UseCase: "internal service token",
			Primary: "EdDSA / Ed25519",
			Why:     "fastest verification, smallest keys, modern choice",
			Command: "smedje jwt eddsa",
			Alternatives: []Alternative{
				{Name: "ES256", When: "need broader JWT library support"},
			},
		},
		{
			UseCase: "symmetric only",
			Primary: "HS256",
			Why:     "simplest when signer and verifier share a key",
			Command: "smedje jwt hs256",
			Avoid:   []string{"sharing the key with untrusted parties"},
		},
		{
			UseCase: "legacy RSA integration",
			Primary: "RS256",
			Why:     "widest JWT library compatibility",
			Command: "smedje jwt rs256",
			Avoid:   []string{"RS384/RS512 (no practical security gain over RS256)"},
		},
	},
	"secret": {
		{
			UseCase: "TOTP secret for 2FA",
			Primary: "20-byte HMAC-SHA1 key",
			Why:     "matches authenticator app expectations (SHA-1 block size)",
			Command: "smedje totp",
		},
		{
			UseCase: "API secret key",
			Primary: "NanoID with 48 characters",
			Why:     "URL-safe, high entropy, no encoding issues",
			Command: "smedje nanoid --length 48",
		},
		{
			UseCase: "pre-shared key (PSK)",
			Primary: "32-byte random (base64)",
			Why:     "256 bits matches most symmetric cipher key sizes",
			Command: "smedje network ipsec-psk",
		},
	},
	"vpn-key": {
		{
			UseCase: "WireGuard tunnel",
			Primary: "Curve25519 keypair",
			Why:     "WireGuard only supports Curve25519 — no choice needed",
			Command: "smedje wireguard keypair",
		},
		{
			UseCase: "IPsec pre-shared key",
			Primary: "High-entropy random string",
			Why:     "PSK must be long enough to resist brute force",
			Command: "smedje password --length 64 --charset alphanum",
			Alternatives: []Alternative{
				{Name: "certificate-based", When: "scale beyond a handful of peers"},
			},
		},
		{
			UseCase: "multi-site WireGuard",
			Primary: "WireGuard mesh",
			Why:     "generates N peer configs with cross-referenced public keys",
			Command: "smedje wireguard mesh --peers 3",
		},
		{
			UseCase: "legacy site-to-site VPN",
			Primary: "IPsec pre-shared key",
			Why:     "256-bit hex PSK for IKEv2 tunnels",
			Command: "smedje network ipsec-psk",
			Alternatives: []Alternative{
				{Name: "certificate-based", When: "scaling beyond a handful of peers"},
			},
		},
	},
	"network-secret": {
		{
			UseCase: "site-to-site VPN",
			Primary: "IPsec PSK (hex, 32 bytes)",
			Why:     "standard pre-shared key for IKEv2 tunnels, 256-bit entropy",
			Command: "smedje network ipsec-psk",
			Alternatives: []Alternative{
				{Name: "certificate-based", When: "scaling beyond a handful of peers"},
			},
			Avoid: []string{"shared secrets over 64 bytes (diminishing returns)"},
		},
		{
			UseCase: "AAA / RADIUS",
			Primary: "RADIUS shared secret (base64, 24 bytes)",
			Why:     "authenticates RADIUS client-server communication",
			Command: "smedje network radius-secret",
		},
		{
			UseCase: "legacy monitoring",
			Primary: "SNMPv3 community string",
			Why:     "alphanumeric, avoids special-char issues in SNMP configs",
			Command: "smedje network snmp-community",
			Avoid:   []string{"SNMPv1/v2c community strings in production (cleartext)"},
		},
		{
			UseCase: "OpenVPN HMAC firewall",
			Primary: "OpenVPN tls-auth key",
			Why:     "HMAC firewall drops unauthenticated packets before TLS handshake",
			Command: "smedje network openvpn-tls-auth",
		},
	},
	"email-auth": {
		{
			UseCase: "domain email signing",
			Primary: "DKIM RSA-2048 keypair",
			Why:     "signs outgoing email, DNS TXT record for verification",
			Command: "smedje email dkim --selector mail --domain example.com",
			Avoid:   []string{"DKIM keys shorter than 1024 bits (insecure)", "RSA-4096 DKIM (DNS TXT record too large for some providers)"},
		},
		{
			UseCase: "email policy enforcement",
			Primary: "DMARC DNS record",
			Why:     "instructs receivers how to handle SPF/DKIM failures",
			Command: "smedje email dmarc --domain example.com --policy quarantine",
			Avoid:   []string{"p=none long-term (provides no protection)"},
		},
	},
	"age": {
		{
			UseCase: "file encryption",
			Primary: "age X25519 keypair",
			Why:     "simple, modern, no configuration needed — the GPG replacement",
			Command: "smedje age x25519",
			Avoid:   []string{"GPG for new projects (complexity, key management burden)"},
		},
	},
	"storage-id": {
		{
			UseCase: "SAN target naming",
			Primary: "iSCSI IQN",
			Why:     "globally unique, human-readable, includes authority and date",
			Command: "smedje network iqn --authority com.example --target storage.lun0",
		},
		{
			UseCase: "Fibre Channel WWN",
			Primary: "WWPN (NAA 5 format)",
			Why:     "48-bit random in NAA 5 format for FC fabric addressing",
			Command: "smedje network wwpn",
		},
		{
			UseCase: "VM / container NIC",
			Primary: "OUI-based MAC address",
			Why:     "vendor-prefixed avoids collisions with real hardware",
			Command: "smedje oui-mac --oui 00:50:56",
			Avoid:   []string{"random MACs without OUI prefix (potential collisions with real hardware)"},
		},
	},
}
