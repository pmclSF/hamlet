// terrain — test system intelligence platform.
//
// Primary commands (canonical user journeys):
//
//	terrain analyze              What is the state of our test system?
//	terrain impact               What validations matter for this change?
//	terrain insights             What should we fix in our test system?
//	terrain explain <target>     Why did Terrain make this decision?
//
// Supporting commands:
//
//	terrain init                 detect data files and print recommended analyze command
//	terrain summary              executive summary with risk, trends, benchmark readiness
//	terrain focus                prioritized next actions
//	terrain posture              detailed posture breakdown with evidence
//	terrain portfolio            portfolio intelligence (cost, breadth, leverage, redundancy)
//	terrain metrics              aggregate metrics scorecard
//	terrain compare              compare two snapshots for trend tracking
//	terrain select-tests         recommend protective test set for a change
//	terrain pr                   PR/change-scoped analysis
//	terrain show <entity> <id>   drill into test, unit, owner, or finding
//	terrain migration <sub>      readiness, blockers, or preview
//	terrain policy check         evaluate local policy rules
//	terrain export benchmark     privacy-safe JSON export for benchmarking
//
// Advanced / debug:
//
//	terrain debug graph          dependency graph statistics
//	terrain debug coverage       structural coverage analysis
//	terrain debug fanout         high-fanout node analysis
//	terrain debug duplicates     duplicate test cluster analysis
//	terrain debug depgraph       full dependency graph analysis (all engines)
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pmclSF/terrain/internal/analysis"
	"github.com/pmclSF/terrain/internal/analyze"
	"github.com/pmclSF/terrain/internal/benchmark"
	"github.com/pmclSF/terrain/internal/changescope"
	"github.com/pmclSF/terrain/internal/comparison"
	"github.com/pmclSF/terrain/internal/depgraph"
	"github.com/pmclSF/terrain/internal/engine"
	"github.com/pmclSF/terrain/internal/explain"
	"github.com/pmclSF/terrain/internal/insights"
	"github.com/pmclSF/terrain/internal/governance"
	"github.com/pmclSF/terrain/internal/graph"
	"github.com/pmclSF/terrain/internal/heatmap"
	"github.com/pmclSF/terrain/internal/impact"
	"github.com/pmclSF/terrain/internal/matrix"
	"github.com/pmclSF/terrain/internal/metrics"
	"github.com/pmclSF/terrain/internal/migration"
	"github.com/pmclSF/terrain/internal/models"
	"github.com/pmclSF/terrain/internal/policy"
	"github.com/pmclSF/terrain/internal/reporting"
	"github.com/pmclSF/terrain/internal/stability"
	"github.com/pmclSF/terrain/internal/summary"
)

// Build-time variables set via ldflags.
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

const defaultSlowThresholdMs = 5000.0

const (
	exitOK              = 0
	exitError           = 1
	exitPolicyViolation = 2
)

