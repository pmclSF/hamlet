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
//	hamlet posture              detailed posture breakdown with evidence
//	hamlet posture --json       JSON posture snapshot
//	hamlet summary              executive summary with risk, trends, benchmark readiness
//	hamlet summary --json       JSON executive summary
//	hamlet compare              compare two snapshots
//	hamlet compare --json       JSON comparison output
//	hamlet migration readiness   migration readiness assessment
//	hamlet migration blockers   list migration blockers
//	hamlet migration preview    preview migration for a file or scope
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
	"github.com/pmclSF/hamlet/internal/impact"
	"github.com/pmclSF/hamlet/internal/metrics"
	"github.com/pmclSF/hamlet/internal/migration"
	"github.com/pmclSF/hamlet/internal/models"
	"github.com/pmclSF/hamlet/internal/policy"
	"github.com/pmclSF/hamlet/internal/quality"
	"github.com/pmclSF/hamlet/internal/reporting"
	"github.com/pmclSF/hamlet/internal/signals"
	"github.com/pmclSF/hamlet/internal/summary"
)

// Build-time variables set via ldflags.
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	engine.EngineVersion = version

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
		coverageFlag := analyzeCmd.String("coverage", "", "path to coverage file or directory (LCOV, Istanbul JSON)")
		runtimeFlag := analyzeCmd.String("runtime", "", "path to runtime artifact (JUnit XML, Jest JSON); comma-separated for multiple")
		slowThreshold := analyzeCmd.Float64("slow-threshold", 0, "slow test threshold in ms (default: 5000)")
		analyzeCmd.Parse(os.Args[2:])
		if err := runAnalyze(*rootFlag, *jsonFlag, *writeSnapshot, *coverageFlag, *runtimeFlag, *slowThreshold); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "impact":
		impactCmd := flag.NewFlagSet("impact", flag.ExitOnError)
		rootFlag := impactCmd.String("root", ".", "repository root to analyze")
		baseRef := impactCmd.String("base", "", "git base ref for diff (default: HEAD~1)")
		jsonFlag := impactCmd.Bool("json", false, "output JSON impact result")
		showFlag := impactCmd.String("show", "", "drill-down view: units, gaps, tests, owners")
		ownerFlag := impactCmd.String("owner", "", "filter results by owner")
		impactCmd.Parse(os.Args[2:])
		if err := runImpact(*rootFlag, *baseRef, *jsonFlag, *showFlag, *ownerFlag); err != nil {
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

	case "posture":
		postureCmd := flag.NewFlagSet("posture", flag.ExitOnError)
		rootFlag := postureCmd.String("root", ".", "repository root to analyze")
		jsonFlag := postureCmd.Bool("json", false, "output JSON posture snapshot")
		postureCmd.Parse(os.Args[2:])
		if err := runPosture(*rootFlag, *jsonFlag); err != nil {
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

	case "migration":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: hamlet migration <readiness|blockers|preview> [flags]")
			os.Exit(2)
		}
		subCmd := os.Args[2]
		migCmd := flag.NewFlagSet("migration "+subCmd, flag.ExitOnError)
		rootFlag := migCmd.String("root", ".", "repository root to analyze")
		jsonFlag := migCmd.Bool("json", false, "output JSON")
		fileFlag := migCmd.String("file", "", "file path for preview (relative to root)")
		scopeFlag := migCmd.String("scope", "", "directory scope for preview")
		migCmd.Parse(os.Args[3:])
		if err := runMigration(subCmd, *rootFlag, *jsonFlag, *fileFlag, *scopeFlag); err != nil {
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

	case "version", "--version", "-v":
		fmt.Printf("hamlet %s (commit %s, built %s)\n", version, commit, date)

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
	fmt.Fprintln(os.Stderr, "Quick start:")
	fmt.Fprintln(os.Stderr, "  hamlet analyze           see what Hamlet finds in your test suite")
	fmt.Fprintln(os.Stderr, "  hamlet summary           leadership-ready overview")
	fmt.Fprintln(os.Stderr, "  hamlet posture           evidence-backed posture by dimension")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  analyze [flags]          full test suite analysis")
	fmt.Fprintln(os.Stderr, "  impact [flags]           impact analysis for changed code")
	fmt.Fprintln(os.Stderr, "  summary [flags]          executive summary with risk, trends, benchmark readiness")
	fmt.Fprintln(os.Stderr, "  posture [flags]          detailed posture breakdown with measurement evidence")
	fmt.Fprintln(os.Stderr, "  metrics [flags]          aggregate metrics scorecard")
	fmt.Fprintln(os.Stderr, "  migration readiness      migration readiness assessment")
	fmt.Fprintln(os.Stderr, "  migration blockers       list migration blockers by type and area")
	fmt.Fprintln(os.Stderr, "  migration preview        preview migration for a file or scope")
	fmt.Fprintln(os.Stderr, "  compare [flags]          compare two snapshots for trend tracking")
	fmt.Fprintln(os.Stderr, "  policy check [flags]     evaluate local policy rules")
	fmt.Fprintln(os.Stderr, "  export benchmark [flags] privacy-safe JSON export for benchmarking")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Common flags (all commands):")
	fmt.Fprintln(os.Stderr, "  --root PATH              repository root (default: current directory)")
	fmt.Fprintln(os.Stderr, "  --json                   machine-readable JSON output")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Impact-specific flags:")
	fmt.Fprintln(os.Stderr, "  --base REF               git base ref for diff (default: HEAD~1)")
	fmt.Fprintln(os.Stderr, "  --show VIEW              drill-down: units, gaps, tests, owners")
	fmt.Fprintln(os.Stderr, "  --owner NAME             filter by owner")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Analyze-specific flags:")
	fmt.Fprintln(os.Stderr, "  --write-snapshot         persist snapshot for trend tracking")
	fmt.Fprintln(os.Stderr, "  --coverage PATH          ingest coverage data (LCOV, Istanbul JSON)")
	fmt.Fprintln(os.Stderr, "  --runtime PATH           ingest runtime artifacts (JUnit XML, Jest JSON; comma-separated)")
	fmt.Fprintln(os.Stderr, "  --slow-threshold MS      slow test threshold in ms (default: 5000)")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Compare-specific flags:")
	fmt.Fprintln(os.Stderr, "  --from PATH              baseline snapshot (default: auto-detected)")
	fmt.Fprintln(os.Stderr, "  --to PATH                current snapshot (default: auto-detected)")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Typical flow:")
	fmt.Fprintln(os.Stderr, "  1. hamlet analyze                    see findings")
	fmt.Fprintln(os.Stderr, "  2. hamlet summary                    get the leadership view")
	fmt.Fprintln(os.Stderr, "  3. hamlet posture                    understand the evidence")
	fmt.Fprintln(os.Stderr, "  4. hamlet analyze --write-snapshot    save for trend tracking")
	fmt.Fprintln(os.Stderr, "  5. hamlet compare                    see what changed")
}

func runAnalyze(root string, jsonOutput bool, writeSnap bool, coveragePath string, runtimePaths string, slowThreshold float64) error {
	opt := engine.PipelineOptions{CoveragePath: coveragePath, SlowTestThresholdMs: slowThreshold}
	if runtimePaths != "" {
		for _, p := range strings.Split(runtimePaths, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				opt.RuntimePaths = append(opt.RuntimePaths, p)
			}
		}
	}
	var opts []engine.PipelineOptions
	if opt.CoveragePath != "" || len(opt.RuntimePaths) > 0 || opt.SlowTestThresholdMs > 0 {
		opts = append(opts, opt)
	}
	result, err := engine.RunPipeline(root, opts...)
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

// runImpact performs impact analysis against a git diff.
func runImpact(root, baseRef string, jsonOutput bool, show, ownerFilter string) error {
	result, err := engine.RunPipeline(root)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	scope, err := impact.ChangeScopeFromGitDiff(absRoot, baseRef)
	if err != nil {
		return fmt.Errorf("failed to determine changed files: %w", err)
	}

	impactResult := impact.Analyze(scope, result.Snapshot)

	// Apply owner filter if specified.
	if ownerFilter != "" {
		impactResult = impact.FilterByOwner(impactResult, ownerFilter)
	}

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(impactResult)
	}

	switch show {
	case "units":
		reporting.RenderImpactUnits(os.Stdout, impactResult)
	case "gaps":
		reporting.RenderImpactGaps(os.Stdout, impactResult)
	case "tests":
		reporting.RenderImpactTests(os.Stdout, impactResult)
	case "owners":
		reporting.RenderImpactOwners(os.Stdout, impactResult)
	case "":
		reporting.RenderImpactReport(os.Stdout, impactResult)
	default:
		return fmt.Errorf("unknown --show value: %q (valid: units, gaps, tests, owners)", show)
	}
	return nil
}

