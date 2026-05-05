// Package entropy wraps crypto/rand with a hook point for a future
// --paranoid mode that could mix in additional entropy sources.
package entropy

import (
	"crypto/rand"
	"io"
)

// Reader is the entropy source used by all generators. It defaults to
// crypto/rand.Reader and exists so a future --paranoid mode can wrap it.
// TODO: --paranoid flag for additional entropy mixing (deferred until real demand).
var Reader io.Reader = rand.Reader

// Read fills b with cryptographically secure random bytes.
func Read(b []byte) (int, error) {
	return io.ReadFull(Reader, b)
}
