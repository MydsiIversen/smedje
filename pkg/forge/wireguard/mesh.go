package wireguard

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/smedje/smedje/internal/bench"
	"github.com/smedje/smedje/internal/entropy"
	"github.com/smedje/smedje/pkg/forge"

	"golang.org/x/crypto/curve25519"
)

func init() {
	forge.Register(&Mesh{})
}

// Mesh generates a full WireGuard mesh configuration for N peers.
// Each peer receives a unique Curve25519 keypair and a wg-quick .conf file
// that references every other peer by public key and endpoint.
type Mesh struct{}

func (m *Mesh) Name() string             { return "mesh" }
func (m *Mesh) Group() string            { return "wireguard" }
func (m *Mesh) Description() string      { return "Generate a WireGuard mesh configuration for N peers" }
func (m *Mesh) Category() forge.Category { return forge.CategoryCrypto }

type meshKey struct {
	priv [32]byte
	pub  []byte
}

// Generate produces N WireGuard keypairs and a wg-quick .conf for each peer.
// Supported params: "peers" (int, 2-255), "endpoint" (comma-separated host:port
// list), "dns" (DNS server address for the Interface section).
func (m *Mesh) Generate(ctx context.Context, opts forge.Options) (*forge.Output, error) {
	n := 3
	if v, ok := opts.Params["peers"]; ok {
		fmt.Sscanf(v, "%d", &n)
	}
	if n < 2 || n > 255 {
		return nil, fmt.Errorf("wireguard: peers must be 2-255, got %d", n)
	}

	var endpoints []string
	if v, ok := opts.Params["endpoint"]; ok && v != "" {
		endpoints = strings.Split(v, ",")
		for i := range endpoints {
			endpoints[i] = strings.TrimSpace(endpoints[i])
		}
	}

	dns := ""
	if v, ok := opts.Params["dns"]; ok {
		dns = v
	}

	keys := make([]meshKey, n)
	for i := range keys {
		if _, err := entropy.Read(keys[i].priv[:]); err != nil {
			return nil, fmt.Errorf("wireguard: keygen: %w", err)
		}
		// Clamp the private key per Curve25519 convention, matching wg(8).
		keys[i].priv[0] &= 248
		keys[i].priv[31] = (keys[i].priv[31] & 127) | 64

		pub, err := curve25519.X25519(keys[i].priv[:], curve25519.Basepoint)
		if err != nil {
			return nil, fmt.Errorf("wireguard: pubkey: %w", err)
		}
		keys[i].pub = pub
	}

	artifacts := make([]forge.Artifact, n)
	for i := range n {
		privB64 := base64.StdEncoding.EncodeToString(keys[i].priv[:])
		pubB64 := base64.StdEncoding.EncodeToString(keys[i].pub)

		var sb strings.Builder
		sb.WriteString("[Interface]\n")
		sb.WriteString(fmt.Sprintf("PrivateKey = %s\n", privB64))
		sb.WriteString(fmt.Sprintf("Address = 10.0.0.%d/24\n", i+1))
		if dns != "" {
			sb.WriteString(fmt.Sprintf("DNS = %s\n", dns))
		}

		for j := range n {
			if j == i {
				continue
			}
			sb.WriteString("\n[Peer]\n")
			sb.WriteString(fmt.Sprintf("PublicKey = %s\n", base64.StdEncoding.EncodeToString(keys[j].pub)))

			endpoint := fmt.Sprintf("<endpoint-%d>", j+1)
			if j < len(endpoints) {
				endpoint = endpoints[j]
			}
			sb.WriteString(fmt.Sprintf("Endpoint = %s\n", endpoint))
			sb.WriteString(fmt.Sprintf("AllowedIPs = 10.0.0.%d/32\n", j+1))
		}

		label := fmt.Sprintf("peer-%d", i+1)
		artifacts[i] = forge.Artifact{
			Label:    label,
			Filename: label + ".conf",
			Fields: []forge.Field{
				{Key: "private-key", Value: privB64, Sensitive: true},
				{Key: "public-key", Value: pubB64},
				{Key: "config", Value: sb.String()},
			},
		}
	}

	return &forge.Output{Name: "wireguard-mesh", Artifacts: artifacts}, nil
}

// Flags describes the generator-specific CLI flags.
func (m *Mesh) Flags() []forge.FlagDef {
	return []forge.FlagDef{
		{Name: "peers", Type: "int", Default: "3", Description: "Number of peers in the mesh (2-255). Each gets a unique keypair and config"},
		{Name: "endpoint", Type: "string", Description: "Public endpoints, comma-separated (e.g. vpn1.example.com:51820,vpn2.example.com:51820)"},
		{Name: "dns", Type: "string", Description: "DNS server for all peers (e.g. 1.1.1.1)"},
	}
}

// Bench runs a self-benchmark for the mesh generator.
func (m *Mesh) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.RunLegacy(ctx, m, 0)
}
