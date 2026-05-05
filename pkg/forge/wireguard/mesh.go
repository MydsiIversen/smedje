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

	subnet := "10.0.0"
	mask := 24
	if v, ok := opts.Params["subnet"]; ok && v != "" {
		subnet, mask = parseSubnet(v)
	}

	port := 51820
	if v, ok := opts.Params["port"]; ok && v != "" {
		fmt.Sscanf(v, "%d", &port)
	}
	if port < 1 || port > 65535 {
		return nil, fmt.Errorf("wireguard: port must be 1-65535, got %d", port)
	}

	keepalive := 0
	if v, ok := opts.Params["keepalive"]; ok && v != "" {
		fmt.Sscanf(v, "%d", &keepalive)
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
		sb.WriteString(fmt.Sprintf("Address = %s.%d/%d\n", subnet, i+1, mask))
		sb.WriteString(fmt.Sprintf("ListenPort = %d\n", port))
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
			sb.WriteString(fmt.Sprintf("AllowedIPs = %s.%d/32\n", subnet, j+1))
			if keepalive > 0 {
				sb.WriteString(fmt.Sprintf("PersistentKeepalive = %d\n", keepalive))
			}
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
		{Name: "subnet", Type: "string", Default: "10.0.0.0/24", Description: "Mesh subnet (e.g. 10.10.0.0/24, 172.16.0.0/24)"},
		{Name: "port", Type: "int", Default: "51820", Description: "WireGuard listen port"},
		{Name: "endpoint", Type: "string", Description: "Public endpoints, comma-separated (e.g. vpn1.example.com:51820,vpn2.example.com:51820)"},
		{Name: "dns", Type: "string", Description: "DNS server for all peers (e.g. 1.1.1.1)"},
		{Name: "keepalive", Type: "int", Default: "0", Description: "PersistentKeepalive interval in seconds (0 = disabled, 25 = typical for NAT)"},
	}
}

// parseSubnet extracts the base address prefix and mask from a CIDR string.
// "10.10.0.0/24" returns ("10.10.0", 24). On parse failure, returns defaults.
func parseSubnet(cidr string) (string, int) {
	parts := strings.SplitN(cidr, "/", 2)
	addr := parts[0]
	mask := 24
	if len(parts) == 2 {
		fmt.Sscanf(parts[1], "%d", &mask)
	}
	octets := strings.Split(addr, ".")
	if len(octets) >= 3 {
		return strings.Join(octets[:3], "."), mask
	}
	return "10.0.0", mask
}

// Bench runs a self-benchmark for the mesh generator.
func (m *Mesh) Bench(ctx context.Context) (*forge.BenchResult, error) {
	return bench.RunLegacy(ctx, m, 0)
}
