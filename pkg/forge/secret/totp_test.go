package secret

import (
	"context"
	"encoding/base32"
	"strings"
	"testing"

	"github.com/smedje/smedje/pkg/forge"
)

func TestTOTPGenerate(t *testing.T) {
	g := &TOTP{}
	out, err := g.Generate(context.Background(), forge.Options{})
	if err != nil {
		t.Fatal(err)
	}

	if len(out.PrimaryFields()) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(out.PrimaryFields()))
	}

	secretField := out.PrimaryFields()[0]
	uriField := out.PrimaryFields()[1]

	if !secretField.Sensitive {
		t.Error("secret should be marked sensitive")
	}

	decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secretField.Value)
	if err != nil {
		t.Fatalf("secret not valid base32: %v", err)
	}
	if len(decoded) != 20 {
		t.Errorf("secret decoded length = %d, want 20", len(decoded))
	}

	if !strings.HasPrefix(uriField.Value, "otpauth://totp/") {
		t.Errorf("URI missing otpauth prefix: %s", uriField.Value)
	}
	if !strings.Contains(uriField.Value, "secret="+secretField.Value) {
		t.Error("URI should contain the secret")
	}
}

func TestTOTPCustomParams(t *testing.T) {
	g := &TOTP{}
	opts := forge.Options{
		Params: map[string]string{
			"issuer":  "MyApp",
			"account": "alice@test.com",
			"digits":  "8",
			"period":  "60",
		},
	}
	out, err := g.Generate(context.Background(), opts)
	if err != nil {
		t.Fatal(err)
	}

	uri := out.PrimaryFields()[1].Value
	if !strings.Contains(uri, "MyApp") {
		t.Error("URI should contain issuer")
	}
	if !strings.Contains(uri, "alice") {
		t.Error("URI should contain account")
	}
	if !strings.Contains(uri, "digits=8") {
		t.Error("URI should contain digits=8")
	}
	if !strings.Contains(uri, "period=60") {
		t.Error("URI should contain period=60")
	}
}

func TestTOTPUniqueness(t *testing.T) {
	g := &TOTP{}
	secrets := make(map[string]struct{}, 50)
	for i := 0; i < 50; i++ {
		out, err := g.Generate(context.Background(), forge.Options{})
		if err != nil {
			t.Fatal(err)
		}
		s := out.PrimaryFields()[0].Value
		if _, exists := secrets[s]; exists {
			t.Fatalf("duplicate secret at iteration %d", i)
		}
		secrets[s] = struct{}{}
	}
}

func TestTOTPMetadata(t *testing.T) {
	g := &TOTP{}
	if g.Name() != "totp" {
		t.Errorf("Name() = %q, want %q", g.Name(), "totp")
	}
	if g.Category() != forge.CategorySecret {
		t.Errorf("Category() = %q, want %q", g.Category(), forge.CategorySecret)
	}
}
