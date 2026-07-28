package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/lemon07r/SlopScoreTool/slop"
)

const usageText = `sst - analyze prose for AI-associated writing tendencies

Usage:
  sst [flags] [FILE ...]      analyze files, or stdin when none are given
  sst --compare BEFORE AFTER  compare two drafts

Flags:
  --format text|json   output format (default "text")
  --max-findings N     findings per group in the report (default 10)
  --full               include all findings, no per-group limit
  --exclude-term TERM  omit findings whose text equals TERM (repeatable)
  --compare            compare exactly two drafts (positional BEFORE AFTER)
  --version            print version information
  -h, --help           show this help

Notes:
  A score orients review but never proves AI authorship or writing quality.
  Adverse findings are a successful analysis: the exit code stays 0.
  JSON goes to stdout; diagnostics go to stderr.
`

type cliOptions struct {
	format       string
	maxFindings  int
	full         bool
	excludeTerms []string
	compare      bool
	version      bool
	help         bool
	files        []string
}

// stringList collects repeatable flags.
type stringList []string

func (s *stringList) String() string { return strings.Join(*s, ",") }
func (s *stringList) Set(v string) error {
	*s = append(*s, v)
	return nil
}

func parseFlags(argv []string) (cliOptions, error) {
	opts := cliOptions{format: "text", maxFindings: 10}
	fs := flag.NewFlagSet("sst", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&opts.format, "format", "text", "output format: text or json")
	fs.IntVar(&opts.maxFindings, "max-findings", 10, "findings per group")
	fs.BoolVar(&opts.full, "full", false, "no finding limit")
	fs.Var((*stringList)(&opts.excludeTerms), "exclude-term", "omit a finding text (repeatable)")
	fs.BoolVar(&opts.compare, "compare", false, "compare two drafts")
	fs.BoolVar(&opts.version, "version", false, "print version")
	fs.BoolVar(&opts.help, "help", false, "show help")
	fs.BoolVar(&opts.help, "h", false, "show help")
	if err := fs.Parse(argv); err != nil {
		return opts, err
	}
	opts.files = fs.Args()
	switch opts.format {
	case "text", "json":
	default:
		return opts, fmt.Errorf("unknown --format %q (want text or json)", opts.format)
	}
	if opts.maxFindings < 0 {
		return opts, fmt.Errorf("--max-findings must be >= 0")
	}
	if opts.compare && len(opts.files) != 2 {
		return opts, &usageError{msg: "--compare needs exactly two files (BEFORE AFTER)"}
	}
	stdinRefs := 0
	for _, f := range opts.files {
		if f == "-" {
			stdinRefs++
		}
	}
	if stdinRefs > 1 {
		return opts, &usageError{msg: "stdin ('-') can be used at most once"}
	}
	return opts, nil
}

// readInputs resolves positional files (or default stdin) into named texts.
func readInputs(files []string, stdin io.Reader) ([]slop.Input, error) {
	if len(files) == 0 {
		body, err := io.ReadAll(stdin)
		if err != nil {
			return nil, fmt.Errorf("read stdin: %w", err)
		}
		return []slop.Input{{Name: "stdin", Text: string(body)}}, nil
	}
	inputs := make([]slop.Input, 0, len(files))
	for _, name := range files {
		var body []byte
		var err error
		if name == "-" {
			body, err = io.ReadAll(stdin)
			name = "stdin"
		} else {
			body, err = os.ReadFile(name)
		}
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", name, err)
		}
		inputs = append(inputs, slop.Input{Name: name, Text: string(body)})
	}
	return inputs, nil
}

func execute(opts cliOptions, stdin io.Reader, stdout io.Writer) error {
	analyzer, err := slop.NewAnalyzer()
	if err != nil {
		return fmt.Errorf("load analyzer: %w", err)
	}
	inputs, err := readInputs(opts.files, stdin)
	if err != nil {
		return err
	}
	analyzeOpts := slop.Options{
		MaxFindings:  opts.maxFindings,
		Full:         opts.full,
		ExcludeTerms: opts.excludeTerms,
	}

	if opts.compare {
		before, err := analyzer.Analyze(inputs[0].Name, inputs[0].Text, analyzeOpts)
		if err != nil {
			return err
		}
		after, err := analyzer.Analyze(inputs[1].Name, inputs[1].Text, analyzeOpts)
		if err != nil {
			return err
		}
		cmp := analyzer.Compare(before, after, opts.maxFindings)
		return renderComparison(stdout, opts.format, cmp)
	}

	report, err := analyzer.AnalyzeAll(inputs, analyzeOpts)
	if err != nil {
		return err
	}
	return renderReport(stdout, opts.format, report)
}