func main() {
	// Backwards-compatibility: warn if invoked as "hamlet" (deprecated alias).
	if base := filepath.Base(os.Args[0]); base == "hamlet" || base == "hamlet.exe" {
		fmt.Fprintln(os.Stderr, "WARNING: The 'hamlet' command has been renamed to 'terrain'.")
		fmt.Fprintln(os.Stderr, "         The 'hamlet' alias is deprecated and will be removed in a future release.")
		fmt.Fprintln(os.Stderr)
	}

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "analyze":
		analyzeCmd := flag.NewFlagSet("analyze", flag.ExitOnError)
		rootFlag := analyzeCmd.String("root", ".", "repository root to analyze")
		jsonFlag := analyzeCmd.Bool("json", false, "output JSON snapshot")
		formatFlag := analyzeCmd.String("format", "", "output format: json or text")
		verboseFlag := analyzeCmd.Bool("verbose", false, "show all findings in analyze output")
		writeSnapshot := analyzeCmd.Bool("write-snapshot", false, "persist snapshot to .terrain/snapshots/latest.json")
		coverageFlag := analyzeCmd.String("coverage", "", "path to coverage file or directory (LCOV, Istanbul JSON)")
		coverageRunLabelFlag := analyzeCmd.String("coverage-run-label", "", "coverage run label: unit, integration, or e2e")
		runtimeFlag := analyzeCmd.String("runtime", "", "path to runtime artifact (JUnit XML, Jest JSON); comma-separated for multiple")
		slowThreshold := analyzeCmd.Float64("slow-threshold", defaultSlowThresholdMs, "slow test threshold in ms")
		_ = analyzeCmd.Parse(os.Args[2:])
		if err := runAnalyze(*rootFlag, *jsonFlag, *formatFlag, *verboseFlag, *writeSnapshot, *coverageFlag, *coverageRunLabelFlag, *runtimeFlag, *slowThreshold); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "init":
		initCmd := flag.NewFlagSet("init", flag.ExitOnError)
		rootFlag := initCmd.String("root", ".", "repository root to inspect")
		_ = initCmd.Parse(os.Args[2:])
		if err := runInit(*rootFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "impact":
		impactCmd := flag.NewFlagSet("impact", flag.ExitOnError)
		rootFlag := impactCmd.String("root", ".", "repository root to analyze")
		baseRef := impactCmd.String("base", "", "git base ref for diff (default: HEAD~1)")
		jsonFlag := impactCmd.Bool("json", false, "output JSON impact result")
		showFlag := impactCmd.String("show", "", "drill-down view: units, gaps, tests, owners, graph, selected")
		ownerFlag := impactCmd.String("owner", "", "filter results by owner")
		_ = impactCmd.Parse(os.Args[2:])
		if err := runImpact(*rootFlag, *baseRef, *jsonFlag, *showFlag, *ownerFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "policy":
		if len(os.Args) < 3 || os.Args[2] != "check" {
			fmt.Fprintln(os.Stderr, "Usage: terrain policy check [flags]")
			os.Exit(2)
		}
		policyCmd := flag.NewFlagSet("policy check", flag.ExitOnError)
		rootFlag := policyCmd.String("root", ".", "repository root to analyze")
		jsonFlag := policyCmd.Bool("json", false, "output JSON policy check result")
		coverageFlag := policyCmd.String("coverage", "", "path to coverage file or directory (LCOV, Istanbul JSON)")
		coverageRunLabelFlag := policyCmd.String("coverage-run-label", "", "coverage run label: unit, integration, or e2e")
		runtimeFlag := policyCmd.String("runtime", "", "path to runtime artifact (JUnit XML, Jest JSON); comma-separated for multiple")
		slowThreshold := policyCmd.Float64("slow-threshold", defaultSlowThresholdMs, "slow test threshold in ms")
		_ = policyCmd.Parse(os.Args[3:])
		exitCode := runPolicyCheck(*rootFlag, *jsonFlag, *coverageFlag, *coverageRunLabelFlag, *runtimeFlag, *slowThreshold)
		os.Exit(exitCode)

	case "metrics":
		metricsCmd := flag.NewFlagSet("metrics", flag.ExitOnError)
		rootFlag := metricsCmd.String("root", ".", "repository root to analyze")
		jsonFlag := metricsCmd.Bool("json", false, "output JSON metrics snapshot")
		_ = metricsCmd.Parse(os.Args[2:])
		if err := runMetrics(*rootFlag, *jsonFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "posture":
		postureCmd := flag.NewFlagSet("posture", flag.ExitOnError)
		rootFlag := postureCmd.String("root", ".", "repository root to analyze")
		jsonFlag := postureCmd.Bool("json", false, "output JSON posture snapshot")
		_ = postureCmd.Parse(os.Args[2:])
		if err := runPosture(*rootFlag, *jsonFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "portfolio":
		portfolioCmd := flag.NewFlagSet("portfolio", flag.ExitOnError)
		rootFlag := portfolioCmd.String("root", ".", "repository root to analyze")
		jsonFlag := portfolioCmd.Bool("json", false, "output JSON portfolio snapshot")
		_ = portfolioCmd.Parse(os.Args[2:])
		if err := runPortfolio(*rootFlag, *jsonFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "insights":
		insightsCmd := flag.NewFlagSet("insights", flag.ExitOnError)
		rootFlag := insightsCmd.String("root", ".", "repository root to analyze")
		jsonFlag := insightsCmd.Bool("json", false, "output JSON insights")
		_ = insightsCmd.Parse(os.Args[2:])
		if err := runInsights(*rootFlag, *jsonFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "explain":
		explainCmd := flag.NewFlagSet("explain", flag.ExitOnError)
		rootFlag := explainCmd.String("root", ".", "repository root to analyze")
		baseRef := explainCmd.String("base", "", "git base ref for diff (default: HEAD~1)")
		jsonFlag := explainCmd.Bool("json", false, "output JSON")
		_ = explainCmd.Parse(os.Args[2:])
		explainArgs := explainCmd.Args()
		if len(explainArgs) == 0 {
			fmt.Fprintln(os.Stderr, "Usage: terrain explain <test-path|selection>")
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, "Explain why Terrain made a decision.")
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, "  terrain explain <test-path>   explain why a test was selected")
			fmt.Fprintln(os.Stderr, "  terrain explain selection     explain overall test selection strategy")
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, "See: docs/examples/explain-report.md")
			os.Exit(2)
		}
		if err := runExplain(explainArgs[0], *rootFlag, *baseRef, *jsonFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "summary":
		summaryCmd := flag.NewFlagSet("summary", flag.ExitOnError)
		rootFlag := summaryCmd.String("root", ".", "repository root to analyze")
		jsonFlag := summaryCmd.Bool("json", false, "output JSON summary with heatmap")
		_ = summaryCmd.Parse(os.Args[2:])
		if err := runSummary(*rootFlag, *jsonFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "focus":
		focusCmd := flag.NewFlagSet("focus", flag.ExitOnError)
		rootFlag := focusCmd.String("root", ".", "repository root to analyze")
		jsonFlag := focusCmd.Bool("json", false, "output JSON focus summary")
		_ = focusCmd.Parse(os.Args[2:])
		if err := runFocus(*rootFlag, *jsonFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "compare":
		compareCmd := flag.NewFlagSet("compare", flag.ExitOnError)
		fromFlag := compareCmd.String("from", "", "path to baseline snapshot JSON")
		toFlag := compareCmd.String("to", "", "path to current snapshot JSON")
		rootFlag := compareCmd.String("root", ".", "repository root (used to find .terrain/snapshots/)")
		jsonFlag := compareCmd.Bool("json", false, "output JSON comparison")
		_ = compareCmd.Parse(os.Args[2:])
		if err := runCompare(*fromFlag, *toFlag, *rootFlag, *jsonFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "migration":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: terrain migration <readiness|blockers|preview> [flags]")
			os.Exit(2)
		}
		subCmd := os.Args[2]
		migCmd := flag.NewFlagSet("migration "+subCmd, flag.ExitOnError)
		rootFlag := migCmd.String("root", ".", "repository root to analyze")
		jsonFlag := migCmd.Bool("json", false, "output JSON")
		fileFlag := migCmd.String("file", "", "file path for preview (relative to root)")
		scopeFlag := migCmd.String("scope", "", "directory scope for preview")
		_ = migCmd.Parse(os.Args[3:])
		if err := runMigration(subCmd, *rootFlag, *jsonFlag, *fileFlag, *scopeFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "select-tests":
		stCmd := flag.NewFlagSet("select-tests", flag.ExitOnError)
		rootFlag := stCmd.String("root", ".", "repository root to analyze")
		baseRef := stCmd.String("base", "", "git base ref for diff (default: HEAD~1)")
		jsonFlag := stCmd.Bool("json", false, "output JSON protective test set")
		_ = stCmd.Parse(os.Args[2:])
		if err := runSelectTests(*rootFlag, *baseRef, *jsonFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "pr":
		prCmd := flag.NewFlagSet("pr", flag.ExitOnError)
		rootFlag := prCmd.String("root", ".", "repository root to analyze")
		baseRef := prCmd.String("base", "", "git base ref for diff (default: HEAD~1)")
		jsonFlag := prCmd.Bool("json", false, "output JSON PR analysis")
		formatFlag := prCmd.String("format", "", "output format: markdown, comment, annotation")
		_ = prCmd.Parse(os.Args[2:])
		if err := runPR(*rootFlag, *baseRef, *jsonFlag, *formatFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "show":
		if len(os.Args) >= 3 {
			switch os.Args[2] {
			case "--help", "-h", "help":
				printShowUsage()
				return
			}
		}
		if len(os.Args) < 3 {
			printShowUsage()
			os.Exit(2)
		}
		showSubCmd := os.Args[2]
		showCmd := flag.NewFlagSet("show", flag.ExitOnError)
		rootFlag := showCmd.String("root", ".", "repository root to analyze")
		jsonFlag := showCmd.Bool("json", false, "output JSON")
		_ = showCmd.Parse(os.Args[3:])
		showArgs := showCmd.Args()
		showID := ""
		if len(showArgs) > 0 {
			showID = showArgs[0]
		}
		if err := runShow(showSubCmd, showID, *rootFlag, *jsonFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "export":
		if len(os.Args) < 3 || os.Args[2] != "benchmark" {
			fmt.Fprintln(os.Stderr, "Usage: terrain export benchmark [flags]")
			os.Exit(2)
		}
		exportCmd := flag.NewFlagSet("export benchmark", flag.ExitOnError)
		rootFlag := exportCmd.String("root", ".", "repository root to analyze")
		_ = exportCmd.Parse(os.Args[3:])
		if err := runExportBenchmark(*rootFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "debug":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: terrain debug <graph|coverage|fanout|duplicates|depgraph> [flags]")
			os.Exit(2)
		}
		debugSub := os.Args[2]
		if debugSub == "depgraph" {
			// Full depgraph analysis under debug namespace.
			dgCmd := flag.NewFlagSet("debug depgraph", flag.ExitOnError)
			rootFlag := dgCmd.String("root", ".", "repository root to analyze")
			jsonFlag := dgCmd.Bool("json", false, "output JSON")
			showFlag := dgCmd.String("show", "", "sub-view: stats, coverage, duplicates, fanout, impact, profile")
			changedFlag := dgCmd.String("changed", "", "comma-separated changed files for impact analysis")
			_ = dgCmd.Parse(os.Args[3:])
			if err := runDepgraph(*rootFlag, *jsonFlag, *showFlag, *changedFlag); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			return
		}
		debugCmd := flag.NewFlagSet("debug "+debugSub, flag.ExitOnError)
		rootFlag := debugCmd.String("root", ".", "repository root to analyze")
		jsonFlag := debugCmd.Bool("json", false, "output JSON")
		changedFlag := debugCmd.String("changed", "", "comma-separated changed files for impact analysis")
		_ = debugCmd.Parse(os.Args[3:])
		showView := ""
		switch debugSub {
		case "graph":
			showView = "stats"
		case "coverage":
			showView = "coverage"
		case "fanout":
			showView = "fanout"
		case "duplicates":
			showView = "duplicates"
		default:
			fmt.Fprintf(os.Stderr, "unknown debug command: %s\n", debugSub)
			fmt.Fprintln(os.Stderr, "Available: graph, coverage, fanout, duplicates, depgraph")
			os.Exit(2)
		}
		if err := runDepgraph(*rootFlag, *jsonFlag, showView, *changedFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "depgraph":
		// Backward-compat alias for "debug depgraph".
		dgCmd := flag.NewFlagSet("depgraph", flag.ExitOnError)
		rootFlag := dgCmd.String("root", ".", "repository root to analyze")
		jsonFlag := dgCmd.Bool("json", false, "output JSON")
		showFlag := dgCmd.String("show", "", "sub-view: stats, coverage, duplicates, fanout, impact, profile")
		changedFlag := dgCmd.String("changed", "", "comma-separated changed files for impact analysis")
		_ = dgCmd.Parse(os.Args[2:])
		if err := runDepgraph(*rootFlag, *jsonFlag, *showFlag, *changedFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "version", "--version", "-v":
		fmt.Printf("terrain %s (commit %s, built %s)\n", version, commit, date)

	case "--help", "-h", "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(2)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Terrain — test system intelligence platform")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Primary commands:")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "  analyze  [flags]         What is the state of our test system?")
	fmt.Fprintln(os.Stderr, "                           Example: terrain analyze --root ./myproject")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "  impact   [flags]         What validations matter for this change?")
	fmt.Fprintln(os.Stderr, "                           Example: terrain impact --base main")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "  insights [flags]         What should we fix in our test system?")
	fmt.Fprintln(os.Stderr, "                           Example: terrain insights --json")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "  explain  <target>        Why did Terrain make this decision?")
	fmt.Fprintln(os.Stderr, "                           Example: terrain explain src/auth/login.test.ts")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Supporting commands:")
	fmt.Fprintln(os.Stderr, "  init [flags]             detect data paths and print recommended analyze command")
	fmt.Fprintln(os.Stderr, "  summary [flags]          executive summary with risk, trends, benchmark readiness")
	fmt.Fprintln(os.Stderr, "  focus [flags]            prioritized next actions")
	fmt.Fprintln(os.Stderr, "  posture [flags]          detailed posture breakdown with measurement evidence")
	fmt.Fprintln(os.Stderr, "  portfolio [flags]        portfolio intelligence: cost, breadth, leverage, redundancy")
	fmt.Fprintln(os.Stderr, "  select-tests [flags]     recommend protective test set for a change")
	fmt.Fprintln(os.Stderr, "  pr [flags]               PR/change-scoped analysis")
	fmt.Fprintln(os.Stderr, "  show <entity> <id>       drill into test, unit/codeunit, owner, or finding")
	fmt.Fprintln(os.Stderr, "  metrics [flags]          aggregate metrics scorecard")
	fmt.Fprintln(os.Stderr, "  compare [flags]          compare two snapshots for trend tracking")
	fmt.Fprintln(os.Stderr, "  migration <sub> [flags]  readiness, blockers, or preview")
	fmt.Fprintln(os.Stderr, "  policy check [flags]     evaluate local policy rules")
	fmt.Fprintln(os.Stderr, "  export benchmark [flags] privacy-safe JSON export for benchmarking")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Advanced / debug:")
	fmt.Fprintln(os.Stderr, "  debug graph [flags]      dependency graph statistics")
	fmt.Fprintln(os.Stderr, "  debug coverage [flags]   structural coverage analysis")
	fmt.Fprintln(os.Stderr, "  debug fanout [flags]     high-fanout node analysis")
	fmt.Fprintln(os.Stderr, "  debug duplicates [flags] duplicate test cluster analysis")
	fmt.Fprintln(os.Stderr, "  debug depgraph [flags]   full dependency graph analysis (all engines)")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Common flags:")
	fmt.Fprintln(os.Stderr, "  --root PATH              repository root (default: current directory)")
	fmt.Fprintln(os.Stderr, "  --json                   machine-readable JSON output")
	fmt.Fprintln(os.Stderr, "  --base REF               git base ref for diff (impact, pr, select-tests)")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Typical flow:")
	fmt.Fprintln(os.Stderr, "  1. terrain analyze                    understand your test system")
	fmt.Fprintln(os.Stderr, "  2. terrain insights                   find what to improve")
	fmt.Fprintln(os.Stderr, "  3. terrain impact                     see what a change affects")
	fmt.Fprintln(os.Stderr, "  4. terrain explain <target>           understand why")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Docs: docs/examples/{analyze,summary,insights,explain,focus,impact}-report.md")
}

func printShowUsage() {
	fmt.Fprintln(os.Stderr, "Usage: terrain show <test|unit|codeunit|owner|finding> <id-or-path> [--root PATH] [--json]")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Examples:")
	fmt.Fprintln(os.Stderr, "  terrain show test src/auth/login.test.js")
	fmt.Fprintln(os.Stderr, "  terrain show codeunit src/auth/login.ts:authenticate --json")
	fmt.Fprintln(os.Stderr, "  terrain show owner platform")
}

func defaultPipelineOptions() engine.PipelineOptions {
	return engine.PipelineOptions{
		EngineVersion: version,
	}
}

func analysisPipelineOptions(coveragePath, coverageRunLabel string, runtimePaths []string, slowThreshold float64) engine.PipelineOptions {
	opt := defaultPipelineOptions()
	opt.CoveragePath = coveragePath
	opt.CoverageRunLabel = strings.TrimSpace(coverageRunLabel)
	opt.RuntimePaths = runtimePaths
	opt.SlowTestThresholdMs = slowThreshold
	return opt
}

func runInit(root string) error {
	rootInfo, err := os.Stat(root)
	if err != nil {
		return fmt.Errorf("invalid --root path %q: %w", root, err)
	}
	if !rootInfo.IsDir() {
		return fmt.Errorf("invalid --root path %q: not a directory", root)
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("resolve root path: %w", err)
	}

	coveragePath := detectFirstExisting(absRoot, []string{
		"coverage/lcov.info",
		"coverage/coverage-final.json",
		"coverage-final.json",
		"coverage.out",
		"coverage.lcov",
		"lcov.info",
	})
	runtimePath := detectFirstExisting(absRoot, []string{
		"junit.xml",
		"test-results.xml",
		"test-results.json",
		"reports/junit.xml",
		"jest-results.json",
		"junit/junit.xml",
	})

	fmt.Println("Terrain Init")
	fmt.Println()
	fmt.Printf("Root: %s\n", absRoot)
	if coveragePath != "" {
		fmt.Printf("Coverage data: found at %s\n", coveragePath)
	} else {
		fmt.Println("Coverage data: not found")
	}
	if runtimePath != "" {
		fmt.Printf("Runtime data: found at %s\n", runtimePath)
	} else {
		fmt.Println("Runtime data: not found")
	}
	fmt.Println()
	fmt.Println("Recommended command:")
	fmt.Printf("  terrain analyze --root %q", root)
	if coveragePath != "" {
		fmt.Printf(" --coverage %q", coveragePath)
	}
	if runtimePath != "" {
		fmt.Printf(" --runtime %q", runtimePath)
	}
	fmt.Println()

	if coveragePath == "" || runtimePath == "" {
		fmt.Println()
		fmt.Println("To unlock fuller analysis:")
		if coveragePath == "" {
			fmt.Println("  1. Generate coverage artifacts and rerun with --coverage <path>")
		}
		if runtimePath == "" {
			if coveragePath == "" {
				fmt.Println("  2. Generate runtime artifacts and rerun with --runtime <path>")
			} else {
				fmt.Println("  1. Generate runtime artifacts and rerun with --runtime <path>")
			}
		}
	}

	return nil
}

func detectFirstExisting(root string, candidates []string) string {
	for _, rel := range candidates {
		p := filepath.Join(root, rel)
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return p
		}
	}
	return ""
}

func runAnalyze(root string, jsonOutput bool, format string, verbose bool, writeSnap bool, coveragePath, coverageRunLabel string, runtimePaths string, slowThreshold float64) error {
	parsedRuntime := parseRuntimePaths(runtimePaths)
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
	result, err := engine.RunPipeline(root, opt)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
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
		// Verbose mode: use legacy renderer with full signal detail.
		reporting.RenderAnalyzeReport(os.Stdout, result.Snapshot, reporting.AnalyzeReportOptions{
			Verbose: true,
		})
	} else {
		reporting.RenderAnalyzeReportV2(os.Stdout, report)
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

// runImpact performs impact analysis against a git diff.
func runImpact(root, baseRef string, jsonOutput bool, show, ownerFilter string) error {
	result, err := engine.RunPipeline(root, defaultPipelineOptions())
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	cs, err := impact.ChangeSetFromGitDiff(absRoot, baseRef)
	if err != nil {
		return fmt.Errorf("failed to determine changed files: %w", err)
	}

	impactResult := impact.AnalyzeChangeSet(cs, result.Snapshot)

	// Apply edge-case policy to adjust confidence and add warnings.
	snapshot := result.Snapshot
	dg := depgraph.Build(snapshot)
	dgCov := depgraph.AnalyzeCoverage(dg)
	dgDupes := depgraph.DetectDuplicates(dg)
	dgFanout := depgraph.AnalyzeFanout(dg, depgraph.DefaultFanoutThreshold)
	ms := metrics.Derive(snapshot)
	pi := depgraph.ProfileInsights{
		Coverage:   &dgCov,
		Duplicates: &dgDupes,
		Fanout:     &dgFanout,
		Snapshot:   analyze.BuildSnapshotProfileData(snapshot),
	}
	dgProfile := depgraph.AnalyzeProfile(dg, pi)
	depgraph.EnrichProfileWithHealthRatios(&dgProfile, ms.Health.SkippedTestRatio, ms.Health.FlakyTestRatio)
	dgEdgeCases := depgraph.DetectEdgeCases(dgProfile, dg, pi)
	if len(dgEdgeCases) > 0 {
		dgPolicy := depgraph.ApplyEdgeCasePolicy(dgEdgeCases, dgProfile)
		impactResult.ApplyEdgeCasePolicy(dgPolicy.ConfidenceAdjustment, dgPolicy.RiskElevated, dgPolicy.Recommendations)
	}

	// Apply manual coverage overlay to annotate protection gaps.
	if len(snapshot.ManualCoverage) > 0 {
		impactResult.ApplyManualCoverageOverlay(snapshot.ManualCoverage)
	}

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
	case "graph":
		reporting.RenderImpactGraph(os.Stdout, impactResult)
	case "selected":
		reporting.RenderProtectiveSet(os.Stdout, impactResult)
	case "":
		reporting.RenderImpactReport(os.Stdout, impactResult)
	default:
		return fmt.Errorf("unknown --show value: %q (valid: units, gaps, tests, owners, graph, selected)", show)
	}
	return nil
}

// runSelectTests performs impact analysis and outputs the protective test set.
func runSelectTests(root, baseRef string, jsonOutput bool) error {
	result, err := engine.RunPipeline(root, defaultPipelineOptions())
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	cs, err := impact.ChangeSetFromGitDiff(absRoot, baseRef)
	if err != nil {
		return fmt.Errorf("failed to determine changed files: %w", err)
	}

	impactResult := impact.AnalyzeChangeSet(cs, result.Snapshot)

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(impactResult.ProtectiveSet)
	}

	reporting.RenderProtectiveSet(os.Stdout, impactResult)
	return nil
}

// runPortfolio performs analysis and outputs portfolio intelligence.
func runPortfolio(root string, jsonOutput bool) error {
	result, err := engine.RunPipeline(root, defaultPipelineOptions())
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result.Snapshot.Portfolio)
	}

	reporting.RenderPortfolioReport(os.Stdout, result.Snapshot)
	return nil
}

// runPosture performs analysis and outputs a detailed posture breakdown.
func runPosture(root string, jsonOutput bool) error {
	result, err := engine.RunPipeline(root, defaultPipelineOptions())
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
	result, err := engine.RunPipeline(root, defaultPipelineOptions())
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
	result, err := engine.RunPipeline(root, defaultPipelineOptions())
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}
	snapshot := result.Snapshot

	// Build graph, heatmap (enriched with graph data), and derive metrics.
	g := graph.Build(snapshot)
	h := heatmap.BuildWithGraph(snapshot, g)
	ms := metrics.Derive(snapshot)

	// Attempt to load prior snapshot for trend comparison.
	var comp *comparison.SnapshotComparison
	absRoot, _ := filepath.Abs(root)
	snapDir := filepath.Join(absRoot, ".terrain", "snapshots")
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

// runFocus performs analysis and emits a compact action-first view.
func runFocus(root string, jsonOutput bool) error {
	result, err := engine.RunPipeline(root, defaultPipelineOptions())
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}
	snapshot := result.Snapshot

	g := graph.Build(snapshot)
	h := heatmap.BuildWithGraph(snapshot, g)
	ms := metrics.Derive(snapshot)
	seg := &benchmark.BuildExport(snapshot, ms, result.HasPolicy).Segment

	es := summary.Build(&summary.BuildInput{
		Snapshot:  snapshot,
		Heatmap:   h,
		Metrics:   ms,
		Segment:   seg,
		HasPolicy: result.HasPolicy,
	})

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(map[string]any{
			"recommendedFocus": es.RecommendedFocus,
			"topRiskAreas":     es.TopRiskAreas,
			"recommendations":  es.Recommendations,
			"posture":          es.Posture,
		})
	}

	fmt.Println("Terrain Focus")
	fmt.Println()
	if es.RecommendedFocus != "" {
		fmt.Printf("Now: %s\n", es.RecommendedFocus)
	} else {
		fmt.Println("Now: No immediate focus area detected.")
	}

	if len(es.TopRiskAreas) > 0 {
		fmt.Println()
		fmt.Println("Top Risk Areas")
		for i, area := range es.TopRiskAreas {
			fmt.Printf("  %d. %s (%s)\n", i+1, area.Name, area.Band)
			if area.RiskType != "" {
				fmt.Printf("     risk: %s (%d signal(s))\n", area.RiskType, area.SignalCount)
			}
		}
	}

	if len(es.Recommendations) > 0 {
		fmt.Println()
		fmt.Println("Recommended Actions")
		for i, r := range es.Recommendations {
			fmt.Printf("  %d. %s\n", i+1, r.What)
			if r.Why != "" {
				fmt.Printf("     why: %s\n", r.Why)
			}
			if r.Where != "" {
				fmt.Printf("     where: %s\n", r.Where)
			}
		}
	}

	fmt.Println()
	fmt.Println("Next: terrain posture    see detailed evidence by dimension")
	return nil
}

// runInsights aggregates all insight engines into a single actionable report.
// It combines executive summary, depgraph profile, and portfolio findings.
func runInsights(root string, jsonOutput bool) error {
	result, err := engine.RunPipeline(root, defaultPipelineOptions())
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}
	snapshot := result.Snapshot

	// Derive metrics for health ratios.
	ms := metrics.Derive(snapshot)

	// Build depgraph insights with a preflight guard for very large repos.
	const maxInsightsGraphNodes = 150000
	estimatedGraphNodes := len(snapshot.TestCases) + len(snapshot.CodeUnits) + len(snapshot.TestFiles)

	input := &insights.BuildInput{
		Snapshot:  snapshot,
		HasPolicy: result.HasPolicy,
	}

	if estimatedGraphNodes > maxInsightsGraphNodes {
		input.DepgraphSkipped = true
		input.DepgraphSkipReason = fmt.Sprintf(
			"depgraph analysis skipped for estimated graph size %d (limit %d)",
			estimatedGraphNodes, maxInsightsGraphNodes,
		)
		input.Duplicates = depgraph.DuplicateResult{
			Skipped:       true,
			SkipReason:    input.DepgraphSkipReason,
			TestsAnalyzed: len(snapshot.TestCases),
		}
		input.Fanout = depgraph.FanoutResult{
			Skipped:      true,
			SkipReason:   input.DepgraphSkipReason,
			NodeCount:    estimatedGraphNodes,
			Threshold:    depgraph.DefaultFanoutThreshold,
			FlaggedCount: 0,
		}
		input.Profile = depgraph.RepoProfile{
			TestVolume:         "large",
			CIPressure:         "high",
			CoverageConfidence: "low",
			RedundancyLevel:    "low",
			FanoutBurden:       "low",
			SkipBurden:         "low",
			FlakeBurden:        "low",
		}
		input.Policy = depgraph.Policy{
			FallbackLevel:        depgraph.FallbackSmokeRegression,
			ConfidenceAdjustment: 0.6,
			RiskElevated:         true,
			Recommendations: []string{
				"Depgraph analysis skipped due to repository scale; narrow scope for full dependency insights.",
			},
		}
		input.EdgeCases = []depgraph.EdgeCase{{
			Type:        depgraph.EdgeCaseLowGraphVisibility,
			Severity:    "warning",
			Description: input.DepgraphSkipReason,
		}}
		input.Coverage = depgraph.CoverageResult{BandCounts: map[depgraph.CoverageBand]int{}}
	} else {
		dg := depgraph.Build(snapshot)
		input.Coverage = depgraph.AnalyzeCoverage(dg)
		input.Duplicates = depgraph.DetectDuplicates(dg)
		input.Fanout = depgraph.AnalyzeFanout(dg, depgraph.DefaultFanoutThreshold)
		dgInsights := depgraph.ProfileInsights{
			Coverage:   &input.Coverage,
			Duplicates: &input.Duplicates,
			Fanout:     &input.Fanout,
			Snapshot:   analyze.BuildSnapshotProfileData(snapshot),
		}
		input.Profile = depgraph.AnalyzeProfile(dg, dgInsights)
		depgraph.EnrichProfileWithHealthRatios(&input.Profile, ms.Health.SkippedTestRatio, ms.Health.FlakyTestRatio)
		input.EdgeCases = depgraph.DetectEdgeCases(input.Profile, dg, dgInsights)
		input.Policy = depgraph.ApplyEdgeCasePolicy(input.EdgeCases, input.Profile)
		dgRedundancy := depgraph.AnalyzeRedundancy(dg)
		if len(dgRedundancy.Clusters) > 0 {
			input.BehaviorRedundancy = &dgRedundancy
		}
		input.StabilityClusters = stability.DetectClusters(dg, snapshot.Signals)
		matrixResult := matrix.Analyze(dg)
		if len(matrixResult.Classes) > 0 {
			input.MatrixCoverage = matrixResult
		}
	}

	report := insights.Build(input)

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(report)
	}

	reporting.RenderInsightsReport(os.Stdout, report)
	return nil
}

