package main

import (
	"strings"
	"testing"

	_ "github.com/smedje/smedje/pkg/forge/id"
	_ "github.com/smedje/smedje/pkg/forge/network"
	_ "github.com/smedje/smedje/pkg/forge/secret"
	_ "github.com/smedje/smedje/pkg/forge/ssh"
	_ "github.com/smedje/smedje/pkg/forge/tls"
	_ "github.com/smedje/smedje/pkg/forge/wireguard"

	"github.com/smedje/smedje/pkg/forge"
)

func TestGeneratorGroupCompletion(t *testing.T) {
	completions := generatorGroupCompletion(nil, nil, "uu")
	found := false
	for _, c := range completions {
		if strings.HasPrefix(c, "uuid") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected uuid in completions for 'uu', got %v", completions)
	}
}

func TestGeneratorVariantCompletion(t *testing.T) {
	fn := generatorVariantCompletion("uuid")
	completions, _ := fn(nil, nil, "v")
	if len(completions) == 0 {
		t.Fatal("expected uuid variant completions for 'v'")
	}
	found := false
	for _, c := range completions {
		if strings.HasPrefix(c, "v7") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected v7 in completions, got %v", completions)
	}
}

func TestFlagValueCompletion(t *testing.T) {
	g, ok := forge.Get(forge.CategorySecret, "password")
	if !ok {
		t.Fatal("password generator not found")
	}
	completions := flagValueCompletion(g, "charset")
	if len(completions) == 0 {
		t.Fatal("expected charset completions")
	}
	found := false
	for _, c := range completions {
		if c == "full" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'full' in charset completions, got %v", completions)
	}
}

func TestFlagValueCompletionNoDescriber(t *testing.T) {
	g, ok := forge.Get(forge.CategoryID, "v7")
	if !ok {
		t.Fatal("uuid v7 generator not found")
	}
	completions := flagValueCompletion(g, "nonexistent")
	if completions != nil {
		t.Errorf("expected nil for non-FlagDescriber, got %v", completions)
	}
}
