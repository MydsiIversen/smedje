# Output formats

Every generator supports multiple output formats via `--format` or
shorthand flags.

## Format selection

```
smedje uuid v7 --format json
smedje uuid v7 --json          # shorthand
smedje uuid v7 --format csv
smedje uuid v7 --format sql
smedje uuid v7 --format env
smedje uuid v7 -q              # quiet (bare values)
```

## Text (default)

- Single value: bare value, one line, no labels
- Multi-value (`--count N`): one per line, no labels
- Multi-field generators (ssh, tls, wireguard): labeled fields

```
$ smedje uuid v7
019dea5d-ed92-778d-8666-cdb34f82e8b3

$ smedje ssh ed25519
private-key: -----BEGIN OPENSSH PRIVATE KEY-----
...
public-key: ssh-ed25519 AAAA...
```

## JSON

- Single value: `{"value": "..."}`
- Multi-value: array of objects `[{"value": "..."}, ...]`
- Multi-field: named fields `{"private_key": "...", "public_key": "..."}`

## CSV

- Header row uses the generator's group name for single-value generators
  (e.g., "uuid", "password"), field keys for multi-field generators.
- RFC 4180 quoting: values containing commas, quotes, or newlines are
  double-quoted with internal quotes escaped as `""`.
- Multi-field generators produce multiple columns.

```
$ smedje uuid v7 --format csv --count 3
uuid
019dea5d-ed92-778d-8666-cdb34f82e8b3
019dea5d-ed92-7a01-9c4a-1234abcd5678
019dea5d-ed92-7b02-8e5b-9876fedc3210
```

## SQL

- Produces a single INSERT statement with all values.
- Default table name is the generator's group name (e.g., "uuid",
  "password").
- Override with `--sql-table NAME`.
- Multi-field generators produce multi-column INSERTs.

```
$ smedje uuid v7 --format sql --count 3
INSERT INTO uuid (value) VALUES
  ('019dea5d-ed92-778d-8666-cdb34f82e8b3'),
  ('019dea5d-ed92-7a01-9c4a-1234abcd5678'),
  ('019dea5d-ed92-7b02-8e5b-9876fedc3210');
```

## Env

- Keys use the generator group name as prefix, uppercased.
- Format: `GROUP_N_FIELD=value`
- Multi-field generators produce multiple lines per item.

```
$ smedje uuid v7 --format env --count 2
UUID_1_VALUE=019dea5d-ed92-778d-8666-cdb34f82e8b3
UUID_2_VALUE=019dea5d-ed92-7a01-9c4a-1234abcd5678
```

## Quiet (`-q` / `--quiet`)

- Bare values only, one per line.
- No labels, no formatting.
- Multi-field generators emit each field value on a separate line.
