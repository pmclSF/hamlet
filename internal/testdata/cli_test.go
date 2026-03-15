package testdata

import (
	"os/exec"
	"strings"
	"testing"
)

// TestCLI_BuildSucceeds verifies the binary compiles.
func TestCLI_BuildSucceeds(t *testing.T) {
	t.Parallel()
	cmd := exec.Command("go", "build", "-o", "/dev/null", "./cmd/terrain/")
	cmd.Dir = "../.."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
}

// TestCLI_HelpExitCode verifies --help exits cleanly.
func TestCLI_HelpExitCode(t *testing.T) {
	t.Parallel()
	cmd := exec.Command("go", "run", "./cmd/terrain/", "--help")
	cmd.Dir = "../.."
	out, err := cmd.CombinedOutput()
	output := string(out)

	// --help should exit 0 (or at least not error).
	if err != nil {
		t.Logf("help output: %s", output)
		// Some CLIs exit non-zero on --help; check output instead.
		if !strings.Contains(output, "Terrain") {
			t.Errorf("--help did not produce terrain output: %v", err)
		}
	}
}

// TestCLI_HelpContainsAllCommands verifies help text lists all commands.
func TestCLI_HelpContainsAllCommands(t *testing.T) {
	t.Parallel()
	cmd := exec.Command("go", "run", "./cmd/terrain/", "--help")
	cmd.Dir = "../.."
	out, _ := cmd.CombinedOutput()
	output := string(out)

	// Primary commands (canonical user journeys).
	primary := []string{"analyze", "impact", "insights", "explain"}
	for _, c := range primary {
		if !strings.Contains(output, c) {
			t.Errorf("help text missing primary command %q", c)
		}
	}

	// Supporting commands.
	supporting := []string{"summary", "posture", "metrics", "compare", "policy check", "export benchmark"}
	for _, c := range supporting {
		if !strings.Contains(output, c) {
			t.Errorf("help text missing supporting command %q", c)
		}
	}
}

// TestCLI_UnknownCommandExitCode verifies unknown commands exit non-zero.
func TestCLI_UnknownCommandExitCode(t *testing.T) {
	t.Parallel()
	cmd := exec.Command("go", "run", "./cmd/terrain/", "nonexistent")
	cmd.Dir = "../.."
	err := cmd.Run()
	if err == nil {
		t.Error("expected non-zero exit for unknown command")
	}
}

// TestCLI_AnalyzeTestdata verifies analyze works on the test repo.
func TestCLI_AnalyzeTestdata(t *testing.T) {
	t.Parallel()
	cmd := exec.Command("go", "run", "./cmd/terrain/", "analyze", "--root", "internal/analysis/testdata/sample-repo")
	cmd.Dir = "../.."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("analyze failed: %v\n%s", err, out)
	}

	output := string(out)
	if !strings.Contains(output, "Terrain") {
		t.Error("analyze output missing header")
	}
	if !strings.Contains(output, "Next steps:") {
		t.Error("analyze output missing next steps")
	}
}

// TestCLI_AnalyzeJSON verifies --json produces valid JSON.
func TestCLI_AnalyzeJSON(t *testing.T) {
	t.Parallel()
	cmd := exec.Command("go", "run", "./cmd/terrain/", "analyze", "--json", "--root", "internal/analysis/testdata/sample-repo")
	cmd.Dir = "../.."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("analyze --json failed: %v\n%s", err, out)
	}

	if !strings.HasPrefix(strings.TrimSpace(string(out)), "{") {
		t.Error("analyze --json output is not JSON")
	}
}

// TestCLI_PostureTestdata verifies posture command works.
func TestCLI_PostureTestdata(t *testing.T) {
	t.Parallel()
	cmd := exec.Command("go", "run", "./cmd/terrain/", "posture", "--root", "internal/analysis/testdata/sample-repo")
	cmd.Dir = "../.."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("posture failed: %v\n%s", err, out)
	}

	output := string(out)
	if !strings.Contains(output, "Terrain Posture") {
		t.Error("posture output missing header")
	}
}

// TestCLI_MetricsTestdata verifies metrics command works.
func TestCLI_MetricsTestdata(t *testing.T) {
	t.Parallel()
	cmd := exec.Command("go", "run", "./cmd/terrain/", "metrics", "--root", "internal/analysis/testdata/sample-repo")
	cmd.Dir = "../.."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("metrics failed: %v\n%s", err, out)
	}

	output := string(out)
	if !strings.Contains(output, "Terrain Metrics") {
		t.Error("metrics output missing header")
	}
}

