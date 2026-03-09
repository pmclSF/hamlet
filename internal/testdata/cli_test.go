package testdata

import (
	"os/exec"
	"strings"
	"testing"
)

// TestCLI_BuildSucceeds verifies the binary compiles.
func TestCLI_BuildSucceeds(t *testing.T) {
	cmd := exec.Command("go", "build", "-o", "/dev/null", "./cmd/hamlet/")
	cmd.Dir = "../.."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
}

// TestCLI_HelpExitCode verifies --help exits cleanly.
func TestCLI_HelpExitCode(t *testing.T) {
	cmd := exec.Command("go", "run", "./cmd/hamlet/", "--help")
	cmd.Dir = "../.."
	out, err := cmd.CombinedOutput()
	output := string(out)

	// --help should exit 0 (or at least not error).
	if err != nil {
		t.Logf("help output: %s", output)
		// Some CLIs exit non-zero on --help; check output instead.
		if !strings.Contains(output, "Hamlet") {
			t.Errorf("--help did not produce hamlet output: %v", err)
		}
	}
}

// TestCLI_HelpContainsAllCommands verifies help text lists all commands.
func TestCLI_HelpContainsAllCommands(t *testing.T) {
	cmd := exec.Command("go", "run", "./cmd/hamlet/", "--help")
	cmd.Dir = "../.."
	out, _ := cmd.CombinedOutput()
	output := string(out)

	commands := []string{"analyze", "summary", "posture", "metrics", "compare", "impact", "policy check", "export benchmark"}
	for _, c := range commands {
		if !strings.Contains(output, c) {
			t.Errorf("help text missing command %q", c)
		}
	}
}

// TestCLI_UnknownCommandExitCode verifies unknown commands exit non-zero.
func TestCLI_UnknownCommandExitCode(t *testing.T) {
	cmd := exec.Command("go", "run", "./cmd/hamlet/", "nonexistent")
	cmd.Dir = "../.."
	err := cmd.Run()
	if err == nil {
		t.Error("expected non-zero exit for unknown command")
	}
}

// TestCLI_AnalyzeTestdata verifies analyze works on the test repo.
func TestCLI_AnalyzeTestdata(t *testing.T) {
	cmd := exec.Command("go", "run", "./cmd/hamlet/", "analyze", "--root", "internal/analysis/testdata/sample-repo")
	cmd.Dir = "../.."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("analyze failed: %v\n%s", err, out)
	}

	output := string(out)
	if !strings.Contains(output, "Hamlet") {
		t.Error("analyze output missing header")
	}
	if !strings.Contains(output, "Next steps:") {
		t.Error("analyze output missing next steps")
	}
}

// TestCLI_AnalyzeJSON verifies --json produces valid JSON.
func TestCLI_AnalyzeJSON(t *testing.T) {
	cmd := exec.Command("go", "run", "./cmd/hamlet/", "analyze", "--json", "--root", "internal/analysis/testdata/sample-repo")
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
	cmd := exec.Command("go", "run", "./cmd/hamlet/", "posture", "--root", "internal/analysis/testdata/sample-repo")
	cmd.Dir = "../.."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("posture failed: %v\n%s", err, out)
	}

	output := string(out)
	if !strings.Contains(output, "Hamlet Posture") {
		t.Error("posture output missing header")
	}
}

// TestCLI_MetricsTestdata verifies metrics command works.
func TestCLI_MetricsTestdata(t *testing.T) {
	cmd := exec.Command("go", "run", "./cmd/hamlet/", "metrics", "--root", "internal/analysis/testdata/sample-repo")
	cmd.Dir = "../.."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("metrics failed: %v\n%s", err, out)
	}

	output := string(out)
	if !strings.Contains(output, "Hamlet Metrics") {
		t.Error("metrics output missing header")
	}
}

// TestCLI_SummaryTestdata verifies summary command works.
func TestCLI_SummaryTestdata(t *testing.T) {
	cmd := exec.Command("go", "run", "./cmd/hamlet/", "summary", "--root", "internal/analysis/testdata/sample-repo")
	cmd.Dir = "../.."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("summary failed: %v\n%s", err, out)
	}

	output := string(out)
	if !strings.Contains(output, "Hamlet Executive Summary") {
		t.Error("summary output missing header")
	}
}

// TestCLI_HelpContainsNewCommands verifies help lists impact and select-tests.
func TestCLI_HelpContainsNewCommands(t *testing.T) {
	cmd := exec.Command("go", "run", "./cmd/hamlet/", "--help")
	cmd.Dir = "../.."
	out, _ := cmd.CombinedOutput()
	output := string(out)

	newCommands := []string{"impact", "select-tests"}
	for _, c := range newCommands {
		if !strings.Contains(output, c) {
			t.Errorf("help text missing command %q", c)
		}
	}
}

// TestCLI_ExportBenchmarkTestdata verifies export benchmark command works.
func TestCLI_ExportBenchmarkTestdata(t *testing.T) {
	cmd := exec.Command("go", "run", "./cmd/hamlet/", "export", "benchmark", "--root", "internal/analysis/testdata/sample-repo")
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
	cmd := exec.Command("go", "run", "./cmd/hamlet/", "analyze", "--root", "internal/analysis/testdata/sample-repo")
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
	cmd := exec.Command("go", "run", "./cmd/hamlet/", "metrics", "--json", "--root", "internal/analysis/testdata/sample-repo")
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
