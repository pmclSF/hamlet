package engine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/pmclSF/hamlet/internal/analysis"
	"github.com/pmclSF/hamlet/internal/coverage"
	"github.com/pmclSF/hamlet/internal/health"
	"github.com/pmclSF/hamlet/internal/measurement"
	"github.com/pmclSF/hamlet/internal/models"
	"github.com/pmclSF/hamlet/internal/ownership"
	"github.com/pmclSF/hamlet/internal/policy"
	"github.com/pmclSF/hamlet/internal/portfolio"
	"github.com/pmclSF/hamlet/internal/runtime"
	"github.com/pmclSF/hamlet/internal/scoring"
)

// DefaultEngineVersion is used when PipelineOptions.EngineVersion is not set.
const DefaultEngineVersion = "dev"

// PipelineResult holds the output of a full analysis pipeline run.
type PipelineResult struct {
	Snapshot         *models.TestSuiteSnapshot
	HasPolicy        bool
	DataCompleteness DataCompleteness
	Diagnostics      *PipelineDiagnostics // populated when CollectDiagnostics is true
}

// DataCompleteness captures which analysis inputs were present and usable.
type DataCompleteness struct {
	SourceAvailable   bool
	CoverageProvided  bool
	CoverageAvailable bool
	RuntimeProvided   bool
	RuntimeAvailable  bool
	PolicyAvailable   bool
}

// PipelineOptions configures optional pipeline behavior.
type PipelineOptions struct {
	// CoveragePath is the path to a coverage file or directory.
	// When set, coverage data is ingested and attributed to code units.
	CoveragePath string

	// CoverageRunLabel classifies the coverage artifact source
	// (for example: unit, integration, e2e). When empty, label inference
	// falls back to artifact file names and default heuristics.
	CoverageRunLabel string

	// RuntimePaths are paths to runtime artifact files (JUnit XML, Jest JSON).
	// When set, runtime data is ingested for health signal detection.
	RuntimePaths []string

	// SlowTestThresholdMs overrides the default slow test threshold.
	SlowTestThresholdMs float64

	// CollectDiagnostics enables pipeline timing and count diagnostics.
	CollectDiagnostics bool

	// EngineVersion is stamped into SnapshotMeta.EngineVersion.
	// If empty, DefaultEngineVersion is used.
	EngineVersion string
}

// RunPipeline executes the full analysis pipeline:
//  1. Static analysis (file discovery, framework detection, code units)
//  2. Policy loading
//  3. Runtime ingestion (optional)
//  4. Coverage ingestion (optional)
//  5. Signal detection via the detector registry
//  6. Ownership resolution
//  7. Risk scoring
//  8. Measurement-layer posture
//  9. Portfolio intelligence
//  10. Deterministic sorting
//
// This replaces the duplicated detector invocation across CLI commands.
func RunPipeline(root string, opts ...PipelineOptions) (*PipelineResult, error) {
	return RunPipelineContext(context.Background(), root, opts...)
}

