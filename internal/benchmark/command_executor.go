package benchmark

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// PerCommandTimeout is the per-command timeout. Set from the CLI.
var PerCommandTimeout = 60 * time.Second

// CommandSpec describes a command to run against a repo.
type CommandSpec struct {
	Name    string   // Display name (e.g., "analyze", "impact")
	Args    []string // CLI arguments
	NeedGit bool     // Whether the command requires a git repo
}

// CommandResult captures the raw output of a single CLI command.
type CommandResult struct {
	RepoName  string   `json:"repoName"`
	Command   string   `json:"command"`
	Args      []string `json:"args,omitempty"`
	ExitCode  int      `json:"exitCode"`
	Stdout    string   `json:"stdout"`
	Stderr    string   `json:"stderr"`
	RuntimeMs int64    `json:"runtimeMs"`
	Error     string   `json:"error,omitempty"`
}

// DetectCommands probes the hamlet binary to find which commands are available
// and returns the appropriate primary and debug command specs.
func DetectCommands(hamletBin string) (primary []CommandSpec, debug []CommandSpec, err error) {
	out, err := exec.Command(hamletBin, "help").CombinedOutput()
	if err != nil {
		return nil, nil, fmt.Errorf("running hamlet help: %w", err)
	}
	helpText := string(out)

	// Primary canonical commands.
	if strings.Contains(helpText, "analyze") {
		primary = append(primary, CommandSpec{
			Name: "analyze",
			Args: []string{"analyze", "--json"},
		})
	}

	if strings.Contains(helpText, "impact") {
		primary = append(primary, CommandSpec{
			Name:    "impact",
			Args:    []string{"impact", "--json", "--base", "HEAD~1"},
			NeedGit: true,
		})
	}

	// Use native insights command if available, otherwise fall back to summary.
	if strings.Contains(helpText, "insights") {
		primary = append(primary, CommandSpec{
			Name: "insights",
			Args: []string{"insights", "--json"},
		})
	} else if strings.Contains(helpText, "summary") {
		primary = append(primary, CommandSpec{
			Name: "insights",
			Args: []string{"summary", "--json"},
		})
	}

	// explain is handled specially (needs a test ID from impact first).
	primary = append(primary, CommandSpec{
		Name: "explain",
		Args: nil,
	})

	// Debug commands — use native debug namespace if available, otherwise depgraph.
	hasDebug := strings.Contains(helpText, "debug graph")
	if hasDebug {
		for _, view := range []string{"graph", "coverage", "fanout", "duplicates"} {
			debug = append(debug, CommandSpec{
				Name: "debug:" + view,
				Args: []string{"debug", view, "--json"},
			})
		}
	} else if strings.Contains(helpText, "depgraph") {
		for _, view := range []string{"stats", "coverage", "fanout", "duplicates"} {
			debug = append(debug, CommandSpec{
				Name: "depgraph:" + view,
				Args: []string{"depgraph", "--json", "--show", view},
			})
		}
	}

	return primary, debug, nil
}

// RunCommand executes a single hamlet command against a repo.
func RunCommand(ctx context.Context, hamletBin string, repoPath string, spec CommandSpec) CommandResult {
	result := CommandResult{
		Command: spec.Name,
		Args:    spec.Args,
	}

	args := make([]string, len(spec.Args))
	copy(args, spec.Args)
	args = append(args, "--root", repoPath)

	cmdCtx, cancel := context.WithTimeout(ctx, PerCommandTimeout)
	defer cancel()

	start := time.Now()
	cmd := exec.CommandContext(cmdCtx, hamletBin, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result.RuntimeMs = time.Since(start).Milliseconds()
	result.Stdout = stdout.String()
	result.Stderr = stderr.String()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1
			result.Error = err.Error()
		}
	}

	return result
}

