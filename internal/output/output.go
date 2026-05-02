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
		m := make(map[string]string, len(o.Fields))
		for _, f := range o.Fields {
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
	var headers []string
	if len(outputs[0].Fields) == 1 {
		headers = []string{outputs[0].Name}
	} else {
		for _, f := range outputs[0].Fields {
			headers = append(headers, f.Key)
		}
	}
	fmt.Fprintln(w, strings.Join(headers, ","))

	for _, o := range outputs {
		var vals []string
		for _, f := range o.Fields {
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

	var cols []string
	for _, f := range outputs[0].Fields {
		cols = append(cols, f.Key)
	}

	fmt.Fprintf(w, "INSERT INTO %s (%s) VALUES\n", table, strings.Join(cols, ", "))
	for i, o := range outputs {
		var vals []string
		for _, f := range o.Fields {
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
		prefix := fmt.Sprintf("%s_%d", strings.ToUpper(o.Name), i+1)
		prefix = strings.ReplaceAll(prefix, "-", "_")
		for _, f := range o.Fields {
			key := prefix + "_" + strings.ToUpper(f.Key)
			key = strings.ReplaceAll(key, "-", "_")
			fmt.Fprintf(w, "%s=%s\n", key, f.Value)
		}
	}
	return nil
}
