package engine

import (
	"fmt"
	"os"
	"time"

	"github.com/pmclSF/hamlet/internal/analysis"
	"github.com/pmclSF/hamlet/internal/coverage"
	"github.com/pmclSF/hamlet/internal/health"
	"github.com/pmclSF/hamlet/internal/measurement"
	"github.com/pmclSF/hamlet/internal/models"
	"github.com/pmclSF/hamlet/internal/ownership"
	"github.com/pmclSF/hamlet/internal/policy"
	"github.com/pmclSF/hamlet/internal/runtime"
	"github.com/pmclSF/hamlet/internal/scoring"
)

// EngineVersion is set by the CLI at startup to embed build version
// into every snapshot's SnapshotMeta. Defaults to "dev".
var EngineVersion = "dev"

// PipelineResult holds the output of a full analysis pipeline run.
type PipelineResult struct {
	Snapshot    *models.TestSuiteSnapshot
	HasPolicy   bool
	Diagnostics *PipelineDiagnostics // populated when CollectDiagnostics is true
}

// PipelineOptions configures optional pipeline behavior.
type PipelineOptions struct {
	// CoveragePath is the path to a coverage file or directory.
	// When set, coverage data is ingested and attributed to code units.
	CoveragePath string

	// RuntimePaths are paths to runtime artifact files (JUnit XML, Jest JSON).
	// When set, runtime data is ingested for health signal detection.
	RuntimePaths []string

	// SlowTestThresholdMs overrides the default slow test threshold.
	SlowTestThresholdMs float64

	// CollectDiagnostics enables pipeline timing and count diagnostics.
	CollectDiagnostics bool
}

// RunPipeline executes the full analysis pipeline:
//  1. Static analysis (file discovery, framework detection, code units)
//  2. Signal detection via the detector registry
//  3. Ownership resolution
//  4. Risk scoring
//
// This replaces the duplicated detector invocation across CLI commands.
func RunPipeline(root string, opts ...PipelineOptions) (*PipelineResult, error) {
	var opt PipelineOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	var diag *PipelineDiagnostics
	if opt.CollectDiagnostics {
		diag = &PipelineDiagnostics{}
	}
	pipelineStart := time.Now()

	// Step 1: Static analysis.
	stepStart := time.Now()
	analyzer := analysis.New(root)
	snapshot, err := analyzer.Analyze()
	if err != nil {
		return nil, err
	}
	if diag != nil {
		diag.add("static-analysis", time.Since(stepStart), len(snapshot.TestFiles))
	}

	// Step 2: Load policy config (needed to configure governance detector).
	policyResult, _ := policy.Load(root)
	hasPolicy := policyResult != nil && policyResult.Found

	var policyCfg *policy.Config
	if hasPolicy {
		policyCfg = policyResult.Config
	}

	// Step 3: Build detector registry and run all detectors.
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

	// Populate snapshot provenance metadata.
	detectorIDs := make([]string, 0, registry.Len())
	for _, reg := range registry.All() {
		detectorIDs = append(detectorIDs, reg.Meta.ID)
	}
	snapshot.SnapshotMeta = models.SnapshotMeta{
		SchemaVersion: models.SnapshotSchemaVersion,
		EngineVersion: EngineVersion,
		DetectorCount: registry.Len(),
		Detectors:     detectorIDs,
	}

	// Step 4: Propagate ownership to signals.
	stepStart = time.Now()
	resolver := ownership.NewResolver(root)
	for i := range snapshot.Signals {
		if snapshot.Signals[i].Owner == "" && snapshot.Signals[i].Location.File != "" {
			snapshot.Signals[i].Owner = resolver.Resolve(snapshot.Signals[i].Location.File)
		}
	}
	if diag != nil {
		diag.add("ownership-resolution", time.Since(stepStart), len(snapshot.Signals))
	}

	// Step 5: Runtime ingestion and health detection (optional).
	if len(opt.RuntimePaths) > 0 {
		stepStart = time.Now()
		if err := ingestRuntime(snapshot, opt.RuntimePaths, opt.SlowTestThresholdMs); err != nil {
			fmt.Fprintf(os.Stderr, "warning: runtime ingestion failed: %v\n", err)
		}
		if diag != nil {
			diag.add("runtime-ingestion", time.Since(stepStart), len(opt.RuntimePaths))
		}
	}

	// Step 6: Compute risk surfaces from signals (including runtime-backed ones).
	stepStart = time.Now()
	snapshot.Risk = scoring.ComputeRisk(snapshot)
	if diag != nil {
		diag.add("risk-scoring", time.Since(stepStart), len(snapshot.Risk))
	}

	// Step 7: Coverage ingestion (optional).
	if opt.CoveragePath != "" {
		stepStart = time.Now()
		if err := ingestCoverage(snapshot, opt.CoveragePath); err != nil {
			fmt.Fprintf(os.Stderr, "warning: coverage ingestion failed: %v\n", err)
		}
		if diag != nil {
			diag.add("coverage-ingestion", time.Since(stepStart), len(snapshot.CoverageInsights))
		}
	}

	// Step 8: Compute measurement-layer posture.
	stepStart = time.Now()
	measRegistry := measurement.DefaultRegistry()
	measSnap := measRegistry.ComputeSnapshot(snapshot)
	snapshot.Measurements = measSnap.ToModel()
	if diag != nil {
		diag.add("measurement", time.Since(stepStart), len(snapshot.Measurements.Posture))
	}

	// Step 9: Sort all snapshot slices into canonical order for determinism.
	models.SortSnapshot(snapshot)

	if diag != nil {
		diag.Total = time.Since(pipelineStart)
	}

	return &PipelineResult{
		Snapshot:    snapshot,
		HasPolicy:   hasPolicy,
		Diagnostics: diag,
	}, nil
}

