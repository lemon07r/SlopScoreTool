// Command sst analyzes prose for AI-associated writing tendencies
// and prints structured evidence as text or JSON. Findings are a successful
// analysis: nonzero exits are reserved for usage and operational failures.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/lemon07r/SlopScoreTool/slop"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

const (
	exitOK    = 0
	exitError = 1
	exitUsage = 2
)

// usageError distinguishes invalid invocation from operational failure.
type usageError struct{ msg string }

func (e *usageError) Error() string { return e.msg }

func run(argv []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(argv) > 0 {
		switch argv[0] {
		case "skill":
			return runSkill(argv[1:], stdout, stderr)
		case "mcp":
			return runMCP(argv[1:], stdin, stdout, stderr)
		}
	}
	opts, err := parseFlags(argv)
	if err != nil {
		fmt.Fprintln(stderr, "sst:", err)
		fmt.Fprintln(stderr, "try 'sst --help'")
		return exitUsage
	}
	if opts.help {
		fmt.Fprint(stdout, usageText)
		return exitOK
	}
	if opts.version {
		fmt.Fprintf(stdout, "%s %s (schema v%d, %s)\n",
			slop.AnalyzerName, slop.Version, slop.SchemaVersion, slop.CompatibilityStage1Fallback)
		return exitOK
	}
	if err := execute(opts, stdin, stdout); err != nil {
		var uErr *usageError
		if errors.As(err, &uErr) {
			fmt.Fprintln(stderr, "sst:", uErr)
			fmt.Fprintln(stderr, "try 'sst --help'")
			return exitUsage
		}
		fmt.Fprintln(stderr, "sst:", err)
		return exitError
	}
	return exitOK
}
