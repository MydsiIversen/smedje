package id

const whyUUIDv7 = `About UUIDv7 (RFC 9562, 2024):
  Time-ordered UUID with millisecond precision. Leading 48 bits are
  a Unix timestamp, followed by 4 version bits, 12 random bits, 2
  variant bits, and 62 random bits.

Why default:
  For new applications, v7 is generally preferred over v4. The time
  prefix improves database index performance dramatically (B-trees
  don't fragment) and the timestamp aids debugging.

Alternatives:
  uuid.v4   non-sortable, fully random — use if time leakage matters
  ulid      shorter (26 chars), Crockford base32, similar properties

See also: smedje recommend id
`

const whyUUIDv4 = `About UUIDv4 (RFC 9562, 2005):
  Random UUID with 122 bits from crypto/rand. No time component,
  no ordering guarantees.

Why use it:
  When time leakage is unacceptable or when you need compatibility
  with systems that only support v4. Also suitable when natural
  ordering is irrelevant (e.g., idempotency keys).

Preferred alternative:
  uuid.v7   time-ordered, better index performance — preferred for
            new applications unless time leakage is a concern

See also: smedje recommend id
`

const whyUUIDv1 = `About UUIDv1 (RFC 9562, 2005):
  Time-based UUID with 60-bit Gregorian timestamp (100-nanosecond
  intervals since 1582-10-15), 14-bit clock sequence, and 48-bit
  node ID. Smedje uses a random node with multicast bit set to avoid
  leaking the host MAC address.

Why use it:
  Legacy compatibility. v1 is widespread in existing systems.

Preferred alternatives:
  uuid.v7   better sort order (ms timestamp in leading bits)
  uuid.v6   v1 bits reordered for natural sort — drop-in if you
            control both producer and consumer

See also: smedje recommend id
`

const whyUUIDv6 = `About UUIDv6 (RFC 9562, 2024):
  Reordered variant of v1 — same 60-bit Gregorian timestamp, but
  the high bits come first so byte-order sorting matches time order.

Why use it:
  Migration path from v1 when you need byte-sortability but must
  keep the Gregorian timestamp semantics.

Preferred alternative:
  uuid.v7   simpler (Unix ms), widely adopted, better ecosystem support

See also: smedje recommend id
`

const whyUUIDv8 = `About UUIDv8 (RFC 9562, 2024):
  Custom-layout UUID. Only the version (4 bits) and variant (2 bits)
  are fixed; all other 122 bits are user-defined. Smedje fills them
  with crypto/rand when no explicit payload is given.

Why use it:
  When you need a custom encoding inside the UUID structure —
  embedding a shard key, a custom timestamp format, or other metadata
  while staying within the UUID type system.

Note:
  If you just need randomness, use v4. If you need time-ordering,
  use v7. v8 is for custom schemes that don't fit v1-v7.

See also: smedje recommend id
`

const whyUUIDNil = `About UUID nil (RFC 9562):
  The all-zeros UUID (00000000-0000-0000-0000-000000000000).
  Used as a sentinel value meaning "no UUID" or "unset."

When to use:
  Default/placeholder values in databases and APIs. Never use
  as a generated identifier — it's not unique by definition.
`

const whyUUIDMax = `About UUID max (RFC 9562):
  The all-ones UUID (ffffffff-ffff-ffff-ffff-ffffffffffff).
  Sorts after every other UUID in byte order.

When to use:
  Sentinel for "maximum possible" in range queries or as an
  end-of-range marker. Never use as a generated identifier.
`

const whyULID = `About ULID (github.com/ulid/spec):
  26-character Crockford Base32 identifier. 48-bit millisecond
  timestamp prefix + 80 bits of randomness. Lexicographically
  sortable, URL-safe, case-insensitive.

Why use it:
  When you need time-sortable IDs that are shorter than UUIDs (26
  chars vs 36) and case-insensitive. Popular in systems that pass
  IDs in URLs or where the shorter representation matters.

Alternatives:
  uuid.v7     standard UUID format, wider ecosystem support
  nanoid      even shorter if you don't need time-sorting

See also: smedje recommend id
`

const whyNanoID = `About NanoID:
  URL-safe random identifier with configurable length and alphabet.
  Default: 21 characters from A-Za-z0-9_- providing ~126 bits of
  entropy. Uses crypto/rand with rejection sampling for uniform
  distribution.

Why use it:
  When you need short, URL-safe, random identifiers. Good for
  user-facing IDs, session tokens, and URL slugs. The custom
  alphabet option lets you exclude ambiguous characters.

Alternatives:
  uuid.v4     128-bit, standard format, wider recognition
  ulid        time-sortable if ordering matters

See also: smedje recommend id
`

const whySnowflake = `About Snowflake (Twitter, 2010):
  64-bit integer ID: 41 bits timestamp (ms since epoch) + 10 bits
  worker ID + 12 bits sequence. Fits in a signed int64. Ordered
  within the same worker. Epoch: 2024-01-01T00:00:00Z.

Why use it:
  When you need compact 64-bit integer IDs that embed a timestamp.
  Natural fit for databases with int64 primary keys and distributed
  systems with worker-based sharding.

Alternatives:
  uuid.v7     128-bit, standard UUID format, no worker coordination
  ulid        128-bit, no worker ID needed, base32 encoded

See also: smedje recommend id
`

func (u *UUIDv7) Why() string    { return whyUUIDv7 }
func (u *UUIDv4) Why() string    { return whyUUIDv4 }
func (u *UUIDv1) Why() string    { return whyUUIDv1 }
func (u *UUIDv6) Why() string    { return whyUUIDv6 }
func (u *UUIDv8) Why() string    { return whyUUIDv8 }
func (u *UUIDNil) Why() string   { return whyUUIDNil }
func (u *UUIDMax) Why() string   { return whyUUIDMax }
func (u *ULID) Why() string      { return whyULID }
func (n *NanoID) Why() string    { return whyNanoID }
func (s *Snowflake) Why() string { return whySnowflake }
