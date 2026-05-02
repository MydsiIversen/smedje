package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(recommendCmd)

	recommendCmd.Flags().String("use-case", "", "Filter to a specific use case")
	recommendCmd.Flags().Bool("json", false, "Output as JSON")
	recommendCmd.Flags().Bool("markdown", false, "Output as Markdown")
}

var recommendCmd = &cobra.Command{
	Use:   "recommend <topic>",
	Short: "Opinionated recommendations for common use cases",
	Long: `Available topics: id, ssh-key, tls-cert, password, hash, jwt, secret, vpn-key

Examples:
  smedje recommend id
  smedje recommend id --use-case "user-facing API"
  smedje recommend ssh-key --json`,
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"id", "ssh-key", "tls-cert", "password", "hash", "jwt", "secret", "vpn-key"},
	RunE: func(cmd *cobra.Command, args []string) error {
		topic := args[0]
		recs, ok := recommendations[topic]
		if !ok {
			topics := make([]string, 0, len(recommendations))
			for k := range recommendations {
				topics = append(topics, k)
			}
			return fmt.Errorf("unknown topic %q. Available: %s", topic, strings.Join(topics, ", "))
		}

		useCase, _ := cmd.Flags().GetString("use-case")
		if useCase != "" {
			filtered := filterByUseCase(recs, useCase)
			if len(filtered) == 0 {
				var cases []string
				for _, r := range recs {
					cases = append(cases, r.UseCase)
				}
				return fmt.Errorf("no use case matching %q. Available:\n  %s",
					useCase, strings.Join(cases, "\n  "))
			}
			recs = filtered
		}

		jsonFlag, _ := cmd.Flags().GetBool("json")
		mdFlag, _ := cmd.Flags().GetBool("markdown")

		if jsonFlag {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(recs)
		}
		if mdFlag {
			return renderRecommendationsMD(topic, recs)
		}
		return renderRecommendationsText(topic, recs)
	},
}

type recommendation struct {
	UseCase      string        `json:"use_case"`
	Primary      string        `json:"primary"`
	Why          string        `json:"why"`
	Command      string        `json:"command"`
	Alternatives []alternative `json:"alternatives,omitempty"`
	Avoid        []string      `json:"avoid,omitempty"`
}

type alternative struct {
	Name string `json:"name"`
	When string `json:"when"`
}

func filterByUseCase(recs []recommendation, query string) []recommendation {
	query = strings.ToLower(query)
	var out []recommendation
	for _, r := range recs {
		if strings.Contains(strings.ToLower(r.UseCase), query) {
			out = append(out, r)
		}
	}
	return out
}

func renderRecommendationsText(topic string, recs []recommendation) error {
	fmt.Printf("Recommendations: %s\n", topic)
	fmt.Println(strings.Repeat("─", 60))
	for i, r := range recs {
		if i > 0 {
			fmt.Println()
		}
		fmt.Printf("  Use case: %s\n", r.UseCase)
		fmt.Printf("  Recommended: %s\n", r.Primary)
		fmt.Printf("  Why: %s\n", r.Why)
		fmt.Printf("  Command: %s\n", r.Command)
		if len(r.Alternatives) > 0 {
			fmt.Printf("  Alternatives:\n")
			for _, a := range r.Alternatives {
				fmt.Printf("    - %s — %s\n", a.Name, a.When)
			}
		}
		if len(r.Avoid) > 0 {
			fmt.Printf("  Avoid: %s\n", strings.Join(r.Avoid, "; "))
		}
	}
	return nil
}

func renderRecommendationsMD(topic string, recs []recommendation) error {
	fmt.Printf("# Recommendations: %s\n\n", topic)
	for _, r := range recs {
		fmt.Printf("## %s\n\n", r.UseCase)
		fmt.Printf("**Recommended:** %s\n\n", r.Primary)
		fmt.Printf("**Why:** %s\n\n", r.Why)
		fmt.Printf("**Command:** `%s`\n\n", r.Command)
		if len(r.Alternatives) > 0 {
			fmt.Printf("**Alternatives:**\n\n")
			for _, a := range r.Alternatives {
				fmt.Printf("- %s — %s\n", a.Name, a.When)
			}
			fmt.Println()
		}
		if len(r.Avoid) > 0 {
			fmt.Printf("**Avoid:** %s\n\n", strings.Join(r.Avoid, "; "))
		}
	}
	return nil
}

