package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lemon07r/SlopScoreTool/slop"
)

func runCLI(t *testing.T, argv []string, stdin string) (code int, stdout, stderr string) {
	t.Helper()
	var outBuf, errBuf bytes.Buffer
	code = run(argv, strings.NewReader(stdin), &outBuf, &errBuf)
	return code, outBuf.String(), errBuf.String()
}

func writeTemp(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

const sampleProse = `The rain did not stop. It was not just weather, but a verdict on the season.
Mira counted the umbrellas twice, then a third time for luck. Nothing changed.
It was not courage that kept her there. It was stubbornness, plain and simple.`

func TestStdinDefault(t *testing.T) {
	code, stdout, _ := runCLI(t, nil, sampleProse)
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(stdout, "== stdin ==") {
		t.Errorf("missing stdin header:\n%s", stdout)
	}
	if !strings.Contains(stdout, "slop score:") {
		t.Errorf("missing score line:\n%s", stdout)
	}
}

func TestJSONModeCleanStreams(t *testing.T) {
	code, stdout, stderr := runCLI(t, []string{"--format", "json"}, sampleProse)
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if stderr != "" {
		t.Errorf("stderr not clean: %q", stderr)
	}
	var report slop.Report
	if err := json.Unmarshal([]byte(stdout), &report); err != nil {
		t.Fatalf("stdout is not clean JSON: %v\n%s", err, stdout)
	}
	if report.SchemaVersion != 1 {
		t.Errorf("schema_version = %d", report.SchemaVersion)
	}
	if len(report.Documents) != 1 {
		t.Fatalf("documents = %d", len(report.Documents))
	}
	doc := report.Documents[0]
	if doc.Score == nil {
		t.Fatal("score is nil for valid prose")
	}
	if doc.Score.Basis != slop.CompatibilityStage1Fallback {
		t.Errorf("basis = %q", doc.Score.Basis)
	}
	if report.Method.Compatibility != slop.CompatibilityStage1Fallback {
		t.Errorf("method.compatibility = %q", report.Method.Compatibility)
	}
}

func TestFileInput(t *testing.T) {
	path := writeTemp(t, "draft.txt", sampleProse)
	code, stdout, _ := runCLI(t, []string{path}, "")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(stdout, "== "+path+" ==") {
		t.Errorf("missing file header:\n%s", stdout)
	}
}

func TestMultiFileCorpus(t *testing.T) {
	a := writeTemp(t, "a.txt", sampleProse)
	b := writeTemp(t, "b.txt", "The quiet road went on. Nothing moved for an hour.")
	code, stdout, _ := runCLI(t, []string{"--format", "json", a, b}, "")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	var report slop.Report
	if err := json.Unmarshal([]byte(stdout), &report); err != nil {
		t.Fatal(err)
	}
	if len(report.Documents) != 2 {
		t.Fatalf("documents = %d", len(report.Documents))
	}
	if report.Corpus == nil {
		t.Fatal("corpus missing for multi-file input")
	}
	if report.Corpus.Documents != 2 {
		t.Errorf("corpus.documents = %d", report.Corpus.Documents)
	}
	if report.Corpus.Score == nil {
		t.Error("corpus score summary missing")
	}
}

func TestCompare(t *testing.T) {
	before := writeTemp(t, "before.txt", "It was not just late, but completely forgotten. The silence was not empty. It was full.")
	after := writeTemp(t, "after.txt", "The train ran three hours late. The platform was quiet.")
	code, stdout, _ := runCLI(t, []string{"--compare", before, after}, "")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(stdout, "compare:") {
		t.Errorf("missing compare header:\n%s", stdout)
	}
	if !strings.Contains(stdout, "orientation only") {
		t.Errorf("compare should disclaim number-chasing:\n%s", stdout)
	}
}

func TestCompareArity(t *testing.T) {
	one := writeTemp(t, "one.txt", sampleProse)
	if code, _, _ := runCLI(t, []string{"--compare", one}, ""); code != 2 {
		t.Errorf("compare with one file: exit %d, want 2", code)
	}
}

func TestUsageErrors(t *testing.T) {
	cases := [][]string{
		{"--format", "yaml"},
		{"--max-findings", "-1"},
		{"--bogus-flag"},
	}
	for _, argv := range cases {
		if code, _, _ := runCLI(t, argv, "x"); code != 2 {
			t.Errorf("%v: exit %d, want 2", argv, code)
		}
	}
}

func TestUnreadableFile(t *testing.T) {
	code, _, stderr := runCLI(t, []string{filepath.Join(t.TempDir(), "missing.txt")}, "")
	if code != 1 {
		t.Errorf("exit %d, want 1", code)
	}
	if stderr == "" {
		t.Error("expected diagnostic on stderr")
	}
}

func TestEmptyInput(t *testing.T) {
	code, stdout, _ := runCLI(t, []string{"--format", "json"}, "")
	if code != 0 {
		t.Fatalf("exit %d (empty input is a successful analysis)", code)
	}
	var report slop.Report
	if err := json.Unmarshal([]byte(stdout), &report); err != nil {
		t.Fatal(err)
	}
	doc := report.Documents[0]
	if doc.Score != nil {
		t.Error("score should be null for empty input")
	}
	if len(doc.Warnings) == 0 {
		t.Error("expected empty-input warning")
	}
}

func TestFindingLimitAndFull(t *testing.T) {
	var b strings.Builder
	for range 30 {
		b.WriteString("The silence was not empty. It was full. ")
	}
	text := b.String()

	_, limitedOut, _ := runCLI(t, []string{"--format", "json", "--max-findings", "1"}, text)
	var limited slop.Report
	if err := json.Unmarshal([]byte(limitedOut), &limited); err != nil {
		t.Fatal(err)
	}
	if limited.Documents[0].Truncation == nil {
		t.Fatal("expected truncation metadata with --max-findings 1")
	}

	_, fullOut, _ := runCLI(t, []string{"--format", "json", "--full"}, text)
	var full slop.Report
	if err := json.Unmarshal([]byte(fullOut), &full); err != nil {
		t.Fatal(err)
	}
	if full.Documents[0].Truncation != nil {
		t.Error("--full should not truncate")
	}
	if len(full.Documents[0].Findings) <= len(limited.Documents[0].Findings) {
		t.Error("--full should surface more findings than the limited run")
	}
}

func TestExcludeTerm(t *testing.T) {
	_, out, _ := runCLI(t, []string{"--format", "json", "--exclude-term", "silence"}, sampleProse)
	var report slop.Report
	if err := json.Unmarshal([]byte(out), &report); err != nil {
		t.Fatal(err)
	}
	for _, f := range report.Documents[0].Findings {
		if strings.EqualFold(f.Text, "silence") {
			t.Errorf("excluded term still present: %+v", f)
		}
	}
}

func TestDeterminism(t *testing.T) {
	_, first, _ := runCLI(t, []string{"--format", "json"}, sampleProse)
	_, second, _ := runCLI(t, []string{"--format", "json"}, sampleProse)
	if first != second {
		t.Error("repeated runs differ; output must be deterministic")
	}
}

func TestAgentFormat(t *testing.T) {
	code, stdout, stderr := runCLI(t, []string{"--format", "agent"}, sampleProse)
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if stderr != "" {
		t.Errorf("stderr not clean: %q", stderr)
	}
	for _, want := range []string{"sst ", "== stdin ==", "score:", "stats:", "guide:"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("agent output missing %q:\n%s", want, stdout)
		}
	}
	if strings.Contains(stdout, "lists_sha256") || strings.Contains(stdout, "schema_version") {
		t.Errorf("agent output leaked full-contract detail:\n%s", stdout)
	}
	_, second, _ := runCLI(t, []string{"--format", "agent"}, sampleProse)
	if stdout != second {
		t.Error("agent output must be deterministic")
	}
}

