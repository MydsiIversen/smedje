package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
)

const projectFileName = ".smedje.toml"
const userFileName = "defaults.toml"

// userConfigDir returns the platform-appropriate user config directory.
func userConfigDir() string {
	if runtime.GOOS == "windows" {
		if appdata := os.Getenv("APPDATA"); appdata != "" {
			return filepath.Join(appdata, "smedje")
		}
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "smedje")
}

// userConfigPath returns the full path to the user config file.
func userConfigPath() string {
	return filepath.Join(userConfigDir(), userFileName)
}

// findProjectConfig walks from cwd upward looking for .smedje.toml.
func findProjectConfig() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		candidate := filepath.Join(dir, projectFileName)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// tomlData is the intermediate structure for TOML parsing. We flatten the
// nested map into dotted keys.
type tomlData = map[string]any

// loadTOML reads a TOML file and flattens it into a dotted-key map.
func loadTOML(path string) (map[string]string, error) {
	var raw tomlData
	if _, err := toml.DecodeFile(path, &raw); err != nil {
		return nil, fmt.Errorf("config: parse %s: %w", path, err)
	}
	result := make(map[string]string)
	flatten("", raw, result)
	return result, nil
}

func flatten(prefix string, m map[string]any, out map[string]string) {
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		switch val := v.(type) {
		case map[string]any:
			flatten(key, val, out)
		default:
			out[key] = fmt.Sprintf("%v", val)
		}
	}
}

// writeTOML writes a flat config map as grouped TOML to path.
func writeTOML(path string, values map[string]string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	grouped := make(map[string]map[string]string)
	for k, v := range values {
		section, key, ok := splitKey(k)
		if !ok {
			section = ""
			key = k
		}
		if grouped[section] == nil {
			grouped[section] = make(map[string]string)
		}
		grouped[section][key] = v
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	for section, keys := range grouped {
		if section != "" {
			fmt.Fprintf(f, "[%s]\n", section)
		}
		for k, v := range keys {
			fmt.Fprintf(f, "%s = %q\n", k, v)
		}
		fmt.Fprintln(f)
	}
	return nil
}

func splitKey(key string) (section, name string, ok bool) {
	for i := len(key) - 1; i >= 0; i-- {
		if key[i] == '.' {
			return key[:i], key[i+1:], true
		}
	}
	return "", key, false
}
