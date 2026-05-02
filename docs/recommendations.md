# Recommendations

Smedje includes opinionated recommendations for common cryptographic
and identity decisions. Access them via:

```
smedje recommend <topic> [--use-case <case>] [--json] [--markdown]
```

## Available topics

| Topic | Coverage |
|-------|----------|
| id | UUIDs, ULIDs, NanoIDs, Snowflakes — which format for which job |
| ssh-key | Key type selection (Ed25519 vs RSA) |
| tls-cert | Self-signed, internal CA, public trust |
| password | Length, charset, and format by context |
| hash | Password hashing and general-purpose integrity |
| jwt | Algorithm selection by use case |
| secret | TOTP, API keys, PSKs |
| vpn-key | WireGuard and IPsec keying |

## Philosophy

Recommendations are concrete and actionable: each includes the exact
`smedje` command to run. Where the recommended tool isn't built yet,
the output says "(planned for v0.X)" with the nearest available
alternative.

## What we don't take a position on

- Secrets management strategy (vault vs. env vars vs. sealed secrets)
- Certificate authority software selection (step-ca vs. CFSSL vs. Vault PKI)
- Password manager choice
- Cloud provider KMS selection
- Token format for your specific auth architecture

These depend on organizational context that a CLI tool can't assess.
Smedje generates the raw material; you decide where to store and
how to distribute it.
