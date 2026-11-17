// Package main provides the ffparity CLI tool for verifying feature-flag API parity.
//
// Usage:
//
//	ffparity <config-path>
//
// Output format (stable for CI grepping):
//
//	PASS                                       – zero divergences; exit 0
//	FAIL: N divergences                        – summary line when N > 0; exit non-zero
//	DIVERGE feature=<f> provider=<p> ...       – one detail line per diverging context
//
// Errors (missing or malformed config) are printed to stderr; exit non-zero.
// The PASS and FAIL: summary lines are stable and grep-safe; do not add decoration.
package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/kyma-project/cloud-manager/pkg/feature"
)

func main() {
	os.Exit(run(os.Args, os.Stdout, os.Stderr))
}

// run is the testable entrypoint. It returns the process exit code.
// stdout receives PASS / FAIL / DIVERGE lines; stderr receives error messages.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) < 2 {
		fmt.Fprintln(stderr, "usage: ffparity <config-path>")
		return 1
	}
	configPath := args[1]

	divergences, err := feature.CheckApiParity(context.Background(), configPath)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	if len(divergences) == 0 {
		fmt.Fprintln(stdout, "PASS")
		return 0
	}

	fmt.Fprintf(stdout, "FAIL: %d divergences\n", len(divergences))
	for _, d := range divergences {
		fmt.Fprintf(stdout, "DIVERGE %s\n", d.String())
	}
	return 1
}