var recommendations = map[string][]recommendation{
	"id": {
		{
			UseCase: "user-facing API",
			Primary: "Stripe-style prefixed UUIDv7",
			Why:     "time-ordered for DB performance, prefix identifies resource type",
			Command: "smedje uuid v7 (typeid prefix planned for v0.4)",
			Alternatives: []alternative{
				{Name: "nanoid", When: "shorter URL-safe slugs without time ordering"},
			},
			Avoid: []string{"uuid.v4 for new APIs (no ordering, worse index perf)"},
		},
		{
			UseCase: "internal database primary key",
			Primary: "UUIDv7",
			Why:     "time-ordered, B-tree friendly, 128-bit, standard format",
			Command: "smedje uuid v7",
			Alternatives: []alternative{
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
			Alternatives: []alternative{
				{Name: "uuid.v7", When: "no worker coordination needed"},
			},
		},
		{
			UseCase: "URL slug, short",
			Primary: "NanoID with custom alphabet",
			Why:     "configurable length, URL-safe, no padding characters",
			Command: "smedje nanoid --length 12",
			Alternatives: []alternative{
				{Name: "ulid", When: "need time-sorting in the slug"},
			},
		},
		{
			UseCase: "session token",
			Primary: "NanoID with high entropy",
			Why:     "URL-safe, no encoding needed, configurable length",
			Command: "smedje nanoid --length 32",
			Alternatives: []alternative{
				{Name: "uuid.v4", When: "standard format preferred"},
			},
			Avoid: []string{"short IDs (< 128 bits entropy for security tokens)"},
		},
		{
			UseCase: "log correlation",
			Primary: "ULID",
			Why:     "time-sortable, compact, grep-friendly (no hyphens)",
			Command: "smedje ulid",
			Alternatives: []alternative{
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
			Alternatives: []alternative{
				{Name: "RSA-3072", When: "legacy systems without Ed25519 support (planned for v0.4)"},
			},
			Avoid: []string{"RSA-1024 (broken)", "DSA (deprecated)", "ECDSA with P-192"},
		},
		{
			UseCase: "legacy compatibility",
			Primary: "RSA-3072 (planned for v0.4)",
			Why:     "widest compatibility across old SSH servers",
			Command: "smedje ssh ed25519 (use Ed25519 where supported)",
			Alternatives: []alternative{
				{Name: "Ed25519", When: "server supports it (OpenSSH 6.5+)"},
			},
		},
		{
			UseCase: "code signing",
			Primary: "Ed25519",
			Why:     "deterministic signatures, fast verification, compact",
			Command: "smedje ssh ed25519",
		},
	},
	"tls-cert": {
		{
			UseCase: "local development",
			Primary: "Self-signed Ed25519, 825-day validity",
			Why:     "fast generation, no external deps, Apple trust-store compatible",
			Command: "smedje tls self-signed --days 825",
			Alternatives: []alternative{
				{Name: "mkcert", When: "want automatic trust-store installation"},
			},
		},
		{
			UseCase: "internal mTLS",
			Primary: "Internal CA + leaf certs (planned for v0.4)",
			Why:     "one root in each service's trust store, sign many leaves",
			Command: "smedje tls self-signed (interim; CA planned for v0.4)",
		},
		{
			UseCase: "publicly-trusted",
			Primary: "Let's Encrypt or your organization's CA",
			Why:     "Smedje does not issue publicly-trusted certs",
			Command: "smedje tls self-signed (for CSR generation, planned for v0.4)",
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
			Alternatives: []alternative{
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
			Primary: "ES256 / P-256 (planned for v0.4)",
			Why:     "required by many OIDC providers, compact signatures",
			Command: "(planned for v0.4)",
		},
		{
			UseCase: "internal service token",
			Primary: "EdDSA / Ed25519 (planned for v0.4)",
			Why:     "fastest verification, smallest keys, modern choice",
			Command: "(planned for v0.4)",
			Alternatives: []alternative{
				{Name: "ES256", When: "need broader JWT library support"},
			},
		},
		{
			UseCase: "symmetric only",
			Primary: "HS256 (planned for v0.4)",
			Why:     "simplest when signer and verifier share a key",
			Command: "(planned for v0.4)",
			Avoid:   []string{"sharing the key with untrusted parties"},
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
			Command: "smedje password --length 44 --charset alphanum (interim; PSK gen planned)",
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
			Alternatives: []alternative{
				{Name: "certificate-based", When: "scale beyond a handful of peers"},
			},
		},
	},
}
