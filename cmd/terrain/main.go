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
//	terrain serve                local HTTP server with HTML report and JSON API
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
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	conv "github.com/pmclSF/terrain/internal/convert"
	"github.com/pmclSF/terrain/internal/engine"
	"github.com/pmclSF/terrain/internal/logging"
	"github.com/pmclSF/terrain/internal/mcp"
	"github.com/pmclSF/terrain/internal/models"
	"github.com/pmclSF/terrain/internal/server"
	"github.com/pmclSF/terrain/internal/telemetry"
)

// Build-time variables set via ldflags.
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

const defaultSlowThresholdMs = 5000.0

// Exit codes. CI scripts can distinguish failure modes from these
// without parsing stderr. Codes 0–2 preserve their pre-0.1.2 meanings;
// codes 4+ are additive.
//
//	0 — success
//	1 — runtime / analysis error (file not found, parse failed, IO error)
//	2 — usage error OR policy violation (overloaded for back-compat;
//	     both meanings retained because at least one consumer pattern-
//	     matches `exit 2 == policy fail today`)
//	3 — reserved for "policy violation" once code 2's overload is
//	     split. The split is a behavior-breaking change that needs a
//	     migration window; do not use for new codepaths until then.
//	4 — AI gate block. Returned by `terrain ai run --baseline` when the
//	     `actionBlock` decision fires (e.g., a high-severity AI signal
//	     introduced vs. baseline). Reserved by `exitAIGateBlock` so a
//	     standalone `terrain ai gate` command later can use the same
//	     code without breaking CI scripts that already branch on it.
//	5 — Not-found. Returned by `terrain show <kind> <id>` and
//	     `terrain explain <target>` when the entity doesn't exist.
//	     Lets CI distinguish "the thing you asked about isn't here"
//	     from "the analysis crashed."
//	6 — Severity gate block. Returned by `terrain analyze --fail-on`
//	     when the report contains at least one finding at or above the
//	     requested severity. Same pattern as code 4 (AI gate); CI
//	     scripts can branch on "the analysis succeeded but the gate
//	     blocked us" without parsing stderr.
//
// Splitting code 2 cleanly into "usage" vs "policy" is a behavior-
// breaking change that needs a migration window — deferred until that
// runway is in place.
const (
	exitOK              = 0
	exitError           = 1
	exitUsageError      = 2
	exitPolicyViolation = 2 // overloaded with exitUsageError; split deferred behind a migration window
	exitAIGateBlock     = 4
	// exitNotFound distinguishes "the entity you asked about does not
	// exist in this repo" from "analysis itself failed." Used by
	// `terrain show` and `terrain explain` so CI scripts can branch on
	// "missing entity" without parsing stderr text. Earlier revisions
	// collapsed not-found into exit 1, indistinguishable from a real
	// analysis crash.
	exitNotFound = 5
	// exitSeverityGateBlock signals that `--fail-on` blocked a
	// successful analysis. Code 6 leaves room for code 3 (planned for
	// "policy fail" once 2 is split) without colliding.
	exitSeverityGateBlock = 6
)

