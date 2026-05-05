package flags

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/smedje/smedje/pkg/forge"
)

// AddWhyFlag registers --why on a command.
func AddWhyFlag(cmd *cobra.Command) {
	cmd.Flags().Bool("why", false, "Generate one example and explain the format")
}

// GetWhy reads the --why flag value.
func GetWhy(cmd *cobra.Command) bool {
	v, _ := cmd.Flags().GetBool("why")
	return v
}

// RunWhy generates one example and prints the rationale. Returns true if
// --why was handled, false if not set.
func RunWhy(cmd *cobra.Command, g forge.Generator, opts forge.Options) (bool, error) {
	if !GetWhy(cmd) {
		return false, nil
	}

	out, err := g.Generate(context.Background(), opts)
	if err != nil {
		return true, err
	}

	w := io.Writer(os.Stdout)
	for _, f := range out.PrimaryFields() {
		if f.Key == "value" {
			fmt.Fprintln(w, f.Value)
		} else {
			fmt.Fprintf(w, "%s: %s\n", f.Key, f.Value)
		}
	}

	if e, ok := g.(forge.Explainer); ok {
		fmt.Fprintln(w)
		fmt.Fprint(w, e.Why())
	} else {
		fmt.Fprintf(w, "\nNo rationale available for %s.\n", forge.Address(g))
	}

	return true, nil
}
