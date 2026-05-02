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
smedje snowflake               # 309052152221794304
smedje ssh ed25519             # OpenSSH keypair to stdout
smedje tls self-signed         # self-signed cert + key
smedje wireguard keypair       # Curve25519 keypair (base64)
smedje password                # DJa}#fM%$7e(5|wyL2hl]*28
smedje totp                    # TOTP secret + otpauth URI
smedje mac                     # 42:88:8a:b6:fe:3c

# Output modes
smedje uuid v7 --json          # {"value": "..."}
smedje uuid v7 --quiet         # raw value only
smedje uuid v7 --bench         # self-benchmark
```

## What it forges

| Command | What | Key bits |
|---|---|---|
| `uuid v7` | RFC 9562 UUIDv7 (time-ordered) | 74-bit random |
| `snowflake` | Twitter-style 64-bit Snowflake ID | 12-bit sequence |
| `ssh ed25519` | OpenSSH Ed25519 keypair | 256-bit |
| `tls self-signed` | Self-signed X.509 leaf (Ed25519) | 256-bit |
| `wireguard keypair` | Curve25519 WireGuard keypair | 256-bit |
| `password` | Random password (configurable) | depends on length/charset |
| `totp` | TOTP secret + otpauth URI | 160-bit (SHA-1 block) |
| `mac` | Locally-administered unicast MAC | 46-bit random |

## Compared to

| Need | Before | Smedje |
|---|---|---|
| UUIDv7 | `uuidgen` (no v7 support) | `smedje uuid v7` |
| Snowflake ID | custom script | `smedje snowflake` |
| SSH key | `ssh-keygen -t ed25519` | `smedje ssh ed25519` |
| TLS cert | `openssl req -x509 ...` | `smedje tls self-signed` |
| WireGuard key | `wg genkey \| tee priv \| wg pubkey` | `smedje wireguard keypair` |
| Password | `openssl rand -base64 24` | `smedje password` |
| TOTP secret | online generator | `smedje totp` |
| MAC address | `printf` one-liner | `smedje mac` |

## Install

```bash
go install github.com/smedje/smedje/cmd/smedje@latest
```

Requires Go 1.23+. Single static binary, no runtime dependencies.

## Why I built this

I got tired of context-switching between `ssh-keygen`, `openssl`, `wg`,
`uuidgen`, and half-remembered shell one-liners every time I set up a new
environment. Smedje puts all of that behind one command with consistent
output formatting and JSON support for scripting.

## Roadmap

- [ ] More key types (RSA, ECDSA)
- [ ] IPsec PSK generator
- [ ] `--paranoid` mode (additional entropy sources)
- [ ] Profile system (YAML-defined generation sequences)
- [ ] TUI mode (Bubble Tea)
- [ ] Web frontend with network design wizard
- [ ] Multi-vendor config rendering (VyOS, RouterOS, OPNsense)

## License

AGPL-3.0. See [LICENSE](LICENSE).