func main() {
	// Extract --mechanisms.<name>=<state> flags from os.Args before
	// subcommand dispatch. Each occurrence is captured into
	// extractedMechanismOverrides; the args slice is rewritten with the
	// matching entries removed so subcommands don't need to know about
	// this global flag.
	if overrides, rest := extractMechanismOverrides(os.Args[1:]); len(overrides) > 0 {
		extractedMechanismOverrides = overrides
		os.Args = append([]string{os.Args[0]}, rest...)
	}

	// Parse global --log-level flag before subcommand dispatch.
	// Accepted values: quiet, debug (default: info-level).
	initLogging(os.Args[1:])

	startedAt := time.Now()
	commandForTelemetry := telemetryCommandName(os.Args)
	defer func() {
		telemetry.Record(telemetry.Event{
			Timestamp:  startedAt.UTC(),
			Version:    version,
			Command:    commandForTelemetry,
			DurationMs: time.Since(startedAt).Milliseconds(),
		})
	}()

	// No subcommand → discovery report. Friendly first-touch: scans the
	// current directory and prints what Terrain sees (frameworks, AI
	// surfaces, schemas, traces) plus three next-step commands.
	// `terrain --help` still routes to printUsage() (see flag check below).
	if len(os.Args) < 2 {
		if err := runDiscover("."); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Help flags still print full usage.
	if os.Args[1] == "--help" || os.Args[1] == "-h" || os.Args[1] == "help" {
		printUsage()
		return
	}

	// Global --print-network flag prints the unified network of
	// detected surfaces / evals / tests / code units and exits.
	// Must run before subcommand dispatch.
	if os.Args[1] == "--print-network" {
		root := "."
		for i := 2; i < len(os.Args); i++ {
			if os.Args[i] == "--root" && i+1 < len(os.Args) {
				root = os.Args[i+1]
			} else if strings.HasPrefix(os.Args[i], "--root=") {
				// The `--root=<dir>` form must be honored too: silently auditing
				// the CWD while reporting success would undermine the very
				// zero-outbound-network claim this command exists to verify.
				root = strings.TrimPrefix(os.Args[i], "--root=")
			}
		}
		if err := runPrintNetwork(root); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	switch os.Args[1] {
	case "fix":
		fixCmd := flag.NewFlagSet("fix", flag.ExitOnError)
		rootFlag := fixCmd.String("root", ".", "repository root to fix")
		applyFlag := fixCmd.Bool("apply", false, "write the validated fixes to disk (default: dry-run — show the diff, change nothing)")
		_ = fixCmd.Parse(os.Args[2:])
		mountPositionalAsRoot("fix", fixCmd.Args(), rootFlag)
		if err := runFix(*rootFlag, *applyFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(exitError)
		}

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
		promptfooFlag := analyzeCmd.String("promptfoo-results", "", "path to Promptfoo --output result file(s); comma-separated for multiple")
		deepevalFlag := analyzeCmd.String("deepeval-results", "", "path to DeepEval result JSON file(s); comma-separated for multiple")
		ragasFlag := analyzeCmd.String("ragas-results", "", "path to Ragas eval result file(s); comma-separated for multiple")
		greatExpectationsFlag := analyzeCmd.String("great-expectations-results", "", "path to Great Expectations validation result JSON file(s); comma-separated for multiple")
		baselineFlag := analyzeCmd.String("baseline", "", "path to a previous snapshot JSON file; enables regression-aware detectors and --new-findings-only filtering")
		slowThreshold := analyzeCmd.Float64("slow-threshold", defaultSlowThresholdMs, "slow test threshold in ms")
		redactPathsFlag := analyzeCmd.Bool("redact-paths", false, "rewrite absolute paths in --format=sarif output to repo-relative form (or basename if outside repo)")
		failOnFlag := analyzeCmd.String("fail-on", "", "exit "+fmt.Sprintf("%d", exitSeverityGateBlock)+" when at least one finding is at or above this severity (critical|high|medium)")
		timeoutFlag := analyzeCmd.Duration("timeout", 0, "abort the analysis after this duration (e.g. 5m); 0 means no timeout")
		suppressionsFlag := analyzeCmd.String("suppressions", "", "path to .terrain/suppressions.yaml (default: $root/.terrain/suppressions.yaml; missing file is fine)")
		newOnlyFlag := analyzeCmd.Bool("new-findings-only", false, "filter signals to those NOT present in --baseline (lets established repos with debt adopt --fail-on without bricking CI)")
		previewFlag := analyzeCmd.Bool("preview", false, "enable preview-tier AI detectors (default off; behavior may change between releases)")
		diagFlag := analyzeCmd.Bool("diag", false, "print per-step pipeline timing diagnostics to stderr (for performance investigation)")
		baseRefFlag := analyzeCmd.String("base", "", "git base ref enabling aiPromptSchemaDrift (e.g. main, origin/main); compares schemas at HEAD against this ref")
		trustFloorFlag := analyzeCmd.Bool("trust-floor", false, "force the remediation-validity gate ON (it is the default): only findings whose remediation is closed-loop validated block CI (--fail-on)")
		noTrustFloorFlag := analyzeCmd.Bool("no-trust-floor", false, "opt out of the default remediation-validity gate: restore severity-only gating so any --fail-on match blocks CI. Also settable via trust_floor: false in terrain.yaml")
		_ = analyzeCmd.Parse(os.Args[2:])
		mountPositionalAsRoot("analyze", analyzeCmd.Args(), rootFlag)
		gate, gateErr := parseSeverityGate(*failOnFlag)
		if gateErr != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", gateErr)
			os.Exit(exitUsageError)
		}
		// Negative timeouts have no meaning; reject explicitly so the
		// user gets a clear error rather than an immediate
		// context.DeadlineExceeded that looks like an analysis failure.
		if *timeoutFlag < 0 {
			fmt.Fprintf(os.Stderr, "error: --timeout must be non-negative (got %s)\n", *timeoutFlag)
			os.Exit(exitUsageError)
		}
		if *newOnlyFlag && *baselineFlag == "" {
			fmt.Fprintln(os.Stderr, "error: --new-findings-only requires --baseline <path>")
			os.Exit(exitUsageError)
		}
		analyzeOpts := analyzeRunOpts{
			Root:                   *rootFlag,
			JSONOutput:             *jsonFlag,
			Format:                 *formatFlag,
			Verbose:                *verboseFlag,
			WriteSnapshot:          *writeSnapshot,
			CoveragePath:           *coverageFlag,
			CoverageRunLabel:       *coverageRunLabelFlag,
			RuntimePaths:           *runtimeFlag,
			GauntletPaths:          *gauntletFlag,
			PromptfooPaths:         *promptfooFlag,
			DeepEvalPaths:          *deepevalFlag,
			RagasPaths:             *ragasFlag,
			GreatExpectationsPaths: *greatExpectationsFlag,
			BaselinePath:           *baselineFlag,
			SlowThreshold:          *slowThreshold,
			RedactPaths:            *redactPathsFlag,
			Gate:                   gate,
			Timeout:                *timeoutFlag,
			SuppressionsPath:       *suppressionsFlag,
			NewFindingsOnly:        *newOnlyFlag,
			EnablePreview:          *previewFlag,
			Diag:                   *diagFlag,
			BaseRef:                *baseRefFlag,
			TrustFloor:             *trustFloorFlag,
			NoTrustFloor:           *noTrustFloorFlag,
		}
		if err := runAnalyze(analyzeOpts); err != nil {
			if errors.Is(err, errSeverityGateBlocked) {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(exitSeverityGateBlock)
			}
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "init":
		initCmd := flag.NewFlagSet("init", flag.ExitOnError)
		rootFlag := initCmd.String("root", ".", "repository root to inspect")
		jsonFlag := initCmd.Bool("json", false, "output JSON init result")
		_ = initCmd.Parse(os.Args[2:])
		if err := runInit(*rootFlag, *jsonFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "impact":
		legacyDeprecationNotice("impact", "report impact")
		impactCmd := flag.NewFlagSet("impact", flag.ExitOnError)
		rootFlag := impactCmd.String("root", ".", "repository root to analyze")
		baseRef := impactCmd.String("base", "", "git base ref for diff (default: HEAD~1)")
		jsonFlag := impactCmd.Bool("json", false, "output JSON impact result")
		showFlag := impactCmd.String("show", "", "drill-down view: units, gaps, tests, owners, graph, selected")
		ownerFlag := impactCmd.String("owner", "", "filter results by owner")
		explainFlag := impactCmd.Bool("explain-selection", false, "render the selection explanation: which tests matter for this PR — and why")
		_ = impactCmd.Parse(os.Args[2:])
		mountPositionalAsRoot("impact", impactCmd.Args(), rootFlag)
		if err := runImpact(*rootFlag, *baseRef, *jsonFlag, *showFlag, *ownerFlag, *explainFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "convert":
		// 0.2: `terrain convert` shares the canonical-verb table with
		// `terrain migrate`, but unknown first args fall through to
		// runConvertCLI (per-file converter) so the historical
		// `terrain convert <file> --to <framework>` shape keeps working.
		if err := runConvertNamespaceCLI(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(exitCodeForCLIError(err))
		}

	case "convert-config":
		legacyDeprecationNotice("convert-config", "migrate config")
		if err := runConvertConfigCLI(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(exitCodeForCLIError(err))
		}

	case "list", "list-conversions":
		legacyDeprecationNotice("list-conversions", "migrate list")
		if err := runListConversionsCLI(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(exitCodeForCLIError(err))
		}

	case "shorthands":
		legacyDeprecationNotice("shorthands", "migrate shorthands")
		if err := runShorthandsCLI(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(exitCodeForCLIError(err))
		}

	case "detect":
		legacyDeprecationNotice("detect", "migrate detect")
		if err := runDetectCLI(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(exitCodeForCLIError(err))
		}

	case "migrate":
		// `terrain migrate` is itself canonical (the namespace dispatcher).
		// No deprecation notice — it's the recommended shape.
		if err := runMigrateNamespaceCLI(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(exitCodeForCLIError(err))
		}

	case "estimate":
		legacyDeprecationNotice("estimate", "migrate estimate")
		if err := runEstimateCLI(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(exitCodeForCLIError(err))
		}

	case "status":
		legacyDeprecationNotice("status", "migrate status")
		if err := runStatusCLI(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(exitCodeForCLIError(err))
		}

	case "checklist":
		legacyDeprecationNotice("checklist", "migrate checklist")
		if err := runChecklistCLI(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(exitCodeForCLIError(err))
		}

	case "doctor":
		os.Exit(runDoctorCLI(os.Args[2:]))

	case "mechanisms":
		// Maintainer-only inspection of detector internals. Hidden from
		// --help and gated on TERRAIN_DEV so end-user CLI surfaces
		// don't expose internal mechanism IDs.
		if os.Getenv("TERRAIN_DEV") == "" {
			fmt.Fprintln(os.Stderr, "error: unknown command \"mechanisms\"")
			os.Exit(2)
		}
		if err := runMechanismsCLI(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(exitCodeForCLIError(err))
		}

	case "plugins":
		if err := runPlugins(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "inject":
		if err := runInject(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(exitCodeForCLIError(err))
		}

	case "scaffold":
		if err := runScaffold(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(exitCodeForCLIError(err))
		}

	case "mcp":
		root := "."
		for i := 2; i < len(os.Args); i++ {
			a := os.Args[i]
			if a == "--root" && i+1 < len(os.Args) {
				root = os.Args[i+1]
				i++
				continue
			}
			if len(a) > len("--root=") && a[:len("--root=")] == "--root=" {
				root = a[len("--root="):]
				continue
			}
			if isHelpArg(a) {
				fmt.Fprintln(os.Stderr, "Usage: terrain mcp [--root <dir>]")
				fmt.Fprintln(os.Stderr)
				fmt.Fprintln(os.Stderr, "Starts the MCP server on stdio. Loads the most-recent")
				fmt.Fprintln(os.Stderr, "`terrain analyze` artifacts from .terrain/ when present.")
				return
			}
		}
		// Report the real build version in the MCP handshake (the package
		// default is a stale placeholder; ldflags set only main.version).
		mcp.ServerVersion = version
		if err := runMCPCommand(root); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "webhook":
		// Self-host surface: receives GitHub webhook deliveries,
		// parses slash commands, dispatches through a Dispatcher
		// (default: informational, no GitHub write-back). Adopters
		// override the dispatcher in their integration code.
		// Gated behind TERRAIN_DEV so adopters don't accidentally run
		// the no-write-back default in production. The error below
		// names the gate explicitly so an adopter who's following
		// docs/integrations/github-checks.md sees the right cause.
		if os.Getenv("TERRAIN_DEV") == "" {
			fmt.Fprintln(os.Stderr, "error: `terrain webhook` is opt-in. Re-run with TERRAIN_DEV=1.")
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, "See docs/integrations/github-checks.md for the production deploy guide,")
			fmt.Fprintln(os.Stderr, "including the Dockerfile that sets TERRAIN_DEV=1 by default.")
			os.Exit(2)
		}
		addr := ":4242"
		for i := 2; i < len(os.Args); i++ {
			a := os.Args[i]
			if a == "--addr" && i+1 < len(os.Args) {
				addr = os.Args[i+1]
				i++
				continue
			}
			if len(a) > len("--addr=") && a[:len("--addr=")] == "--addr=" {
				addr = a[len("--addr="):]
				continue
			}
			if isHelpArg(a) {
				fmt.Fprintln(os.Stderr, "Usage: terrain webhook [--addr=:4242]")
				fmt.Fprintln(os.Stderr)
				fmt.Fprintln(os.Stderr, "Start the GitHub webhook server.")
				fmt.Fprintln(os.Stderr, "Requires TERRAIN_WEBHOOK_SECRET (same value GitHub uses).")
				return
			}
		}
		if err := runWebhook(addr); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "reset":
		legacyDeprecationNotice("reset", "config reset")
		if err := runResetCLI(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(exitCodeForCLIError(err))
		}

	case "policy":
		legacyDeprecationNotice("policy check", "analyze --policy=<file>")
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
		legacyDeprecationNotice("metrics", "report metrics")
		metricsCmd := flag.NewFlagSet("metrics", flag.ExitOnError)
		rootFlag := metricsCmd.String("root", ".", "repository root to analyze")
		jsonFlag := metricsCmd.Bool("json", false, "output JSON metrics snapshot")
		verboseFlag := metricsCmd.Bool("verbose", false, "show detailed metric breakdowns")
		_ = metricsCmd.Parse(os.Args[2:])
		mountPositionalAsRoot("metrics", metricsCmd.Args(), rootFlag)
		if err := runMetrics(*rootFlag, *jsonFlag, *verboseFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "posture":
		legacyDeprecationNotice("posture", "report posture")
		postureCmd := flag.NewFlagSet("posture", flag.ExitOnError)
		rootFlag := postureCmd.String("root", ".", "repository root to analyze")
		jsonFlag := postureCmd.Bool("json", false, "output JSON posture snapshot")
		verboseFlag := postureCmd.Bool("verbose", false, "show measurement values and thresholds")
		_ = postureCmd.Parse(os.Args[2:])
		mountPositionalAsRoot("posture", postureCmd.Args(), rootFlag)
		if err := runPosture(*rootFlag, *jsonFlag, *verboseFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "portfolio":
		legacyDeprecationNotice("portfolio", "report portfolio")
		portfolioCmd := flag.NewFlagSet("portfolio", flag.ExitOnError)
		rootFlag := portfolioCmd.String("root", ".", "repository root to analyze")
		fromFlag := portfolioCmd.String("from", "", "multi-repo manifest path (.terrain/repos.yaml)")
		jsonFlag := portfolioCmd.Bool("json", false, "output JSON portfolio snapshot")
		verboseFlag := portfolioCmd.Bool("verbose", false, "show per-asset details")
		_ = portfolioCmd.Parse(os.Args[2:])
		mountPositionalAsRoot("portfolio", portfolioCmd.Args(), rootFlag)
		if err := runPortfolioWithManifest(*rootFlag, *fromFlag, *jsonFlag, *verboseFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "insights":
		legacyDeprecationNotice("insights", "report insights")
		insightsCmd := flag.NewFlagSet("insights", flag.ExitOnError)
		rootFlag := insightsCmd.String("root", ".", "repository root to analyze")
		jsonFlag := insightsCmd.Bool("json", false, "output JSON insights")
		verboseFlag := insightsCmd.Bool("verbose", false, "show per-finding evidence and file details")
		_ = insightsCmd.Parse(os.Args[2:])
		mountPositionalAsRoot("insights", insightsCmd.Args(), rootFlag)
		if err := runInsights(*rootFlag, *jsonFlag, *verboseFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "explain":
		legacyDeprecationNotice("explain", "report explain")
		explainCmd := flag.NewFlagSet("explain", flag.ExitOnError)
		rootFlag := explainCmd.String("root", ".", "repository root to analyze")
		baseRef := explainCmd.String("base", "", "git base ref for diff (default: HEAD~1)")
		jsonFlag := explainCmd.Bool("json", false, "output JSON")
		verboseFlag := explainCmd.Bool("verbose", false, "show detection evidence, tiers, and confidence details")
		explainFlagsWithValue := map[string]bool{
			"--root": true, "--base": true,
		}
		_ = explainCmd.Parse(reorderCLIArgs(os.Args[2:], explainFlagsWithValue))
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
			fmt.Fprintln(os.Stderr, "  terrain explain <finding-id>    explain a finding (e.g. weakAssertion@path:Sym#hash)")
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
			os.Exit(exitCodeForCLIError(err))
		}

	case "suppress":
		// `terrain suppress <finding-id> --reason "why" [--expires YYYY-MM-DD] [--owner @who]`
		// appends a suppression entry to .terrain/suppressions.yaml.
		// No legacy alias: this command is new in 0.2.0 and lives in the
		// canonical surface as a Gate-pillar primitive.
		suppressCmd := flag.NewFlagSet("suppress", flag.ExitOnError)
		rootFlag := suppressCmd.String("root", ".", "repository root")
		reasonFlag := suppressCmd.String("reason", "", "why this finding is being suppressed (required)")
		expiresFlag := suppressCmd.String("expires", "", "ISO date YYYY-MM-DD when the suppression should expire (default: per-scope; instance=+90d, file=+180d, directory=+180d, repo=+365d)")
		ownerFlag := suppressCmd.String("owner", "", "owner pointer for review (optional)")
		scopeFlag := suppressCmd.String("scope", "", "instance | file | directory | repo (default: instance)")
		// Go's flag package stops at the first non-flag token, so the
		// `suppress <finding-id> --reason "..."` order would silently drop the
		// flags. Extract the leading finding-id positional, then parse the rest.
		rawArgs := os.Args[2:]
		var findingID string
		flagArgs := make([]string, 0, len(rawArgs))
		for _, a := range rawArgs {
			if findingID == "" && a != "" && !strings.HasPrefix(a, "-") {
				findingID = a
				continue
			}
			flagArgs = append(flagArgs, a)
		}
		_ = suppressCmd.Parse(flagArgs)
		if findingID == "" {
			// fall back to any positional that survived (e.g. flags-first order)
			if rest := suppressCmd.Args(); len(rest) > 0 {
				findingID = rest[0]
			}
		}
		if findingID == "" {
			fmt.Fprintln(os.Stderr, "Usage: terrain suppress <finding-id> --reason \"why\" [--expires YYYY-MM-DD] [--owner @who] [--scope instance|file|directory|repo]")
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, "Find a finding's ID with: terrain explain <finding-id>")
			os.Exit(exitUsageError)
		}
		if err := runSuppress(findingID, *reasonFlag, *expiresFlag, *ownerFlag, *scopeFlag, *rootFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(exitCodeForCLIError(err))
		}

	case "summary":
		legacyDeprecationNotice("summary", "report summary")
		summaryCmd := flag.NewFlagSet("summary", flag.ExitOnError)
		rootFlag := summaryCmd.String("root", ".", "repository root to analyze")
		jsonFlag := summaryCmd.Bool("json", false, "output JSON summary with heatmap")
		verboseFlag := summaryCmd.Bool("verbose", false, "show detailed heatmap breakdown")
		_ = summaryCmd.Parse(os.Args[2:])
		mountPositionalAsRoot("summary", summaryCmd.Args(), rootFlag)
		if err := runSummary(*rootFlag, *jsonFlag, *verboseFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "focus":
		legacyDeprecationNotice("focus", "report focus")
		focusCmd := flag.NewFlagSet("focus", flag.ExitOnError)
		rootFlag := focusCmd.String("root", ".", "repository root to analyze")
		jsonFlag := focusCmd.Bool("json", false, "output JSON focus summary")
		verboseFlag := focusCmd.Bool("verbose", false, "show full rationale and dependency chains")
		_ = focusCmd.Parse(os.Args[2:])
		mountPositionalAsRoot("focus", focusCmd.Args(), rootFlag)
		if err := runFocus(*rootFlag, *jsonFlag, *verboseFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "compare":
		legacyDeprecationNotice("compare", "report compare")
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
		legacyDeprecationNotice("migration", "migrate")
		if len(os.Args) < 3 {
			printMigrationUsage()
			os.Exit(2)
		}
		if isHelpArg(os.Args[2]) {
			printMigrationUsage()
			return
		}
		subCmd := os.Args[2]
		var err error
		switch subCmd {
		case "readiness", "blockers":
			migCmd := flag.NewFlagSet("migration "+subCmd, flag.ExitOnError)
			rootFlag := migCmd.String("root", ".", "repository root to analyze")
			jsonFlag := migCmd.Bool("json", false, "output JSON")
			_ = migCmd.Parse(os.Args[3:])
			err = runMigration(subCmd, *rootFlag, *jsonFlag, "", "")
		case "preview":
			migCmd := flag.NewFlagSet("migration preview", flag.ExitOnError)
			rootFlag := migCmd.String("root", ".", "repository root to analyze")
			jsonFlag := migCmd.Bool("json", false, "output JSON")
			fileFlag := migCmd.String("file", "", "file path for preview (relative to root)")
			scopeFlag := migCmd.String("scope", "", "directory scope for preview")
			_ = migCmd.Parse(os.Args[3:])
			err = runMigration(subCmd, *rootFlag, *jsonFlag, *fileFlag, *scopeFlag)
		default:
			fmt.Fprintf(os.Stderr, "unknown migration subcommand: %q (valid: readiness, blockers, preview)\n", subCmd)
			os.Exit(2)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "select-tests":
		legacyDeprecationNotice("select-tests", "report select-tests")
		stCmd := flag.NewFlagSet("select-tests", flag.ExitOnError)
		rootFlag := stCmd.String("root", ".", "repository root to analyze")
		baseRef := stCmd.String("base", "", "git base ref for diff (default: HEAD~1)")
		jsonFlag := stCmd.Bool("json", false, "output JSON protective test set")
		formatFlag := stCmd.String("format", "", "output format: paths (one test path per line), json, or text")
		_ = stCmd.Parse(os.Args[2:])
		mountPositionalAsRoot("select-tests", stCmd.Args(), rootFlag)
		if err := runSelectTests(*rootFlag, *baseRef, *jsonFlag, *formatFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "pr":
		legacyDeprecationNotice("pr", "report pr")
		prCmd := flag.NewFlagSet("pr", flag.ExitOnError)
		rootFlag := prCmd.String("root", ".", "repository root to analyze")
		baseRef := prCmd.String("base", "", "git base ref for diff (default: HEAD~1)")
		jsonFlag := prCmd.Bool("json", false, "output JSON PR analysis")
		formatFlag := prCmd.String("format", "", "output format: markdown, comment, annotation")
		failOnFlag := prCmd.String("fail-on", "", "exit "+fmt.Sprintf("%d", exitSeverityGateBlock)+" when at least one finding (NewFindings + AI BlockingSignals) is at or above this severity (critical|high|medium)")
		baselineFlag := prCmd.String("baseline", "", "path to a previous snapshot JSON file; enables --new-findings-only filtering")
		newOnlyFlag := prCmd.Bool("new-findings-only", false, "filter signals to those NOT present in --baseline before PR analysis")
		_ = prCmd.Parse(os.Args[2:])
		mountPositionalAsRoot("pr", prCmd.Args(), rootFlag)
		gate, gateErr := parseSeverityGate(*failOnFlag)
		if gateErr != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", gateErr)
			os.Exit(exitUsageError)
		}
		if *newOnlyFlag && *baselineFlag == "" {
			fmt.Fprintln(os.Stderr, "error: --new-findings-only requires --baseline <path>")
			os.Exit(exitUsageError)
		}
		if err := runPR(prRunOpts{
			Root:            *rootFlag,
			BaseRef:         *baseRef,
			JSONOutput:      *jsonFlag,
			Format:          *formatFlag,
			Gate:            gate,
			BaselinePath:    *baselineFlag,
			NewFindingsOnly: *newOnlyFlag,
		}); err != nil {
			if errors.Is(err, errSeverityGateBlocked) {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(exitSeverityGateBlock)
			}
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "show":
		legacyDeprecationNotice("show", "report show")
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
		// Separate flags from positional args manually so flags can appear
		// in any position (e.g., "terrain show --json test foo",
		// "terrain show test foo --json", or "terrain show test --json foo").
		var showPositional []string
		showJSON := false
		showRoot := "."
		for _, arg := range os.Args[2:] {
			switch {
			case arg == "--json" || arg == "-json":
				showJSON = true
			case strings.HasPrefix(arg, "--root="):
				showRoot = strings.TrimPrefix(arg, "--root=")
			case strings.HasPrefix(arg, "-root="):
				showRoot = strings.TrimPrefix(arg, "-root=")
			case arg == "--root" || arg == "-root":
				// Next arg would be the root value — handled below.
				showRoot = ""
			default:
				if showRoot == "" {
					showRoot = arg // consume the value after --root
				} else {
					showPositional = append(showPositional, arg)
				}
			}
		}
		if showRoot == "" {
			showRoot = "." // --root provided without value; fall back to cwd
		}
		showSubCmd := ""
		showID := ""
		if len(showPositional) > 0 {
			showSubCmd = showPositional[0]
		}
		if len(showPositional) > 1 {
			showID = showPositional[1]
		}
		if showSubCmd == "" {
			printShowUsage()
			os.Exit(2)
		}
		rootFlag := &showRoot
		jsonFlag := &showJSON
		if err := runShow(showSubCmd, showID, *rootFlag, *jsonFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(exitCodeForCLIError(err))
		}

	case "export":
		legacyDeprecationNotice("export benchmark", "report export-benchmark")
		if len(os.Args) < 3 {
			printExportUsage()
			os.Exit(2)
		}
		if isHelpArg(os.Args[2]) {
			printExportUsage()
			return
		}
		if os.Args[2] != "benchmark" {
			fmt.Fprintf(os.Stderr, "unknown export subcommand: %q\n\n", os.Args[2])
			printExportUsage()
			os.Exit(2)
		}
		exportCmd := flag.NewFlagSet("export benchmark", flag.ExitOnError)
		rootFlag := exportCmd.String("root", ".", "repository root to analyze")
		// --json is accepted but a no-op: export benchmark always
		// emits JSON. Kept for flag parity with other commands.
		_ = exportCmd.Bool("json", false, "machine-readable output (default; this command always emits JSON)")
		_ = exportCmd.Parse(os.Args[3:])
		if err := runExportBenchmark(*rootFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "report":
		if err := runReportNamespaceCLI(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(exitCodeForCLIError(err))
		}

	case "config":
		if err := runConfigNamespaceCLI(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(exitCodeForCLIError(err))
		}

	case "debug":
		if len(os.Args) < 3 {
			printDebugUsage()
			os.Exit(2)
		}
		if isHelpArg(os.Args[2]) {
			printDebugUsage()
			return
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
		legacyDeprecationNotice("depgraph", "debug depgraph")
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
			printAIUsage()
			os.Exit(2)
		}
		if isHelpArg(os.Args[2]) {
			printAIUsage()
			return
		}
		aiSub := os.Args[2]
		switch aiSub {
		case "list":
			aiCmd := flag.NewFlagSet("ai list", flag.ExitOnError)
			rootFlag := aiCmd.String("root", ".", "repository root to analyze")
			jsonFlag := aiCmd.Bool("json", false, "output JSON")
			verboseFlag := aiCmd.Bool("verbose", false, "show detection evidence per surface")
			_ = aiCmd.Parse(os.Args[3:])
			if err := runAIList(*rootFlag, *jsonFlag, *verboseFlag); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		case "findings":
			aiCmd := flag.NewFlagSet("ai findings", flag.ExitOnError)
			rootFlag := aiCmd.String("root", ".", "repository root to analyze")
			jsonFlag := aiCmd.Bool("json", false, "output JSON")
			verboseFlag := aiCmd.Bool("verbose", false, "show per-evidence-atom weight scores")
			postureFlag := aiCmd.String("posture", "observability",
				"emission posture: observability | gate")
			ruleFlag := aiCmd.String("rule", "",
				"rule to evaluate (default: ai.surface.missing_eval; see docs/rules/ai/)")
			_ = aiCmd.Parse(os.Args[3:])
			if err := runAIFindings(*rootFlag, *jsonFlag, *verboseFlag, *postureFlag, *ruleFlag); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		case "run":
			aiCmd := flag.NewFlagSet("ai run", flag.ExitOnError)
			rootFlag := aiCmd.String("root", ".", "repository root to analyze")
			jsonFlag := aiCmd.Bool("json", false, "output JSON")
			baseRef := aiCmd.String("base", "", "git base ref for impact-based scenario selection")
			fullFlag := aiCmd.Bool("full", false, "run all scenarios (skip impact selection)")
			dryRunFlag := aiCmd.Bool("dry-run", false, "show what would run without executing")
			timeoutFlag := aiCmd.Duration("timeout", 0, "abort eval execution after this duration (e.g. 10m); 0 means no timeout")
			_ = aiCmd.Parse(os.Args[3:])
			if *timeoutFlag < 0 {
				fmt.Fprintf(os.Stderr, "error: --timeout must be non-negative (got %s)\n", *timeoutFlag)
				os.Exit(exitUsageError)
			}
			if err := runAIRunWithTimeout(*rootFlag, *jsonFlag, *baseRef, *fullFlag, *dryRunFlag, *timeoutFlag); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(exitCodeForCLIError(err))
			}
		case "replay":
			aiCmd := flag.NewFlagSet("ai replay", flag.ExitOnError)
			rootFlag := aiCmd.String("root", ".", "repository root to analyze")
			jsonFlag := aiCmd.Bool("json", false, "output JSON")
			_ = aiCmd.Parse(os.Args[3:])
			args := aiCmd.Args()
			if len(args) == 0 {
				// Default to latest artifact.
				args = []string{filepath.Join(*rootFlag, ".terrain", "artifacts", "ai-run-latest.json")}
			}
			if err := runAIReplay(*rootFlag, *jsonFlag, args[0]); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		case "record":
			aiCmd := flag.NewFlagSet("ai record", flag.ExitOnError)
			rootFlag := aiCmd.String("root", ".", "repository root to analyze")
			jsonFlag := aiCmd.Bool("json", false, "output JSON")
			_ = aiCmd.Parse(os.Args[3:])
			if err := runAIRecord(*rootFlag, *jsonFlag); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		case "baseline":
			aiCmd := flag.NewFlagSet("ai baseline", flag.ExitOnError)
			rootFlag := aiCmd.String("root", ".", "repository root to analyze")
			jsonFlag := aiCmd.Bool("json", false, "output JSON")
			_ = aiCmd.Parse(os.Args[3:])
			// Check for sub-subcommand: terrain ai baseline compare
			baselineArgs := aiCmd.Args()
			if len(baselineArgs) > 0 && baselineArgs[0] == "compare" {
				if err := runAIBaselineCompare(*rootFlag, *jsonFlag); err != nil {
					fmt.Fprintf(os.Stderr, "error: %v\n", err)
					os.Exit(1)
				}
				return
			}
			if err := runAIBaseline(*rootFlag, *jsonFlag); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		case "doctor":
			aiCmd := flag.NewFlagSet("ai doctor", flag.ExitOnError)
			rootFlag := aiCmd.String("root", ".", "repository root to analyze")
			jsonFlag := aiCmd.Bool("json", false, "output JSON")
			_ = aiCmd.Parse(os.Args[3:])
			if err := runAIDoctor(*rootFlag, *jsonFlag); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		default:
			fmt.Fprintf(os.Stderr, "error: unknown ai subcommand: %q\n\n", aiSub)
			printAIUsage()
			os.Exit(exitUsageError)
		}

	case "feedback":
		legacyDeprecationNotice("feedback", "config feedback")
		url := "https://github.com/pmclSF/terrain/issues/new?template=feedback.md&title=Feedback:+&labels=feedback"
		fmt.Println("Open the following URL to share feedback:")
		fmt.Println()
		fmt.Printf("  %s\n", url)
		fmt.Println()
		fmt.Println("Or email: terrain-feedback@pmcl.dev")

	case "telemetry":
		legacyDeprecationNotice("telemetry", "config telemetry")
		if len(os.Args) < 3 {
			fmt.Println("Telemetry:", telemetry.Status())
			fmt.Println()
			fmt.Println("Usage:")
			fmt.Println("  terrain telemetry --on     enable local telemetry")
			fmt.Println("  terrain telemetry --off    disable local telemetry")
			fmt.Println("  terrain telemetry --status show current state")
			fmt.Println()
			fmt.Println("Telemetry records command name and duration.")
			fmt.Println("Some commands may add repo size band, languages, and signal count.")
			fmt.Println("Events are written to ~/.terrain/telemetry.jsonl.")
			fmt.Println("No file paths, repo URLs, or PII are recorded.")
			fmt.Println("Override with TERRAIN_TELEMETRY=on|off environment variable.")
			return
		}
		switch os.Args[2] {
		case "--on", "on":
			if err := telemetry.SaveConfig(telemetry.Config{Enabled: true}); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Telemetry enabled. Events will be written to ~/.terrain/telemetry.jsonl")
		case "--off", "off":
			if err := telemetry.SaveConfig(telemetry.Config{Enabled: false}); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Telemetry disabled.")
		case "--status", "status":
			fmt.Println("Telemetry:", telemetry.Status())
		default:
			fmt.Fprintf(os.Stderr, "unknown telemetry subcommand: %q\n", os.Args[2])
			os.Exit(1)
		}

	case "version", "--version", "-v":
		versionCmd := flag.NewFlagSet("version", flag.ExitOnError)
		jsonFlag := versionCmd.Bool("json", false, "output JSON version info")
		_ = versionCmd.Parse(os.Args[2:])
		if *jsonFlag {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			// schemaVersion is the snapshot-JSON contract version this
			// binary produces; CI tooling that pins on the snapshot
			// shape can gate on this field. Earlier revisions only
			// carried version/commit/date — consumers had to load a
			// snapshot and read its `meta.schemaVersion` to find out.
			_ = enc.Encode(map[string]string{
				"version":       version,
				"commit":        commit,
				"date":          date,
				"schemaVersion": models.SnapshotSchemaVersion,
			})
			return
		}
		fmt.Printf("terrain %s (commit %s, built %s; snapshot schema %s)\n",
			version, commit, date, models.SnapshotSchemaVersion)

	case "serve":
		serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
		rootFlag := serveCmd.String("root", ".", "repository root to analyze")
		portFlag := serveCmd.Int("port", server.DefaultPort, "port to listen on")
		hostFlag := serveCmd.String("host", server.DefaultHost, "bind host (default 127.0.0.1; non-localhost values are unauthenticated and warned about)")
		readOnlyFlag := serveCmd.Bool("read-only", true, "reject any non-GET/HEAD/OPTIONS request with 405; pass --read-only=false to opt out for local experiments")
		_ = serveCmd.Parse(os.Args[2:])
		if err := runServe(*rootFlag, *portFlag, *hostFlag, *readOnlyFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "test":
		testCmd := flag.NewFlagSet("test", flag.ExitOnError)
		rootFlag := testCmd.String("root", ".", "repository root to analyze")
		selectorFlag := testCmd.String("selector", "", "rule selector (e.g., 'regression/test-failed' or 'coverage/*')")
		jsonFlag := testCmd.Bool("json", false, "emit findings.json instead of human-readable output")
		junitFlag := testCmd.String("junit", "", "write JUnit XML to the given path")
		summaryFlag := testCmd.String("summary", "", "write Step Summary markdown to the given path (set to $GITHUB_STEP_SUMMARY in GitHub Actions)")
		failOnFlag := testCmd.String("fail-on", "", "exit "+fmt.Sprintf("%d", exitSeverityGateBlock)+" when at least one finding is at or above this severity (critical|high|medium)")
		baselineFlag := testCmd.String("baseline", "", "path to a previous snapshot JSON file; enables --new-findings-only filtering")
		newOnlyFlag := testCmd.Bool("new-findings-only", false, "filter signals to those NOT present in --baseline before gating")
		noTrustFloorFlag := testCmd.Bool("no-trust-floor", false, "opt out of the default remediation-validity gate: any --fail-on match blocks CI (matches terrain analyze --no-trust-floor)")
		_ = testCmd.Parse(os.Args[2:])
		gate, gateErr := parseSeverityGate(*failOnFlag)
		if gateErr != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", gateErr)
			os.Exit(exitUsageError)
		}
		if *newOnlyFlag && *baselineFlag == "" {
			fmt.Fprintln(os.Stderr, "error: --new-findings-only requires --baseline <path>")
			os.Exit(exitUsageError)
		}
		if err := runTestCommand(testRunOpts{
			Root:            *rootFlag,
			Selector:        *selectorFlag,
			JSONOutput:      *jsonFlag,
			JUnitPath:       *junitFlag,
			SummaryPath:     *summaryFlag,
			Gate:            gate,
			BaselinePath:    *baselineFlag,
			NewFindingsOnly: *newOnlyFlag,
			NoTrustFloor:    *noTrustFloorFlag,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(exitCodeForCLIError(err))
		}

	case "describe":
		descCmd := flag.NewFlagSet("describe", flag.ExitOnError)
		rootFlag := descCmd.String("root", ".", "repository root to analyze")
		writeFlag := descCmd.Bool("write", false, "write detected surfaces into terrain.yaml (prompts for overwrite)")
		_ = descCmd.Parse(os.Args[2:])
		mountPositionalAsRoot("describe", descCmd.Args(), rootFlag)
		if err := runDescribeCommand(*rootFlag, *writeFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "accept-snapshot":
		acceptCmd := flag.NewFlagSet("accept-snapshot", flag.ExitOnError)
		rootFlag := acceptCmd.String("root", ".", "repository root")
		yesFlag := acceptCmd.Bool("yes", false, "skip the interactive confirmation prompt")
		_ = acceptCmd.Parse(os.Args[2:])
		if err := runAcceptSnapshotCommand(*rootFlag, *yesFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "--help", "-h", "help":
		printUsage()
	default:
		if _, ok := conv.LookupShorthand(os.Args[1]); ok {
			if err := runShorthandCLI(os.Args[1], os.Args[2:]); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(exitCodeForCLIError(err))
			}
			return
		}
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		if suggestions := didYouMean(os.Args[1], 3); len(suggestions) > 0 {
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, "Did you mean one of these?")
			for _, s := range suggestions {
				fmt.Fprintf(os.Stderr, "  terrain %s\n", s)
			}
			fmt.Fprintln(os.Stderr)
		}
		fmt.Fprintln(os.Stderr, "Run `terrain --help` for the full command surface.")
		os.Exit(exitUsageError)
	}
}

// knownCommands lists every top-level command the dispatcher accepts. Kept
// in sync with the switch statement in main(); when you add a new case,
// add it here so `terrain mistyped-name` can suggest it.
var knownCommands = []string{
	"analyze", "init", "impact", "explain", "insights", "summary",
	"focus", "posture", "portfolio", "metrics", "compare",
	"select-tests", "pr", "show", "policy", "export",
	"convert", "convert-config", "list", "list-conversions",
	"shorthands", "detect",
	"migration", "migrate", "estimate", "status", "checklist",
	"doctor", "reset",
	"ai", "feedback", "telemetry",
	"debug", "depgraph",
	"version", "serve", "help", "--help", "-h",
	"mcp", "webhook",
	"suppress", "test", "describe", "accept-snapshot",
	"report", "config",
	"plugins", "inject", "scaffold", "fix",
}

func telemetryCommandName(args []string) string {
	if len(args) < 2 {
		return "discover"
	}
	cmd := args[1]
	for _, known := range knownCommands {
		if cmd == known {
			return cmd
		}
	}
	if cmd == "--print-network" {
		return cmd
	}
	return "unknown"
}

// didYouMean returns up to maxResults command names from knownCommands
// closest to candidate by Levenshtein distance, sorted nearest-first.
// Suggestions are emitted only for distance <= 2 — any further away is
// noisy more often than helpful.
// mountPositionalAsRoot mounts the first non-flag positional as the
// `--root` value. This makes `terrain <command> <path>` work alongside
// `terrain <command> --root=<path>` for every analysis-style command.
//
// Earlier revisions had most analysis commands silently ignore
// positionals — a user typing `terrain analyze ./myproj` got cwd
// analysis with no warning. The fix is now uniform across the family:
// analyze, ai run, ai list, ai doctor, debug graph, debug coverage,
// report impact, report insights.
//
// Errors out with exit 2 (usage error) if more than one positional was
// supplied. Callers must pass the FlagSet's args slice (post-Parse).
// argHasFlag reports whether args contains exactly the flag --name or
// -name (with or without an attached =value). Used to detect when a
// user reaches for a flag that this command doesn't support, so we
// can emit a helpful redirect before the stdlib flag parser dumps
// the full flag list.
//
// Matches `--name`, `-name`, `--name=foo`, and `-name=foo`. Does NOT
// match `--namesake` or `-named` — exact match only on the flag name.
func argHasFlag(args []string, name string) bool {
	for _, a := range args {
		if a == "" {
			continue
		}
		// Strip a single leading dash, then optionally another. We
		// intentionally accept both -base and --base because the
		// stdlib flag package treats them as equivalent.
		s := a
		if len(s) > 0 && s[0] == '-' {
			s = s[1:]
		}
		if len(s) > 0 && s[0] == '-' {
			s = s[1:]
		}
		// Trim a value suffix.
		if i := strings.IndexByte(s, '='); i >= 0 {
			s = s[:i]
		}
		if s == name {
			return true
		}
	}
	return false
}

func mountPositionalAsRoot(commandName string, args []string, root *string) {
	if len(args) == 0 {
		return
	}
	if args[0] != "" {
		*root = args[0]
	}
	if len(args) > 1 {
		fmt.Fprintf(os.Stderr, "error: terrain %s takes at most one positional path; got %d (%s)\n",
			commandName, len(args), strings.Join(args, " "))
		os.Exit(2)
	}
}

func didYouMean(candidate string, maxResults int) []string {
	candidate = strings.ToLower(candidate)
	type scored struct {
		name string
		dist int
	}
	var ranked []scored
	for _, cmd := range knownCommands {
		d := levenshtein(candidate, strings.ToLower(cmd))
		if d <= 2 {
			ranked = append(ranked, scored{name: cmd, dist: d})
		}
	}
	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].dist != ranked[j].dist {
			return ranked[i].dist < ranked[j].dist
		}
		return ranked[i].name < ranked[j].name
	})
	if len(ranked) > maxResults {
		ranked = ranked[:maxResults]
	}
	out := make([]string, 0, len(ranked))
	for _, s := range ranked {
		out = append(out, s.name)
	}
	return out
}

// levenshtein returns the Levenshtein edit distance between a and b.
// Standard dynamic-programming implementation; O(len(a) * len(b)) time and
// O(min(len(a), len(b))) space.
func levenshtein(a, b string) int {
	if a == b {
		return 0
	}
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}
	if len(a) < len(b) {
		a, b = b, a
	}
	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			del := prev[j] + 1
			ins := curr[j-1] + 1
			sub := prev[j-1] + cost
			best := del
			if ins < best {
				best = ins
			}
			if sub < best {
				best = sub
			}
			curr[j] = best
		}
		prev, curr = curr, prev
	}
	return prev[len(b)]
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

func isHelpArg(arg string) bool {
	return arg == "--help" || arg == "-h" || arg == "help"
}

func printUsage() {
	fmt.Fprintln(os.Stdout, "Terrain — pre-flight checks for AI/ML systems and the tests around them.")
	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, "Treats unit tests, integration tests, e2e tests, and AI/ML evals as one")
	fmt.Fprintln(os.Stdout, "dependency graph; gates pull requests on AI-specific risks plus")
	fmt.Fprintln(os.Stdout, "test-system regressions. No API key required.")
	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, "Run `terrain` with no arguments for a discovery report.")
	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, "Canonical commands (recommended):")
	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, "  analyze [path] [flags]    What is the state of our test system?")
	fmt.Fprintln(os.Stdout, "  test [flags]              CI-mode wrapper around analyze: emits JUnit XML")
	fmt.Fprintln(os.Stdout, "                            + step-summary markdown alongside the report")
	fmt.Fprintln(os.Stdout, "  fix [path] [--apply]      apply validated remediations (dry-run by default)")
	fmt.Fprintln(os.Stdout, "  suppress <id> [flags]     waive a finding with a reason (writes to")
	fmt.Fprintln(os.Stdout, "                            .terrain/suppressions.yaml)")
	fmt.Fprintln(os.Stdout, "  init [path]               set up Terrain in a repository")
	fmt.Fprintln(os.Stdout, "  report <verb> [flags]     read-side queries: summary, insights, metrics,")
	fmt.Fprintln(os.Stdout, "                            explain, show, impact, pr, posture, select-tests")
	fmt.Fprintln(os.Stdout, "  migrate <verb> [flags]    framework conversion + migration:")
	fmt.Fprintln(os.Stdout, "                            run, config, list, detect, shorthands, estimate,")
	fmt.Fprintln(os.Stdout, "                            status, checklist, readiness, blockers, preview")
	fmt.Fprintln(os.Stdout, "  ai <verb> [flags]         eval scenarios: list, run, doctor, record,")
	fmt.Fprintln(os.Stdout, "                            baseline, replay")
	fmt.Fprintln(os.Stdout, "  accept-snapshot [--yes]   accept the current snapshot as the regression")
	fmt.Fprintln(os.Stdout, "                            baseline")
	fmt.Fprintln(os.Stdout, "  scaffold --schema <path>  generate boundary-case test scaffolds from a")
	fmt.Fprintln(os.Stdout, "                            JSON Schema; never calls the model")
	fmt.Fprintln(os.Stdout, "  inject --prompt <path>    generate prompt-injection test scaffolds from a")
	fmt.Fprintln(os.Stdout, "                            prompt template; never calls the model")
	fmt.Fprintln(os.Stdout, "  plugins <verb>            validate third-party plugin manifests:")
	fmt.Fprintln(os.Stdout, "                            manifest, list")
	fmt.Fprintln(os.Stdout, "  config <verb> [flags]     workspace prefs: feedback, telemetry")
	fmt.Fprintln(os.Stdout, "  doctor [path]             diagnostics for current setup")
	fmt.Fprintln(os.Stdout, "  describe [path] [--write] detect AI surfaces; --write records them in")
	fmt.Fprintln(os.Stdout, "                            terrain.yaml")
	fmt.Fprintln(os.Stdout, "  mcp [--root <dir>]        start the MCP server on stdio for AI assistants")
	fmt.Fprintln(os.Stdout, "  debug <verb> [flags]      dependency graph drill-downs:")
	fmt.Fprintln(os.Stdout, "                            graph, coverage, fanout, duplicates, depgraph")
	fmt.Fprintln(os.Stdout, "  portfolio [flags]         test portfolio intelligence; --from aggregates repos")
	fmt.Fprintln(os.Stdout, "  serve [flags]             local HTTP server with HTML report + JSON API")
	fmt.Fprintln(os.Stdout, "  version                   print version info")
	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, "Typical flow:")
	fmt.Fprintln(os.Stdout, "  1. terrain analyze                 understand your test system")
	fmt.Fprintln(os.Stdout, "  2. terrain report insights         find what to improve")
	fmt.Fprintln(os.Stdout, "  3. terrain report impact --base=main  see what a PR affects")
	fmt.Fprintln(os.Stdout, "  4. terrain report explain <target> understand why")
	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, "Common flags (most commands):")
	fmt.Fprintln(os.Stdout, "  --root PATH                       repository root (default: current dir; positional accepted)")
	fmt.Fprintln(os.Stdout, "  --json                            machine-readable output")
	fmt.Fprintln(os.Stdout, "  --base REF                        git base ref for diff (impact / pr / select-tests)")
	fmt.Fprintln(os.Stdout, "  --baseline PATH                   baseline snapshot for regression detectors")
	fmt.Fprintln(os.Stdout, "  --log-level LEVEL                 diagnostic verbosity: quiet, debug (default: info)")
	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, "Legacy command forms (still work; will be removed in a future release):")
	fmt.Fprintln(os.Stdout, "  Conversion subsystem: convert, convert-config, migrate <dir>, estimate,")
	fmt.Fprintln(os.Stdout, "    status, checklist, reset, list-conversions, shorthands, detect")
	fmt.Fprintln(os.Stdout, "  Report verbs (now under `terrain report <verb>`): summary, insights,")
	fmt.Fprintln(os.Stdout, "    metrics, posture, focus, explain, show, impact, pr, select-tests,")
	fmt.Fprintln(os.Stdout, "    compare, policy, export, migration")
	fmt.Fprintln(os.Stdout, "  Config verbs (now under `terrain config <verb>`): feedback, telemetry")
	fmt.Fprintln(os.Stdout, "  Other: depgraph")
	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, "  Set TERRAIN_LEGACY_HINT=1 to surface canonical-shape suggestions on")
	fmt.Fprintln(os.Stdout, "  legacy invocations (default: silent).")
	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, "Docs: docs/examples/{analyze,summary,insights,explain,focus,impact}-report.md")
	fmt.Fprintln(os.Stdout, "      docs/release/feature-status.md  full per-feature status")
}