// runExplain auto-detects the entity type and shows detail with reasoning.
func runExplain(target, root, baseRef string, jsonOutput bool) error {
	result, err := engine.RunPipeline(root, defaultPipelineOptions())
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}
	snap := result.Snapshot

	// Compute impact result for structured explanation.
	impactResult, impactErr := computeImpactForExplain(root, baseRef, snap)

	// "selection" mode: explain overall test selection strategy.
	if target == "selection" {
		if impactErr != nil {
			return fmt.Errorf("impact analysis required for selection explanation: %w", impactErr)
		}
		sel, err := explain.ExplainSelection(impactResult)
		if err != nil {
			return err
		}
		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(sel)
		}
		reporting.RenderSelectionExplanation(os.Stdout, sel)
		return nil
	}

	// Try structured test explanation first (if impact data available).
	if impactErr == nil {
		te, err := explain.ExplainTest(target, impactResult)
		if err == nil {
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(te)
			}
			reporting.RenderTestExplanation(os.Stdout, te)
			return nil
		}
	}

	// Fall back to legacy entity lookup for non-test targets.

	// Try test file.
	for _, tf := range snap.TestFiles {
		if tf.Path == target {
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(tf)
			}
			renderTestDetail(tf, snap)
			return nil
		}
	}

	// Try test case by ID or canonical identity.
	for _, tc := range snap.TestCases {
		if tc.TestID == target || tc.CanonicalIdentity == target {
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(tc)
			}
			renderTestCaseDetail(tc, snap)
			return nil
		}
	}

	// Try code unit.
	for _, cu := range snap.CodeUnits {
		unitID := cu.Path + ":" + cu.Name
		if unitID == target || cu.Name == target || cu.Path == target {
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(cu)
			}
			renderCodeUnitDetail(cu, snap)
			return nil
		}
	}

	// Try owner.
	ownerID := strings.ToLower(target)
	ownerFound := false
	if snap.Ownership != nil {
		for _, owners := range snap.Ownership {
			for _, o := range owners {
				if strings.ToLower(o) == ownerID {
					ownerFound = true
					break
				}
			}
			if ownerFound {
				break
			}
		}
	}
	if ownerFound {
		return showOwner(target, snap, jsonOutput)
	}

	// Try finding.
	if snap.Portfolio != nil {
		for i, f := range snap.Portfolio.Findings {
			findingID := fmt.Sprintf("%d", i)
			if findingID == target || f.Type == target {
				return showFinding(target, snap, jsonOutput)
			}
		}
	}

	return fmt.Errorf("entity not found: %s\n\nTry: a test file path, test ID, or 'selection'", target)
}

