package entropy

import (
	"bytes"
	"testing"
)

func TestSeededReaderDeterministic(t *testing.T) {
	r1 := NewSeededReader("test-seed")
	r2 := NewSeededReader("test-seed")

	buf1 := make([]byte, 1000)
	buf2 := make([]byte, 1000)

	r1.Read(buf1)
	r2.Read(buf2)

	if !bytes.Equal(buf1, buf2) {
		t.Fatal("same seed produced different output")
	}
}

func TestSeededReaderDifferentSeeds(t *testing.T) {
	r1 := NewSeededReader("seed-a")
	r2 := NewSeededReader("seed-b")

	buf1 := make([]byte, 32)
	buf2 := make([]byte, 32)

	r1.Read(buf1)
	r2.Read(buf2)

	if bytes.Equal(buf1, buf2) {
		t.Fatal("different seeds produced same output")
	}
}

func TestSetSeedAndReset(t *testing.T) {
	original := Reader

	SetSeed("my-seed")
	if Reader == original {
		t.Fatal("SetSeed did not change Reader")
	}

	Reset()
	if Reader != original {
		t.Fatal("Reset did not restore Reader")
	}
}
