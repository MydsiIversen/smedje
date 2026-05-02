// Package explain provides format detection for identifiers and secrets.
// Each ID generator can register a Detector that pattern-matches its format
// and decodes embedded parts (e.g., timestamps).
package explain

import (
	"fmt"
	"sync"
)

// Match describes a successful format detection.
type Match struct {
	Format     string
	Confidence float64
	Fields     map[string]string
}

// Detector identifies and decodes a specific format.
type Detector interface {
	Name() string
	Detect(input string) (Match, bool)
}

var (
	mu        sync.RWMutex
	detectors []Detector
)

// Register adds a detector to the chain.
func Register(d Detector) {
	mu.Lock()
	defer mu.Unlock()
	detectors = append(detectors, d)
}

// Identify runs all registered detectors against input and returns the best
// match (highest confidence). Returns nil if nothing matches.
func Identify(input string) *Match {
	mu.RLock()
	defer mu.RUnlock()

	var best *Match
	for _, d := range detectors {
		if m, ok := d.Detect(input); ok {
			if best == nil || m.Confidence > best.Confidence {
				mc := m
				best = &mc
			}
		}
	}
	return best
}

// FormatResult produces a human-readable summary of a detection match.
func FormatResult(m *Match) string {
	if m == nil {
		return "Unknown format"
	}
	s := fmt.Sprintf("Format: %s\n", m.Format)
	for k, v := range m.Fields {
		s += fmt.Sprintf("  %s: %s\n", k, v)
	}
	return s
}