// computeImpactForExplain runs impact analysis using git diff to detect changes.
func computeImpactForExplain(root, baseRef string, snap *models.TestSuiteSnapshot) (*impact.ImpactResult, error) {
	if baseRef == "" {
		baseRef = "HEAD~1"
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	cs, err := impact.ChangeSetFromGitDiff(absRoot, baseRef)
	if err != nil {
		return nil, fmt.Errorf("change detection failed: %w", err)
	}

	return impact.AnalyzeChangeSet(cs, snap), nil
}

// runMigration handles `terrain migration readiness`, `terrain migration blockers`,
// and `terrain migration preview`.
func runMigration(subCmd, root string, jsonOutput bool, file, scope string) error {
	result, err := engine.RunPipeline(root, defaultPipelineOptions())
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
	result, err := engine.RunPipeline(root, defaultPipelineOptions())
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
// snapshots in .terrain/snapshots/.
func runCompare(fromPath, toPath, root string, jsonOutput bool) error {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	// Resolve snapshot paths if not explicitly provided.
	if fromPath == "" || toPath == "" {
		snapDir := filepath.Join(absRoot, ".terrain", "snapshots")
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
		return fmt.Errorf("need at least two snapshots to compare; use --write-snapshot with terrain analyze first")
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
	models.MigrateSnapshotInPlace(&snap)
	return &snap, nil
}

// findRecentSnapshots returns the two most recent snapshot files in the directory.
func findRecentSnapshots(dir string) (latest, previous string, err error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", "", fmt.Errorf("no snapshot history found. Run `terrain analyze --write-snapshot` to begin tracking")
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
			return latestPath, snapFiles[0], nil
		}
		return "", "", fmt.Errorf("need at least 2 snapshots to compare; found %d. Run `terrain analyze --write-snapshot` to save snapshots", len(snapFiles))
	}

	return snapFiles[len(snapFiles)-1], snapFiles[len(snapFiles)-2], nil
}

// persistSnapshot writes the snapshot to .terrain/snapshots/ as both
// latest.json and a timestamped archive file.
func persistSnapshot(snapshot *models.TestSuiteSnapshot, root string) error {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	dir := filepath.Join(absRoot, ".terrain", "snapshots")
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

// runPR performs a PR/change-scoped analysis.
func runPR(root, baseRef string, jsonOutput bool, format string) error {
	result, err := engine.RunPipeline(root, defaultPipelineOptions())
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	cs, err := impact.ChangeSetFromGitDiff(absRoot, baseRef)
	if err != nil {
		return fmt.Errorf("failed to determine changed files: %w", err)
	}

	pr := changescope.AnalyzePRFromChangeSet(cs, result.Snapshot)

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(pr)
	}

	switch format {
	case "markdown", "md":
		changescope.RenderPRSummaryMarkdown(os.Stdout, pr)
	case "comment":
		changescope.RenderPRCommentConcise(os.Stdout, pr)
	case "annotation", "ci":
		changescope.RenderCIAnnotation(os.Stdout, pr)
	default:
		changescope.RenderChangeScopedReport(os.Stdout, pr)
	}
	return nil
}

// runShow handles entity drill-down commands.
func runShow(entity, id, root string, jsonOutput bool) error {
	entity = strings.TrimSpace(entity)
	id = strings.TrimSpace(id)
	switch entity {
	case "test", "unit", "codeunit", "owner", "finding":
	default:
		return fmt.Errorf("unknown entity type: %q (valid: test, unit, codeunit, owner, finding)", entity)
	}
	if id == "" {
		return fmt.Errorf("missing ID for show %q", entity)
	}

	result, err := engine.RunPipeline(root, defaultPipelineOptions())
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}
	snap := result.Snapshot

	switch entity {
	case "test":
		return showTest(id, snap, jsonOutput)
	case "unit", "codeunit":
		return showCodeUnit(id, snap, jsonOutput)
	case "owner":
		return showOwner(id, snap, jsonOutput)
	case "finding":
		return showFinding(id, snap, jsonOutput)
	}
	return nil
}

