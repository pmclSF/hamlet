package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pmclSF/terrain/internal/analyze"
	"github.com/pmclSF/terrain/internal/engine"
	"github.com/pmclSF/terrain/internal/governance"
	"github.com/pmclSF/terrain/internal/logging"
	"github.com/pmclSF/terrain/internal/policy"
	"github.com/pmclSF/terrain/internal/reporting"
)

func runInit(root string) error {
	result, err := engine.RunInit(root)
	if err != nil {
		return err
	}

	sep := strings.Repeat("─", 60)

	fmt.Println("Terrain Init")
	fmt.Println(sep)
	fmt.Println()

	// Section 1: What we detected.
	fmt.Printf("Repository: %s\n", filepath.Base(result.Root))
	fmt.Printf("Test files: %d\n", result.TestFileCount)
	fmt.Println()

	if len(result.Frameworks) > 0 {
		fmt.Println("Frameworks detected:")
		for _, fw := range result.Frameworks {
			fmt.Printf("  %-20s %s (via %s, confidence %.0f%%)\n",
				fw.Name, fw.Language, fw.Source, fw.Confidence*100)
		}
		fmt.Println()
	}

	// Section 2: Artifacts.
	if result.Artifacts != nil {
		if result.Artifacts.CoveragePath != "" {
			fmt.Printf("Coverage:  %s (%s)\n", result.Artifacts.CoveragePath, result.Artifacts.CoverageFormat)
		} else {
			fmt.Println("Coverage:  not found")
		}
		if len(result.Artifacts.RuntimePaths) > 0 {
			for i, p := range result.Artifacts.RuntimePaths {
				fmt.Printf("Runtime:   %s (%s)\n", p, result.Artifacts.RuntimeFormats[i])
			}
		} else {
			fmt.Println("Runtime:   not found")
		}
		fmt.Println()
	}

	// Section 3: What we generated.
	fmt.Println("Configuration")
	fmt.Println(sep)
	if result.PolicyPath != "" {
		fmt.Printf("  Created: %s\n", relativeToRoot(result.PolicyPath, result.Root))
		fmt.Println("           Governance rules (all commented out — edit to enable)")
	} else if result.HasPolicyFile {
		fmt.Println("  Exists:  .terrain/policy.yaml")
	}

	if result.ConfigPath != "" {
		fmt.Printf("  Created: %s\n", relativeToRoot(result.ConfigPath, result.Root))
		fmt.Println("           Manual coverage, scenarios, and CI metadata (template)")
	} else if result.HasTerrainYAML {
		fmt.Println("  Exists:  terrain.yaml")
	}
	fmt.Println()

	// Section 4: What to do next.
	fmt.Println("Next steps")
	fmt.Println(sep)
	fmt.Println("  1. Run analysis:")
	fmt.Printf("     terrain analyze --root %q\n", root)
	fmt.Println()

	step := 2
	if result.Artifacts == nil || result.Artifacts.CoveragePath == "" {
		fmt.Printf("  %d. Generate coverage data to unlock coverage signals:\n", step)
		fmt.Println("     Example: npx jest --coverage")
		fmt.Println("     Example: go test -coverprofile=coverage.out ./...")
		fmt.Println("     Example: pytest --cov --cov-report=lcov")
		step++
		fmt.Println()
	}
	if result.Artifacts == nil || len(result.Artifacts.RuntimePaths) == 0 {
		fmt.Printf("  %d. Generate test result artifacts to unlock health signals:\n", step)
		fmt.Println("     Example: npx jest --json --outputFile=jest-results.json")
		fmt.Println("     Example: pytest --junitxml=junit.xml")
		step++
		fmt.Println()
	}
	if result.PolicyPath != "" {
		fmt.Printf("  %d. Edit .terrain/policy.yaml to enable governance rules\n", step)
		fmt.Println()
	}

	return nil
}

func relativeToRoot(path, root string) string {
	if rel, err := filepath.Rel(root, path); err == nil {
		return rel
	}
	return path
}

// detectFirstExisting checks a list of candidate relative paths against root.
// Returns the first path that exists as a non-empty file, or "".
func detectFirstExisting(root string, candidates []string) string {
	for _, rel := range candidates {
		p := filepath.Join(root, rel)
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return p
		}
	}
	return ""
}

