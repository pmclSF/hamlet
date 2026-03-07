// hamlet — signal-first test intelligence for engineering teams.
//
// Commands:
//
//	hamlet analyze              analyze current directory, human-readable output
//	hamlet analyze --root PATH  analyze a specific directory
//	hamlet analyze --json       JSON output (TestSuiteSnapshot)
//	hamlet analyze --write-snapshot  persist snapshot to .hamlet/snapshots/latest.json
//	hamlet metrics              aggregate metrics scorecard (human-readable)
//	hamlet metrics --json       JSON metrics snapshot
//	hamlet summary              executive summary with risk, trends, benchmark readiness
//	hamlet summary --json       JSON executive summary
//	hamlet compare              compare two snapshots
//	hamlet compare --json       JSON comparison output
//	hamlet policy check         evaluate local policy and report violations
//	hamlet policy check --json  JSON output for policy check
//	hamlet export benchmark     benchmark-safe JSON export
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pmclSF/hamlet/internal/analysis"
	"github.com/pmclSF/hamlet/internal/benchmark"
	"github.com/pmclSF/hamlet/internal/comparison"
	"github.com/pmclSF/hamlet/internal/engine"
	"github.com/pmclSF/hamlet/internal/governance"
	"github.com/pmclSF/hamlet/internal/heatmap"
	"github.com/pmclSF/hamlet/internal/metrics"
	"github.com/pmclSF/hamlet/internal/models"
	"github.com/pmclSF/hamlet/internal/policy"
	"github.com/pmclSF/hamlet/internal/quality"
	"github.com/pmclSF/hamlet/internal/reporting"
	"github.com/pmclSF/hamlet/internal/signals"
	"github.com/pmclSF/hamlet/internal/summary"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "analyze":
		analyzeCmd := flag.NewFlagSet("analyze", flag.ExitOnError)
		rootFlag := analyzeCmd.String("root", ".", "repository root to analyze")
		jsonFlag := analyzeCmd.Bool("json", false, "output JSON snapshot")
		writeSnapshot := analyzeCmd.Bool("write-snapshot", false, "persist snapshot to .hamlet/snapshots/latest.json")
		analyzeCmd.Parse(os.Args[2:])
		if err := runAnalyze(*rootFlag, *jsonFlag, *writeSnapshot); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "policy":
		if len(os.Args) < 3 || os.Args[2] != "check" {
			fmt.Fprintln(os.Stderr, "Usage: hamlet policy check [flags]")
			os.Exit(2)
		}
		policyCmd := flag.NewFlagSet("policy check", flag.ExitOnError)
		rootFlag := policyCmd.String("root", ".", "repository root to analyze")
		jsonFlag := policyCmd.Bool("json", false, "output JSON policy check result")
		policyCmd.Parse(os.Args[3:])
		exitCode := runPolicyCheck(*rootFlag, *jsonFlag)
		os.Exit(exitCode)

	case "metrics":
		metricsCmd := flag.NewFlagSet("metrics", flag.ExitOnError)
		rootFlag := metricsCmd.String("root", ".", "repository root to analyze")
		jsonFlag := metricsCmd.Bool("json", false, "output JSON metrics snapshot")
		metricsCmd.Parse(os.Args[2:])
		if err := runMetrics(*rootFlag, *jsonFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "summary":
		summaryCmd := flag.NewFlagSet("summary", flag.ExitOnError)
		rootFlag := summaryCmd.String("root", ".", "repository root to analyze")
		jsonFlag := summaryCmd.Bool("json", false, "output JSON summary with heatmap")
		summaryCmd.Parse(os.Args[2:])
		if err := runSummary(*rootFlag, *jsonFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "compare":
		compareCmd := flag.NewFlagSet("compare", flag.ExitOnError)
		fromFlag := compareCmd.String("from", "", "path to baseline snapshot JSON")
		toFlag := compareCmd.String("to", "", "path to current snapshot JSON")
		rootFlag := compareCmd.String("root", ".", "repository root (used to find .hamlet/snapshots/)")
		jsonFlag := compareCmd.Bool("json", false, "output JSON comparison")
		compareCmd.Parse(os.Args[2:])
		if err := runCompare(*fromFlag, *toFlag, *rootFlag, *jsonFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "export":
		if len(os.Args) < 3 || os.Args[2] != "benchmark" {
			fmt.Fprintln(os.Stderr, "Usage: hamlet export benchmark [flags]")
			os.Exit(2)
		}
		exportCmd := flag.NewFlagSet("export benchmark", flag.ExitOnError)
		rootFlag := exportCmd.String("root", ".", "repository root to analyze")
		exportCmd.Parse(os.Args[3:])
		if err := runExportBenchmark(*rootFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "--help", "-h", "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(2)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Hamlet — signal-first test intelligence for engineering teams")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  hamlet analyze [flags]       analyze repository test suite")
	fmt.Fprintln(os.Stderr, "  hamlet metrics [flags]       output aggregate metrics scorecard")
	fmt.Fprintln(os.Stderr, "  hamlet summary [flags]       executive summary with risk, trends, and benchmark readiness")
	fmt.Fprintln(os.Stderr, "  hamlet compare [flags]       compare two snapshots")
	fmt.Fprintln(os.Stderr, "  hamlet policy check [flags]  evaluate local policy")
	fmt.Fprintln(os.Stderr, "  hamlet export benchmark      output benchmark-safe JSON export")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Analyze flags:")
	fmt.Fprintln(os.Stderr, "  --root PATH          repository root to analyze (default: current directory)")
	fmt.Fprintln(os.Stderr, "  --json               output JSON snapshot instead of human-readable report")
	fmt.Fprintln(os.Stderr, "  --write-snapshot     persist snapshot to .hamlet/snapshots/latest.json")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Policy check flags:")
	fmt.Fprintln(os.Stderr, "  --root PATH          repository root to analyze (default: current directory)")
	fmt.Fprintln(os.Stderr, "  --json               output JSON result instead of human-readable report")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Metrics flags:")
	fmt.Fprintln(os.Stderr, "  --root PATH          repository root to analyze (default: current directory)")
	fmt.Fprintln(os.Stderr, "  --json               output JSON metrics snapshot")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Summary flags:")
	fmt.Fprintln(os.Stderr, "  --root PATH          repository root to analyze (default: current directory)")
	fmt.Fprintln(os.Stderr, "  --json               output JSON executive summary")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Compare flags:")
	fmt.Fprintln(os.Stderr, "  --root PATH          repository root (default: current directory)")
	fmt.Fprintln(os.Stderr, "  --from PATH          baseline snapshot JSON (default: second-latest in .hamlet/snapshots/)")
	fmt.Fprintln(os.Stderr, "  --to PATH            current snapshot JSON (default: latest in .hamlet/snapshots/)")
	fmt.Fprintln(os.Stderr, "  --json               output JSON comparison")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Export benchmark flags:")
	fmt.Fprintln(os.Stderr, "  --root PATH          repository root to analyze (default: current directory)")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Exit codes (policy check):")
	fmt.Fprintln(os.Stderr, "  0  no policy file or no violations")
	fmt.Fprintln(os.Stderr, "  1  violations found or evaluation error")
	fmt.Fprintln(os.Stderr, "  2  usage error")
}

func runAnalyze(root string, jsonOutput bool, writeSnap bool) error {
	result, err := engine.RunPipeline(root)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result.Snapshot)
	}

	reporting.RenderAnalyzeReport(os.Stdout, result.Snapshot)

	if writeSnap {
		return persistSnapshot(result.Snapshot, root)
	}

	return nil
}

