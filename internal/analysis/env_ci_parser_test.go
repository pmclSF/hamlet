package analysis

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseGitHubActions_MatrixStrategy(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	wfDir := filepath.Join(root, ".github", "workflows")
	os.MkdirAll(wfDir, 0o755)
	os.WriteFile(filepath.Join(wfDir, "ci.yml"), []byte(`
name: CI
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        node-version: [18, 20, 22]
        os: [ubuntu-latest, macos-latest, windows-latest]
    steps:
      - uses: actions/checkout@v4
      - run: npm test
`), 0o644)

	result := ParseCIMatrices(root)

	// Should have node-version and os classes.
	if len(result.EnvironmentClasses) < 2 {
		t.Fatalf("expected at least 2 environment classes, got %d", len(result.EnvironmentClasses))
	}

	classNames := map[string]string{} // classID → dimension
	for _, cls := range result.EnvironmentClasses {
		classNames[cls.ClassID] = cls.Dimension
	}

	if dim, ok := classNames["envclass:gha-node-version"]; !ok {
		t.Error("expected node-version class")
	} else if dim != "runtime" {
		t.Errorf("node-version dimension: want runtime, got %s", dim)
	}

	if dim, ok := classNames["envclass:gha-os"]; !ok {
		t.Error("expected os class")
	} else if dim != "os" {
		t.Errorf("os dimension: want os, got %s", dim)
	}

	// Check member counts.
	for _, cls := range result.EnvironmentClasses {
		if cls.ClassID == "envclass:gha-node-version" && len(cls.MemberIDs) != 3 {
			t.Errorf("node-version: want 3 members, got %d", len(cls.MemberIDs))
		}
		if cls.ClassID == "envclass:gha-os" && len(cls.MemberIDs) != 3 {
			t.Errorf("os: want 3 members, got %d", len(cls.MemberIDs))
		}
	}

	// Verify environments have OS inferred.
	for _, env := range result.Environments {
		if env.EnvironmentID == "env:gha-os-ubuntu-latest" {
			if env.OS != "linux" {
				t.Errorf("ubuntu-latest: want OS linux, got %s", env.OS)
			}
		}
		if env.EnvironmentID == "env:gha-os-macos-latest" {
			if env.OS != "macos" {
				t.Errorf("macos-latest: want OS macos, got %s", env.OS)
			}
		}
	}

	// Verify provenance.
	for _, env := range result.Environments {
		if env.CIProvider != "github-actions" {
			t.Errorf("env %s: want provider github-actions, got %s", env.EnvironmentID, env.CIProvider)
		}
	}
}

func TestParseGitHubActions_RunsOnOnly(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	wfDir := filepath.Join(root, ".github", "workflows")
	os.MkdirAll(wfDir, 0o755)
	os.WriteFile(filepath.Join(wfDir, "build.yml"), []byte(`
name: Build
on: [push]
jobs:
  build:
    runs-on: ubuntu-22.04
    steps:
      - run: echo hello
`), 0o644)

	result := ParseCIMatrices(root)

	found := false
	for _, env := range result.Environments {
		if env.Name == "ubuntu-22.04" {
			found = true
			if env.OS != "linux" {
				t.Errorf("want OS linux, got %s", env.OS)
			}
		}
	}
	if !found {
		t.Error("expected runs-on environment to be detected")
	}
}

func TestParseGitLabCI_ParallelMatrix(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	os.WriteFile(filepath.Join(root, ".gitlab-ci.yml"), []byte(`
stages:
  - test

test:
  stage: test
  parallel:
    matrix:
      - PYTHON_VERSION: ["3.10", "3.11", "3.12"]
        PLATFORM: [linux, macos]
  script:
    - pytest
`), 0o644)

	result := ParseCIMatrices(root)

	if len(result.EnvironmentClasses) < 2 {
		t.Fatalf("expected at least 2 classes, got %d", len(result.EnvironmentClasses))
	}

	for _, cls := range result.EnvironmentClasses {
		if cls.ClassID == "envclass:gitlab-python-version" {
			if len(cls.MemberIDs) != 3 {
				t.Errorf("PYTHON_VERSION: want 3 members, got %d", len(cls.MemberIDs))
			}
			if cls.Dimension != "runtime" {
				t.Errorf("PYTHON_VERSION: want dimension runtime, got %s", cls.Dimension)
			}
		}
	}

	for _, env := range result.Environments {
		if env.CIProvider != "gitlab-ci" {
			t.Errorf("env %s: want provider gitlab-ci, got %s", env.EnvironmentID, env.CIProvider)
		}
	}
}

