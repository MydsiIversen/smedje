package web

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// sseWriter wraps an http.ResponseWriter for Server-Sent Events streaming.
type sseWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

// newSSEWriter configures the response for SSE and returns a writer.
// Returns nil if the ResponseWriter doesn't support flushing.
func newSSEWriter(w http.ResponseWriter) *sseWriter {
	f, ok := w.(http.Flusher)
	if !ok {
		return nil
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	return &sseWriter{w: w, flusher: f}
}

// writeEvent sends a named SSE event with a JSON-encoded data payload.
func (s *sseWriter) writeEvent(event string, data interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("sse marshal: %w", err)
	}
	_, err = fmt.Fprintf(s.w, "event: %s\ndata: %s\n\n", event, b)
	if err != nil {
		return err
	}
	s.flusher.Flush()
	return nil
}

// sseStatus is the data shape for "status" events.
type sseStatus struct {
	Message string `json:"message"`
}

// sseProgress is the data shape for "progress" events.
type sseProgress struct {
	Current   int     `json:"current"`
	Total     int     `json:"total"`
	OpsPerSec float64 `json:"opsPerSec"`
}

// sseArtifact is the data shape for "artifact" events.
type sseArtifact struct {
	Value  string            `json:"value"`
	Fields map[string]string `json:"fields"`
}

// sseDone is the data shape for "done" events.
type sseDone struct {
	DurationMs int64 `json:"durationMs"`
	Count      int   `json:"count"`
}

// sseError is the data shape for "error" events.
type sseError struct {
	Message string `json:"message"`
}
