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
	if len(o.Artifacts) <= 1 {
		fields := o.PrimaryFields()
		if len(fields) == 1 {
			_, err := fmt.Fprintln(w, fields[0].Value)
			return err
		}
		for _, f := range fields {
			if _, err := fmt.Fprintf(w, "%s: %s\n", f.Key, f.Value); err != nil {
				return err
			}
		}
		return nil
	}
	// Multi-artifact: header per artifact.
	for i, a := range o.Artifacts {
		if i > 0 {
			fmt.Fprintln(w)
		}
		label := a.Label
		if label == "" {
			label = fmt.Sprintf("artifact-%d", i+1)
		}
		fmt.Fprintf(w, "--- %s ---\n", label)
		if len(a.Fields) == 1 {
			fmt.Fprintln(w, a.Fields[0].Value)
		} else {
			for _, f := range a.Fields {
				if _, err := fmt.Fprintf(w, "%s: %s\n", f.Key, f.Value); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Quiet writes only the raw values, one per line.
func Quiet(w io.Writer, o *forge.Output) error {
	if len(o.Artifacts) <= 1 {
		for _, f := range o.PrimaryFields() {
			if _, err := fmt.Fprintln(w, f.Value); err != nil {
				return err
			}
		}
		return nil
	}
	// Multi-artifact: blank line between artifacts.
	for i, a := range o.Artifacts {
		if i > 0 {
			fmt.Fprintln(w)
		}
		for _, f := range a.Fields {
			if _, err := fmt.Fprintln(w, f.Value); err != nil {
				return err
			}
		}
	}
	return nil
}

// JSON writes structured JSON output.
func JSON(w io.Writer, o *forge.Output) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	if len(o.Artifacts) <= 1 {
		fields := o.PrimaryFields()
		m := make(map[string]string, len(fields))
		for _, f := range fields {
			m[f.Key] = f.Value
		}
		return enc.Encode(m)
	}
	// Multi-artifact: array of {label, fields}.
	type artifactJSON struct {
		Label  string            `json:"label"`
		Fields map[string]string `json:"fields"`
	}
	items := make([]artifactJSON, 0, len(o.Artifacts))
	for _, a := range o.Artifacts {
		m := make(map[string]string, len(a.Fields))
		for _, f := range a.Fields {
			m[f.Key] = f.Value
		}
		label := a.Label
		if label == "" {
			label = "artifact"
		}
		items = append(items, artifactJSON{Label: label, Fields: m})
	}
	return enc.Encode(items)
}

// Env writes output as KEY=VALUE lines suitable for shell eval.
func Env(w io.Writer, o *forge.Output, prefix string) error {
	if len(o.Artifacts) <= 1 {
		for _, f := range o.PrimaryFields() {
			key := strings.ToUpper(prefix + "_" + f.Key)
			key = strings.ReplaceAll(key, "-", "_")
			if _, err := fmt.Fprintf(w, "%s=%s\n", key, f.Value); err != nil {
				return err
			}
		}
		return nil
	}
	// Multi-artifact: label becomes part of the prefix.
	for _, a := range o.Artifacts {
		label := a.Label
		if label == "" {
			label = "artifact"
		}
		artPrefix := strings.ToUpper(label)
		artPrefix = strings.ReplaceAll(artPrefix, "-", "_")
		for _, f := range a.Fields {
			key := artPrefix + "_" + strings.ToUpper(f.Key)
			key = strings.ReplaceAll(key, "-", "_")
			if _, err := fmt.Fprintf(w, "%s=%s\n", key, f.Value); err != nil {
				return err
			}
		}
	}
	return nil
}

// PEM writes PEM-encoded values as-is (they already contain headers).
func PEM(w io.Writer, o *forge.Output) error {
	if len(o.Artifacts) <= 1 {
		for _, f := range o.PrimaryFields() {
			if _, err := fmt.Fprint(w, f.Value); err != nil {
				return err
			}
		}
		return nil
	}
	// Multi-artifact: concatenate in order.
	for _, a := range o.Artifacts {
		for _, f := range a.Fields {
			if _, err := fmt.Fprint(w, f.Value); err != nil {
				return err
			}
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

// RenderBatch renders multiple outputs in the appropriate batch format.
// JSON wraps in an array; CSV adds a header row; SQL produces INSERT statements.
func RenderBatch(w io.Writer, outputs []*forge.Output, format string, opts BatchOptions) error {
	switch format {
	case "json":
		return jsonBatch(w, outputs)
	case "csv":
		return csvBatch(w, outputs)
	case "sql":
		return sqlBatch(w, outputs, opts.SQLTable)
	case "env":
		return envBatch(w, outputs)
	default:
		for _, o := range outputs {
			if err := Render(w, o, format); err != nil {
				return err
			}
		}
		return nil
	}
}

// BatchOptions controls batch rendering behavior.
type BatchOptions struct {
	SQLTable string
}

func jsonBatch(w io.Writer, outputs []*forge.Output) error {
	items := make([]map[string]string, 0, len(outputs))
	for _, o := range outputs {
		fields := o.PrimaryFields()
		m := make(map[string]string, len(fields))
		for _, f := range fields {
			m[f.Key] = f.Value
		}
		items = append(items, m)
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(items)
}

func csvBatch(w io.Writer, outputs []*forge.Output) error {
	if len(outputs) == 0 {
		return nil
	}
	// Header: for single-field generators use the output name (group name)
	// as the column header; for multi-field use the field keys.
	firstFields := outputs[0].PrimaryFields()
	var headers []string
	if len(firstFields) == 1 {
		headers = []string{outputs[0].Name}
	} else {
		for _, f := range firstFields {
			headers = append(headers, f.Key)
		}
	}
	fmt.Fprintln(w, strings.Join(headers, ","))

	for _, o := range outputs {
		var vals []string
		for _, f := range o.PrimaryFields() {
			vals = append(vals, csvEscape(f.Value))
		}
		fmt.Fprintln(w, strings.Join(vals, ","))
	}
	return nil
}

func csvEscape(s string) string {
	if strings.ContainsAny(s, ",\"\n") {
		return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
	}
	return s
}

func sqlBatch(w io.Writer, outputs []*forge.Output, table string) error {
	if len(outputs) == 0 {
		return nil
	}
	if table == "" {
		table = outputs[0].Name
	}

	firstFields := outputs[0].PrimaryFields()
	var cols []string
	for _, f := range firstFields {
		cols = append(cols, f.Key)
	}

	fmt.Fprintf(w, "INSERT INTO %s (%s) VALUES\n", table, strings.Join(cols, ", "))
	for i, o := range outputs {
		var vals []string
		for _, f := range o.PrimaryFields() {
			vals = append(vals, "'"+strings.ReplaceAll(f.Value, "'", "''")+"'")
		}
		sep := ","
		if i == len(outputs)-1 {
			sep = ";"
		}
		fmt.Fprintf(w, "  (%s)%s\n", strings.Join(vals, ", "), sep)
	}
	return nil
}

func envBatch(w io.Writer, outputs []*forge.Output) error {
	for i, o := range outputs {
		fields := o.PrimaryFields()
		prefix := fmt.Sprintf("%s_%d", strings.ToUpper(o.Name), i+1)
		prefix = strings.ReplaceAll(prefix, "-", "_")
		for _, f := range fields {
			key := prefix + "_" + strings.ToUpper(f.Key)
			key = strings.ReplaceAll(key, "-", "_")
			fmt.Fprintf(w, "%s=%s\n", key, f.Value)
		}
	}
	return nil
}