func printMigrationUsage() {
	fmt.Fprintln(os.Stderr, "Usage: terrain migration <readiness|blockers|preview> [flags]")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Subcommands:")
	fmt.Fprintln(os.Stderr, "  readiness   assess migration readiness and risk")
	fmt.Fprintln(os.Stderr, "  blockers    list migration blockers and highest-risk areas")
	fmt.Fprintln(os.Stderr, "  preview     preview migration for a file or directory scope")
}

func printExportUsage() {
	fmt.Fprintln(os.Stderr, "Usage: terrain export benchmark [flags]")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Exports a privacy-safe benchmark artifact as JSON.")
}

func printDebugUsage() {
	fmt.Fprintln(os.Stderr, "Usage: terrain debug <graph|coverage|fanout|duplicates|depgraph> [flags]")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Subcommands:")
	fmt.Fprintln(os.Stderr, "  graph       dependency graph statistics")
	fmt.Fprintln(os.Stderr, "  coverage    structural coverage analysis")
	fmt.Fprintln(os.Stderr, "  fanout      high-fanout node analysis")
	fmt.Fprintln(os.Stderr, "  duplicates  duplicate test cluster analysis")
	fmt.Fprintln(os.Stderr, "  depgraph    full dependency graph analysis (supports --show: stats, coverage, duplicates, fanout, impact, profile)")
}