// ingestRuntime parses runtime artifact files and runs health detectors.
func ingestRuntime(snapshot *models.TestSuiteSnapshot, paths []string, slowThreshold float64) error {
	var allResults []runtime.TestResult
	for _, p := range paths {
		result, err := runtime.Ingest(p)
		if err != nil {
			return err
		}
		allResults = append(allResults, result.Results...)
	}

	if len(allResults) == 0 {
		return nil
	}

	// Apply runtime stats to matching test files.
	updates := make([]runtime.TestFileUpdate, len(snapshot.TestFiles))
	for i, tf := range snapshot.TestFiles {
		updates[i] = runtime.TestFileUpdate{Path: tf.Path}
	}
	runtime.ApplyToTestFiles(allResults, updates)
	for i, u := range updates {
		if u.AvgRuntimeMs > 0 {
			snapshot.TestFiles[i].RuntimeStats = &models.RuntimeStats{
				AvgRuntimeMs: u.AvgRuntimeMs,
				P95RuntimeMs: u.P95RuntimeMs,
				PassRate:     u.PassRate,
				RetryRate:    u.RetryRate,
			}
		}
	}

	// Run health detectors on runtime data.
	slowDetector := &health.SlowTestDetector{ThresholdMs: slowThreshold}
	flakyDetector := &health.FlakyTestDetector{}
	skippedDetector := &health.SkippedTestDetector{}

	snapshot.Signals = append(snapshot.Signals, slowDetector.Detect(allResults)...)
	snapshot.Signals = append(snapshot.Signals, flakyDetector.Detect(allResults)...)
	snapshot.Signals = append(snapshot.Signals, skippedDetector.Detect(allResults)...)

	return nil
}

// ingestCoverage loads coverage data and populates the snapshot's coverage fields.
func ingestCoverage(snapshot *models.TestSuiteSnapshot, coveragePath string) error {
	info, err := os.Stat(coveragePath)
	if err != nil {
		return fmt.Errorf("coverage path %s: %w", coveragePath, err)
	}

	var artifacts []coverage.CoverageArtifact
	if info.IsDir() {
		artifacts, err = coverage.IngestDirectory(coveragePath, "")
	} else {
		var art *coverage.CoverageArtifact
		art, err = coverage.IngestFile(coveragePath, "")
		if art != nil {
			artifacts = append(artifacts, *art)
		}
	}
	if err != nil {
		return err
	}
	if len(artifacts) == 0 {
		return nil
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
		TotalCodeUnits:     repoSummary.TotalCodeUnits,
		CoveredByUnitTests: repoSummary.CoveredByUnitTests,
		CoveredByE2E:       repoSummary.CoveredByE2E,
		CoveredOnlyByE2E:   repoSummary.CoveredOnlyByE2E,
		UncoveredExported:  repoSummary.UncoveredExported,
		Uncovered:          repoSummary.TotalCodeUnits - repoSummary.CoveredByUnitTests,
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

	return nil
}
