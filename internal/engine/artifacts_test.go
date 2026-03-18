package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func writeArtifact(t *testing.T, root, relPath, content string) {
	t.Helper()
	abs := filepath.Join(root, relPath)
	os.MkdirAll(filepath.Dir(abs), 0o755)
	os.WriteFile(abs, []byte(content), 0o644)
}

// --- DiscoverArtifacts ---

func TestDiscoverArtifacts_FindsLCOV(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeArtifact(t, root, "coverage/lcov.info", "TN:\nSF:src/app.ts\nend_of_record\n")

	d := DiscoverArtifacts(root)
	if d.CoveragePath == "" {
		t.Fatal("expected coverage path to be detected")
	}
	if d.CoverageFormat != "lcov" {
		t.Errorf("format: want lcov, got %s", d.CoverageFormat)
	}
}

func TestDiscoverArtifacts_FindsIstanbulJSON(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeArtifact(t, root, "coverage/coverage-final.json", `{"src/app.ts":{}}`)

	d := DiscoverArtifacts(root)
	if d.CoveragePath == "" {
		t.Fatal("expected coverage path to be detected")
	}
	if d.CoverageFormat != "istanbul" {
		t.Errorf("format: want istanbul, got %s", d.CoverageFormat)
	}
}

func TestDiscoverArtifacts_FindsGoCoverage(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeArtifact(t, root, "coverage.out", "mode: set\npkg/foo.go:10.1,12.1 1 1\n")

	d := DiscoverArtifacts(root)
	if d.CoveragePath == "" {
		t.Fatal("expected Go coverage to be detected")
	}
	if d.CoverageFormat != "go-cover" {
		t.Errorf("format: want go-cover, got %s", d.CoverageFormat)
	}
}

func TestDiscoverArtifacts_FindsJUnitXML(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeArtifact(t, root, "junit.xml", `<?xml version="1.0"?><testsuites><testsuite name="t"/></testsuites>`)

	d := DiscoverArtifacts(root)
	if len(d.RuntimePaths) == 0 {
		t.Fatal("expected runtime path to be detected")
	}
	if d.RuntimeFormats[0] != "junit-xml" {
		t.Errorf("format: want junit-xml, got %s", d.RuntimeFormats[0])
	}
}

func TestDiscoverArtifacts_FindsJestJSON(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeArtifact(t, root, "jest-results.json", `{"testResults":[]}`)

	d := DiscoverArtifacts(root)
	if len(d.RuntimePaths) == 0 {
		t.Fatal("expected Jest JSON to be detected")
	}
	if d.RuntimeFormats[0] != "jest-json" {
		t.Errorf("format: want jest-json, got %s", d.RuntimeFormats[0])
	}
}

func TestDiscoverArtifacts_FindsMultipleRuntime(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeArtifact(t, root, "junit.xml", `<testsuites/>`)
	writeArtifact(t, root, "jest-results.json", `{}`)

	d := DiscoverArtifacts(root)
	if len(d.RuntimePaths) < 2 {
		t.Errorf("expected at least 2 runtime paths, got %d", len(d.RuntimePaths))
	}
}

func TestDiscoverArtifacts_EmptyRepo(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	d := DiscoverArtifacts(root)
	if d.CoveragePath != "" {
		t.Errorf("expected no coverage in empty repo, got %s", d.CoveragePath)
	}
	if len(d.RuntimePaths) != 0 {
		t.Errorf("expected no runtime in empty repo, got %v", d.RuntimePaths)
	}
}

func TestDiscoverArtifacts_SkipsEmptyFiles(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeArtifact(t, root, "coverage/lcov.info", "") // empty file

	d := DiscoverArtifacts(root)
	if d.CoveragePath != "" {
		t.Error("should skip empty coverage files")
	}
}

func TestDiscoverArtifacts_SkipsDirectories(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	// Create a directory named lcov.info (unlikely but defensive).
	os.MkdirAll(filepath.Join(root, "coverage", "lcov.info"), 0o755)

	d := DiscoverArtifacts(root)
	if d.CoveragePath != "" {
		t.Error("should skip directories even if name matches")
	}
}

func TestDiscoverArtifacts_PriorityOrder(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	// Both exist — coverage/lcov.info has higher priority than root lcov.info.
	writeArtifact(t, root, "coverage/lcov.info", "TN:\nSF:a\nend_of_record\n")
	writeArtifact(t, root, "lcov.info", "TN:\nSF:b\nend_of_record\n")

	d := DiscoverArtifacts(root)
	if d.CoveragePath == "" {
		t.Fatal("expected coverage path")
	}
	// Should pick coverage/lcov.info (first in candidate list).
	if !filepath.IsAbs(d.CoveragePath) {
		t.Fatal("expected absolute path")
	}
	if filepath.Base(filepath.Dir(d.CoveragePath)) != "coverage" {
		t.Errorf("expected coverage/ directory to win priority, got %s", d.CoveragePath)
	}
}