// runPolicyCheck evaluates the repository against its local policy.
//
// Exit codes:
//   - 0: no policy file found, or policy exists with no violations
//   - 1: violations found, policy file malformed, or evaluation error
func runPolicyCheck(root string, jsonOutput bool) int {
	// Load policy
	policyResult, err := policy.Load(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	if !policyResult.Found {
		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			enc.Encode(map[string]any{
				"policyFile": nil,
				"pass":       true,
				"violations": []any{},
				"message":    "No policy file found. Create .hamlet/policy.yaml to define policy.",
			})
		} else {
			fmt.Println("Hamlet Policy Check")
			fmt.Println()
			fmt.Println("No policy file found.")
			fmt.Println("Create .hamlet/policy.yaml to define policy rules.")
		}
		return 0
	}

	// Policy check uses a targeted pipeline: analyze + quality detectors + governance.
	// It does not run the full pipeline because its output is specifically the
	// governance evaluation result, not the full snapshot.
	analyzer := analysis.New(root)
	snapshot, err := analyzer.Analyze()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: analysis failed: %v\n", err)
		return 1
	}

	// Run quality detectors (some policy rules reference quality signals).
	signals.RunDetectors(snapshot,
		&quality.WeakAssertionDetector{},
		&quality.MockHeavyDetector{},
		&quality.UntestedExportDetector{},
		&quality.CoverageThresholdDetector{},
	)

	// Evaluate policy.
	govResult := governance.Evaluate(snapshot, policyResult.Config)

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(map[string]any{
			"policyFile": policyResult.Path,
			"pass":       govResult.Pass,
			"violations": govResult.Violations,
		})
	} else {
		reporting.RenderPolicyReport(os.Stdout, policyResult.Path, govResult)
	}

	if !govResult.Pass {
		return 1
	}
	return 0
}

// runMetrics performs analysis and outputs aggregate metrics.
func runMetrics(root string, jsonOutput bool) error {
	result, err := engine.RunPipeline(root)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	ms := metrics.Derive(result.Snapshot)

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(ms)
	}

	reporting.RenderMetricsReport(os.Stdout, ms)
	return nil
}