func showTest(id string, snap *models.TestSuiteSnapshot, jsonOutput bool) error {
	// Search by test ID or file path.
	for _, tf := range snap.TestFiles {
		if tf.Path == id {
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(tf)
			}
			renderTestDetail(tf, snap)
			return nil
		}
	}
	// Search test cases by ID.
	for _, tc := range snap.TestCases {
		if tc.TestID == id || tc.CanonicalIdentity == id {
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(tc)
			}
			renderTestCaseDetail(tc, snap)
			return nil
		}
	}
	return fmt.Errorf("test not found: %s", id)
}

func showCodeUnit(id string, snap *models.TestSuiteSnapshot, jsonOutput bool) error {
	for _, cu := range snap.CodeUnits {
		unitID := cu.Path + ":" + cu.Name
		if unitID == id || cu.Name == id || cu.Path == id {
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(cu)
			}
			renderCodeUnitDetail(cu, snap)
			return nil
		}
	}
	return fmt.Errorf("code unit not found: %s", id)
}

func showOwner(id string, snap *models.TestSuiteSnapshot, jsonOutput bool) error {
	ownerID := strings.ToLower(id)

	// Collect owner's files, signals, test files.
	type ownerData struct {
		Owner       string          `json:"owner"`
		OwnedFiles  []string        `json:"ownedFiles"`
		TestFiles   []string        `json:"testFiles"`
		SignalCount int             `json:"signalCount"`
		Signals     []models.Signal `json:"signals,omitempty"`
	}

	data := ownerData{Owner: id}

	if snap.Ownership != nil {
		for path, owners := range snap.Ownership {
			for _, o := range owners {
				if strings.ToLower(o) == ownerID {
					data.OwnedFiles = append(data.OwnedFiles, path)
				}
			}
		}
	}
	sort.Strings(data.OwnedFiles)

	for _, tf := range snap.TestFiles {
		if strings.ToLower(tf.Owner) == ownerID {
			data.TestFiles = append(data.TestFiles, tf.Path)
		}
	}

	for _, sig := range snap.Signals {
		if strings.ToLower(sig.Owner) == ownerID {
			data.SignalCount++
			if len(data.Signals) < 10 {
				data.Signals = append(data.Signals, sig)
			}
		}
	}

	if len(data.OwnedFiles) == 0 && len(data.TestFiles) == 0 && data.SignalCount == 0 {
		return fmt.Errorf("owner not found: %s", id)
	}

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	}

	fmt.Printf("Owner: %s\n", data.Owner)
	fmt.Printf("Owned files: %d\n", len(data.OwnedFiles))
	fmt.Printf("Test files: %d\n", len(data.TestFiles))
	fmt.Printf("Signals: %d\n", data.SignalCount)
	if len(data.OwnedFiles) > 0 {
		fmt.Println("\nOwned files:")
		limit := 10
		if len(data.OwnedFiles) < limit {
			limit = len(data.OwnedFiles)
		}
		for _, f := range data.OwnedFiles[:limit] {
			fmt.Printf("  %s\n", f)
		}
		if len(data.OwnedFiles) > 10 {
			fmt.Printf("  ... and %d more\n", len(data.OwnedFiles)-10)
		}
	}
	if data.SignalCount > 0 {
		fmt.Println("\nTop signals:")
		for _, sig := range data.Signals {
			fmt.Printf("  [%s] %s — %s\n", sig.Severity, sig.Type, sig.Location.File)
		}
	}
	fmt.Println("\nNext: terrain show test <path>   drill into a specific test file")
	return nil
}