// runPosture performs analysis and outputs a detailed posture breakdown.
func runPosture(root string, jsonOutput bool) error {
	result, err := engine.RunPipeline(root)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result.Snapshot.Measurements)
	}

	reporting.RenderPostureReport(os.Stdout, result.Snapshot)
	return nil
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

// runMigration handles `hamlet migration readiness`, `hamlet migration blockers`,
// and `hamlet migration preview`.
func runMigration(subCmd, root string, jsonOutput bool, file, scope string) error {
	result, err := engine.RunPipeline(root)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	switch subCmd {
	case "readiness":
		readiness := migration.ComputeReadiness(result.Snapshot)
		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(readiness)
		}
		reporting.RenderMigrationReport(os.Stdout, readiness)
		return nil

	case "blockers":
		readiness := migration.ComputeReadiness(result.Snapshot)
		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(map[string]any{
				"totalBlockers":          readiness.TotalBlockers,
				"blockersByType":         readiness.BlockersByType,
				"representativeBlockers": readiness.RepresentativeBlockers,
				"areaAssessments":        readiness.AreaAssessments,
			})
		}
		reporting.RenderMigrationBlockers(os.Stdout, readiness)
		return nil

	case "preview":
		if file != "" {
			preview := migration.PreviewFile(result.Snapshot, file, absRoot)
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(preview)
			}
			reporting.RenderMigrationPreview(os.Stdout, preview)
			return nil
		}
		// Scope-based preview
		previews := migration.PreviewScope(result.Snapshot, scope, absRoot)
		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(previews)
		}
		reporting.RenderMigrationPreviewScope(os.Stdout, previews)
		return nil

	default:
		return fmt.Errorf("unknown migration subcommand: %q (valid: readiness, blockers, preview)", subCmd)
	}
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
