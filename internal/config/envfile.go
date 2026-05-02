package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// loadEnvFile parses a .env file into a map of environment variable names to
// values. Lines starting with # are comments. Empty lines are skipped.
// Supports KEY=VALUE and KEY="VALUE" (strips outer quotes).
func loadEnvFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("config: open env file: %w", err)
	}
	defer f.Close()

	result := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		v = strings.Trim(v, `"'`)
		result[k] = v
	}
	return result, scanner.Err()
}

// loadEnvFileAsConfig parses a .env file and converts SMEDJE_* entries to
// config keys.
func loadEnvFileAsConfig(path string) (map[string]string, error) {
	raw, err := loadEnvFile(path)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for k, v := range raw {
		if strings.HasPrefix(k, envPrefix) {
			result[envToKey(k)] = v
		}
	}
	return result, nil
}