func TestParseCircleCI_WorkflowMatrix(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	os.MkdirAll(filepath.Join(root, ".circleci"), 0o755)
	os.WriteFile(filepath.Join(root, ".circleci", "config.yml"), []byte(`
version: 2.1
jobs:
  test:
    parameters:
      node-version:
        type: string
    docker:
      - image: node:<< parameters.node-version >>
    steps:
      - run: npm test

workflows:
  test-all:
    jobs:
      - test:
          matrix:
            parameters:
              node-version: ["18", "20", "22"]
`), 0o644)

	result := ParseCIMatrices(root)

	found := false
	for _, cls := range result.EnvironmentClasses {
		if cls.ClassID == "envclass:circleci-node-version" {
			found = true
			if len(cls.MemberIDs) != 3 {
				t.Errorf("node-version: want 3 members, got %d", len(cls.MemberIDs))
			}
			if cls.Dimension != "runtime" {
				t.Errorf("want dimension runtime, got %s", cls.Dimension)
			}
		}
	}
	if !found {
		t.Error("expected CircleCI node-version class")
	}
}

func TestParseBuildkite_MatrixSetup(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	os.MkdirAll(filepath.Join(root, ".buildkite"), 0o755)
	os.WriteFile(filepath.Join(root, ".buildkite", "pipeline.yml"), []byte(`
steps:
  - label: "Test"
    command: "make test"
    matrix:
      setup:
        os: [linux, macos, windows]
        runtime: [node-18, node-20]
`), 0o644)

	result := ParseCIMatrices(root)

	if len(result.EnvironmentClasses) < 2 {
		t.Fatalf("expected at least 2 classes, got %d", len(result.EnvironmentClasses))
	}

	for _, cls := range result.EnvironmentClasses {
		if cls.ClassID == "envclass:buildkite-os" {
			if len(cls.MemberIDs) != 3 {
				t.Errorf("os: want 3 members, got %d", len(cls.MemberIDs))
			}
		}
		if cls.ClassID == "envclass:buildkite-runtime" {
			if len(cls.MemberIDs) != 2 {
				t.Errorf("runtime: want 2 members, got %d", len(cls.MemberIDs))
			}
		}
	}
}

func TestParseCIMatrices_NoCI(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	result := ParseCIMatrices(root)
	if len(result.Environments) != 0 {
		t.Errorf("expected 0 environments, got %d", len(result.Environments))
	}
}

func TestSanitizeID(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input string
		want  string
	}{
		{"ubuntu-latest", "ubuntu-latest"},
		{"Node 22.x", "node-22.x"},
		{"3.12", "3.12"},
		{"windows/2022", "windows-2022"},
		{"  spaces  ", "spaces"},
	}
	for _, tc := range cases {
		got := sanitizeID(tc.input)
		if got != tc.want {
			t.Errorf("sanitizeID(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestInferDimension(t *testing.T) {
	t.Parallel()
	cases := []struct {
		key  string
		want string
	}{
		{"os", "os"},
		{"node-version", "runtime"},
		{"python_version", "runtime"},
		{"browser", "browser"},
		{"device", "device"},
		{"arch", "architecture"},
		{"custom-key", "custom-key"},
	}
	for _, tc := range cases {
		got := inferDimension(tc.key)
		if got != tc.want {
			t.Errorf("inferDimension(%q) = %q, want %q", tc.key, got, tc.want)
		}
	}
}

func TestInferOSFromRunner(t *testing.T) {
	t.Parallel()
	cases := []struct {
		runner string
		want   string
	}{
		{"ubuntu-latest", "linux"},
		{"ubuntu-22.04", "linux"},
		{"macos-latest", "macos"},
		{"macos-14", "macos"},
		{"windows-latest", "windows"},
		{"self-hosted", ""},
	}
	for _, tc := range cases {
		got := inferOSFromRunner(tc.runner)
		if got != tc.want {
			t.Errorf("inferOSFromRunner(%q) = %q, want %q", tc.runner, got, tc.want)
		}
	}
}

func TestGitHubActions_MultipleWorkflows(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	wfDir := filepath.Join(root, ".github", "workflows")
	os.MkdirAll(wfDir, 0o755)
	os.WriteFile(filepath.Join(wfDir, "ci.yml"), []byte(`
name: CI
on: [push]
jobs:
  test:
    strategy:
      matrix:
        node-version: [18, 20]
`), 0o644)
	os.WriteFile(filepath.Join(wfDir, "e2e.yml"), []byte(`
name: E2E
on: [push]
jobs:
  e2e:
    strategy:
      matrix:
        browser: [chromium, firefox, webkit]
`), 0o644)

	result := ParseCIMatrices(root)

	classIDs := map[string]bool{}
	for _, cls := range result.EnvironmentClasses {
		classIDs[cls.ClassID] = true
	}

	if !classIDs["envclass:gha-node-version"] {
		t.Error("expected node-version class from ci.yml")
	}
	if !classIDs["envclass:gha-browser"] {
		t.Error("expected browser class from e2e.yml")
	}
}