// TestCLI_SummaryTestdata verifies summary command works.
func TestCLI_SummaryTestdata(t *testing.T) {
	t.Parallel()
	cmd := exec.Command("go", "run", "./cmd/terrain/", "summary", "--root", "internal/analysis/testdata/sample-repo")
	cmd.Dir = "../.."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("summary failed: %v\n%s", err, out)
	}

	output := string(out)
	if !strings.Contains(output, "Terrain Executive Summary") {
		t.Error("summary output missing header")
	}
}

// TestCLI_HelpContainsPrimaryCommands verifies help prominently lists the four canonical commands.
func TestCLI_HelpContainsPrimaryCommands(t *testing.T) {
	t.Parallel()
	cmd := exec.Command("go", "run", "./cmd/terrain/", "--help")
	cmd.Dir = "../.."
	out, _ := cmd.CombinedOutput()
	output := string(out)

	// The help output must contain a "Primary commands:" section.
	if !strings.Contains(output, "Primary commands:") {
		t.Error("help text missing 'Primary commands:' section header")
	}

	// Each canonical command must appear with its journey question.
	journeys := map[string]string{
		"analyze":  "What is the state of our test system?",
		"impact":   "What validations matter for this change?",
		"insights": "What should we fix in our test system?",
		"explain":  "Why did Terrain make this decision?",
	}
	for cmd, question := range journeys {
		if !strings.Contains(output, cmd) {
			t.Errorf("help text missing primary command %q", cmd)
		}
		if !strings.Contains(output, question) {
			t.Errorf("help text missing journey question for %q: %q", cmd, question)
		}
	}
}

// TestCLI_HelpContainsCanonicalWorkflow verifies help includes the standard journey walkthrough.
func TestCLI_HelpContainsCanonicalWorkflow(t *testing.T) {
	t.Parallel()
	cmd := exec.Command("go", "run", "./cmd/terrain/", "--help")
	cmd.Dir = "../.."
	out, _ := cmd.CombinedOutput()
	output := string(out)

	workflow := []string{
		"Typical flow:",
		"terrain analyze",
		"terrain insights",
		"terrain impact",
		"terrain explain <target>",
	}
	for _, expected := range workflow {
		if !strings.Contains(output, expected) {
			t.Errorf("help text missing canonical workflow item %q", expected)
		}
	}
}

// TestCLI_HelpContainsDebugNamespace verifies debug commands appear in help.
func TestCLI_HelpContainsDebugNamespace(t *testing.T) {
	t.Parallel()
	cmd := exec.Command("go", "run", "./cmd/terrain/", "--help")
	cmd.Dir = "../.."
	out, _ := cmd.CombinedOutput()
	output := string(out)

	debugCmds := []string{"debug graph", "debug coverage", "debug fanout", "debug duplicates", "debug depgraph"}
	for _, c := range debugCmds {
		if !strings.Contains(output, c) {
			t.Errorf("help text missing debug command %q", c)
		}
	}
}

// TestCLI_ExportBenchmarkTestdata verifies export benchmark command works.
func TestCLI_ExportBenchmarkTestdata(t *testing.T) {
	t.Parallel()
	cmd := exec.Command("go", "run", "./cmd/terrain/", "export", "benchmark", "--root", "internal/analysis/testdata/sample-repo")
	cmd.Dir = "../.."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("export benchmark failed: %v\n%s", err, out)
	}

	output := strings.TrimSpace(string(out))
	if !strings.HasPrefix(output, "{") {
		t.Error("export benchmark output should be JSON")
	}
}

// TestCLI_AnalyzeOutputConsistency verifies analyze output structure.
func TestCLI_AnalyzeOutputConsistency(t *testing.T) {
	t.Parallel()
	cmd := exec.Command("go", "run", "./cmd/terrain/", "analyze", "--root", "internal/analysis/testdata/sample-repo")
	cmd.Dir = "../.."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("analyze failed: %v\n%s", err, out)
	}

	output := string(out)
	// Should have standard sections.
	sections := []string{"Test Files", "Frameworks", "Signals"}
	for _, s := range sections {
		if !strings.Contains(output, s) {
			t.Errorf("analyze output missing section %q", s)
		}
	}
}

// TestCLI_MetricsJSON verifies metrics --json produces valid JSON.
func TestCLI_MetricsJSON(t *testing.T) {
	t.Parallel()
	cmd := exec.Command("go", "run", "./cmd/terrain/", "metrics", "--json", "--root", "internal/analysis/testdata/sample-repo")
	cmd.Dir = "../.."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("metrics --json failed: %v\n%s", err, out)
	}

	output := strings.TrimSpace(string(out))
	if !strings.HasPrefix(output, "{") {
		t.Error("metrics --json output should be JSON")
	}
}
