package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestKeyToEnv(t *testing.T) {
	tests := []struct {
		key  string
		want string
	}{
		{"tls.days", "SMEDJE_TLS_DAYS"},
		{"password.length", "SMEDJE_PASSWORD_LENGTH"},
		{"bulk.max-count", "SMEDJE_BULK_MAX_COUNT"},
	}
	for _, tt := range tests {
		got := keyToEnv(tt.key)
		if got != tt.want {
			t.Errorf("keyToEnv(%q) = %q, want %q", tt.key, got, tt.want)
		}
	}
}

func TestEnvToKey(t *testing.T) {
	tests := []struct {
		env  string
		want string
	}{
		{"SMEDJE_TLS_DAYS", "tls.days"},
		{"SMEDJE_PASSWORD_LENGTH", "password.length"},
		{"SMEDJE_BULK_MAX_COUNT", "bulk.max-count"},
	}
	for _, tt := range tests {
		got := envToKey(tt.env)
		if got != tt.want {
			t.Errorf("envToKey(%q) = %q, want %q", tt.env, got, tt.want)
		}
	}
}

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load(LoadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if v := cfg.Get("password.length"); v != "24" {
		t.Errorf("password.length = %q, want %q", v, "24")
	}
	e, ok := cfg.GetEntry("password.length")
	if !ok {
		t.Fatal("password.length not found")
	}
	if e.Source != SourceDefault {
		t.Errorf("source = %v, want SourceDefault", e.Source)
	}
}

func TestLoadEnvOverridesDefault(t *testing.T) {
	t.Setenv("SMEDJE_PASSWORD_LENGTH", "32")
	cfg, err := Load(LoadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if v := cfg.Get("password.length"); v != "32" {
		t.Errorf("password.length = %q, want %q", v, "32")
	}
	e, _ := cfg.GetEntry("password.length")
	if e.Source != SourceEnv {
		t.Errorf("source = %v, want SourceEnv", e.Source)
	}
}

func TestLoadFlagsOverrideAll(t *testing.T) {
	t.Setenv("SMEDJE_PASSWORD_LENGTH", "32")
	cfg, err := Load(LoadOptions{
		Flags: map[string]string{"password.length": "48"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if v := cfg.Get("password.length"); v != "48" {
		t.Errorf("password.length = %q, want %q", v, "48")
	}
	e, _ := cfg.GetEntry("password.length")
	if e.Source != SourceFlag {
		t.Errorf("source = %v, want SourceFlag", e.Source)
	}
}

func TestLoadTOML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.toml")
	os.WriteFile(path, []byte(`
[password]
length = "16"
charset = "alphanum"
`), 0o644)

	vals, err := loadTOML(path)
	if err != nil {
		t.Fatal(err)
	}
	if vals["password.length"] != "16" {
		t.Errorf("password.length = %q, want %q", vals["password.length"], "16")
	}
	if vals["password.charset"] != "alphanum" {
		t.Errorf("password.charset = %q, want %q", vals["password.charset"], "alphanum")
	}
}

func TestLoadEnvFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	os.WriteFile(path, []byte(`
# comment
SMEDJE_TLS_DAYS=730
SMEDJE_PASSWORD_LENGTH="48"
UNRELATED=foo
`), 0o644)

	vals, err := loadEnvFileAsConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if vals["tls.days"] != "730" {
		t.Errorf("tls.days = %q, want %q", vals["tls.days"], "730")
	}
	if vals["password.length"] != "48" {
		t.Errorf("password.length = %q, want %q", vals["password.length"], "48")
	}
	if _, ok := vals["unrelated"]; ok {
		t.Error("non-SMEDJE_ key should be excluded")
	}
}

func TestValidate(t *testing.T) {
	cfg, _ := Load(LoadOptions{
		Flags: map[string]string{"password.length": "3"},
	})
	errs := cfg.Validate()
	if len(errs) == 0 {
		t.Error("expected validation error for length=3")
	}
}
