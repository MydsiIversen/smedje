package ssh

import (
	"context"
	"fmt"

	"github.com/smedje/smedje/internal/bench"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	forge.Register(&RSA{})
}

// RSA generates an OpenSSH RSA keypair.
type RSA struct{}

func (r *RSA) Name() string             { return "ssh-rsa" }
func (r *RSA) Group() string            { return "ssh" }
func (r *RSA) Description() string      { return "Generate an RSA OpenSSH keypair" }
func (r *RSA) Category() forge.Category { return forge.CategoryCrypto }

func (r *RSA) Generate(ctx context.Context, opts forge.Options) (*forge.Output, error) {
	bits := 4096
	if v, ok := opts.Params["bits"]; ok {
		fmt.Sscanf(v, "%d", &bits)
	}

	comment := opts.Params["comment"]

	priv, pub, err := generateSSHKeypair("rsa", bits, comment)
	if err != nil {
		return nil, err
	}

	return forge.SingleArtifact("ssh",
		forge.Field{Key: "private-key", Value: priv, Sensitive: true},
		forge.Field{Key: "public-key", Value: pub},
	), nil
}

// Flags returns the configurable parameters for RSA key generation.
func (r *RSA) Flags() []forge.FlagDef {
	return []forge.FlagDef{
		{Name: "bits", Type: "int", Default: "4096", Description: "Key strength (4096 more secure, 2048 faster)", Options: []string{"2048", "4096"}},
		{Name: "comment", Type: "string", Description: "Key comment embedded in the key file (e.g. user@host)"},
	}
}

func (r *RSA) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.RunLegacy(ctx, r, 0)
}