// runSummary performs analysis and outputs an executive summary with
// trend highlights (if prior snapshots exist) and benchmark readiness.
func runSummary(root string, jsonOutput bool) error {
	result, err := engine.RunPipeline(root)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}
	snapshot := result.Snapshot

	// Build heatmap and derive metrics.
	h := heatmap.Build(snapshot)
	ms := metrics.Derive(snapshot)

	// Attempt to load prior snapshot for trend comparison.
	var comp *comparison.SnapshotComparison
	absRoot, _ := filepath.Abs(root)
	snapDir := filepath.Join(absRoot, ".hamlet", "snapshots")
	latest, previous, snapErr := findRecentSnapshots(snapDir)
	if snapErr == nil && latest != "" && previous != "" {
		fromSnap, err1 := loadSnapshot(previous)
		toSnap, err2 := loadSnapshot(latest)
		if err1 == nil && err2 == nil {
			comp = comparison.Compare(fromSnap, toSnap)
		}
	}

	// Build benchmark segment.
	seg := &benchmark.BuildExport(snapshot, ms, result.HasPolicy).Segment

	// Build executive summary.
	es := summary.Build(&summary.BuildInput{
		Snapshot:   snapshot,
		Heatmap:    h,
		Metrics:    ms,
		Comparison: comp,
		Segment:    seg,
		HasPolicy:  result.HasPolicy,
	})

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(es)
	}

	reporting.RenderExecutiveSummary(os.Stdout, es)
	return nil
}

// runExportBenchmark performs analysis and outputs a benchmark-safe JSON export.
func runExportBenchmark(root string) error {
	result, err := engine.RunPipeline(root)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	ms := metrics.Derive(result.Snapshot)
	export := benchmark.BuildExport(result.Snapshot, ms, result.HasPolicy)

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(export)
}

// runCompare loads two snapshots and produces a comparison report.
//
// If --from and --to are not specified, it looks for the two most recent
// snapshots in .hamlet/snapshots/.
func runCompare(fromPath, toPath, root string, jsonOutput bool) error {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	// Resolve snapshot paths if not explicitly provided.
	if fromPath == "" || toPath == "" {
		snapDir := filepath.Join(absRoot, ".hamlet", "snapshots")
		latest, previous, err := findRecentSnapshots(snapDir)
		if err != nil {
			return err
		}
		if toPath == "" {
			toPath = latest
		}
		if fromPath == "" {
			fromPath = previous
		}
	}

	if fromPath == "" || toPath == "" {
		return fmt.Errorf("need at least two snapshots to compare; use --write-snapshot with hamlet analyze first")
	}

	fromSnap, err := loadSnapshot(fromPath)
	if err != nil {
		return fmt.Errorf("failed to load baseline snapshot: %w", err)
	}
	toSnap, err := loadSnapshot(toPath)
	if err != nil {
		return fmt.Errorf("failed to load current snapshot: %w", err)
	}

	comp := comparison.Compare(fromSnap, toSnap)

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(comp)
	}

	reporting.RenderComparisonReport(os.Stdout, comp)
	return nil
}

func loadSnapshot(path string) (*models.TestSuiteSnapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var snap models.TestSuiteSnapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, fmt.Errorf("invalid snapshot JSON in %s: %w", path, err)
	}
	return &snap, nil
}

// findRecentSnapshots returns the two most recent snapshot files in the directory.
func findRecentSnapshots(dir string) (latest, previous string, err error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", "", fmt.Errorf("no snapshot history found. Run `hamlet analyze --write-snapshot` to begin tracking")
	}

	var snapFiles []string
	for _, e := range entries {
		name := e.Name()
		if name == "latest.json" || !strings.HasSuffix(name, ".json") {
			continue
		}
		snapFiles = append(snapFiles, filepath.Join(dir, name))
	}

	sort.Strings(snapFiles) // Timestamped names sort chronologically

	if len(snapFiles) < 2 {
		latestPath := filepath.Join(dir, "latest.json")
		if _, statErr := os.Stat(latestPath); statErr == nil && len(snapFiles) == 1 {
			return snapFiles[0], latestPath, nil
		}
		return "", "", fmt.Errorf("need at least 2 snapshots to compare; found %d. Run `hamlet analyze --write-snapshot` to save snapshots", len(snapFiles))
	}

	return snapFiles[len(snapFiles)-1], snapFiles[len(snapFiles)-2], nil
}

// persistSnapshot writes the snapshot to .hamlet/snapshots/ as both
// latest.json and a timestamped archive file.
func persistSnapshot(snapshot *models.TestSuiteSnapshot, root string) error {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	dir := filepath.Join(absRoot, ".hamlet", "snapshots")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create snapshot directory: %w", err)
	}

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	latestPath := filepath.Join(dir, "latest.json")
	if err := os.WriteFile(latestPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write snapshot: %w", err)
	}

	ts := snapshot.GeneratedAt.UTC().Format("2006-01-02T15-04-05Z")
	archivePath := filepath.Join(dir, ts+".json")
	if err := os.WriteFile(archivePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write archive snapshot: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Snapshot written to %s\n", latestPath)
	fmt.Fprintf(os.Stderr, "Archive written to %s\n", archivePath)
	return nil
}
