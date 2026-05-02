# Configuration

Smedje resolves configuration using a layered precedence system.

## Precedence (highest wins)

1. **CLI flags** — explicit flags passed to any command
2. **`--env-file PATH`** — optional .env file overlay
3. **`SMEDJE_*` environment variables** — flat snake_case from dotted paths
4. **`.smedje.toml`** — project-local config (walks up from cwd)
5. **`~/.config/smedje/defaults.toml`** — user-level defaults (Linux/macOS)
   `%APPDATA%\smedje\defaults.toml` on Windows
6. **Built-in defaults** — hardcoded in the binary

## Config file format

TOML with sections matching the generator category or feature area:

```toml
[password]
length = "32"
charset = "alphanum"

[tls]
cn = "dev.local"
days = "730"

[bulk]
max-count = "10000000"
```

## Environment variable naming

Dotted config keys map to environment variables by:
1. Uppercasing the entire key
2. Replacing `.` and `-` with `_`
3. Prepending `SMEDJE_`

Examples:
- `password.length` → `SMEDJE_PASSWORD_LENGTH`
- `tls.days` → `SMEDJE_TLS_DAYS`
- `bulk.max-count` → `SMEDJE_BULK_MAX_COUNT`

## .env file support

Pass `--env-file PATH` to load additional environment overrides from a
file. Only lines starting with `SMEDJE_` are used. Format:

```
SMEDJE_PASSWORD_LENGTH=48
SMEDJE_TLS_DAYS="730"
```

## Config commands

```
smedje config init       # Write commented defaults to user config dir
smedje config show       # Show all effective values
smedje config show --explain  # Show values with their source
smedje config get KEY    # Get one value
smedje config set KEY VALUE  # Write to user config file
smedje config validate   # Check all values are valid
```

## Project-local config

Place a `.smedje.toml` in your project root (or any parent directory).
Smedje walks up from cwd looking for it. This is useful for team-wide
defaults (e.g., always use 730-day TLS certs for this project).
