package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	// Import all generator packages so init() fires.
	_ "github.com/smedje/smedje/pkg/forge/id"
	_ "github.com/smedje/smedje/pkg/forge/network"
	_ "github.com/smedje/smedje/pkg/forge/secret"
	_ "github.com/smedje/smedje/pkg/forge/ssh"
	_ "github.com/smedje/smedje/pkg/forge/tls"
	_ "github.com/smedje/smedje/pkg/forge/wireguard"
)

func testServer() *Server {
	cfg := DefaultConfig()
	cfg.Version = "test"
	cfg.Commit = "abc123"
	return New(cfg)
}

func TestListGenerators(t *testing.T) {
	s := testServer()
	req := httptest.NewRequest(http.MethodGet, "/api/generators", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var infos []GeneratorInfo
	if err := json.NewDecoder(w.Body).Decode(&infos); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(infos) < 16 {
		t.Errorf("expected at least 16 generators, got %d", len(infos))
		for _, info := range infos {
			t.Logf("  %s", info.Address)
		}
	}
}

func TestGetGeneratorSchema(t *testing.T) {
	s := testServer()
	req := httptest.NewRequest(http.MethodGet, "/api/generators/uuid.v7", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var schema GeneratorSchema
	if err := json.NewDecoder(w.Body).Decode(&schema); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if schema.Name != "v7" {
		t.Errorf("expected name v7, got %q", schema.Name)
	}
	if schema.Group != "uuid" {
		t.Errorf("expected group uuid, got %q", schema.Group)
	}
	if schema.Address != "uuid.v7" {
		t.Errorf("expected address uuid.v7, got %q", schema.Address)
	}
	if schema.Category != "id" {
		t.Errorf("expected category id, got %q", schema.Category)
	}
	if len(schema.Flags) == 0 {
		t.Error("expected flags to be populated")
	}
	if !schema.Supports.Count {
		t.Error("expected count to be supported")
	}
	if !schema.Supports.Seed {
		t.Error("expected seed to be supported for uuid")
	}
	if !schema.Supports.Bench {
		t.Error("expected bench to be supported")
	}
	if schema.ExampleOutput == nil {
		t.Error("expected exampleOutput to be populated")
	}
}

func TestGetGeneratorNotFound(t *testing.T) {
	s := testServer()
	req := httptest.NewRequest(http.MethodGet, "/api/generators/nonexistent.gen", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestExplainUUIDv7(t *testing.T) {
	s := testServer()
	// Generate a UUIDv7 first, then explain it.
	body := `{"value":"019012a0-d8c7-7b2a-8e7f-1234567890ab"}`
	req := httptest.NewRequest(http.MethodPost, "/api/explain", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ExplainResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if resp.Detected == "" || resp.Detected == "unknown" {
		t.Errorf("expected detected format, got %q", resp.Detected)
	}
}

func TestVersionEndpoint(t *testing.T) {
	s := testServer()
	req := httptest.NewRequest(http.MethodGet, "/api/version", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var info VersionInfo
	if err := json.NewDecoder(w.Body).Decode(&info); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if info.Version != "test" {
		t.Errorf("expected version test, got %q", info.Version)
	}
	if info.Commit != "abc123" {
		t.Errorf("expected commit abc123, got %q", info.Commit)
	}
	if info.GoVersion == "" {
		t.Error("expected goVersion to be populated")
	}
}

func TestHealthz(t *testing.T) {
	s := testServer()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("expected status ok, got %q", resp["status"])
	}
}

func TestGenerateSingle(t *testing.T) {
	s := testServer()
	body := `{"generator":"uuid.v7","count":1}`
	req := httptest.NewRequest(http.MethodPost, "/api/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var artifact sseArtifact
	if err := json.NewDecoder(w.Body).Decode(&artifact); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if artifact.Value == "" {
		t.Error("expected non-empty value")
	}
}

func TestGenerateSSE(t *testing.T) {
	s := testServer()
	body := `{"generator":"uuid.v7","count":3}`
	req := httptest.NewRequest(http.MethodPost, "/api/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := w.Body.String()
	if !strings.Contains(resp, "event: artifact") {
		t.Error("expected SSE artifact events in response")
	}
	if !strings.Contains(resp, "event: done") {
		t.Error("expected SSE done event in response")
	}
}

func TestPublicModeMaxCount(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Public = true
	s := New(cfg)

	body := `{"generator":"uuid.v7","count":200}`
	req := httptest.NewRequest(http.MethodPost, "/api/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for count exceeding public max, got %d", w.Code)
	}
}

func TestCryptoGeneratorSeedFalse(t *testing.T) {
	s := testServer()
	req := httptest.NewRequest(http.MethodGet, "/api/generators/ssh.ed25519", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var schema GeneratorSchema
	if err := json.NewDecoder(w.Body).Decode(&schema); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if schema.Supports.Seed {
		t.Error("crypto generator ssh.ed25519 should not support seed")
	}
}

func TestPrivacyPage(t *testing.T) {
	s := testServer()
	req := httptest.NewRequest(http.MethodGet, "/privacy", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "text/html; charset=utf-8" {
		t.Errorf("expected text/html content type, got %q", ct)
	}

	body := w.Body.String()
	for _, want := range []string{
		"Privacy",
		"Your data stays in your browser",
		"analytics.smedje.net",
		"No cookies",
		"Do Not Track",
		"github.com/MydsiIversen/smedje",
		"Back to Smedje",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("privacy page missing expected text: %q", want)
		}
	}
}

func TestAnalyticsTagParsing(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   string
		wantOK bool
	}{
		{
			name:   "valid script URL",
			input:  "https://analytics.smedje.net/script.js?id=abc-123",
			want:   `<script defer data-website-id="abc-123" src="https://analytics.smedje.net/script.js"></script>`,
			wantOK: true,
		},
		{
			name:   "empty string",
			input:  "",
			want:   "",
			wantOK: false,
		},
		{
			name:   "no id param",
			input:  "https://analytics.smedje.net/script.js",
			want:   "",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.AnalyticsScript = tt.input
			s := &Server{cfg: cfg}

			got := s.analyticsTag()
			if tt.wantOK && got != tt.want {
				t.Errorf("analyticsTag() = %q, want %q", got, tt.want)
			}
			if !tt.wantOK && got != "" {
				t.Errorf("analyticsTag() = %q, want empty", got)
			}
		})
	}
}
