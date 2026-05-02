package web

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"github.com/smedje/smedje/internal/explain"
	"github.com/smedje/smedje/pkg/forge"
)

// handleListGenerators returns all registered generators.
func (s *Server) handleListGenerators(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, listGenerators())
}

// handleGetGenerator returns the full schema for a single generator.
func (s *Server) handleGetGenerator(w http.ResponseWriter, r *http.Request) {
	addr := r.PathValue("name")
	if addr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "generator address required",
		})
		return
	}

	g, err := resolveGenerator(addr)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, buildGeneratorSchema(g))
}

// handleGenerate runs a generator and returns the results. For count > 1,
// it streams results as SSE events.
func (s *Server) handleGenerate(w http.ResponseWriter, r *http.Request) {
	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body: " + err.Error(),
		})
		return
	}

	if req.Generator == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "generator address is required",
		})
		return
	}

	g, err := resolveGenerator(req.Generator)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": err.Error(),
		})
		return
	}

	count := req.Count
	if count < 1 {
		count = 1
	}
	if count > s.cfg.MaxCount {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "count exceeds maximum",
		})
		return
	}

	opts := forge.Options{
		Count:  1,
		Format: req.Format,
		Params: req.Params,
	}

	// Single value: return JSON directly.
	if count == 1 {
		out, err := g.Generate(r.Context(), opts)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
			return
		}
		writeJSON(w, http.StatusOK, outputToArtifact(out))
		return
	}

	// Multiple values: stream as SSE.
	sse := newSSEWriter(w)
	if sse == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "streaming not supported",
		})
		return
	}

	_ = sse.writeEvent("status", sseStatus{
		Message: "generating " + req.Generator,
	})

	start := time.Now()
	for i := 0; i < count; i++ {
		select {
		case <-r.Context().Done():
			return
		default:
		}

		out, err := g.Generate(r.Context(), opts)
		if err != nil {
			_ = sse.writeEvent("error", sseError{Message: err.Error()})
			return
		}

		_ = sse.writeEvent("artifact", outputToArtifact(out))

		// Send progress every 100 items or on the last item.
		if (i+1)%100 == 0 || i == count-1 {
			elapsed := time.Since(start)
			var opsPerSec float64
			if elapsed > 0 {
				opsPerSec = float64(i+1) / elapsed.Seconds()
			}
			_ = sse.writeEvent("progress", sseProgress{
				Current:   i + 1,
				Total:     count,
				OpsPerSec: opsPerSec,
			})
		}
	}

	_ = sse.writeEvent("done", sseDone{
		DurationMs: time.Since(start).Milliseconds(),
		Count:      count,
	})
}

// handleExplain identifies and decodes a value.
func (s *Server) handleExplain(w http.ResponseWriter, r *http.Request) {
	var req ExplainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body: " + err.Error(),
		})
		return
	}

	if req.Value == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "value is required",
		})
		return
	}

	m := explain.Identify(req.Value)
	if m == nil {
		writeJSON(w, http.StatusOK, ExplainResponse{
			Detected: "unknown",
			Fields:   map[string]string{},
		})
		return
	}

	writeJSON(w, http.StatusOK, ExplainResponse{
		Detected: m.Format,
		Fields:   m.Fields,
	})
}

// handleRecommend returns placeholder recommendations. Task 6 will extract
// the real recommendation data from cmd/smedje/recommend.go.
func (s *Server) handleRecommend(w http.ResponseWriter, r *http.Request) {
	topic := r.URL.Query().Get("topic")
	if topic == "" {
		writeJSON(w, http.StatusOK, map[string][]string{
			"topics": {"id", "ssh-key", "tls-cert", "password", "hash", "jwt", "secret", "vpn-key"},
		})
		return
	}

	// Placeholder: return an empty array until Task 6 wires real data.
	writeJSON(w, http.StatusOK, []struct{}{})
}

// handleBench runs a generator benchmark.
func (s *Server) handleBench(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Generator string `json:"generator"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body: " + err.Error(),
		})
		return
	}

	if req.Generator == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "generator address is required",
		})
		return
	}

	g, err := resolveGenerator(req.Generator)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": err.Error(),
		})
		return
	}

	result, err := g.Bench(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// handleVersion returns build info and public mode status.
func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, VersionInfo{
		Version:    s.cfg.Version,
		Commit:     s.cfg.Commit,
		GoVersion:  runtime.Version(),
		PublicMode: s.cfg.Public,
	})
}

// handleHealthz is a simple health check endpoint.
func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handlePrivacy is a placeholder for the privacy page (Task 4).
func (s *Server) handlePrivacy(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Privacy policy placeholder. Nothing is collected.",
	})
}

// outputToArtifact converts a forge.Output to an sseArtifact.
func outputToArtifact(out *forge.Output) sseArtifact {
	fields := make(map[string]string, len(out.Fields))
	var value string
	for _, f := range out.Fields {
		fields[f.Key] = f.Value
		if f.Key == "value" {
			value = f.Value
		}
	}
	if value == "" && len(out.Fields) > 0 {
		value = out.Fields[0].Value
	}
	return sseArtifact{Value: value, Fields: fields}
}

// writeJSON marshals v to JSON and writes it to the response.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}
