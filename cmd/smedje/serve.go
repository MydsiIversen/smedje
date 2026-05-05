package main

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/smedje/smedje/internal/web"

	// Ensure all generator packages are registered.
	_ "github.com/smedje/smedje/pkg/forge/crypto"
	_ "github.com/smedje/smedje/pkg/forge/id"
	_ "github.com/smedje/smedje/pkg/forge/network"
	_ "github.com/smedje/smedje/pkg/forge/secret"
	_ "github.com/smedje/smedje/pkg/forge/ssh"
	_ "github.com/smedje/smedje/pkg/forge/tls"
	_ "github.com/smedje/smedje/pkg/forge/wireguard"
)

// publicMode is set via ldflags for public-facing deployments.
var publicMode string

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().Int("port", 8080, "HTTP listen port")
	serveCmd.Flags().String("host", "127.0.0.1", "HTTP listen address")
	serveCmd.Flags().Bool("dev", false, "Enable dev mode (CORS, Vite proxy)")
	serveCmd.Flags().Bool("no-browser", false, "Don't open browser on start")
	serveCmd.Flags().Bool("public", false, "Enable public mode (lower limits)")
	serveCmd.Flags().String("analytics-script", "", "Analytics script URL to inject")
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Smedje web server",
	Long: `Start an HTTP server that wraps the forge generator registry
with a JSON API, SSE streaming, and (eventually) a web frontend.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		port, _ := cmd.Flags().GetInt("port")
		host, _ := cmd.Flags().GetString("host")
		dev, _ := cmd.Flags().GetBool("dev")
		noBrowser, _ := cmd.Flags().GetBool("no-browser")
		public, _ := cmd.Flags().GetBool("public")
		analyticsScript, _ := cmd.Flags().GetString("analytics-script")

		if publicMode == "true" {
			public = true
		}

		cfg := web.DefaultConfig()
		cfg.Port = port
		cfg.Host = host
		cfg.Dev = dev
		cfg.Public = public
		cfg.NoBrowser = noBrowser
		cfg.AnalyticsScript = analyticsScript
		cfg.Version = version
		cfg.Commit = commit

		srv := web.New(cfg)

		if !noBrowser && !public {
			go openBrowser(fmt.Sprintf("http://%s", srv.Addr()))
		}

		fmt.Fprintf(cmd.OutOrStdout(), "smedje server listening on http://%s\n", srv.Addr())
		if dev {
			fmt.Fprintln(cmd.OutOrStdout(), "  dev mode: CORS enabled, proxying to localhost:5173")
		}
		if public {
			fmt.Fprintln(cmd.OutOrStdout(), "  public mode: rate limits and reduced max count")
		}

		return srv.ListenAndServe()
	},
}

// openBrowser attempts to open the given URL in the default browser.
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}
