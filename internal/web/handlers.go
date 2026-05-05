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
	"github.com/smedje/smedje/internal/output"
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
		artifact := outputToArtifact(out)
		if isBatchFormat(req.Format) {
			var buf strings.Builder
			output.RenderBatch(&buf, []*forge.Output{out}, req.Format, output.BatchOptions{
				SQLTable: req.Generator,
			})
			artifact.Formatted = buf.String()
		}
		writeJSON(w, http.StatusOK, artifact)
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

	const maxBatchFormat = 10000
	collectForBatch := isBatchFormat(req.Format) && count <= maxBatchFormat
	var outputs []*forge.Output

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
		if collectForBatch {
			outputs = append(outputs, out)
		}

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

	done := sseDone{
		DurationMs: time.Since(start).Milliseconds(),
		Count:      count,
	}
	if collectForBatch && len(outputs) > 0 {
		var buf strings.Builder
		output.RenderBatch(&buf, outputs, req.Format, output.BatchOptions{
			SQLTable: req.Generator,
		})
		done.Formatted = buf.String()
	}
	_ = sse.writeEvent("done", done)
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
		hexStr := strings.ReplaceAll(lower, "-", "")
		forms["hex"] = hexStr
		forms["urn"] = "urn:uuid:" + lower
		raw := hexToBytes(hexStr)
		if raw != nil {
			forms["base64"] = base64.StdEncoding.EncodeToString(raw)
		}
	case strings.Contains(f, "ulid"):
		trimmed := strings.TrimSpace(input)
		hexStr := ulidToHex(trimmed)
		forms["hex"] = hexStr
		raw := hexToBytes(hexStr)
		if raw != nil {
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

// hexToBytes decodes a hex string to bytes, returning nil on invalid input.
func hexToBytes(s string) []byte {
	if len(s)%2 != 0 {
		return nil
	}
	b := make([]byte, len(s)/2)
	for i := 0; i < len(s); i += 2 {
		hi := hexVal(s[i])
		lo := hexVal(s[i+1])
		if hi < 0 || lo < 0 {
			return nil
		}
		b[i/2] = byte(hi<<4 | lo)
	}
	return b
}

func hexVal(c byte) int {
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

// ulidToHex converts a 26-char ULID (Crockford base32) to a 32-char hex string.
func ulidToHex(s string) string {
	const alphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
	upper := strings.ToUpper(s)

	// 26 Crockford chars encode 130 bits; top 2 must be zero, leaving 128.
	var bits [130]byte
	for i := 0; i < 26; i++ {
		idx := strings.IndexByte(alphabet, upper[i])
		if idx < 0 {
			return ""
		}
		for b := 4; b >= 0; b-- {
			if idx&(1<<b) != 0 {
				bits[i*5+(4-b)] = 1
			}
		}
	}

	// Skip the top 2 bits and convert the remaining 128 to hex nibbles.
	result := make([]byte, 0, 32)
	for i := 2; i < 130; i += 4 {
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

// handlePrivacy renders a standalone HTML privacy page explaining what
// data Smedje does and does not collect.
func (s *Server) handlePrivacy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, privacyHTML)
}

const privacyHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Privacy — Smedje</title>
<style>
  *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
  body {
    background: #0A0B0E;
    color: #E8E6E1;
    font-family: 'Geist', system-ui, sans-serif;
    font-size: 14px;
    line-height: 1.7;
    padding: 2rem 1rem;
  }
  main {
    max-width: 720px;
    margin: 0 auto;
  }
  a { color: #E2683A; text-decoration: none; }
  a:hover { text-decoration: underline; }
  .back { display: inline-block; margin-bottom: 2rem; color: #6B6E78; }
  .back:hover { color: #E8E6E1; }
  h1 {
    font-size: 1.5rem;
    margin-bottom: 1.5rem;
    font-weight: 600;
  }
  .panel {
    background: #13151A;
    border: 1px solid #1F222A;
    border-radius: 0px;
    padding: 1.5rem;
    margin-bottom: 1.5rem;
  }
  h2 {
    font-size: 1rem;
    font-weight: 600;
    margin-bottom: 0.75rem;
    color: #E2683A;
  }
  p { margin-bottom: 0.75rem; color: #E8E6E1; }
  p:last-child { margin-bottom: 0; }
  code {
    font-family: 'Geist Mono', ui-monospace, monospace;
    background: #1F222A;
    padding: 0.15em 0.35em;
    border-radius: 4px;
    font-size: 0.9em;
  }
  .muted { color: #6B6E78; }
</style>
</head>
<body>
<main>
  <a href="/" class="back">&larr; Back to Smedje</a>
  <h1>Privacy</h1>

  <div class="panel">
    <h2>Your data stays in your browser</h2>
    <p>Every value Smedje generates — UUIDs, keys, passwords, certificates —
    is created entirely in your browser or on your local machine. Generated
    values are never transmitted to any server. There is no backend database,
    no telemetry endpoint, and no server-side logging of generated output.</p>
  </div>

  <div class="panel">
    <h2>Analytics</h2>
    <p>Smedje tracks aggregate page views and popular generators using a
    self-hosted <a href="https://umami.is">Umami</a> instance at
    <code>analytics.smedje.net</code>. Umami is open-source, privacy-focused
    analytics software.</p>
    <p>This means:</p>
    <p>— No cookies are set. Ever.<br>
    — No fingerprinting or cross-site tracking.<br>
    — No personally identifiable information (PII) is collected.<br>
    — IP addresses are not stored.<br>
    — The <code>DNT</code> (Do Not Track) header is honored. If your browser
    sends it, no analytics data is recorded for your visit.</p>
  </div>

  <div class="panel">
    <h2>Source code</h2>
    <p>Smedje is open source. You can inspect exactly what the application
    does — including this privacy policy — in the repository:
    <a href="https://github.com/MydsiIversen/smedje">github.com/MydsiIversen/smedje</a>.</p>
  </div>

  <p class="muted" style="margin-top: 1rem; text-align: center;">
    That&rsquo;s it. No legalese, no dark patterns, no surprises.
  </p>
</main>
</body>
</html>`

// isBatchFormat returns true for formats that produce structured batch output
// (SQL, CSV, JSON array, env vars) rather than plain text.
func isBatchFormat(format string) bool {
	switch format {
	case "csv", "sql", "env", "json":
		return true
	}
	return false
}

// outputToArtifact converts a forge.Output to an sseArtifact.
func outputToArtifact(out *forge.Output) sseArtifact {
	fields := make(map[string]string, len(out.PrimaryFields()))
	var value string
	for _, f := range out.PrimaryFields() {
		fields[f.Key] = f.Value
		if f.Key == "value" {
			value = f.Value
		}
	}
	if value == "" && len(out.PrimaryFields()) > 0 {
		value = out.PrimaryFields()[0].Value
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