func showFinding(id string, snap *models.TestSuiteSnapshot, jsonOutput bool) error {
	// Findings are identified by index or type.
	if snap.Portfolio != nil {
		for i, f := range snap.Portfolio.Findings {
			findingID := fmt.Sprintf("%d", i)
			if findingID == id || f.Type == id {
				if jsonOutput {
					enc := json.NewEncoder(os.Stdout)
					enc.SetIndent("", "  ")
					return enc.Encode(f)
				}
				fmt.Printf("Finding: %s\n", f.Type)
				fmt.Printf("Path: %s\n", f.Path)
				fmt.Printf("Confidence: %s\n", f.Confidence)
				fmt.Printf("Explanation: %s\n", f.Explanation)
				if f.SuggestedAction != "" {
					fmt.Printf("Action: %s\n", f.SuggestedAction)
				}
				return nil
			}
		}
	}
	// Also search signals.
	for i, sig := range snap.Signals {
		sigID := fmt.Sprintf("s%d", i)
		if sigID == id || string(sig.Type) == id {
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(sig)
			}
			fmt.Printf("Signal: %s\n", sig.Type)
			fmt.Printf("Category: %s\n", sig.Category)
			fmt.Printf("Severity: %s\n", sig.Severity)
			fmt.Printf("File: %s\n", sig.Location.File)
			fmt.Printf("Explanation: %s\n", sig.Explanation)
			return nil
		}
	}
	return fmt.Errorf("finding not found: %s", id)
}

