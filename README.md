# Smedje

**Forge every key, ID, cert, and config from scratch.**

One binary replaces the dozen single-purpose tools you reach for whenever
you need to generate cryptographic artifacts — UUIDs, Snowflake IDs, SSH
keys, TLS certs, WireGuard configs, TOTP secrets, passwords, MAC addresses,
and more.

## 30-second tour

```bash
# Install
go install github.com/smedje/smedje/cmd/smedje@latest

# Generate things
smedje uuid v7                 # 019dea33-4e8a-7585-ad4d-6a3232718cb3
smedje uuid v4                 # 9f8b3c6a-1d4e-4f8a-a2b7-5c6d8e9f0a1b
smedje ulid                    # 01KQN58A9C978951R7QB2CNDEJ
smedje nanoid                  # V1StGXR8_Z5jdHi6B-myT
smedje snowflake               # 309052152221794304
smedje ssh ed25519             # OpenSSH keypair to stdout
smedje tls self-signed         # self-signed cert + key
smedje wireguard keypair       # Curve25519 keypair (base64)
smedje password                # DJa}#fM%$7e(5|wyL2hl]*28
smedje totp                    # TOTP secret + otpauth URI
smedje mac                     # 42:88:8a:b6:fe:3c

# Bulk generation
smedje uuid v7 --count 1000 --format csv    # 1000 UUIDs as CSV
smedje password --count 5 --json            # 5 passwords as JSON array
smedje nanoid --count 10 --format sql       # SQL INSERT statement

# Configuration
smedje config show --explain   # show all effective values + sources
smedje config set password.length 32

# Explain any value
smedje explain "019dea33-4e8a-7585-ad4d-6a3232718cb3"
# → Format: UUIDv7 (Unix time-ordered)
#     timestamp: 2026-05-02T...

# Recommendations
smedje recommend id            # opinionated ID format advice
smedje recommend ssh-key       # which key type for which job

# Explain why
smedje uuid v7 --why           # generate + rationale + alternatives

# Reproducible output
smedje uuid v7 --seed test --count 5  # same seed = same output

# Benchmarks
smedje bench all               # benchmark every generator
smedje bench compare uuid.v7 ulid nanoid  # side-by-side comparison
smedje bench list              # show all addressable generator names

# Output modes
smedje uuid v7 --json          # {"value": "..."}
smedje uuid v7 --quiet         # raw value only
smedje uuid v7 --format env    # UUID_1_VALUE=...
```

## What it forges

| Command | What | Key bits |
|---|---|---|
| `uuid v1` | RFC 9562 UUIDv1 (time + random node) | 14-bit clock + 48-bit node |
| `uuid v4` | RFC 9562 UUIDv4 (random) | 122-bit random |
| `uuid v6` | RFC 9562 UUIDv6 (reordered time) | sortable + 14-bit clock |
| `uuid v7` | RFC 9562 UUIDv7 (time-ordered) | 74-bit random |
| `uuid v8` | RFC 9562 UUIDv8 (custom payload) | 122-bit custom |
| `uuid nil` | Nil UUID (all zeros) | — |
| `uuid max` | Max UUID (all ones) | — |
| `ulid` | ULID (Crockford Base32, sortable) | 80-bit random |
| `nanoid` | NanoID (URL-safe, configurable) | ~126-bit (21 chars) |
| `snowflake` | Twitter-style 64-bit Snowflake ID | 12-bit sequence |
| `ssh ed25519` | OpenSSH Ed25519 keypair | 256-bit |
| `tls self-signed` | Self-signed X.509 leaf (Ed25519) | 256-bit |
| `wireguard keypair` | Curve25519 WireGuard keypair | 256-bit |
| `password` | Random password (configurable) | depends on length/charset |
| `totp` | TOTP secret + otpauth URI | 160-bit (SHA-1 block) |
| `mac` | Locally-administered unicast MAC | 46-bit random |

## Configuration

Smedje uses layered configuration (highest wins):

1. CLI flags
2. `--env-file` overlay
3. `SMEDJE_*` environment variables
4. `.smedje.toml` (project-local)
5. `~/.config/smedje/defaults.toml`
6. Built-in defaults

```bash
# Set user-level defaults
smedje config set password.length 32
smedje config set tls.days 730

# Show effective config with sources
smedje config show --explain
```

See [docs/configuration.md](docs/configuration.md) for full details.

## Documentation

- [docs/cli-conventions.md](docs/cli-conventions.md) — naming, addressing, flag patterns
- [docs/configuration.md](docs/configuration.md) — precedence, file locations, env naming
- [docs/defaults.md](docs/defaults.md) — all defaults with rationale
- [docs/output-formats.md](docs/output-formats.md) — format conventions per output mode
- [docs/recommendations.md](docs/recommendations.md) — opinionated guidance

## Install

```bash
go install github.com/smedje/smedje/cmd/smedje@latest
```

Requires Go 1.23+. Single static binary, no runtime dependencies.

## Shell completions

```bash
# Bash
source <(smedje completion bash)

# Zsh
smedje completion zsh > "${fpath[1]}/_smedje"

# Fish
smedje completion fish | source
```

## Why I built this

I got tired of context-switching between `ssh-keygen`, `openssl`, `wg`,
`uuidgen`, and half-remembered shell one-liners every time I set up a new
environment. Smedje puts all of that behind one command with consistent
output formatting and JSON support for scripting.

## Roadmap

- [ ] More key types (RSA, ECDSA)
- [ ] IPsec PSK generator
- [ ] JWT key generators
- [ ] TLS CA hierarchy
- [ ] `--paranoid` mode (additional entropy sources)
- [ ] Profile system (YAML-defined generation sequences)
- [ ] TUI mode (Bubble Tea)
- [ ] Web frontend with network design wizard
- [ ] Multi-vendor config rendering (VyOS, RouterOS, OPNsense)

## License

AGPL-3.0. See [LICENSE](LICENSE).
