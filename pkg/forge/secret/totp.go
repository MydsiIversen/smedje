package secret

import (
	"context"
	"encoding/base32"
	"fmt"
	"net/url"

	"github.com/smedje/smedje/internal/bench"
	"github.com/smedje/smedje/internal/entropy"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	forge.Register(&TOTP{})
}

// TOTP generates a TOTP secret and otpauth URI per RFC 6238.
//
// Options:
//
//	issuer:  issuer label (default "Smedje")
//	account: account label (default "user@example.com")
//	digits:  code length, 6 or 8 (default 6)
//	period:  time step in seconds (default 30)
//
// The secret is 20 bytes (160 bits) from crypto/rand, matching the SHA-1
// HMAC block size used by most authenticator apps.
type TOTP struct{}

func (t *TOTP) Name() string             { return "totp" }
func (t *TOTP) Description() string      { return "Generate a TOTP secret and otpauth URI" }
func (t *TOTP) Category() forge.Category { return forge.CategorySecret }

func (t *TOTP) Generate(ctx context.Context, opts forge.Options) (*forge.Output, error) {
	secretBytes := make([]byte, 20)
	if _, err := entropy.Read(secretBytes); err != nil {
		return nil, fmt.Errorf("totp: entropy read: %w", err)
	}

	secret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secretBytes)

	issuer := "Smedje"
	if v, ok := opts.Params["issuer"]; ok {
		issuer = v
	}
	account := "user@example.com"
	if v, ok := opts.Params["account"]; ok {
		account = v
	}
	digits := "6"
	if v, ok := opts.Params["digits"]; ok {
		digits = v
	}
	period := "30"
	if v, ok := opts.Params["period"]; ok {
		period = v
	}

	uri := fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&digits=%s&period=%s",
		url.PathEscape(issuer),
		url.PathEscape(account),
		secret,
		url.QueryEscape(issuer),
		digits,
		period,
	)

	return &forge.Output{
		Name: "totp",
		Fields: []forge.Field{
			{Key: "secret", Value: secret, Sensitive: true},
			{Key: "uri", Value: uri},
		},
	}, nil
}

func (t *TOTP) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.Run(ctx, t, 0)
}