// --- ApplyDiscovery ---

func TestApplyDiscovery_ExplicitFlagTakesPrecedence(t *testing.T) {
	t.Parallel()
	opts := PipelineOptions{
		CoveragePath: "/explicit/coverage.lcov",
		RuntimePaths: []string{"/explicit/junit.xml"},
	}
	discovery := &ArtifactDiscovery{
		CoveragePath: "/auto/coverage/lcov.info",
		RuntimePaths: []string{"/auto/junit.xml"},
	}

	messages := ApplyDiscovery(&opts, discovery)

	// Explicit flags should not be overridden.
	if opts.CoveragePath != "/explicit/coverage.lcov" {
		t.Errorf("explicit coverage overridden: %s", opts.CoveragePath)
	}
	if opts.RuntimePaths[0] != "/explicit/junit.xml" {
		t.Errorf("explicit runtime overridden: %s", opts.RuntimePaths[0])
	}
	if len(messages) != 0 {
		t.Errorf("expected no messages when explicit flags set, got %v", messages)
	}
	if discovery.CoverageAutoDetected {
		t.Error("should not mark as auto-detected when explicit flag present")
	}
}

func TestApplyDiscovery_AppliesWhenNoExplicitFlags(t *testing.T) {
	t.Parallel()
	opts := PipelineOptions{}
	discovery := &ArtifactDiscovery{
		CoveragePath:   "/auto/coverage/lcov.info",
		CoverageFormat: "lcov",
		RuntimePaths:   []string{"/auto/junit.xml"},
		RuntimeFormats: []string{"junit-xml"},
	}

	messages := ApplyDiscovery(&opts, discovery)

	if opts.CoveragePath != "/auto/coverage/lcov.info" {
		t.Errorf("expected auto-discovered coverage, got %s", opts.CoveragePath)
	}
	if len(opts.RuntimePaths) != 1 || opts.RuntimePaths[0] != "/auto/junit.xml" {
		t.Errorf("expected auto-discovered runtime, got %v", opts.RuntimePaths)
	}
	if len(messages) != 2 {
		t.Errorf("expected 2 messages, got %d: %v", len(messages), messages)
	}
	if !discovery.CoverageAutoDetected {
		t.Error("expected CoverageAutoDetected=true")
	}
	if !discovery.RuntimeAutoDetected {
		t.Error("expected RuntimeAutoDetected=true")
	}
}

func TestApplyDiscovery_NilDiscovery(t *testing.T) {
	t.Parallel()
	opts := PipelineOptions{}
	messages := ApplyDiscovery(&opts, nil)
	if len(messages) != 0 {
		t.Errorf("expected no messages for nil discovery, got %v", messages)
	}
}

// --- MissingArtifactHints ---

func TestMissingArtifactHints_BothMissing(t *testing.T) {
	t.Parallel()
	opts := PipelineOptions{}
	hints := MissingArtifactHints(&opts, &ArtifactDiscovery{})
	if len(hints) != 2 {
		t.Errorf("expected 2 hints when both missing, got %d", len(hints))
	}
}

func TestMissingArtifactHints_CoverageProvided(t *testing.T) {
	t.Parallel()
	opts := PipelineOptions{CoveragePath: "/some/lcov.info"}
	hints := MissingArtifactHints(&opts, &ArtifactDiscovery{})
	if len(hints) != 1 {
		t.Errorf("expected 1 hint (runtime missing), got %d", len(hints))
	}
}

func TestMissingArtifactHints_AllProvided(t *testing.T) {
	t.Parallel()
	opts := PipelineOptions{
		CoveragePath: "/some/lcov.info",
		RuntimePaths: []string{"/some/junit.xml"},
	}
	hints := MissingArtifactHints(&opts, &ArtifactDiscovery{})
	if len(hints) != 0 {
		t.Errorf("expected 0 hints when all provided, got %d", len(hints))
	}
}

// --- Determinism ---

func TestDiscoverArtifacts_Deterministic(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeArtifact(t, root, "coverage/lcov.info", "TN:\nend_of_record\n")
	writeArtifact(t, root, "junit.xml", "<testsuites/>")

	d1 := DiscoverArtifacts(root)
	d2 := DiscoverArtifacts(root)

	if d1.CoveragePath != d2.CoveragePath {
		t.Error("coverage path not deterministic")
	}
	if len(d1.RuntimePaths) != len(d2.RuntimePaths) {
		t.Error("runtime path count not deterministic")
	}
}
