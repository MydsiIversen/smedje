// Package config provides layered configuration resolution for Smedje.
//
// Precedence (highest wins):
//  1. CLI flags
//  2. SMEDJE_* environment variables
//  3. .smedje.toml (project-local, walks up from cwd)
//  4. ~/.config/smedje/defaults.toml (user config)
//  5. Built-in defaults
//
// An optional --env-file PATH loads a .env file between #1 and #2.
package config

import (
	"fmt"
	"os"
)

// Source describes where a config value came from.
type Source int

const (
	SourceDefault Source = iota
	SourceUserConfig
	SourceProjectConfig
	SourceEnv
	SourceEnvFile
	SourceFlag
)

func (s Source) String() string {
	switch s {
	case SourceDefault:
		return "default"
	case SourceUserConfig:
		return "user-config"
	case SourceProjectConfig:
		return "project-config"
	case SourceEnv:
		return "env"
	case SourceEnvFile:
		return "env-file"
	case SourceFlag:
		return "flag"
	default:
		return "unknown"
	}
}

// Entry is a resolved config value with its source.
type Entry struct {
	Value  string
	Source Source
}

// Config holds the fully resolved configuration.
type Config struct {
	entries map[string]Entry
}

// Get returns the value for a key, or empty string if unset.
func (c *Config) Get(key string) string {
	if e, ok := c.entries[key]; ok {
		return e.Value
	}
	return ""
}

// GetEntry returns the full entry including source info.
func (c *Config) GetEntry(key string) (Entry, bool) {
	e, ok := c.entries[key]
	return e, ok
}

// All returns every resolved key-value pair.
func (c *Config) All() map[string]Entry {
	out := make(map[string]Entry, len(c.entries))
	for k, v := range c.entries {
		out[k] = v
	}
	return out
}

// Keys returns all config keys in no particular order.
func (c *Config) Keys() []string {
	keys := make([]string, 0, len(c.entries))
	for k := range c.entries {
		keys = append(keys, k)
	}
	return keys
}

// LoadOptions controls config loading behavior.
type LoadOptions struct {
	EnvFilePath string
	Flags       map[string]string
}

// Load resolves configuration using the full precedence chain.
func Load(opts LoadOptions) (*Config, error) {
	entries := make(map[string]Entry)

	// Layer 5: built-in defaults
	for k, v := range Defaults {
		entries[k] = Entry{Value: v, Source: SourceDefault}
	}

	// Layer 4: user config file
	userPath := userConfigPath()
	if fileExists(userPath) {
		vals, err := loadTOML(userPath)
		if err != nil {
			return nil, err
		}
		for k, v := range vals {
			entries[k] = Entry{Value: v, Source: SourceUserConfig}
		}
	}

	// Layer 3: project config file
	projPath := findProjectConfig()
	if projPath != "" {
		vals, err := loadTOML(projPath)
		if err != nil {
			return nil, err
		}
		for k, v := range vals {
			entries[k] = Entry{Value: v, Source: SourceProjectConfig}
		}
	}

	// Layer 2: environment variables
	for k, v := range loadEnv() {
		entries[k] = Entry{Value: v, Source: SourceEnv}
	}

	// Layer 1.5: env file (between env and flags)
	if opts.EnvFilePath != "" {
		vals, err := loadEnvFileAsConfig(opts.EnvFilePath)
		if err != nil {
			return nil, err
		}
		for k, v := range vals {
			entries[k] = Entry{Value: v, Source: SourceEnvFile}
		}
	}

	// Layer 1: CLI flags
	for k, v := range opts.Flags {
		entries[k] = Entry{Value: v, Source: SourceFlag}
	}

	return &Config{entries: entries}, nil
}

// Validate checks that all values in the config are valid.
func (c *Config) Validate() []error {
	var errs []error
	for k, e := range c.entries {
		if err := validateKey(k, e.Value); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func validateKey(key, value string) error {
	switch key {
	case "password.length":
		var n int
		fmt.Sscanf(value, "%d", &n)
		if n < 8 || n > 256 {
			return fmt.Errorf("config: %s must be 8-256, got %s", key, value)
		}
	case "totp.digits":
		if value != "6" && value != "8" {
			return fmt.Errorf("config: %s must be 6 or 8, got %s", key, value)
		}
	case "bulk.max-count":
		var n int
		fmt.Sscanf(value, "%d", &n)
		if n < 1 {
			return fmt.Errorf("config: %s must be positive, got %s", key, value)
		}
	}
	return nil
}

// UserConfigPath exposes the user config path for the init command.
func UserConfigPath() string {
	return userConfigPath()
}

// UserConfigDir exposes the user config directory.
func UserConfigDir() string {
	return userConfigDir()
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