func renderTestDetail(tf models.TestFile, snap *models.TestSuiteSnapshot) {
	fmt.Printf("Test File: %s\n", tf.Path)
	fmt.Printf("Framework: %s\n", tf.Framework)
	if tf.Owner != "" {
		fmt.Printf("Owner: %s\n", tf.Owner)
	}
	fmt.Printf("Tests: %d    Assertions: %d\n", tf.TestCount, tf.AssertionCount)
	if tf.MockCount > 0 {
		fmt.Printf("Mocks: %d\n", tf.MockCount)
	}
	if tf.RuntimeStats != nil {
		fmt.Printf("Runtime: %.0fms    Pass rate: %.0f%%    Retry rate: %.0f%%\n",
			tf.RuntimeStats.AvgRuntimeMs,
			tf.RuntimeStats.PassRate*100,
			tf.RuntimeStats.RetryRate*100)
	}
	if len(tf.LinkedCodeUnits) > 0 {
		fmt.Printf("Covers: %s\n", strings.Join(tf.LinkedCodeUnits, ", "))
	}

	// Show signals for this file.
	var fileSignals []models.Signal
	for _, sig := range snap.Signals {
		if sig.Location.File == tf.Path {
			fileSignals = append(fileSignals, sig)
		}
	}
	if len(fileSignals) > 0 {
		fmt.Printf("\nSignals (%d):\n", len(fileSignals))
		for _, sig := range fileSignals {
			fmt.Printf("  [%s] %s: %s\n", sig.Severity, sig.Type, sig.Explanation)
		}
	}
	fmt.Println("\nNext: terrain impact --show tests   see impact analysis")
}

func renderTestCaseDetail(tc models.TestCase, snap *models.TestSuiteSnapshot) {
	fmt.Printf("Test: %s\n", tc.TestName)
	fmt.Printf("ID: %s\n", tc.TestID)
	fmt.Printf("File: %s\n", tc.FilePath)
	if len(tc.SuiteHierarchy) > 0 {
		fmt.Printf("Suite: %s\n", strings.Join(tc.SuiteHierarchy, " > "))
	}
	fmt.Printf("Framework: %s\n", tc.Framework)
	if tc.TestType != "" {
		fmt.Printf("Type: %s (confidence: %.0f%%)\n", tc.TestType, tc.TestTypeConfidence*100)
	}
	fmt.Printf("Extraction: %s (confidence: %.0f%%)\n", tc.ExtractionKind, tc.Confidence*100)
}

func renderCodeUnitDetail(cu models.CodeUnit, snap *models.TestSuiteSnapshot) {
	fmt.Printf("Code Unit: %s\n", cu.Name)
	fmt.Printf("Path: %s\n", cu.Path)
	fmt.Printf("Kind: %s\n", cu.Kind)
	fmt.Printf("Exported: %v\n", cu.Exported)
	if cu.Owner != "" {
		fmt.Printf("Owner: %s\n", cu.Owner)
	}

	// Find covering tests.
	unitID := cu.Path + ":" + cu.Name
	allowNameOnly := isUniqueCodeUnitName(snap, cu.Name)
	var coveringTests []string
	for _, tf := range snap.TestFiles {
		for _, linked := range tf.LinkedCodeUnits {
			if linked == unitID || (allowNameOnly && linked == cu.Name) {
				coveringTests = append(coveringTests, tf.Path)
				break
			}
		}
	}
	if len(coveringTests) > 0 {
		fmt.Printf("\nCovering tests (%d):\n", len(coveringTests))
		for _, t := range coveringTests {
			fmt.Printf("  %s\n", t)
		}
	} else {
		fmt.Println("\nNo covering tests detected.")
	}
	fmt.Println("\nNext: terrain show test <path>   drill into a covering test")
}

func isUniqueCodeUnitName(snap *models.TestSuiteSnapshot, name string) bool {
	if name == "" {
		return false
	}
	count := 0
	for _, cu := range snap.CodeUnits {
		if cu.Name == name {
			count++
			if count > 1 {
				return false
			}
		}
	}
	return count == 1
}