func TestAgentCompare(t *testing.T) {
	before := writeTemp(t, "before.txt", "It was not just late, but completely forgotten. The silence was not empty. It was full.")
	after := writeTemp(t, "after.txt", "The train ran three hours late. The platform was quiet.")
	code, stdout, _ := runCLI(t, []string{"--format", "agent", "--compare", before, after}, "")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(stdout, "== compare:") || !strings.Contains(stdout, "no winner declared") {
		t.Errorf("unexpected agent compare output:\n%s", stdout)
	}
}

func TestBriefJSON(t *testing.T) {
	code, stdout, _ := runCLI(t, []string{"--format", "brief"}, sampleProse)
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	var brief briefReport
	if err := json.Unmarshal([]byte(stdout), &brief); err != nil {
		t.Fatalf("stdout is not clean JSON: %v\n%s", err, stdout)
	}
	if brief.SchemaVersion != slop.SchemaVersion {
		t.Errorf("schema_version = %d", brief.SchemaVersion)
	}
	if brief.Compatibility != slop.CompatibilityStage1Fallback {
		t.Errorf("compatibility = %q", brief.Compatibility)
	}
	if len(brief.Documents) != 1 || brief.Documents[0].Score == nil {
		t.Fatal("brief report missing scored document")
	}
	for _, leaked := range []string{"lists_sha256", "calibration_range", "excerpt", "signals"} {
		if strings.Contains(stdout, leaked) {
			t.Errorf("brief output leaked %q:\n%s", leaked, stdout)
		}
	}
}

