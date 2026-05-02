# CLI conventions

## Generator addressing

Every generator has a dotted address in the form `<group>.<variant>`.
Single-variant generators are addressable by bare name alone.

### Multi-variant groups

These require the dotted form:

| Address | Generator |
|---------|-----------|
| uuid.v1 | UUIDv1 (time-based, random node) |
| uuid.v4 | UUIDv4 (random) |
| uuid.v6 | UUIDv6 (reordered time) |
| uuid.v7 | UUIDv7 (time-ordered, recommended) |
| uuid.v8 | UUIDv8 (custom payload) |
| uuid.nil | Nil UUID (all zeros) |
| uuid.max | Max UUID (all ones) |
| ssh.ed25519 | Ed25519 OpenSSH keypair |
| tls.self-signed | Self-signed TLS certificate |
| wireguard.keypair | WireGuard Curve25519 keypair |

Using a bare multi-variant group name (e.g., `uuid`) produces an error
listing all available variants.

### Single-variant generators

These work with or without the dotted form:

| Address | Generator |
|---------|-----------|
| ulid | ULID (Crockford Base32, time-sortable) |
| nanoid | NanoID (URL-safe, configurable) |
| snowflake | Snowflake ID (64-bit, time + worker + seq) |
| password | Random password |
| totp | TOTP secret and otpauth URI |
| mac | Random locally-administered MAC address |

### Where addressing applies

- `smedje bench <address>` — benchmark a single generator
- `smedje bench compare <addr1> <addr2> ...` — side-by-side comparison
- `smedje bench list` — print all addressable names

### Listing available generators

```
smedje bench list
```

Prints one address per line, sorted alphabetically.

## Naming

- Command and flag names use kebab-case: `self-signed`, `--no-color`.
- Generator groups match the top-level CLI command: `uuid`, `ssh`, `tls`,
  `wireguard`, `ulid`, `nanoid`, `snowflake`, `password`, `totp`, `mac`.
- Subcommand names match the generator variant: `v7`, `ed25519`,
  `self-signed`, `keypair`.

## Flag patterns

- `--format` selects output format: text (default), json, csv, sql, env.
- `--quiet` strips all decoration, outputs bare values only.
- `--json` is a shorthand for `--format json`.
- `--count N` generates N values in a single invocation.
- `--bench` runs an inline benchmark after generation.
