package output

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/smedje/smedje/pkg/forge"
)

// WriteDir writes each artifact in out as a separate file under dir.
// Returns the list of file paths written.
func WriteDir(dir string, out *forge.Output, format string) ([]string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create output dir: %w", err)
	}

	ext := formatExtension(format)
	var paths []string

	for _, a := range out.Artifacts {
		filename := a.Filename
		if filename == "" {
			label := a.Label
			if label == "" {
				label = out.Name
			}
			filename = label + ext
		}
		path := filepath.Join(dir, filename)

		var buf bytes.Buffer
		if isPEMOrConf(filename) {
			writeRawArtifact(&buf, a)
		} else {
			single := &forge.Output{
				Name:      out.Name,
				Artifacts: []forge.Artifact{a},
			}
			if err := Render(&buf, single, format); err != nil {
				return nil, fmt.Errorf("render %s: %w", a.Label, err)
			}
		}

		if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
			return nil, fmt.Errorf("write %s: %w", path, err)
		}
		paths = append(paths, path)
	}

	return paths, nil
}

func isPEMOrConf(filename string) bool {
	return strings.HasSuffix(filename, ".pem") || strings.HasSuffix(filename, ".conf")
}

// writeRawArtifact writes artifact fields as raw values suitable for PEM
// and config files. For config-type files, only the "config" field is
// written; for PEM files, all PEM blocks are concatenated.
func writeRawArtifact(buf *bytes.Buffer, a forge.Artifact) {
	for _, f := range a.Fields {
		if strings.Contains(f.Value, "-----BEGIN ") {
			buf.WriteString(f.Value)
			if !strings.HasSuffix(f.Value, "\n") {
				buf.WriteByte('\n')
			}
			continue
		}
		if f.Key == "config" {
			buf.WriteString(f.Value)
			if !strings.HasSuffix(f.Value, "\n") {
				buf.WriteByte('\n')
			}
		}
	}
}

func formatExtension(format string) string {
	switch format {
	case "json":
		return ".json"
	case "pem":
		return ".pem"
	case "csv":
		return ".csv"
	case "sql":
		return ".sql"
	case "env":
		return ".env"
	default:
		return ".txt"
	}
}
