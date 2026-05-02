# Security Policy

## Reporting a vulnerability

If you discover a security vulnerability in Smedje, please report it
responsibly. Do **not** open a public issue.

Email: mydsi@mydsi.cc

Include:
- Description of the vulnerability
- Steps to reproduce
- Impact assessment

You should receive a response within 72 hours.

## Scope

Smedje generates cryptographic material but does not store or manage it.
Security-relevant areas include:

- Entropy sources (must use crypto/rand)
- Key generation (must use well-vetted primitives)
- Output handling (must not leak keys to logs or stdout unintentionally)

## Supported versions

Only the latest release is supported with security fixes.
