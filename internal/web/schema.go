// Package web provides an HTTP server that wraps the forge generator
// registry with a JSON API and SSE streaming.
package web

import (
	"context"
	"sort"
	"strings"

	"github.com/smedje/smedje/pkg/forge"
)

// GeneratorInfo is the summary representation of a generator in list responses.
type GeneratorInfo struct {
	Name        string `json:"name"`
	Address     string `json:"address"`
	Group       string `json:"group"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Rationale   string `json:"rationale,omitempty"`
}

// GeneratorSchema is the full representation of a generator, including flags
// and feature support.
type GeneratorSchema struct {
	GeneratorInfo
	Flags         []FlagDef         `json:"flags"`
	Supports      SupportedFeatures `json:"supports"`
	ExampleOutput interface{}       `json:"exampleOutput"`
}

// FlagDef describes a generator-specific flag.
type FlagDef struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Default     string   `json:"default,omitempty"`
	Description string   `json:"description"`
	Options     []string `json:"options,omitempty"`
}

// SupportedFeatures describes which optional features a generator supports.
type SupportedFeatures struct {
	Count bool `json:"count"`
	Seed  bool `json:"seed"`
	Bench bool `json:"bench"`
}

// GenerateRequest is the JSON body for POST /api/generate.
type GenerateRequest struct {
	Generator string            `json:"generator"`
	Count     int               `json:"count"`
	Format    string            `json:"format"`
	Params    map[string]string `json:"params,omitempty"`
	Seed      string            `json:"seed,omitempty"`
}

// ExplainRequest is the JSON body for POST /api/explain.
type ExplainRequest struct {
	Value string `json:"value"`
}

// ExplainResponse is the response body for POST /api/explain.
type ExplainResponse struct {
	Detected       string            `json:"detected"`
	Spec           string            `json:"spec,omitempty"`
	Layout         []LayoutSegment   `json:"layout"`
	Fields         map[string]string `json:"fields"`
	AlternateForms map[string]string `json:"alternateForms,omitempty"`
}

// LayoutSegment describes a decoded segment of an identified value.
type LayoutSegment struct {
	Start       int    `json:"start"`
	End         int    `json:"end"`
	Label       string `json:"label"`
	Type        string `json:"type"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

// VersionInfo is the response body for GET /api/version.
type VersionInfo struct {
	Version    string `json:"version"`
	Commit     string `json:"commit"`
	GoVersion  string `json:"goVersion"`
	PublicMode bool   `json:"publicMode"`
}

// commonFlags are the flags that every generator accepts.
var commonFlags = []FlagDef{
	{
		Name:        "count",
		Type:        "int",
		Default:     "1",
		Description: "Number of values to generate",
	},
	{
		Name:        "format",
		Type:        "string",
		Default:     "text",
		Description: "Output format",
		Options:     []string{"text", "json", "quiet", "csv", "sql", "env"},
	},
}

// generatorFlags returns the flags for a generator. Common flags (count,
// format) are always included; generator-specific flags come from the
// FlagDescriber interface when the generator implements it.
func generatorFlags(g forge.Generator) []FlagDef {
	flags := make([]FlagDef, len(commonFlags))
	copy(flags, commonFlags)

	if fd, ok := g.(forge.FlagDescriber); ok {
		for _, f := range fd.Flags() {
			flags = append(flags, FlagDef{
				Name:        f.Name,
				Type:        f.Type,
				Default:     f.Default,
				Description: f.Description,
				Options:     f.Options,
			})
		}
	}

	return flags
}

// generatorSupports returns the feature support flags for a generator.
// Crypto generators (ssh, tls, wireguard) don't support seeding.
func generatorSupports(g forge.Generator) SupportedFeatures {
	cat := g.Category()
	return SupportedFeatures{
		Count: true,
		Seed:  cat != forge.CategoryCrypto,
		Bench: true,
	}
}

// generatorRationale returns the --why rationale if the generator implements
// forge.Explainer.
func generatorRationale(g forge.Generator) string {
	if e, ok := g.(forge.Explainer); ok {
		return e.Why()
	}
	return ""
}

// generatorExampleOutput generates a single value with default options and
// returns it as the example output. Returns nil on error.
func generatorExampleOutput(g forge.Generator) interface{} {
	out, err := g.Generate(context.Background(), forge.Options{
		Count:  1,
		Format: "text",
	})
	if err != nil {
		return nil
	}
	fields := make(map[string]string, len(out.PrimaryFields()))
	for _, f := range out.PrimaryFields() {
		fields[f.Key] = f.Value
	}
	return fields
}

// buildGeneratorInfo returns a GeneratorInfo for the given generator.
func buildGeneratorInfo(g forge.Generator) GeneratorInfo {
	return GeneratorInfo{
		Name:        g.Name(),
		Address:     forge.Address(g),
		Group:       g.Group(),
		Category:    string(g.Category()),
		Description: g.Description(),
		Rationale:   generatorRationale(g),
	}
}

// buildGeneratorSchema returns a full GeneratorSchema for the given generator.
func buildGeneratorSchema(g forge.Generator) GeneratorSchema {
	return GeneratorSchema{
		GeneratorInfo: buildGeneratorInfo(g),
		Flags:         generatorFlags(g),
		Supports:      generatorSupports(g),
		ExampleOutput: generatorExampleOutput(g),
	}
}

// listGenerators returns sorted GeneratorInfo for all registered generators.
func listGenerators() []GeneratorInfo {
	all := forge.All()
	sort.Slice(all, func(i, j int) bool {
		ai, aj := forge.Address(all[i]), forge.Address(all[j])
		return ai < aj
	})

	infos := make([]GeneratorInfo, len(all))
	for i, g := range all {
		infos[i] = buildGeneratorInfo(g)
	}
	return infos
}

// isCryptoGenerator returns true for generators that should not use seeded entropy.
func isCryptoGenerator(g forge.Generator) bool {
	return g.Category() == forge.CategoryCrypto
}

// resolveGenerator looks up a generator by its dotted address (e.g. "uuid.v7").
func resolveGenerator(address string) (forge.Generator, error) {
	// Normalize: the API uses dots, Resolve already handles dots.
	address = strings.TrimSpace(address)
	gens, err := forge.Resolve(address)
	if err != nil {
		return nil, err
	}
	return gens[0], nil
}
