package ssh

import (
	"context"

	"github.com/smedje/smedje/internal/bench"
	"github.com/smedje/smedje/pkg/forge"
)

func init() {
	forge.Register(&ECDSA{})
}

// ECDSA generates an OpenSSH ECDSA keypair.
type ECDSA struct{}

func (e *ECDSA) Name() string             { return "ssh-ecdsa" }
func (e *ECDSA) Group() string            { return "ssh" }
func (e *ECDSA) Description() string      { return "Generate an ECDSA OpenSSH keypair" }
func (e *ECDSA) Category() forge.Category { return forge.CategoryCrypto }

func (e *ECDSA) Generate(ctx context.Context, opts forge.Options) (*forge.Output, error) {
	curve := "p256"
	if v, ok := opts.Params["curve"]; ok {
		curve = v
	}

	priv, pub, err := generateSSHKeypair("ecdsa-"+curve, 0)
	if err != nil {
		return nil, err
	}

	return forge.SingleArtifact("ssh",
		forge.Field{Key: "private-key", Value: priv, Sensitive: true},
		forge.Field{Key: "public-key", Value: pub},
	), nil
}

// Flags returns the configurable parameters for ECDSA key generation.
func (e *ECDSA) Flags() []forge.FlagDef {
	return []forge.FlagDef{
		{Name: "curve", Type: "string", Default: "p256", Description: "ECDSA curve", Options: []string{"p256", "p384"}},
	}
}

func (e *ECDSA) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.RunLegacy(ctx, e, 0)
}
