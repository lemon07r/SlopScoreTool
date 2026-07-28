package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func connectMCP(t *testing.T) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()
	server, err := newMCPServer()
	if err != nil {
		t.Fatal(err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0"}, nil)
	st, ct := mcp.NewInMemoryTransports()
	if _, err := server.Connect(ctx, st, nil); err != nil {
		t.Fatal(err)
	}
	cs, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cs.Close() })
	return cs
}

func callTool(t *testing.T, cs *mcp.ClientSession, name string, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("call %s: %v", name, err)
	}
	if res.IsError {
		t.Fatalf("%s returned a tool error: %+v", name, res.Content)
	}
	return res
}

func resultText(t *testing.T, res *mcp.CallToolResult) string {
	t.Helper()
	if len(res.Content) != 1 {
		t.Fatalf("content items = %d, want 1", len(res.Content))
	}
	text, ok := res.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("content is %T, want TextContent", res.Content[0])
	}
	return text.Text
}

func TestMCPToolsListed(t *testing.T) {
	cs := connectMCP(t)
	res, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, tool := range res.Tools {
		names[tool.Name] = true
	}
	for _, want := range []string{"analyze", "compare"} {
		if !names[want] {
			t.Errorf("tool %q not listed (got %v)", want, names)
		}
	}
}

func TestMCPAnalyze(t *testing.T) {
	cs := connectMCP(t)
	res := callTool(t, cs, "analyze", map[string]any{"text": sampleProse})

	text := resultText(t, res)
	if !strings.Contains(text, "score:") || !strings.Contains(text, "== text ==") {
		t.Errorf("unexpected analyze text content:\n%s", text)
	}

	raw, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	var brief briefDocument
	if err := json.Unmarshal(raw, &brief); err != nil {
		t.Fatalf("structured content is not a brief document: %v", err)
	}
	if brief.Score == nil {
		t.Fatal("structured content missing score")
	}
	if brief.Score.Basis != "eqbench-stage1-fallback" {
		t.Errorf("basis = %q", brief.Score.Basis)
	}
}

func TestMCPCompare(t *testing.T) {
	cs := connectMCP(t)
	res := callTool(t, cs, "compare", map[string]any{
		"before": "It was not just late, but completely forgotten. The silence was not empty. It was full.",
		"after":  "The train ran three hours late. The platform was quiet.",
	})

	text := resultText(t, res)
	if !strings.Contains(text, "== compare:") || !strings.Contains(text, "no winner declared") {
		t.Errorf("unexpected compare text content:\n%s", text)
	}

	raw, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	var brief briefComparison
	if err := json.Unmarshal(raw, &brief); err != nil {
		t.Fatalf("structured content is not a brief comparison: %v", err)
	}
	if brief.Deltas.Score == nil {
		t.Error("structured content missing score delta")
	}
}

func TestMCPAnalyzeRequiresText(t *testing.T) {
	cs := connectMCP(t)
	// Schema validation rejects the call before the handler; the SDK delivers
	// that as a tool-error result rather than a protocol error.
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "analyze", Arguments: map[string]any{}})
	if err == nil && (res == nil || !res.IsError) {
		t.Error("analyze without required text should fail schema validation")
	}
}