func printAIUsage() {
	fmt.Fprintln(os.Stderr, "Usage: terrain ai <list|findings|run|replay|record|baseline|doctor> [flags]")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Inventory + execution:")
	fmt.Fprintln(os.Stderr, "  list       list detected AI/eval scenarios and surfaces")
	fmt.Fprintln(os.Stderr, "  run        execute eval scenarios and collect results")
	fmt.Fprintln(os.Stderr, "  replay     replay and verify a previous run artifact")
	fmt.Fprintln(os.Stderr, "  record     record eval run results as a baseline snapshot")
	fmt.Fprintln(os.Stderr, "  baseline   show the latest baseline snapshot (default)")
	fmt.Fprintln(os.Stderr, "  baseline compare   diff the latest baseline against the prior one")
	fmt.Fprintln(os.Stderr, "  doctor     validate AI/eval setup and surface configuration issues")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Per-rule AI findings:")
	fmt.Fprintln(os.Stderr, "  findings   emit AI eval-gap findings for the specified rule")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Categorical AI quality checks ship via 'terrain analyze'")
	fmt.Fprintln(os.Stderr, "and 'terrain report pr'. Run 'terrain analyze --help' for the full")
	fmt.Fprintln(os.Stderr, "list, or browse docs/rules/ai/ for per-rule documentation.")
}

