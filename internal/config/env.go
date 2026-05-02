package config

import (
	"os"
	"strings"
)

// envPrefix is the prefix for all Smedje environment variables.
const envPrefix = "SMEDJE_"

// keyToEnv converts a dotted config key to an environment variable name.
// Example: "tls.validity" → "SMEDJE_TLS_VALIDITY"
func keyToEnv(key string) string {
	s := strings.ToUpper(key)
	s = strings.ReplaceAll(s, ".", "_")
	s = strings.ReplaceAll(s, "-", "_")
	return envPrefix + s
}

// envToKey converts an environment variable name back to a dotted config key.
// Example: "SMEDJE_TLS_DAYS" → "tls.days"
// This is lossy: we can't distinguish "." from "-" in the original. The
// canonical form uses dots between sections and hyphens within a key name.
func envToKey(env string) string {
	s := strings.TrimPrefix(env, envPrefix)
	s = strings.ToLower(s)
	// First underscore after a known section prefix is the dot separator.
	// For simplicity, we treat the first segment as the section.
	parts := strings.SplitN(s, "_", 2)
	if len(parts) == 2 {
		return parts[0] + "." + strings.ReplaceAll(parts[1], "_", "-")
	}
	return s
}

// loadEnv reads all SMEDJE_* environment variables into a map.
func loadEnv() map[string]string {
	result := make(map[string]string)
	for _, env := range os.Environ() {
		k, v, ok := strings.Cut(env, "=")
		if !ok {
			continue
		}
		if strings.HasPrefix(k, envPrefix) {
			result[envToKey(k)] = v
		}
	}
	return result
}