// runDepgraph builds the dependency graph and runs the requested analysis.
func runDepgraph(root string, jsonOutput bool, show string, changed string) error {
	analyzer := analysis.New(root)
	snapshot, err := analyzer.Analyze()
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	// Build the dependency graph from the snapshot.
	dg := depgraph.Build(snapshot)
	stats := dg.Stats()

	// Run impact if changed files specified.
	var impactResult *depgraph.ImpactResult
	computeImpact := func() {
		if changed == "" || impactResult != nil {
			return
		}
		files := strings.Split(changed, ",")
		ir := depgraph.AnalyzeImpact(dg, files)
		impactResult = &ir
	}

	var (
		coverageComputed   bool
		coverageResult     depgraph.CoverageResult
		duplicatesComputed bool
		duplicatesResult   depgraph.DuplicateResult
		fanoutComputed     bool
		fanoutResult       depgraph.FanoutResult
		profileComputed    bool
		profileResult      depgraph.RepoProfile
		edgeCasesResult    []depgraph.EdgeCase
		policyResult       depgraph.Policy
	)
	computeCoverage := func() depgraph.CoverageResult {
		if !coverageComputed {
			coverageResult = depgraph.AnalyzeCoverage(dg)
			coverageComputed = true
		}
		return coverageResult
	}
	computeDuplicates := func() depgraph.DuplicateResult {
		if !duplicatesComputed {
			duplicatesResult = depgraph.DetectDuplicates(dg)
			duplicatesComputed = true
		}
		return duplicatesResult
	}
	computeFanout := func() depgraph.FanoutResult {
		if !fanoutComputed {
			fanoutResult = depgraph.AnalyzeFanout(dg, depgraph.DefaultFanoutThreshold)
			fanoutComputed = true
		}
		return fanoutResult
	}
	computeProfile := func() (depgraph.RepoProfile, []depgraph.EdgeCase, depgraph.Policy) {
		if !profileComputed {
			coverage := computeCoverage()
			duplicates := computeDuplicates()
			fanout := computeFanout()
			insights := depgraph.ProfileInsights{
				Coverage:   &coverage,
				Duplicates: &duplicates,
				Fanout:     &fanout,
			}
			profileResult = depgraph.AnalyzeProfile(dg, insights)
			dgMetrics := metrics.Derive(snapshot)
			depgraph.EnrichProfileWithHealthRatios(&profileResult, dgMetrics.Health.SkippedTestRatio, dgMetrics.Health.FlakyTestRatio)
			edgeCasesResult = depgraph.DetectEdgeCases(profileResult, dg, insights)
			policyResult = depgraph.ApplyEdgeCasePolicy(edgeCasesResult, profileResult)
			profileComputed = true
		}
		return profileResult, edgeCasesResult, policyResult
	}

	if show != "" {
		switch show {
		case "stats":
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(stats)
			}
			fmt.Println("Terrain Dependency Graph Stats")
			fmt.Println(strings.Repeat("─", 50))
			fmt.Printf("  Nodes: %d    Edges: %d    Density: %.4f\n", stats.NodeCount, stats.EdgeCount, stats.Density)
			fmt.Println()
			fmt.Println("Node Types:")
			for _, nt := range sortedMapKeys(stats.NodesByType) {
				fmt.Printf("  %-20s %d\n", nt, stats.NodesByType[nt])
			}
			return nil
		case "coverage":
			coverage := computeCoverage()
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(coverage)
			}
			fmt.Println("Coverage (structural):")
			fmt.Printf("  Sources: %d   High: %d   Medium: %d   Low: %d\n",
				coverage.SourceCount,
				coverage.BandCounts[depgraph.CoverageBandHigh],
				coverage.BandCounts[depgraph.CoverageBandMedium],
				coverage.BandCounts[depgraph.CoverageBandLow])
			return nil
		case "duplicates":
			duplicates := computeDuplicates()
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(duplicates)
			}
			fmt.Println("Duplicates:")
			fmt.Printf("  Tests analyzed: %d   Duplicates: %d   Clusters: %d\n",
				duplicates.TestsAnalyzed, duplicates.DuplicateCount, len(duplicates.Clusters))
			if duplicates.Skipped {
				fmt.Printf("  Note: %s\n", duplicates.SkipReason)
			}
			if len(duplicates.Clusters) > 0 {
				top := duplicates.Clusters[0]
				fmt.Printf("  Top cluster: %d tests, similarity %.2f\n", len(top.Tests), top.Similarity)
			}
			return nil
		case "fanout":
			fanout := computeFanout()
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(fanout)
			}
			fmt.Println("Fanout:")
			fmt.Printf("  Nodes: %d   Flagged: %d   Threshold: %d\n",
				fanout.NodeCount, fanout.FlaggedCount, fanout.Threshold)
			if fanout.Skipped {
				fmt.Printf("  Note: %s\n", fanout.SkipReason)
			}
			if len(fanout.Entries) > 0 {
				top := fanout.Entries[0]
				fmt.Printf("  Highest: %s (transitive: %d)\n", top.NodeID, top.TransitiveFanout)
			}
			return nil
		case "impact":
			computeImpact()
			if impactResult == nil {
				return fmt.Errorf("impact view requires --changed")
			}
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(impactResult)
			}
			fmt.Println("Impact:")
			fmt.Printf("  Changed files: %d   Impacted tests: %d\n",
				len(impactResult.ChangedFiles), len(impactResult.Tests))
			fmt.Printf("  High: %d   Medium: %d   Low: %d\n",
				impactResult.LevelCounts["high"],
				impactResult.LevelCounts["medium"],
				impactResult.LevelCounts["low"])
			return nil
		case "profile":
			profile, edgeCases, pol := computeProfile()
			if jsonOutput {
				out := map[string]any{
					"profile":   profile,
					"edgeCases": edgeCases,
					"policy":    pol,
				}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(out)
			}
			fmt.Println("Repository Profile:")
			fmt.Printf("  Test Volume:          %s\n", profile.TestVolume)
			fmt.Printf("  CI Pressure:          %s\n", profile.CIPressure)
			fmt.Printf("  Coverage Confidence:  %s\n", profile.CoverageConfidence)
			fmt.Printf("  Redundancy Level:     %s\n", profile.RedundancyLevel)
			fmt.Printf("  Fanout Burden:        %s\n", profile.FanoutBurden)
			fmt.Println()
			if len(edgeCases) > 0 {
				fmt.Println("Edge Cases:")
				for _, ec := range edgeCases {
					fmt.Printf("  [%s] %s: %s\n", ec.Severity, ec.Type, ec.Description)
				}
				fmt.Println()
			}
			if len(pol.Recommendations) > 0 {
				fmt.Println("Recommendations:")
				for _, r := range pol.Recommendations {
					fmt.Printf("  • %s\n", r)
				}
				fmt.Println()
			}
			fmt.Printf("Policy: fallback=%s  confidence=%.2f  optimization=%s\n",
				pol.FallbackLevel, pol.ConfidenceAdjustment,
				map[bool]string{true: "disabled", false: "enabled"}[pol.OptimizationDisabled])
			return nil
		default:
			return fmt.Errorf("unknown view: %s (available: stats, coverage, duplicates, fanout, impact, profile)", show)
		}
	}

	// Full report mode (all engines).
	coverage := computeCoverage()
	duplicates := computeDuplicates()
	fanout := computeFanout()
	profile, edgeCases, pol := computeProfile()
	computeImpact()

	// JSON output.
	if jsonOutput {
		out := map[string]any{
			"stats":      stats,
			"coverage":   coverage,
			"duplicates": duplicates,
			"fanout":     fanout,
			"profile":    profile,
			"edgeCases":  edgeCases,
			"policy":     pol,
		}
		if impactResult != nil {
			out["impact"] = impactResult
		}

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	}

	// Text output.
	fmt.Println("Terrain Dependency Graph")
	fmt.Println(strings.Repeat("─", 50))
	fmt.Printf("  Nodes: %d    Edges: %d    Density: %.4f\n", stats.NodeCount, stats.EdgeCount, stats.Density)
	fmt.Println()

	// Node breakdown.
	fmt.Println("Node Types:")
	for _, nt := range sortedMapKeys(stats.NodesByType) {
		fmt.Printf("  %-20s %d\n", nt, stats.NodesByType[nt])
	}
	fmt.Println()

	// Coverage summary.
	fmt.Println("Coverage (structural):")
	fmt.Printf("  Sources: %d   High: %d   Medium: %d   Low: %d\n",
		coverage.SourceCount,
		coverage.BandCounts[depgraph.CoverageBandHigh],
		coverage.BandCounts[depgraph.CoverageBandMedium],
		coverage.BandCounts[depgraph.CoverageBandLow])
	fmt.Println()

	// Duplicates summary.
	fmt.Println("Duplicates:")
	fmt.Printf("  Tests analyzed: %d   Duplicates: %d   Clusters: %d\n",
		duplicates.TestsAnalyzed, duplicates.DuplicateCount, len(duplicates.Clusters))
	if duplicates.Skipped {
		fmt.Printf("  Note: %s\n", duplicates.SkipReason)
	}
	if len(duplicates.Clusters) > 0 {
		top := duplicates.Clusters[0]
		fmt.Printf("  Top cluster: %d tests, similarity %.2f\n", len(top.Tests), top.Similarity)
	}
	fmt.Println()

	// Fanout summary.
	fmt.Println("Fanout:")
	fmt.Printf("  Nodes: %d   Flagged: %d   Threshold: %d\n",
		fanout.NodeCount, fanout.FlaggedCount, fanout.Threshold)
	if fanout.Skipped {
		fmt.Printf("  Note: %s\n", fanout.SkipReason)
	}
	if len(fanout.Entries) > 0 {
		top := fanout.Entries[0]
		fmt.Printf("  Highest: %s (transitive: %d)\n", top.NodeID, top.TransitiveFanout)
	}
	fmt.Println()

	// Impact (if requested).
	if impactResult != nil {
		fmt.Println("Impact:")
		fmt.Printf("  Changed files: %d   Impacted tests: %d\n",
			len(impactResult.ChangedFiles), len(impactResult.Tests))
		fmt.Printf("  High: %d   Medium: %d   Low: %d\n",
			impactResult.LevelCounts["high"],
			impactResult.LevelCounts["medium"],
			impactResult.LevelCounts["low"])
		fmt.Println()
	}

	// Profile.
	fmt.Println("Repository Profile:")
	fmt.Printf("  Test Volume:          %s\n", profile.TestVolume)
	fmt.Printf("  CI Pressure:          %s\n", profile.CIPressure)
	fmt.Printf("  Coverage Confidence:  %s\n", profile.CoverageConfidence)
	fmt.Printf("  Redundancy Level:     %s\n", profile.RedundancyLevel)
	fmt.Printf("  Fanout Burden:        %s\n", profile.FanoutBurden)
	fmt.Println()

	// Edge cases and policy.
	if len(edgeCases) > 0 {
		fmt.Println("Edge Cases:")
		for _, ec := range edgeCases {
			fmt.Printf("  [%s] %s: %s\n", ec.Severity, ec.Type, ec.Description)
		}
		fmt.Println()
	}

	if len(pol.Recommendations) > 0 {
		fmt.Println("Recommendations:")
		for _, r := range pol.Recommendations {
			fmt.Printf("  • %s\n", r)
		}
		fmt.Println()
	}

	fmt.Printf("Policy: fallback=%s  confidence=%.2f  optimization=%s\n",
		pol.FallbackLevel, pol.ConfidenceAdjustment,
		map[bool]string{true: "disabled", false: "enabled"}[pol.OptimizationDisabled])

	return nil
}

func sortedMapKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
