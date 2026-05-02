package web

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/smedje/smedje/internal/entropy"
	"github.com/smedje/smedje/internal/explain"
	"github.com/smedje/smedje/internal/recommend"
	"github.com/smedje/smedje/pkg/forge"
)

// seedMu serializes seeded generation requests since entropy.SetSeed
// uses global state. Fine for demo use; not suitable for high-concurrency.
var seedMu sync.Mutex

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

	if req.Seed != "" && !isCryptoGenerator(g) {
		seedMu.Lock()
		entropy.SetSeed(req.Seed)
		defer func() {
			entropy.Reset()
			seedMu.Unlock()
		}()
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

	resp := ExplainResponse{
		Detected:       m.Format,
		Fields:         m.Fields,
		Layout:         convertLayout(m.Layout),
		Spec:           specURL(m.Format),
		AlternateForms: alternateForms(req.Value, m.Format),
	}
	writeJSON(w, http.StatusOK, resp)
}

// convertLayout maps explain.LayoutSegment values to the web LayoutSegment type.
func convertLayout(segs []explain.LayoutSegment) []LayoutSegment {
	if len(segs) == 0 {
		return nil
	}
	out := make([]LayoutSegment, len(segs))
	for i, s := range segs {
		out[i] = LayoutSegment{
			Start:       s.Start,
			End:         s.End,
			Label:       s.Label,
			Type:        s.Type,
			Value:       s.Value,
			Description: s.Description,
		}
	}
	return out
}

// specURL returns the specification URL for the given format string.
func specURL(format string) string {
	f := strings.ToLower(format)
	switch {
	case strings.Contains(f, "uuid"):
		return "https://www.rfc-editor.org/rfc/rfc9562"
	case strings.Contains(f, "ulid"):
		return "https://github.com/ulid/spec"
	default:
		return ""
	}
}

// alternateForms returns alternate string representations of the input value.
func alternateForms(input, format string) map[string]string {
	f := strings.ToLower(format)
	forms := make(map[string]string)

	switch {
	case strings.Contains(f, "uuid"):
		lower := strings.ToLower(strings.TrimSpace(input))
		hex := strings.ReplaceAll(lower, "-", "")
		forms["hex"] = hex
		forms["urn"] = "urn:uuid:" + lower
		raw, err := decodeHex(hex)
		if err == nil {
			forms["base64"] = base64.StdEncoding.EncodeToString(raw)
		}
	case strings.Contains(f, "ulid"):
		trimmed := strings.TrimSpace(input)
		forms["hex"] = ulidToHex(trimmed)
		raw, err := decodeHex(forms["hex"])
		if err == nil {
			forms["base64"] = base64.StdEncoding.EncodeToString(raw)
		}
	case strings.Contains(f, "snowflake"):
		trimmed := strings.TrimSpace(input)
		var n uint64
		for _, c := range trimmed {
			n = n*10 + uint64(c-'0')
		}
		forms["hex"] = fmt.Sprintf("%x", n)
		forms["binary"] = fmt.Sprintf("%b", n)
	}

	if len(forms) == 0 {
		return nil
	}
	return forms
}

// decodeHex is a small wrapper around hex.DecodeString for readability.
func decodeHex(s string) ([]byte, error) {
	return hexDecodeString(s)
}

// hexDecodeString decodes a hex string to bytes.
var hexDecodeString = func(s string) ([]byte, error) {
	b := make([]byte, len(s)/2)
	for i := 0; i < len(s); i += 2 {
		hi := unhex(s[i])
		lo := unhex(s[i+1])
		if hi < 0 || lo < 0 {
			return nil, fmt.Errorf("invalid hex char")
		}
		b[i/2] = byte(hi<<4 | lo)
	}
	return b, nil
}

func unhex(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c - 'a' + 10)
	case c >= 'A' && c <= 'F':
		return int(c - 'A' + 10)
	default:
		return -1
	}
}

// ulidToHex converts a ULID (Crockford base32) to a hex string.
func ulidToHex(s string) string {
	const alphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
	upper := strings.ToUpper(s)

	// ULID is 128 bits = 16 bytes. Decode the 26 Crockford chars.
	var bits [128]byte
	bitIdx := 0
	for i := 0; i < 26; i++ {
		idx := strings.IndexByte(alphabet, upper[i])
		if idx < 0 {
			return ""
		}
		// Each char encodes 5 bits, except first encodes only 2 (total: 2 + 25*5 = 127, but spec says 128).
		// Actually: 26 * 5 = 130 bits, top 2 bits must be zero.
		nBits := 5
		for b := nBits - 1; b >= 0 && bitIdx < 128; b-- {
			if bitIdx < 2 {
				// first 2 positions come from the top bits of the first char
			}
			if idx&(1<<b) != 0 {
				bits[bitIdx] = 1
			}
			bitIdx++
		}
	}

	// Convert bits to hex.
	result := make([]byte, 0, 32)
	for i := 0; i < 128; i += 4 {
		val := bits[i]<<3 | bits[i+1]<<2 | bits[i+2]<<1 | bits[i+3]
		if val < 10 {
			result = append(result, '0'+val)
		} else {
			result = append(result, 'a'+val-10)
		}
	}
	return string(result)
}

// handleRecommend returns opinionated recommendations for a topic.
func (s *Server) handleRecommend(w http.ResponseWriter, r *http.Request) {
	topic := r.URL.Query().Get("topic")
	if topic == "" {
		writeJSON(w, http.StatusOK, map[string][]string{
			"topics": recommend.Topics(),
		})
		return
	}

	recs, ok := recommend.Recommendations[topic]
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":  fmt.Sprintf("unknown topic %q", topic),
			"topics": recommend.Topics(),
		})
		return
	}

	useCase := r.URL.Query().Get("use-case")
	if useCase != "" {
		recs = recommend.FilterByUseCase(recs, useCase)
	}

	writeJSON(w, http.StatusOK, recs)
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