func runAnalyze(root string, jsonOutput bool, format string, verbose bool, writeSnap bool, coveragePath, coverageRunLabel string, runtimePaths string, gauntletPaths string, slowThreshold float64) error {
	parsedRuntime := parseRuntimePaths(runtimePaths)
	parsedGauntlet := parseRuntimePaths(gauntletPaths) // same comma-split logic
	if err := validateCommandInputs(root, coveragePath, parsedRuntime); err != nil {
		return err
	}
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "":
	case "json":
		jsonOutput = true
	case "text":
		jsonOutput = false
	default:
		return fmt.Errorf("invalid --format %q (valid: json, text)", format)
	}

	opt := analysisPipelineOptions(coveragePath, coverageRunLabel, parsedRuntime, slowThreshold)
	opt.GauntletPaths = parsedGauntlet
	opt.OnProgress = newProgressFunc(jsonOutput)
	result, err := engine.RunPipeline(root, opt)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	// Report auto-discovered artifacts via structured logging.
	for _, msg := range result.DiscoveryMessages {
		logging.L().Info(msg)
	}

	// Build the structured analyze report (includes depgraph analysis).
	report := analyze.Build(&analyze.BuildInput{
		Snapshot:  result.Snapshot,
		HasPolicy: result.HasPolicy,
	})

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(report)
	}

	if verbose {
		reporting.RenderAnalyzeReport(os.Stdout, result.Snapshot, reporting.AnalyzeReportOptions{
			Verbose: true,
		})
	} else {
		reporting.RenderAnalyzeReportV2(os.Stdout, report)
	}

	// Show hints for missing artifacts after the report.
	hints := engine.MissingArtifactHints(&opt, result.ArtifactDiscovery)
	if len(hints) > 0 {
		fmt.Println()
		fmt.Println("Unlock more:")
		for _, hint := range hints {
			fmt.Printf("  %s\n", hint)
		}
	}

	if writeSnap {
		return persistSnapshot(result.Snapshot, root)
	}

	return nil
}

// runPolicyCheck evaluates the repository against its local policy.
//
// Exit codes:
//   - 0: no policy file found, or policy exists with no violations
//   - 1: policy file malformed or evaluation/runtime error
//   - 2: policy violations found
func runPolicyCheck(root string, jsonOutput bool, coveragePath, coverageRunLabel string, runtimePaths string, slowThreshold float64) int {
	parsedRuntime := parseRuntimePaths(runtimePaths)
	if err := validateCommandInputs(root, coveragePath, parsedRuntime); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return exitError
	}

	// Load policy
	policyResult, err := policy.Load(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return exitError
	}

	if !policyResult.Found {
		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			if err := enc.Encode(map[string]any{
				"policyFile": nil,
				"pass":       true,
				"violations": []any{},
				"message":    "No policy file found. Create .terrain/policy.yaml to define policy.",
			}); err != nil {
				fmt.Fprintf(os.Stderr, "error: failed to render policy output: %v\n", err)
				return exitError
			}
		} else {
			fmt.Println("Terrain Policy Check")
			fmt.Println()
			fmt.Println("No policy file found.")
			fmt.Println("Create .terrain/policy.yaml to define policy rules.")
		}
		return exitOK
	}

	opt := analysisPipelineOptions(coveragePath, coverageRunLabel, parsedRuntime, slowThreshold)

	// Reuse the main analysis pipeline so policy evaluation can use runtime and
	// coverage artifacts when provided.
	result, err := engine.RunPipeline(root, opt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: analysis failed: %v\n", err)
		return exitError
	}

	// Evaluate policy.
	govResult := governance.Evaluate(result.Snapshot, policyResult.Config)

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(map[string]any{
			"policyFile": policyResult.Path,
			"pass":       govResult.Pass,
			"violations": govResult.Violations,
			"message":    policyStatusMessage(govResult.Pass),
		}); err != nil {
			fmt.Fprintf(os.Stderr, "error: failed to render policy output: %v\n", err)
			return exitError
		}
	} else {
		reporting.RenderPolicyReport(os.Stdout, policyResult.Path, govResult)
	}

	if !govResult.Pass {
		return exitPolicyViolation
	}
	return exitOK
}

func parseRuntimePaths(runtimePaths string) []string {
	var paths []string
	if runtimePaths == "" {
		return paths
	}
	for _, p := range strings.Split(runtimePaths, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			paths = append(paths, p)
		}
	}
	return paths
}

func validateCommandInputs(root, coveragePath string, runtimePaths []string) error {
	rootInfo, err := os.Stat(root)
	if err != nil {
		return fmt.Errorf("invalid --root %q: %w", root, err)
	}
	if !rootInfo.IsDir() {
		return fmt.Errorf("invalid --root %q: expected a directory", root)
	}

	if coveragePath != "" {
		if _, err := os.Stat(coveragePath); err != nil {
			return fmt.Errorf("invalid --coverage %q: %w", coveragePath, err)
		}
	}

	for _, p := range runtimePaths {
		if _, err := os.Stat(p); err != nil {
			return fmt.Errorf("invalid --runtime path %q: %w", p, err)
		}
	}
	return nil
}

func policyStatusMessage(pass bool) string {
	if pass {
		return "Policy checks passed."
	}
	return "Policy violations detected."
}
