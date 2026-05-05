package network

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/smedje/smedje/internal/bench"
	"github.com/smedje/smedje/internal/entropy"
	"github.com/smedje/smedje/pkg/forge"
)

func init() { forge.Register(&OpenVPNTLSAuth{}) }

// OpenVPNTLSAuth generates a static tls-auth key in OpenVPN's native format.
// The default is 2048 bits (256 bytes), which matches the output of
// `openvpn --genkey --secret ta.key`. The key is hex-encoded and wrapped
// in OpenVPN's PEM-style header and footer.
type OpenVPNTLSAuth struct{}

func (o *OpenVPNTLSAuth) Name() string             { return "openvpn-tls-auth" }
func (o *OpenVPNTLSAuth) Group() string            { return "network" }
func (o *OpenVPNTLSAuth) Description() string      { return "Generate an OpenVPN tls-auth key" }
func (o *OpenVPNTLSAuth) Category() forge.Category { return forge.CategoryCrypto }

// Generate returns the OpenVPN static key in its native format.
func (o *OpenVPNTLSAuth) Generate(ctx context.Context, opts forge.Options) (*forge.Output, error) {
	bits := 2048
	if v, ok := opts.Params["bits"]; ok {
		fmt.Sscanf(v, "%d", &bits)
	}
	byteLen := bits / 8

	b := make([]byte, byteLen)
	if _, err := entropy.Read(b); err != nil {
		return nil, fmt.Errorf("openvpn: %w", err)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "#\n# %d bit OpenVPN static key\n#\n", bits)
	sb.WriteString("-----BEGIN OpenVPN Static key V1-----\n")
	hexStr := hex.EncodeToString(b)
	for i := 0; i < len(hexStr); i += 32 {
		end := i + 32
		if end > len(hexStr) {
			end = len(hexStr)
		}
		sb.WriteString(hexStr[i:end])
		sb.WriteByte('\n')
	}
	sb.WriteString("-----END OpenVPN Static key V1-----\n")

	return forge.SingleArtifact("openvpn-tls-auth",
		forge.Field{Key: "key", Value: sb.String(), Sensitive: true},
	), nil
}

func (o *OpenVPNTLSAuth) Flags() []forge.FlagDef {
	return []forge.FlagDef{
		{Name: "bits", Type: "int", Default: "2048", Description: "Key size in bits (must be multiple of 8). 2048 matches openvpn --genkey default"},
	}
}

func (o *OpenVPNTLSAuth) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.RunLegacy(ctx, o, 0)
}
