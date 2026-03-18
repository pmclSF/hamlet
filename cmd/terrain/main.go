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
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pmclSF/terrain/internal/engine"
	"github.com/pmclSF/terrain/internal/logging"
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
	// Parse global --log-level flag before subcommand dispatch.
	// Accepted values: quiet, debug (default: info-level).
	initLogging(os.Args[1:])

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
		gauntletFlag := analyzeCmd.String("gauntlet", "", "path to Gauntlet eval result artifact (JSON); comma-separated for multiple")
		slowThreshold := analyzeCmd.Float64("slow-threshold", defaultSlowThresholdMs, "slow test threshold in ms")
		_ = analyzeCmd.Parse(os.Args[2:])
		if err := runAnalyze(*rootFlag, *jsonFlag, *formatFlag, *verboseFlag, *writeSnapshot, *coverageFlag, *coverageRunLabelFlag, *runtimeFlag, *gauntletFlag, *slowThreshold); err != nil {
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
		verboseFlag := explainCmd.Bool("verbose", false, "show detection evidence, tiers, and confidence details")
		_ = explainCmd.Parse(os.Args[2:])
		explainArgs := explainCmd.Args()
		if len(explainArgs) == 0 {
			fmt.Fprintln(os.Stderr, "Usage: terrain explain <target> [flags]")
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, "Explain why Terrain made a decision. Target is auto-detected:")
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, "  terrain explain <test-path>     explain a test file")
			fmt.Fprintln(os.Stderr, "  terrain explain <test-id>       explain a test case by ID")
			fmt.Fprintln(os.Stderr, "  terrain explain <code-unit>     explain a code unit (path:name)")
			fmt.Fprintln(os.Stderr, "  terrain explain <owner>         explain an owner's scope")
			fmt.Fprintln(os.Stderr, "  terrain explain <scenario-id>   explain an AI/eval scenario")
			fmt.Fprintln(os.Stderr, "  terrain explain selection       explain overall test selection")
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, "Flags:")
			fmt.Fprintln(os.Stderr, "  --verbose    show detection evidence, tiers, and confidence details")
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, "See: docs/examples/explain-report.md")
			os.Exit(2)
		}
		if err := runExplain(explainArgs[0], *rootFlag, *baseRef, *jsonFlag, *verboseFlag); err != nil {
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

	case "ai":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: terrain ai <list|run|record|baseline|doctor> [flags]")
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, "Commands:")
			fmt.Fprintln(os.Stderr, "  list       list detected AI/eval scenarios and surfaces")
			fmt.Fprintln(os.Stderr, "  run        execute eval scenarios and collect results")
			fmt.Fprintln(os.Stderr, "  replay     replay and verify a previous run artifact")
			fmt.Fprintln(os.Stderr, "  record     record eval run results as a baseline snapshot")
			fmt.Fprintln(os.Stderr, "  baseline   manage eval baselines (show, compare, promote)")
			fmt.Fprintln(os.Stderr, "  doctor     validate AI/eval setup and surface configuration issues")
			os.Exit(2)
		}
		aiSub := os.Args[2]
		aiCmd := flag.NewFlagSet("ai "+aiSub, flag.ExitOnError)
		rootFlag := aiCmd.String("root", ".", "repository root to analyze")
		jsonFlag := aiCmd.Bool("json", false, "output JSON")
		aiVerboseFlag := aiCmd.Bool("verbose", false, "show detection evidence per surface")
		baseRef := aiCmd.String("base", "", "git base ref for impact-based scenario selection")
		fullFlag := aiCmd.Bool("full", false, "run all scenarios (skip impact selection)")
		dryRunFlag := aiCmd.Bool("dry-run", false, "show what would run without executing")
		_ = aiCmd.Parse(os.Args[3:])
		if aiSub == "run" {
			if err := runAIRun(*rootFlag, *jsonFlag, *baseRef, *fullFlag, *dryRunFlag); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		} else if aiSub == "replay" {
			args := aiCmd.Args()
			if len(args) == 0 {
				// Default to latest artifact.
				args = []string{filepath.Join(*rootFlag, ".terrain", "artifacts", "ai-run-latest.json")}
			}
			if err := runAIReplay(*rootFlag, *jsonFlag, args[0]); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		} else if aiSub == "list" {
			if err := runAIList(*rootFlag, *jsonFlag, *aiVerboseFlag); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		} else if err := runAI(aiSub, *rootFlag, *jsonFlag); err != nil {
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

// initLogging scans args for --log-level=<level> or --log-level <level>
// and configures the global structured logger. This runs before subcommand
// parsing so all commands inherit the configured level.
func initLogging(args []string) {
	for i, arg := range args {
		var level string
		if strings.HasPrefix(arg, "--log-level=") {
			level = strings.TrimPrefix(arg, "--log-level=")
		} else if arg == "--log-level" && i+1 < len(args) {
			level = args[i+1]
		}
		if level != "" {
			logging.Init(logging.ParseLevel(level))
			return
		}
	}
	// Default: info level (already set by logging.init()).
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
	fmt.Fprintln(os.Stderr, "AI / eval:")
	fmt.Fprintln(os.Stderr, "  ai list [flags]          list detected AI/eval scenarios and surfaces")
	fmt.Fprintln(os.Stderr, "  ai run [flags]           execute eval scenarios and collect results")
	fmt.Fprintln(os.Stderr, "  ai record [flags]        record eval run results as a baseline snapshot")
	fmt.Fprintln(os.Stderr, "  ai baseline [flags]      manage eval baselines (show, compare, promote)")
	fmt.Fprintln(os.Stderr, "  ai doctor [flags]        validate AI/eval setup and configuration")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Advanced / debug:")
	fmt.Fprintln(os.Stderr, "  debug graph [flags]      dependency graph statistics")
	fmt.Fprintln(os.Stderr, "  debug coverage [flags]   structural coverage analysis")
	fmt.Fprintln(os.Stderr, "  debug fanout [flags]     high-fanout node analysis")
	fmt.Fprintln(os.Stderr, "  debug duplicates [flags] duplicate test cluster analysis")
	fmt.Fprintln(os.Stderr, "  debug depgraph [flags]   full dependency graph analysis (all engines)")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Benchmark / validation (separate binaries):")
	fmt.Fprintln(os.Stderr, "  terrain-bench            run benchmark suite across repos (go run ./cmd/terrain-bench)")
	fmt.Fprintln(os.Stderr, "  terrain-truthcheck       validate output against ground truth (go run ./cmd/terrain-truthcheck)")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Common flags:")
	fmt.Fprintln(os.Stderr, "  --root PATH              repository root (default: current directory)")
	fmt.Fprintln(os.Stderr, "  --json                   machine-readable JSON output")
	fmt.Fprintln(os.Stderr, "  --base REF               git base ref for diff (impact, pr, select-tests)")
	fmt.Fprintln(os.Stderr, "  --log-level LEVEL        diagnostic verbosity: quiet, debug (default: info)")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Typical flow:")
	fmt.Fprintln(os.Stderr, "  1. terrain analyze                    understand your test system")
	fmt.Fprintln(os.Stderr, "  2. terrain insights                   find what to improve")
	fmt.Fprintln(os.Stderr, "  3. terrain impact                     see what a change affects")
	fmt.Fprintln(os.Stderr, "  4. terrain explain <target>           understand why")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Docs: docs/examples/{analyze,summary,insights,explain,focus,impact}-report.md")
}

func defaultPipelineOptions() engine.PipelineOptions {
	return engine.PipelineOptions{
		EngineVersion: version,
	}
}

// defaultPipelineOptionsWithProgress returns pipeline options with progress
// reporting enabled for interactive terminals. Pass jsonOutput=true to
// suppress progress (keeps stdout clean for JSON).
func defaultPipelineOptionsWithProgress(jsonOutput bool) engine.PipelineOptions {
	return engine.PipelineOptions{
		EngineVersion: version,
		OnProgress:    newProgressFunc(jsonOutput),
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