// RunPipelineContext executes the full analysis pipeline with cancellation
// support via context.
func RunPipelineContext(ctx context.Context, root string, opts ...PipelineOptions) (*PipelineResult, error) {
	var opt PipelineOptions
	if len(opts) > 1 {
		return nil, fmt.Errorf("expected at most one PipelineOptions value, got %d", len(opts))
	}
	if len(opts) == 1 {
		opt = opts[0]
	}
	if err := validatePipelineOptions(root, opt); err != nil {
		return nil, err
	}
	engineVersion := strings.TrimSpace(opt.EngineVersion)
	if engineVersion == "" {
		engineVersion = DefaultEngineVersion
	}

	var diag *PipelineDiagnostics
	if opt.CollectDiagnostics {
		diag = &PipelineDiagnostics{}
	}
	pipelineStart := time.Now()
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Group 1: perform independent preparation work concurrently.
	// - static analysis (required for all downstream stages)
	// - policy loading
	// - runtime artifact ingestion (if provided)
	// - coverage artifact ingestion (if provided)
	var (
		snapshot          *models.TestSuiteSnapshot
		policyResult      *policy.LoadResult
		policyErr         error
		ownerResolver     *ownership.Resolver
		runtimeResults    []runtime.TestResult
		runtimeIngestErr  error
		coverageArtifacts []coverage.CoverageArtifact
		coverageIngestErr error

		staticAnalysisDuration  time.Duration
		policyLoadDuration      time.Duration
		ownershipLoadDuration   time.Duration
		runtimeIngestDuration   time.Duration
		coverageIngestDuration  time.Duration
		staticAnalysisTestFiles int
	)

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	fatalErrCh := make(chan error, 1)
	recordFatal := func(err error) {
		if err == nil {
			return
		}
		select {
		case fatalErrCh <- err:
		default:
		}
		cancel()
	}

	startTask := func(wg *sync.WaitGroup, task func(context.Context) error) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := task(runCtx); err != nil {
				recordFatal(err)
			}
		}()
	}

	var prepWG sync.WaitGroup
	startTask(&prepWG, func(taskCtx context.Context) error {
		if err := taskCtx.Err(); err != nil {
			return err
		}
		stepStart := time.Now()
		analyzer := analysis.New(root)
		analyzedSnapshot, err := analyzer.Analyze()
		staticAnalysisDuration = time.Since(stepStart)
		if err != nil {
			return err
		}
		snapshot = analyzedSnapshot
		staticAnalysisTestFiles = len(analyzedSnapshot.TestFiles)
		return nil
	})
	startTask(&prepWG, func(taskCtx context.Context) error {
		if err := taskCtx.Err(); err != nil {
			return err
		}
		stepStart := time.Now()
		policyResult, policyErr = policy.Load(root)
		policyLoadDuration = time.Since(stepStart)
		return nil
	})
	startTask(&prepWG, func(taskCtx context.Context) error {
		if err := taskCtx.Err(); err != nil {
			return err
		}
		stepStart := time.Now()
		ownerResolver = ownership.NewResolver(root)
		ownershipLoadDuration = time.Since(stepStart)
		return nil
	})
	if len(opt.RuntimePaths) > 0 {
		startTask(&prepWG, func(taskCtx context.Context) error {
			if err := taskCtx.Err(); err != nil {
				return err
			}
			stepStart := time.Now()
			runtimeResults, runtimeIngestErr = ingestRuntimeArtifacts(taskCtx, opt.RuntimePaths)
			runtimeIngestDuration = time.Since(stepStart)
			return nil
		})
	}
	if opt.CoveragePath != "" {
		startTask(&prepWG, func(taskCtx context.Context) error {
			if err := taskCtx.Err(); err != nil {
				return err
			}
			stepStart := time.Now()
			coverageArtifacts, coverageIngestErr = ingestCoverageArtifacts(taskCtx, opt.CoveragePath, opt.CoverageRunLabel)
			coverageIngestDuration = time.Since(stepStart)
			return nil
		})
	}
	prepWG.Wait()
	select {
	case err := <-fatalErrCh:
		return nil, err
	default:
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if snapshot == nil {
		return nil, fmt.Errorf("static analysis produced nil snapshot")
	}
	if diag != nil {
		diag.add("static-analysis", staticAnalysisDuration, staticAnalysisTestFiles)
		diag.add("policy-load", policyLoadDuration, 1)
		diag.add("ownership-load", ownershipLoadDuration, 1)
		if len(opt.RuntimePaths) > 0 {
			diag.add("runtime-ingestion", runtimeIngestDuration, len(opt.RuntimePaths))
		}
		if opt.CoveragePath != "" {
			diag.add("coverage-ingestion", coverageIngestDuration, len(coverageArtifacts))
		}
	}

	// Step 2: Load policy config result and attach data-source metadata.
	hasPolicy := policyResult != nil && policyResult.Found
	var stepStart time.Time

	var policyCfg *policy.Config
	if hasPolicy {
		policyCfg = policyResult.Config
		snapshot.Policies = policyConfigMap(policyCfg)
		snapshot.DataSources = append(snapshot.DataSources, models.DataSource{
			Name:   "policy",
			Status: models.DataSourceAvailable,
			Detail: ".hamlet/policy.yaml",
		})
	} else if policyErr != nil {
		snapshot.DataSources = append(snapshot.DataSources, models.DataSource{
			Name:   "policy",
			Status: models.DataSourceError,
			Detail: policyErr.Error(),
			Impact: "Governance checks will not run. Policy violations cannot be detected.",
		})
	} else {
		snapshot.DataSources = append(snapshot.DataSources, models.DataSource{
			Name:   "policy",
			Status: models.DataSourceUnavailable,
			Impact: "No .hamlet/policy.yaml found. Governance checks will not run.",
		})
	}

	// Step 3: Runtime ingestion and health detection (optional).
	if len(opt.RuntimePaths) > 0 {
		stepStart = time.Now()
		if runtimeIngestErr != nil {
			fmt.Fprintf(os.Stderr, "warning: runtime ingestion failed: %v\n", runtimeIngestErr)
			snapshot.DataSources = append(snapshot.DataSources, models.DataSource{
				Name:   "runtime",
				Status: models.DataSourceError,
				Detail: runtimeIngestErr.Error(),
				Impact: "Health measurements (flaky_share, slow_test_share) will report unknown. Risk scoring lacks runtime context.",
			})
		} else {
			applyRuntimeResults(snapshot, runtimeResults, opt.SlowTestThresholdMs)
			snapshot.DataSources = append(snapshot.DataSources, models.DataSource{
				Name:   "runtime",
				Status: models.DataSourceAvailable,
				Detail: fmt.Sprintf("%d artifact(s) ingested", len(opt.RuntimePaths)),
			})
		}
		if diag != nil {
			diag.add("runtime-apply", time.Since(stepStart), len(runtimeResults))
		}
	} else {
		snapshot.DataSources = append(snapshot.DataSources, models.DataSource{
			Name:   "runtime",
			Status: models.DataSourceUnavailable,
			Impact: "Health measurements (flaky_share, slow_test_share) will report unknown. Portfolio cost estimates use type heuristics instead of observed runtime.",
		})
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Step 4: Coverage ingestion (optional).
	// Moved before risk scoring so risk surfaces can account for coverage context.
	if opt.CoveragePath != "" {
		stepStart = time.Now()
		if coverageIngestErr != nil {
			fmt.Fprintf(os.Stderr, "warning: coverage ingestion failed: %v\n", coverageIngestErr)
			snapshot.DataSources = append(snapshot.DataSources, models.DataSource{
				Name:   "coverage",
				Status: models.DataSourceError,
				Detail: fmt.Sprintf("path: %s, error: %v", opt.CoveragePath, coverageIngestErr),
				Impact: "Coverage measurements will report unknown. Portfolio breadth estimates use module heuristics. untestedExport signals rely on import graph only.",
			})
		} else {
			applyCoverageArtifacts(snapshot, coverageArtifacts)
			snapshot.DataSources = append(snapshot.DataSources, models.DataSource{
				Name:   "coverage",
				Status: models.DataSourceAvailable,
				Detail: fmt.Sprintf("path: %s", opt.CoveragePath),
			})
		}
		if diag != nil {
			diag.add("coverage-apply", time.Since(stepStart), len(snapshot.CoverageInsights))
		}
	} else {
		snapshot.DataSources = append(snapshot.DataSources, models.DataSource{
			Name:   "coverage",
			Status: models.DataSourceUnavailable,
			Impact: "Coverage measurements will report unknown. Portfolio breadth estimates use module heuristics. No coverage-based insights will be generated.",
		})
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Step 5: Build detector registry and run all detectors.
	stepStart = time.Now()
	registry := DefaultRegistry(Config{
		RepoRoot:     root,
		PolicyConfig: policyCfg,
	})
	signalsBefore := len(snapshot.Signals)
	registry.Run(snapshot)
	if diag != nil {
		diag.add("signal-detection", time.Since(stepStart), len(snapshot.Signals)-signalsBefore)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Populate snapshot provenance metadata.
	detectorIDs := make([]string, 0, registry.Len())
	for _, reg := range registry.All() {
		detectorIDs = append(detectorIDs, reg.Meta.ID)
	}
	snapshot.SnapshotMeta = models.SnapshotMeta{
		SchemaVersion: models.SnapshotSchemaVersion,
		EngineVersion: engineVersion,
		DetectorCount: registry.Len(),
		Detectors:     detectorIDs,
	}

	// Step 6: Propagate ownership after all signal-producing stages so
	// runtime and coverage-derived signals receive owners consistently.
	stepStart = time.Now()
	if ownerResolver == nil {
		ownerResolver = ownership.NewResolver(root)
	}
	ownership.Propagate(ownerResolver, snapshot)
	if diag != nil {
		diag.add("ownership-resolution", time.Since(stepStart), len(snapshot.Signals))
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	normalizeSignalMetadata(snapshot)

	// Attach file-scoped signals to their corresponding test files.
	attachSignalsToTestFiles(snapshot)

	// Step 7: Compute risk surfaces from signals.
	stepStart = time.Now()
	snapshot.Risk = scoring.ComputeRisk(snapshot)
	if diag != nil {
		diag.add("risk-scoring", time.Since(stepStart), len(snapshot.Risk))
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Step 8: Compute measurement-layer posture.
	stepStart = time.Now()
	measRegistry := measurement.DefaultRegistry()
	measSnap := measRegistry.ComputeSnapshot(snapshot)
	snapshot.Measurements = measSnap.ToModel()
	snapshot.SnapshotMeta.MethodologyFingerprint = methodologyFingerprint(
		snapshot.SnapshotMeta.Detectors,
		measurementDefinitionIDs(measRegistry.All()),
		scoring.RiskModelVersion,
	)
	if diag != nil {
		diag.add("measurement", time.Since(stepStart), len(snapshot.Measurements.Posture))
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Step 9: Compute portfolio intelligence.
	stepStart = time.Now()
	portfolioSummary := portfolio.Analyze(snapshot)
	snapshot.Portfolio = portfolioSummary.ToModel()
	compactRiskContributingSignals(snapshot)
	if diag != nil {
		diag.add("portfolio", time.Since(stepStart), len(portfolioSummary.Findings))
	}

	populateSnapshotMetadata(snapshot, opt, hasPolicy)

	// Step 10: Sort all snapshot slices into canonical order for determinism.
	models.SortSnapshot(snapshot)

	if err := models.ValidateSnapshot(snapshot); err != nil {
		return nil, fmt.Errorf("invalid snapshot produced by pipeline: %w", err)
	}

	if diag != nil {
		diag.Total = time.Since(pipelineStart)
	}

	return &PipelineResult{
		Snapshot:         snapshot,
		HasPolicy:        hasPolicy,
		DataCompleteness: deriveDataCompleteness(snapshot, opt, hasPolicy),
		Diagnostics:      diag,
	}, nil
}

func validatePipelineOptions(root string, opt PipelineOptions) error {
	if root == "" {
		return fmt.Errorf("repository root cannot be empty")
	}
	if opt.SlowTestThresholdMs < 0 {
		return fmt.Errorf("slow test threshold must be >= 0, got %f", opt.SlowTestThresholdMs)
	}
	if label := strings.ToLower(strings.TrimSpace(opt.CoverageRunLabel)); label != "" {
		switch label {
		case "unit", "integration", "e2e":
		default:
			return fmt.Errorf("coverage run label must be one of unit, integration, e2e; got %q", opt.CoverageRunLabel)
		}
	}
	return nil
}

func attachSignalsToTestFiles(snapshot *models.TestSuiteSnapshot) {
	if snapshot == nil || len(snapshot.TestFiles) == 0 || len(snapshot.Signals) == 0 {
		return
	}

	byFile := map[string][]models.Signal{}
	for _, s := range snapshot.Signals {
		if s.Location.File == "" {
			continue
		}
		byFile[s.Location.File] = append(byFile[s.Location.File], s)
	}

	for i := range snapshot.TestFiles {
		path := snapshot.TestFiles[i].Path
		if path == "" {
			continue
		}
		signals := byFile[path]
		if len(signals) == 0 {
			snapshot.TestFiles[i].Signals = []models.Signal{}
			continue
		}
		snapshot.TestFiles[i].Signals = append([]models.Signal(nil), signals...)
	}
}

func populateSnapshotMetadata(snapshot *models.TestSuiteSnapshot, opt PipelineOptions, hasPolicy bool) {
	if snapshot == nil {
		return
	}

	available := 0
	unavailable := 0
	errors := 0
	for _, ds := range snapshot.DataSources {
		switch ds.Status {
		case models.DataSourceAvailable:
			available++
		case models.DataSourceUnavailable:
			unavailable++
		case models.DataSourceError:
			errors++
		}
	}

	testFilesWithLinkedUnits := 0
	testFilesWithSignals := 0
	for _, tf := range snapshot.TestFiles {
		if len(tf.LinkedCodeUnits) > 0 {
			testFilesWithLinkedUnits++
		}
		if len(tf.Signals) > 0 {
			testFilesWithSignals++
		}
	}

	snapshot.Metadata = map[string]any{
		"runtimeArtifactsProvided":  len(opt.RuntimePaths),
		"coverageInputProvided":     opt.CoveragePath != "",
		"coverageRunLabel":          strings.TrimSpace(opt.CoverageRunLabel),
		"policyConfigLoaded":        hasPolicy,
		"dataSourcesAvailable":      available,
		"dataSourcesUnavailable":    unavailable,
		"dataSourcesError":          errors,
		"testFilesWithLinkedUnits":  testFilesWithLinkedUnits,
		"testFilesWithFileSignals":  testFilesWithSignals,
		"totalSignals": len(snapshot.Signals),
	}
}

func normalizeSignalMetadata(snapshot *models.TestSuiteSnapshot) {
	if snapshot == nil || len(snapshot.Signals) == 0 {
		return
	}
	for i := range snapshot.Signals {
		s := &snapshot.Signals[i]
		if s.Metadata == nil {
			s.Metadata = map[string]any{}
		}
		if _, ok := s.Metadata["signalType"]; !ok {
			s.Metadata["signalType"] = string(s.Type)
		}
		if _, ok := s.Metadata["category"]; !ok {
			s.Metadata["category"] = string(s.Category)
		}
		if _, ok := s.Metadata["severity"]; !ok {
			s.Metadata["severity"] = string(s.Severity)
		}
		if s.EvidenceSource != "" {
			if _, ok := s.Metadata["evidenceSource"]; !ok {
				s.Metadata["evidenceSource"] = string(s.EvidenceSource)
			}
		}
		if s.Confidence > 0 {
			if _, ok := s.Metadata["confidence"]; !ok {
				s.Metadata["confidence"] = s.Confidence
			}
		}
		if _, ok := s.Metadata["scope"]; !ok {
			switch {
			case s.Location.File != "":
				s.Metadata["scope"] = "file"
			case s.Location.Package != "":
				s.Metadata["scope"] = "package"
			case s.Location.Repository != "":
				s.Metadata["scope"] = "repository"
			default:
				s.Metadata["scope"] = "unknown"
			}
		}
	}
}

func compactRiskContributingSignals(snapshot *models.TestSuiteSnapshot) {
	if snapshot == nil || len(snapshot.Risk) == 0 {
		return
	}
	const maxSignalsPerRiskSurface = 25
	for i := range snapshot.Risk {
		rs := &snapshot.Risk[i]
		if len(rs.ContributingSignals) == 0 {
			continue
		}
		seen := map[string]bool{}
		compact := make([]models.Signal, 0, len(rs.ContributingSignals))
		for _, s := range rs.ContributingSignals {
			key := string(s.Type) + "|" + s.Location.File + "|" + s.Location.Repository + "|" + s.Explanation
			if seen[key] {
				continue
			}
			seen[key] = true
			compact = append(compact, models.Signal{
				Type:        s.Type,
				Category:    s.Category,
				Severity:    s.Severity,
				Location:    s.Location,
				Owner:       s.Owner,
				Explanation: s.Explanation,
			})
			if len(compact) >= maxSignalsPerRiskSurface {
				break
			}
		}
		rs.ContributingSignals = compact
	}
}

func policyConfigMap(cfg *policy.Config) map[string]any {
	if cfg == nil {
		return nil
	}
	out := map[string]any{}
	rules := map[string]any{}
	if cfg.Rules.DisallowSkippedTests != nil {
		rules["disallow_skipped_tests"] = *cfg.Rules.DisallowSkippedTests
	}
	if len(cfg.Rules.DisallowFrameworks) > 0 {
		frameworks := append([]string(nil), cfg.Rules.DisallowFrameworks...)
		sort.Strings(frameworks)
		rules["disallow_frameworks"] = frameworks
	}
	if cfg.Rules.MaxTestRuntimeMs != nil {
		rules["max_test_runtime_ms"] = *cfg.Rules.MaxTestRuntimeMs
	}
	if cfg.Rules.MinimumCoveragePercent != nil {
		rules["minimum_coverage_percent"] = *cfg.Rules.MinimumCoveragePercent
	}
	if cfg.Rules.MaxWeakAssertions != nil {
		rules["max_weak_assertions"] = *cfg.Rules.MaxWeakAssertions
	}
	if cfg.Rules.MaxMockHeavyTests != nil {
		rules["max_mock_heavy_tests"] = *cfg.Rules.MaxMockHeavyTests
	}
	if len(rules) == 0 {
		return nil
	}
	out["rules"] = rules
	return out
}

func measurementDefinitionIDs(defs []measurement.Definition) []string {
	if len(defs) == 0 {
		return nil
	}
	ids := make([]string, 0, len(defs))
	for _, d := range defs {
		if d.ID != "" {
			ids = append(ids, d.ID)
		}
	}
	sort.Strings(ids)
	return ids
}

func methodologyFingerprint(detectors, measurementIDs []string, riskModelVersion string) string {
	parts := []string{
		"schema=" + models.SnapshotSchemaVersion,
		"riskModel=" + riskModelVersion,
		"detectors=" + strings.Join(detectors, ","),
		"measurements=" + strings.Join(measurementIDs, ","),
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(sum[:])
}

func deriveDataCompleteness(snapshot *models.TestSuiteSnapshot, opt PipelineOptions, hasPolicy bool) DataCompleteness {
	dc := DataCompleteness{
		SourceAvailable:   snapshot != nil,
		CoverageProvided:  strings.TrimSpace(opt.CoveragePath) != "",
		RuntimeProvided:   len(opt.RuntimePaths) > 0,
		PolicyAvailable:   hasPolicy,
		CoverageAvailable: false,
		RuntimeAvailable:  false,
	}
	if snapshot == nil {
		return dc
	}
	if len(snapshot.TestFiles) == 0 && len(snapshot.CodeUnits) == 0 {
		dc.SourceAvailable = false
	}
	for _, ds := range snapshot.DataSources {
		switch ds.Name {
		case "coverage":
			dc.CoverageAvailable = ds.Status == models.DataSourceAvailable
		case "runtime":
			dc.RuntimeAvailable = ds.Status == models.DataSourceAvailable
		case "policy":
			dc.PolicyAvailable = ds.Status == models.DataSourceAvailable
		}
	}
	return dc
}

// ingestRuntime parses runtime artifact files and runs health detectors.
func ingestRuntime(snapshot *models.TestSuiteSnapshot, paths []string, slowThreshold float64) error {
	allResults, err := ingestRuntimeArtifacts(context.Background(), paths)
	if err != nil {
		return err
	}
	applyRuntimeResults(snapshot, allResults, slowThreshold)
	return nil
}

func ingestRuntimeArtifacts(ctx context.Context, paths []string) ([]runtime.TestResult, error) {
	var allResults []runtime.TestResult
	for _, p := range paths {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		result, err := runtime.Ingest(p)
		if err != nil {
			return nil, err
		}
		allResults = append(allResults, result.Results...)
	}
	return allResults, nil
}

func applyRuntimeResults(snapshot *models.TestSuiteSnapshot, allResults []runtime.TestResult, slowThreshold float64) {
	if snapshot == nil || len(allResults) == 0 {
		return
	}

	// Resolve stable test IDs by joining runtime results to extracted test cases.
	runtime.ResolveTestIDs(allResults, snapshot.TestCases)

	// Apply runtime stats to matching test files.
	updates := make([]runtime.TestFileUpdate, len(snapshot.TestFiles))
	for i, tf := range snapshot.TestFiles {
		updates[i] = runtime.TestFileUpdate{Path: tf.Path}
	}
	runtime.ApplyToTestFiles(allResults, updates)
	for i, u := range updates {
		if u.AvgRuntimeMs > 0 {
			snapshot.TestFiles[i].RuntimeStats = &models.RuntimeStats{
				AvgRuntimeMs:    u.AvgRuntimeMs,
				P95RuntimeMs:    u.P95RuntimeMs,
				PassRate:        u.PassRate,
				RetryRate:       u.RetryRate,
				RuntimeVariance: u.RuntimeVariance,
			}
		}
	}

	// Run health detectors on runtime data.
	slowDetector := &health.SlowTestDetector{ThresholdMs: slowThreshold}
	flakyDetector := &health.FlakyTestDetector{}
	skippedDetector := &health.SkippedTestDetector{}
	deadDetector := &health.DeadTestDetector{}
	unstableDetector := &health.UnstableSuiteDetector{}

	snapshot.Signals = append(snapshot.Signals, slowDetector.Detect(allResults)...)
	snapshot.Signals = append(snapshot.Signals, flakyDetector.Detect(allResults)...)
	snapshot.Signals = append(snapshot.Signals, skippedDetector.Detect(allResults)...)
	snapshot.Signals = append(snapshot.Signals, deadDetector.Detect(allResults)...)
	snapshot.Signals = append(snapshot.Signals, unstableDetector.Detect(allResults)...)
}

// ingestCoverage loads coverage data and populates the snapshot's coverage fields.
func ingestCoverage(snapshot *models.TestSuiteSnapshot, coveragePath string) error {
	artifacts, err := ingestCoverageArtifacts(context.Background(), coveragePath, "")
	if err != nil {
		return err
	}
	applyCoverageArtifacts(snapshot, artifacts)
	return nil
}

func ingestCoverageArtifacts(ctx context.Context, coveragePath string, runLabel string) ([]coverage.CoverageArtifact, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	info, err := os.Stat(coveragePath)
	if err != nil {
		return nil, fmt.Errorf("coverage path %s: %w", coveragePath, err)
	}

	var artifacts []coverage.CoverageArtifact
	if info.IsDir() {
		artifacts, err = coverage.IngestDirectory(coveragePath, runLabel)
		if err != nil {
			var warn *coverage.IngestWarning
			if errors.As(err, &warn) {
				fmt.Fprintf(os.Stderr, "warning: %v\n", warn)
				err = nil
			}
		}
	} else {
		var art *coverage.CoverageArtifact
		art, err = coverage.IngestFile(coveragePath, runLabel)
		if art != nil {
			artifacts = append(artifacts, *art)
		}
	}
	if err != nil {
		return nil, err
	}
	return artifacts, nil
}

func applyCoverageArtifacts(snapshot *models.TestSuiteSnapshot, artifacts []coverage.CoverageArtifact) {
	if snapshot == nil {
		return
	}
	if len(artifacts) == 0 {
		return
	}

	// Merge all artifacts.
	merged := coverage.Merge(artifacts)

	// Attribute to code units.
	unitCov := coverage.AttributeToCodeUnits(merged, snapshot.CodeUnits)

	// Compute type-based coverage.
	typeCov := coverage.ComputeByType(artifacts, snapshot.CodeUnits)

	// Build repo summary.
	repoSummary := coverage.BuildRepoSummary(typeCov, snapshot.CodeUnits)

	// Populate snapshot coverage summary.
	snapshot.CoverageSummary = &models.CoverageSummary{
		TotalCodeUnits:       repoSummary.TotalCodeUnits,
		CoveredByUnitTests:   repoSummary.CoveredByUnitTests,
		CoveredByIntegration: repoSummary.CoveredByIntegration,
		CoveredByE2E:         repoSummary.CoveredByE2E,
		CoveredOnlyByE2E:     repoSummary.CoveredOnlyByE2E,
		UncoveredExported:    repoSummary.UncoveredExported,
		Uncovered:            repoSummary.TotalCodeUnits - repoSummary.CoveredByUnitTests,
	}

	// Compute overall line/branch coverage from merged data.
	var totalLines, coveredLines, totalBranches, coveredBranches int
	for _, rec := range merged.ByFile {
		totalLines += rec.LineTotalCount
		coveredLines += rec.LineCoveredCount
		totalBranches += rec.BranchTotalCount
		coveredBranches += rec.BranchCoveredCount
	}
	if totalLines > 0 {
		snapshot.CoverageSummary.LineCoveragePct = float64(coveredLines) / float64(totalLines) * 100
	}
	if totalBranches > 0 {
		snapshot.CoverageSummary.BranchCoveragePct = float64(coveredBranches) / float64(totalBranches) * 100
	}

	// Derive insights.
	covInsights := coverage.DeriveInsights(typeCov, snapshot.CodeUnits)
	unitInsights := coverage.DeriveUnitInsights(unitCov)
	for _, ci := range append(covInsights, unitInsights...) {
		snapshot.CoverageInsights = append(snapshot.CoverageInsights, models.CoverageInsight{
			Type:            ci.Type,
			Severity:        ci.Severity,
			Description:     ci.Description,
			Path:            ci.Path,
			UnitID:          ci.UnitID,
			SuggestedAction: ci.SuggestedAction,
		})
	}
}