// RunExplain implements the multi-step explain workflow:
//  1. Run analyze --json to get test IDs (always available)
//  2. If no test IDs found, fall back to impact --json
//  3. Run explain <id> --json for the chosen test
func RunExplain(ctx context.Context, hamletBin string, repoPath string) CommandResult {
	result := CommandResult{
		Command: "explain",
	}

	start := time.Now()

	cmdCtx, cancel := context.WithTimeout(ctx, PerCommandTimeout)
	defer cancel()

	// Step 1: Get a test ID from analyze output.
	// We use analyze (not impact) because impact requires meaningful git
	// changes and often finds no impacted tests in benchmark scenarios.
	testID := ""
	analyzeArgs := []string{"analyze", "--json", "--root", repoPath}
	analyzeCmd := exec.CommandContext(cmdCtx, hamletBin, analyzeArgs...)
	var analyzeOut bytes.Buffer
	analyzeCmd.Stdout = &analyzeOut
	if err := analyzeCmd.Run(); err == nil {
		testID = extractTestIDFromAnalyze(analyzeOut.String())
	}

	// Step 2: Fall back to impact if analyze found nothing.
	if testID == "" {
		impactArgs := []string{"impact", "--json", "--base", "HEAD~1", "--root", repoPath}
		impactCmd := exec.CommandContext(cmdCtx, hamletBin, impactArgs...)
		var impactOut bytes.Buffer
		impactCmd.Stdout = &impactOut
		if err := impactCmd.Run(); err == nil {
			testID = ExtractTestID(impactOut.String())
		}
	}

	if testID == "" {
		result.RuntimeMs = time.Since(start).Milliseconds()
		result.Stdout = `{"error":"no test IDs found","detail":"neither impact nor analyze produced a test ID to explain"}`
		result.ExitCode = 0
		return result
	}

	// Step 3: Run explain --json --root <path> <id>.
	// Flags must come before positional args for Go's flag package.
	explainArgs := []string{"explain", "--json", "--root", repoPath, testID}
	explainCmd := exec.CommandContext(cmdCtx, hamletBin, explainArgs...)
	var explainOut, explainErr bytes.Buffer
	explainCmd.Stdout = &explainOut
	explainCmd.Stderr = &explainErr

	err := explainCmd.Run()
	result.RuntimeMs = time.Since(start).Milliseconds()
	result.Args = explainArgs
	result.Stdout = explainOut.String()
	result.Stderr = explainErr.String()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1
			result.Error = err.Error()
		}
	}

	return result
}

func extractTestIDFromAnalyze(jsonOutput string) string {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonOutput), &data); err != nil {
		return ""
	}

	// Shape: { "testCases": [{ "testId": "..." }] }
	if cases, ok := data["testCases"].([]interface{}); ok {
		for _, c := range cases {
			if tc, ok := c.(map[string]interface{}); ok {
				if id, ok := tc["testId"].(string); ok {
					return id
				}
				if id, ok := tc["id"].(string); ok {
					return id
				}
			}
		}
	}

	return ""
}

// ExtractTestID tries to find a test ID from impact JSON output.
func ExtractTestID(jsonOutput string) string {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonOutput), &data); err != nil {
		return ""
	}

	// Shape 1: { "impactedTests": [{ "id": "..." }] }
	if tests, ok := data["impactedTests"].([]interface{}); ok && len(tests) > 0 {
		if t, ok := tests[0].(map[string]interface{}); ok {
			if id, ok := t["id"].(string); ok {
				return id
			}
			if id, ok := t["testId"].(string); ok {
				return id
			}
		}
	}

	// Shape 2: { "tests": [{ "id": "..." }] }
	if tests, ok := data["tests"].([]interface{}); ok && len(tests) > 0 {
		if t, ok := tests[0].(map[string]interface{}); ok {
			if id, ok := t["id"].(string); ok {
				return id
			}
		}
	}

	// Shape 3: { "selectedTests": ["id1", "id2"] }
	if selected, ok := data["selectedTests"].([]interface{}); ok && len(selected) > 0 {
		if id, ok := selected[0].(string); ok {
			return id
		}
	}

	// Shape 4: look in impact.impacted array
	if impact, ok := data["impact"].(map[string]interface{}); ok {
		if impacted, ok := impact["impacted"].([]interface{}); ok && len(impacted) > 0 {
			if t, ok := impacted[0].(map[string]interface{}); ok {
				if id, ok := t["id"].(string); ok {
					return id
				}
				if id, ok := t["testId"].(string); ok {
					return id
				}
			}
		}
	}

	return ""
}
