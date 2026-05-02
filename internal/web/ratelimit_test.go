package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimiterAllowsBurst(t *testing.T) {
	rl := newRateLimiter()
	ip := "192.0.2.1"

	for i := 0; i < 60; i++ {
		if !rl.allow(ip) {
			t.Fatalf("request %d should have been allowed", i+1)
		}
	}
}

func TestRateLimiterDeniesAfterBurst(t *testing.T) {
	rl := newRateLimiter()
	ip := "192.0.2.1"

	// Exhaust the burst.
	for i := 0; i < 60; i++ {
		rl.allow(ip)
	}

	if rl.allow(ip) {
		t.Fatal("61st request should have been denied")
	}
}

func TestRateLimiterSeparateIPs(t *testing.T) {
	rl := newRateLimiter()

	// Exhaust burst for one IP.
	for i := 0; i < 60; i++ {
		rl.allow("192.0.2.1")
	}

	// A different IP should still be allowed.
	if !rl.allow("192.0.2.2") {
		t.Fatal("different IP should not be rate limited")
	}
}

func TestExtractIPFromXForwardedFor(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Forwarded-For", "203.0.113.50, 70.41.3.18, 150.172.238.178")

	ip := extractIP(r)
	if ip != "203.0.113.50" {
		t.Errorf("expected 203.0.113.50, got %q", ip)
	}
}

func TestExtractIPFromXForwardedForSingle(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Forwarded-For", "203.0.113.50")

	ip := extractIP(r)
	if ip != "203.0.113.50" {
		t.Errorf("expected 203.0.113.50, got %q", ip)
	}
}

func TestExtractIPFromRemoteAddr(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "192.0.2.1:12345"

	ip := extractIP(r)
	if ip != "192.0.2.1" {
		t.Errorf("expected 192.0.2.1, got %q", ip)
	}
}

func TestRateLimitMiddleware429(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Public = true
	s := New(cfg)

	handler := s.Handler()

	// Exhaust the burst.
	for i := 0; i < 60; i++ {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		req.RemoteAddr = "192.0.2.1:12345"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, w.Code)
		}
	}

	// Next request should be 429.
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.RemoteAddr = "192.0.2.1:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["error"] != "rate_limited" {
		t.Errorf("expected error rate_limited, got %q", resp["error"])
	}
	if w.Header().Get("Retry-After") != "30" {
		t.Errorf("expected Retry-After 30, got %q", w.Header().Get("Retry-After"))
	}
}

func TestNoRateLimitWhenNotPublic(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Public = false
	s := New(cfg)

	handler := s.Handler()

	// Send 70 requests -- all should succeed since rate limiting is off.
	for i := 0; i < 70; i++ {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		req.RemoteAddr = "192.0.2.1:12345"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, w.Code)
		}
	}
}