func defaultPipelineOptions() engine.PipelineOptions {
	return engine.PipelineOptions{
		EngineVersion:      version,
		MechanismOverrides: mechanismOverrides(),
	}
}

// defaultPipelineOptionsWithProgress returns pipeline options with progress
// reporting enabled for interactive terminals. Pass jsonOutput=true to
// suppress progress (keeps stdout clean for JSON).
func defaultPipelineOptionsWithProgress(jsonOutput bool) engine.PipelineOptions {
	return engine.PipelineOptions{
		EngineVersion:      version,
		MechanismOverrides: mechanismOverrides(),
		OnProgress:         newProgressFunc(jsonOutput),
	}
}

// extractedMechanismOverrides holds parsed --mechanisms.<name>=<state>
// CLI flag values, captured by extractMechanismOverrides at startup.
var extractedMechanismOverrides []string

func mechanismOverrides() []string {
	if len(extractedMechanismOverrides) == 0 {
		return nil
	}
	out := make([]string, len(extractedMechanismOverrides))
	copy(out, extractedMechanismOverrides)
	return out
}

// extractMechanismOverrides scans args for --mechanisms.<name>=<state>
// entries (both --mechanisms.x=on and --mechanisms.x on forms),
// captures each as "name=state", and returns the args with those
// entries removed. Allows subcommand dispatch to remain agnostic of
// the global mechanisms flag.
func extractMechanismOverrides(args []string) (overrides, rest []string) {
	const prefix = "--mechanisms."
	rest = make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		a := args[i]
		if !strings.HasPrefix(a, prefix) {
			rest = append(rest, a)
			continue
		}
		body := a[len(prefix):]
		if eq := strings.IndexByte(body, '='); eq >= 0 {
			overrides = append(overrides, body)
			continue
		}
		// "--mechanisms.x on" — consume the next arg as the state.
		if i+1 < len(args) {
			overrides = append(overrides, body+"="+args[i+1])
			i++
			continue
		}
		// Malformed — leave as-is so the runtime registry validates and
		// surfaces the error to the user.
		overrides = append(overrides, body+"=")
	}
	return overrides, rest
}

func analysisPipelineOptions(coveragePath, coverageRunLabel string, runtimePaths []string, slowThreshold float64) engine.PipelineOptions {
	opt := defaultPipelineOptions()
	opt.CoveragePath = coveragePath
	opt.CoverageRunLabel = strings.TrimSpace(coverageRunLabel)
	opt.RuntimePaths = runtimePaths
	opt.SlowTestThresholdMs = slowThreshold
	return opt
}
