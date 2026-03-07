package policy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_ValidFile(t *testing.T) {
	dir := t.TempDir()
	hamletDir := filepath.Join(dir, ".hamlet")
	os.MkdirAll(hamletDir, 0o755)

	content := `rules:
  disallow_skipped_tests: true
  disallow_frameworks:
    - jest
    - mocha
  max_test_runtime_ms: 5000
  minimum_coverage_percent: 80
  max_weak_assertions: 5
  max_mock_heavy_tests: 3
`
	os.WriteFile(filepath.Join(hamletDir, "policy.yaml"), []byte(content), 0o644)

	result, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Found {
		t.Fatal("expected Found=true")
	}
	if result.Config == nil {
		t.Fatal("expected non-nil Config")
	}
	if result.Config.Rules.DisallowSkippedTests == nil || !*result.Config.Rules.DisallowSkippedTests {
		t.Error("expected disallow_skipped_tests=true")
	}
	if len(result.Config.Rules.DisallowFrameworks) != 2 {
		t.Errorf("expected 2 disallowed frameworks, got %d", len(result.Config.Rules.DisallowFrameworks))
	}
	if result.Config.Rules.MaxTestRuntimeMs == nil || *result.Config.Rules.MaxTestRuntimeMs != 5000 {
		t.Error("expected max_test_runtime_ms=5000")
	}
	if result.Config.Rules.MinimumCoveragePercent == nil || *result.Config.Rules.MinimumCoveragePercent != 80 {
		t.Error("expected minimum_coverage_percent=80")
	}
	if result.Config.Rules.MaxWeakAssertions == nil || *result.Config.Rules.MaxWeakAssertions != 5 {
		t.Error("expected max_weak_assertions=5")
	}
	if result.Config.Rules.MaxMockHeavyTests == nil || *result.Config.Rules.MaxMockHeavyTests != 3 {
		t.Error("expected max_mock_heavy_tests=3")
	}
}

func TestLoad_PartialFile(t *testing.T) {
	dir := t.TempDir()
	hamletDir := filepath.Join(dir, ".hamlet")
	os.MkdirAll(hamletDir, 0o755)

	content := `rules:
  disallow_frameworks:
    - jest
`
	os.WriteFile(filepath.Join(hamletDir, "policy.yaml"), []byte(content), 0o644)

	result, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Found {
		t.Fatal("expected Found=true")
	}
	if result.Config.Rules.DisallowSkippedTests != nil {
		t.Error("expected disallow_skipped_tests to be nil (unset)")
	}
	if len(result.Config.Rules.DisallowFrameworks) != 1 {
		t.Errorf("expected 1 disallowed framework, got %d", len(result.Config.Rules.DisallowFrameworks))
	}
	if result.Config.Rules.MaxTestRuntimeMs != nil {
		t.Error("expected max_test_runtime_ms to be nil (unset)")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	dir := t.TempDir()

	result, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Found {
		t.Error("expected Found=false for missing file")
	}
	if result.Config != nil {
		t.Error("expected nil Config for missing file")
	}
}

func TestLoad_MalformedFile(t *testing.T) {
	dir := t.TempDir()
	hamletDir := filepath.Join(dir, ".hamlet")
	os.MkdirAll(hamletDir, 0o755)

	content := `rules:
  disallow_frameworks: [[[invalid yaml
`
	os.WriteFile(filepath.Join(hamletDir, "policy.yaml"), []byte(content), 0o644)

	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error for malformed YAML")
	}
}

func TestConfig_IsEmpty(t *testing.T) {
	cfg := &Config{}
	if !cfg.IsEmpty() {
		t.Error("empty config should report IsEmpty=true")
	}

	trueVal := true
	cfg.Rules.DisallowSkippedTests = &trueVal
	if cfg.IsEmpty() {
		t.Error("config with a rule set should report IsEmpty=false")
	}
}
