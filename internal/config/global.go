package config

import "sync"

var (
	globalMu  sync.RWMutex
	globalCfg *Config
)

// SetGlobal sets the globally accessible config instance.
// Called once during CLI initialization.
func SetGlobal(c *Config) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalCfg = c
}

// Global returns the global config instance.
// Returns nil if not yet initialized (library usage without CLI).
func Global() *Config {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalCfg
}

// GetDefault returns the effective config value for a key, falling back to
// the built-in default if no config is loaded.
func GetDefault(key string) string {
	if c := Global(); c != nil {
		if v := c.Get(key); v != "" {
			return v
		}
	}
	if v, ok := Defaults[key]; ok {
		return v
	}
	return ""
}
