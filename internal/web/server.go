package web

import (
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

// ServerConfig holds the configuration for the HTTP server.
type ServerConfig struct {
	Port            int
	Host            string
	Dev             bool
	Public          bool
	NoBrowser       bool
	AnalyticsScript string
	MaxCount        int
	RequestTimeout  time.Duration
	Version         string
	Commit          string
}

// DefaultConfig returns a ServerConfig with sensible defaults for normal
// (non-public) operation.
func DefaultConfig() ServerConfig {
	return ServerConfig{
		Port:           8080,
		Host:           "127.0.0.1",
		MaxCount:       100_000_000,
		RequestTimeout: 60 * time.Second,
		Version:        "dev",
		Commit:         "none",
	}
}

// Server is the Smedje HTTP server.
type Server struct {
	cfg     ServerConfig
	mux     *http.ServeMux
	limiter *rateLimiter
}

// New creates a new Server with the given configuration and registers all
// routes.
func New(cfg ServerConfig) *Server {
	if cfg.Public {
		if cfg.MaxCount == 0 || cfg.MaxCount > 100 {
			cfg.MaxCount = 100
		}
		if cfg.RequestTimeout == 0 || cfg.RequestTimeout > 5*time.Second {
			cfg.RequestTimeout = 5 * time.Second
		}
	}
	if cfg.MaxCount == 0 {
		cfg.MaxCount = 100_000_000
	}
	if cfg.RequestTimeout == 0 {
		cfg.RequestTimeout = 60 * time.Second
	}

	s := &Server{cfg: cfg, mux: http.NewServeMux()}
	if cfg.Public {
		s.limiter = newRateLimiter()
	}
	s.registerRoutes()
	return s
}

// registerRoutes wires API handlers and static file serving.
func (s *Server) registerRoutes() {
	// API routes.
	s.mux.HandleFunc("GET /api/generators", s.handleListGenerators)
	s.mux.HandleFunc("GET /api/generators/{name...}", s.handleGetGenerator)
	s.mux.HandleFunc("POST /api/generate", s.handleGenerate)
	s.mux.HandleFunc("POST /api/explain", s.handleExplain)
	s.mux.HandleFunc("GET /api/recommend", s.handleRecommend)
	s.mux.HandleFunc("POST /api/bench", s.handleBench)
	s.mux.HandleFunc("GET /api/version", s.handleVersion)
	s.mux.HandleFunc("GET /healthz", s.handleHealthz)
	s.mux.HandleFunc("GET /privacy", s.handlePrivacy)

	// Static files or dev proxy for everything else.
	if s.cfg.Dev {
		s.mux.Handle("/", s.devProxy())
	} else {
		s.mux.Handle("/", s.embeddedFS())
	}
}

// Handler returns the root http.Handler with middleware applied.
func (s *Server) Handler() http.Handler {
	var h http.Handler = s.mux
	h = s.apiContentType(h)
	if s.cfg.Dev {
		h = s.corsMiddleware(h)
	}
	if s.limiter != nil {
		h = s.rateLimitMiddleware(h)
	}
	return h
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      s.Handler(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: s.cfg.RequestTimeout + 5*time.Second,
		IdleTimeout:  120 * time.Second,
	}
	return srv.ListenAndServe()
}

// Addr returns the listen address string.
func (s *Server) Addr() string {
	return fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
}

// corsMiddleware adds permissive CORS headers in dev mode.
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// apiContentType sets Content-Type to application/json for API routes.
func (s *Server) apiContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Content-Type", "application/json")
		}
		next.ServeHTTP(w, r)
	})
}

// devProxy returns a reverse proxy to the Vite dev server at localhost:5173.
func (s *Server) devProxy() http.Handler {
	target, _ := url.Parse("http://localhost:5173")
	return httputil.NewSingleHostReverseProxy(target)
}

// embeddedFS serves static files from the embedded dist directory with SPA
// fallback: if a file is not found, serve index.html.
func (s *Server) embeddedFS() http.Handler {
	sub, err := fs.Sub(frontendFS, "dist")
	if err != nil {
		// Should never happen with a valid embed directive.
		panic("web: embedded dist fs: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(sub))

	// Pre-build the injected index.html once at startup if analytics is configured.
	var injectedIndex []byte
	if tag := s.analyticsTag(); tag != "" {
		raw, err := fs.ReadFile(sub, "index.html")
		if err == nil {
			injectedIndex = []byte(strings.Replace(string(raw), "</head>", tag+"\n</head>", 1))
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to open the file. If it doesn't exist, serve index.html
		// for SPA client-side routing.
		path := r.URL.Path
		if path == "/" {
			path = "index.html"
		} else {
			path = strings.TrimPrefix(path, "/")
		}

		if _, err := fs.Stat(sub, path); err != nil {
			// SPA fallback: serve index.html.
			path = "index.html"
			r.URL.Path = "/"
		}

		// Serve the injected index.html if analytics is configured.
		if path == "index.html" && injectedIndex != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(injectedIndex)
			return
		}

		fileServer.ServeHTTP(w, r)
	})
}

// analyticsTag parses the --analytics-script flag value and returns a
// <script> tag suitable for injection before </head>. Returns empty
// string if the flag is unset or malformed.
func (s *Server) analyticsTag() string {
	raw := s.cfg.AnalyticsScript
	if raw == "" {
		return ""
	}

	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}

	websiteID := u.Query().Get("id")
	if websiteID == "" {
		return ""
	}

	// Reconstruct the script src without the id query param.
	q := u.Query()
	q.Del("id")
	u.RawQuery = q.Encode()
	src := u.String()

	return `<script defer data-website-id="` + websiteID + `" src="` + src + `"></script>`
}