func TestSkillPrint(t *testing.T) {
	code, stdout, stderr := runCLI(t, []string{"skill"}, "")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, stderr)
	}
	if !strings.Contains(stdout, "name: review-writing-slop") {
		t.Errorf("skill output missing frontmatter:\n%.200s", stdout)
	}
}

func TestSkillInstall(t *testing.T) {
	dir := t.TempDir()
	code, stdout, _ := runCLI(t, []string{"skill", "--install", dir}, "")
	if code != 0 {
		t.Fatalf("install exit %d", code)
	}
	target := filepath.Join(dir, "review-writing-slop", "SKILL.md")
	if !strings.Contains(stdout, target) {
		t.Errorf("install output should name the written file: %q", stdout)
	}
	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != string(skillDoc) {
		t.Error("installed file differs from embedded skill")
	}
	if _, err := os.Stat(filepath.Join(dir, "review-writing-slop", "agents", "openai.yaml")); err != nil {
		t.Error("auxiliary skill files were not installed:", err)
	}

	if code, _, _ := runCLI(t, []string{"skill", "--install", dir}, ""); code != 0 {
		t.Errorf("reinstall of identical file: exit %d, want 0", code)
	}

	if err := os.WriteFile(target, []byte("local edits"), 0o600); err != nil {
		t.Fatal(err)
	}
	if code, _, _ := runCLI(t, []string{"skill", "--install", dir}, ""); code != 1 {
		t.Errorf("conflicting file without --force: exit %d, want 1", code)
	}
	if code, _, _ := runCLI(t, []string{"skill", "--install", dir, "--force"}, ""); code != 0 {
		t.Errorf("conflicting file with --force: exit %d, want 0", code)
	}
}

func TestVersion(t *testing.T) {
	code, stdout, _ := runCLI(t, []string{"--version"}, "")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(stdout, slop.Version) || !strings.Contains(stdout, slop.CompatibilityStage1Fallback) {
		t.Errorf("unexpected version output: %q", stdout)
	}
}

func TestInvalidUTF8(t *testing.T) {
	code, stdout, _ := runCLI(t, []string{"--format", "json"}, "valid text \xff\xfe more text.")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	var report slop.Report
	if err := json.Unmarshal([]byte(stdout), &report); err != nil {
		t.Fatalf("invalid UTF-8 broke JSON output: %v", err)
	}
	found := false
	for _, w := range report.Documents[0].Warnings {
		if strings.HasPrefix(w, "invalid_utf8") {
			found = true
		}
	}
	if !found {
		t.Error("expected invalid_utf8 warning")
	}
}
