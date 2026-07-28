package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/lemon07r/SlopScoreTool/slop"
)

const mcpInstructions = `Analyze prose for AI-associated writing tendencies: repetitive, canned, or statistically overused language. Results are editorial evidence for review, never proof of AI authorship or writing quality. Weigh repeated and concentrated findings over single hits, read findings in their source context, and keep intentional, quoted, or domain-necessary language.`

const defaultMaxFindings = 10

type analyzeArgs struct {
	Text         string   `json:"text" jsonschema:"the prose to analyze"`
	Name         string   `json:"name,omitempty" jsonschema:"label for the input (default \"text\")"`
	MaxFindings  int      `json:"max_findings,omitempty" jsonschema:"findings per group (default 10)"`
	Full         bool     `json:"full,omitempty" jsonschema:"include all findings with no per-group limit"`
	ExcludeTerms []string `json:"exclude_terms,omitempty" jsonschema:"finding texts to omit; only for vocabulary necessary to the subject"`
}

type compareArgs struct {
	Before      string `json:"before" jsonschema:"the earlier draft"`
	After       string `json:"after" jsonschema:"the revised draft"`
	MaxFindings int    `json:"max_findings,omitempty" jsonschema:"finding deltas to show (default 10)"`
}

// newMCPServer builds the analyzer-backed MCP server. Transport-agnostic so
// tests can connect over in-memory transports.
func newMCPServer() (*mcp.Server, error) {
	analyzer, err := slop.NewAnalyzer()
	if err != nil {
		return nil, err
	}
	server := mcp.NewServer(
		&mcp.Implementation{Name: slop.AnalyzerName, Version: slop.Version},
		&mcp.ServerOptions{Instructions: mcpInstructions},
	)

	mcp.AddTool(server, &mcp.Tool{
		Name: "analyze",
		Description: "Analyze prose for AI-associated writing tendencies. " +
			"Returns a compact report (orientation score, warnings, located findings) " +
			"plus structured brief JSON. Evidence for editorial judgment, not a verdict.",
		Annotations: &mcp.ToolAnnotations{Title: "Analyze prose", ReadOnlyHint: true},
	}, func(_ context.Context, _ *mcp.CallToolRequest, args analyzeArgs) (*mcp.CallToolResult, briefDocument, error) {
		name := args.Name
		if name == "" {
			name = "text"
		}
		opts := slop.Options{
			MaxFindings:  mcpMaxFindings(args.MaxFindings),
			Full:         args.Full,
			ExcludeTerms: args.ExcludeTerms,
		}
		report, err := analyzer.AnalyzeAll([]slop.Input{{Name: name, Text: args.Text}}, opts)
		if err != nil {
			return nil, briefDocument{}, err
		}
		var buf bytes.Buffer
		if err := renderAgentReport(&buf, report); err != nil {
			return nil, briefDocument{}, err
		}
		brief := briefReportFrom(report)
		return textResult(buf.String()), brief.Documents[0], nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "compare",
		Description: "Compare two drafts of prose for changes in AI-associated writing tendencies. " +
			"Returns score, component, and finding deltas without declaring a winner.",
		Annotations: &mcp.ToolAnnotations{Title: "Compare drafts", ReadOnlyHint: true},
	}, func(_ context.Context, _ *mcp.CallToolRequest, args compareArgs) (*mcp.CallToolResult, briefComparison, error) {
		opts := slop.Options{MaxFindings: mcpMaxFindings(args.MaxFindings)}
		before, err := analyzer.Analyze("before", args.Before, opts)
		if err != nil {
			return nil, briefComparison{}, err
		}
		after, err := analyzer.Analyze("after", args.After, opts)
		if err != nil {
			return nil, briefComparison{}, err
		}
		cmp := analyzer.Compare(before, after, mcpMaxFindings(args.MaxFindings))
		var buf bytes.Buffer
		if err := renderAgentComparison(&buf, cmp); err != nil {
			return nil, briefComparison{}, err
		}
		return textResult(buf.String()), briefComparisonFrom(cmp), nil
	})

	return server, nil
}

func mcpMaxFindings(v int) int {
	if v <= 0 {
		return defaultMaxFindings
	}
	return v
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: text}}}
}

const mcpUsage = `sst mcp - serve the analyzer over MCP (stdio transport)

Expose the analyze and compare tools to an MCP-capable agent harness.
Example client configuration: {"mcpServers": {"sst": {"command": "sst", "args": ["mcp"]}}}
`

func runMCP(argv []string, _ io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("mcp", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	help := fs.Bool("help", false, "show help")
	fs.BoolVar(help, "h", false, "show help")
	if err := fs.Parse(argv); err != nil {
		fmt.Fprintln(stderr, "sst mcp:", err)
		return exitUsage
	}
	if *help {
		fmt.Fprint(stdout, mcpUsage)
		return exitOK
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(stderr, "sst mcp: unexpected positional arguments")
		return exitUsage
	}
	server, err := newMCPServer()
	if err != nil {
		fmt.Fprintln(stderr, "sst mcp:", err)
		return exitError
	}
	// StdioTransport owns os.Stdin/os.Stdout; nothing else may write stdout.
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		fmt.Fprintln(stderr, "sst mcp:", err)
		return exitError
	}
	return exitOK
}
