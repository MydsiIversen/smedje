// Package output renders forge.Output in different formats.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/smedje/smedje/pkg/forge"
)

// Text writes human-readable output.
func Text(w io.Writer, o *forge.Output) error {
	if len(o.Fields) == 1 {
		_, err := fmt.Fprintln(w, o.Fields[0].Value)
		return err
	}
	for _, f := range o.Fields {
		if _, err := fmt.Fprintf(w, "%s: %s\n", f.Key, f.Value); err != nil {
			return err
		}
	}
	return nil
}

// Quiet writes only the raw values, one per line.
func Quiet(w io.Writer, o *forge.Output) error {
	for _, f := range o.Fields {
		if _, err := fmt.Fprintln(w, f.Value); err != nil {
			return err
		}
	}
	return nil
}

// JSON writes structured JSON output.
func JSON(w io.Writer, o *forge.Output) error {
	m := make(map[string]string, len(o.Fields))
	for _, f := range o.Fields {
		m[f.Key] = f.Value
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(m)
}

// Env writes output as KEY=VALUE lines suitable for shell eval.
func Env(w io.Writer, o *forge.Output, prefix string) error {
	for _, f := range o.Fields {
		key := strings.ToUpper(prefix + "_" + f.Key)
		key = strings.ReplaceAll(key, "-", "_")
		if _, err := fmt.Fprintf(w, "%s=%s\n", key, f.Value); err != nil {
			return err
		}
	}
	return nil
}

// PEM writes PEM-encoded values as-is (they already contain headers).
func PEM(w io.Writer, o *forge.Output) error {
	for _, f := range o.Fields {
		if _, err := fmt.Fprint(w, f.Value); err != nil {
			return err
		}
	}
	return nil
}

// Render dispatches to the appropriate renderer based on format string.
func Render(w io.Writer, o *forge.Output, format string) error {
	switch format {
	case "json":
		return JSON(w, o)
	case "quiet":
		return Quiet(w, o)
	case "env":
		return Env(w, o, o.Name)
	case "pem":
		return PEM(w, o)
	default:
		return Text(w, o)
	}
}
