package entropy

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"io"
	"sync"
)

// SeededReader produces deterministic bytes from an HMAC-SHA256 PRG keyed by
// a seed string. NOT cryptographically secure — use only for reproducible
// test/demo output.
type SeededReader struct {
	key     []byte
	counter uint64
	buf     []byte
	mu      sync.Mutex
}

// NewSeededReader creates a deterministic reader keyed by seed.
func NewSeededReader(seed string) *SeededReader {
	h := hmac.New(sha256.New, []byte("smedje-prg-key"))
	h.Write([]byte(seed))
	return &SeededReader{key: h.Sum(nil)}
}

func (s *SeededReader) Read(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	n := 0
	for n < len(p) {
		if len(s.buf) == 0 {
			s.buf = s.nextBlock()
		}
		copied := copy(p[n:], s.buf)
		s.buf = s.buf[copied:]
		n += copied
	}
	return n, nil
}

func (s *SeededReader) nextBlock() []byte {
	h := hmac.New(sha256.New, s.key)
	var ctr [8]byte
	binary.BigEndian.PutUint64(ctr[:], s.counter)
	h.Write(ctr[:])
	s.counter++
	return h.Sum(nil)
}

// SetSeed replaces the global Reader with a deterministic SeededReader.
// Call Reset() to restore crypto/rand. This is safe for single-threaded
// CLI use only.
func SetSeed(seed string) {
	Reader = NewSeededReader(seed)
}

// Reset restores the global Reader to crypto/rand.
func Reset() {
	Reader = defaultReader
}

var defaultReader io.Reader

func init() {
	defaultReader = Reader
}
