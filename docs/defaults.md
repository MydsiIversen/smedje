# Configurable Defaults

Every default below can be overridden via config file, environment
variable, or CLI flag.

| Key | Default | Rationale |
|-----|---------|-----------|
| `password.length` | `24` | OWASP minimum is 8; 24 balances security and usability for most use cases |
| `password.charset` | `full` | Includes upper, lower, digits, symbols. Maximum entropy per character |
| `totp.issuer` | `Smedje` | Placeholder; users should override per-project |
| `totp.account` | `user@example.com` | Placeholder; users should override per-project |
| `totp.digits` | `6` | Industry standard per RFC 6238; most authenticator apps expect 6 |
| `totp.period` | `30` | 30-second window is the de facto standard |
| `tls.cn` | `localhost` | Safe default for local dev; override for real certs |
| `tls.days` | `365` | One year; matches Let's Encrypt renewal mental model |
| `snowflake.worker` | `0` | Single-instance default; set per-machine in production |
| `mac.format` | `colon` | Most common human-readable format (aa:bb:cc:dd:ee:ff) |
| `nanoid.length` | `21` | Standard NanoID length providing ~126 bits of entropy |
| `bulk.max-count` | `100000000` | Safety cap to prevent accidental OOM; override with config |
| `bench.duration` | `2s` | Long enough for stable results without blocking interactive use |
| `bench.warmup` | `500ms` | Warm CPU caches before measuring |
| `bench.repeat` | `1` | Single run by default; increase for variance analysis |
| `bench.cores` | `0` | 0 means use all available cores (runtime.NumCPU()) |
