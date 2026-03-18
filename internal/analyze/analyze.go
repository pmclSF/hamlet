// Package analyze aggregates Terrain's engines into a single "first-run"
// report for `terrain analyze`. It combines the pipeline snapshot with
// depgraph analysis (coverage, duplicates, fanout, profile) to produce
// a structured report suitable for both human rendering and JSON output.
package analyze

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pmclSF/terrain/internal/depgraph"
	"github.com/pmclSF/terrain/internal/matrix"
	"github.com/pmclSF/terrain/internal/models"
	"github.com/pmclSF/terrain/internal/stability"
)

// Report is the structured output of `terrain analyze`.
// Every field is concrete and JSON-serializable.
// AnalyzeReportSchemaVersion is the current schema version for analyze reports.
const AnalyzeReportSchemaVersion = "1"

type Report struct {
	// SchemaVersion identifies the report JSON schema version.
	SchemaVersion string `json:"schemaVersion"`

	// Repository metadata.
	Repository RepositoryInfo `json:"repository"`

	// DataCompleteness shows which data sources are available.
	DataCompleteness []DataSource `json:"dataCompleteness"`

	// TestsDetected summarizes test/validation targets.
	TestsDetected TestSummary `json:"testsDetected"`

	// RepoProfile classifies the repository along key dimensions.
	RepoProfile ProfileSummary `json:"repoProfile"`

	// CoverageConfidence summarizes structural coverage bands.
	CoverageConfidence CoverageSummary `json:"coverageConfidence"`

	// DuplicateClusters summarizes redundancy findings.
	DuplicateClusters DuplicateSummary `json:"duplicateClusters"`

	// HighFanout summarizes high-fanout fixtures/helpers.
	HighFanout FanoutSummary `json:"highFanout"`

	// SkippedTestBurden summarizes skipped test load.
	SkippedTestBurden SkipSummary `json:"skippedTestBurden"`

	// WeakCoverageAreas lists source areas with poor structural coverage.
	WeakCoverageAreas []WeakArea `json:"weakCoverageAreas,omitempty"`

	// CIOptimization estimates potential CI improvements.
	CIOptimization CIOptimizationSummary `json:"ciOptimization"`

	// TopInsight is the single biggest opportunity or risk.
	// Retained for backward compatibility; superseded by KeyFindings.
	TopInsight string `json:"topInsight"`

	// KeyFindings are the top 3 prioritized findings surfaced inline in
	// the analyze output so users see the most important issues immediately.
	// These are derived from the same logic as `terrain insights`.
	KeyFindings []KeyFinding `json:"keyFindings,omitempty"`

	// TotalFindingCount is the total number of findings (including those
	// not shown in KeyFindings). Enables "N more via terrain insights".
	TotalFindingCount int `json:"totalFindingCount,omitempty"`

	// RiskPosture summarizes risk by dimension.
	RiskPosture []RiskDimension `json:"riskPosture,omitempty"`

	// SignalSummary breaks down detected signals.
	SignalSummary SignalBreakdown `json:"signalSummary"`

	// BehaviorRedundancy detects behavior-aware test redundancy.
	BehaviorRedundancy *depgraph.RedundancyResult `json:"behaviorRedundancy,omitempty"`

	// StabilityClusters groups unstable tests by shared root cause.
	StabilityClusters *stability.ClusterResult `json:"stabilityClusters,omitempty"`

	// MatrixCoverage holds device/environment matrix analysis results.
	MatrixCoverage *matrix.MatrixResult `json:"matrixCoverage,omitempty"`

	// ManualCoverage summarizes manual validation overlays.
	ManualCoverage *ManualCoverageSummary `json:"manualCoverage,omitempty"`

	// EdgeCases are structural anomalies that affect analysis confidence.
	EdgeCases []depgraph.EdgeCase `json:"edgeCases,omitempty"`

	// Policy captures the derived policy from edge case analysis.
	Policy *depgraph.Policy `json:"policy,omitempty"`

	// Limitations notes where analysis is incomplete.
	Limitations []string `json:"limitations,omitempty"`
}

// KeyFinding is a prioritized finding surfaced in the analyze output.
// This is a lightweight summary — the full finding with description,
// scope, and metric lives in the insights package.
type KeyFinding struct {
	// Title is a short description of the issue.
	Title string `json:"title"`

	// Severity is critical, high, medium, or low.
	Severity string `json:"severity"`

	// Category is optimization, reliability, architecture_debt, or coverage_debt.
	Category string `json:"category"`

	// Metric is the key number (e.g., "340 duplicates", "12 flaky tests").
	Metric string `json:"metric,omitempty"`
}

// RepositoryInfo captures repo metadata.
type RepositoryInfo struct {
	Name            string   `json:"name"`
	Branch          string   `json:"branch,omitempty"`
	CommitSHA       string   `json:"commitSha,omitempty"`
	Languages       []string `json:"languages,omitempty"`
	PackageManagers []string `json:"packageManagers,omitempty"`
	CISystems       []string `json:"ciSystems,omitempty"`
}

// DataSource describes a data source's availability.
type DataSource struct {
	Name      string `json:"name"`
	Available bool   `json:"available"`
}

// TestSummary counts validation assets — tests, scenarios, and code surfaces.
type TestSummary struct {
	TestFileCount    int              `json:"testFileCount"`
	TestCaseCount    int              `json:"testCaseCount"`
	CodeUnitCount    int              `json:"codeUnitCount"`
	ScenarioCount    int              `json:"scenarioCount,omitempty"`
	CodeSurfaceCount int              `json:"codeSurfaceCount,omitempty"`
	PromptCount      int              `json:"promptCount,omitempty"`
	DatasetCount     int              `json:"datasetCount,omitempty"`
	Frameworks       []FrameworkCount `json:"frameworks"`
}

// FrameworkCount is a framework with its file count and type.
type FrameworkCount struct {
	Name      string `json:"name"`
	FileCount int    `json:"fileCount"`
	Type      string `json:"type,omitempty"`
}

// ProfileSummary classifies the repository.
type ProfileSummary struct {
	TestVolume         string  `json:"testVolume"`
	CIPressure         string  `json:"ciPressure"`
	CoverageConfidence string  `json:"coverageConfidence"`
	RedundancyLevel    string  `json:"redundancyLevel"`
	FanoutBurden       string  `json:"fanoutBurden"`
	SkipBurden             string  `json:"skipBurden,omitempty"`
	FlakeBurden            string  `json:"flakeBurden,omitempty"`
	ManualCoveragePresence string  `json:"manualCoveragePresence,omitempty"`
	GraphDensity           float64 `json:"graphDensity"`
}

// CoverageSummary breaks down structural coverage by band.
type CoverageSummary struct {
	HighCount   int `json:"highCount"`
	MediumCount int `json:"mediumCount"`
	LowCount    int `json:"lowCount"`
	TotalFiles  int `json:"totalFiles"`
}

// DuplicateSummary summarizes duplicate clusters.
type DuplicateSummary struct {
	ClusterCount       int     `json:"clusterCount"`
	RedundantTestCount int     `json:"redundantTestCount"`
	HighestSimilarity  float64 `json:"highestSimilarity,omitempty"`
}

// FanoutSummary summarizes high-fanout nodes.
type FanoutSummary struct {
	FlaggedCount int          `json:"flaggedCount"`
	Threshold    int          `json:"threshold"`
	TopNodes     []FanoutNode `json:"topNodes,omitempty"`
}

// FanoutNode is a high-fanout node.
type FanoutNode struct {
	Path             string `json:"path"`
	NodeType         string `json:"nodeType"`
	TransitiveFanout int    `json:"transitiveFanout"`
}

// SkipSummary summarizes skipped tests.
type SkipSummary struct {
	SkippedCount int     `json:"skippedCount"`
	TotalTests   int     `json:"totalTests"`
	SkipRatio    float64 `json:"skipRatio"`
}

// WeakArea is a source file or directory with poor coverage.
type WeakArea struct {
	Path      string `json:"path"`
	TestCount int    `json:"testCount"`
	Band      string `json:"band"`
}

// CIOptimizationSummary estimates potential CI improvements.
type CIOptimizationSummary struct {
	DuplicateTestsRemovable int    `json:"duplicateTestsRemovable"`
	SkippedTestsReviewable  int    `json:"skippedTestsReviewable"`
	HighFanoutNodes         int    `json:"highFanoutNodes"`
	Recommendation          string `json:"recommendation"`
}

// RiskDimension is one risk assessment dimension.
type RiskDimension struct {
	Dimension string `json:"dimension"`
	Band      string `json:"band"`
}

// SignalBreakdown counts signals by severity and category.
type SignalBreakdown struct {
	Total    int            `json:"total"`
	Critical int            `json:"critical"`
	High     int            `json:"high"`
	Medium   int            `json:"medium"`
	Low      int            `json:"low"`
	ByCategory map[string]int `json:"byCategory"`
}

// ManualCoverageSummary summarizes manual validation overlays.
// Manual coverage supplements automated CI coverage but is never treated
// as executable validation.
type ManualCoverageSummary struct {
	// ArtifactCount is the total number of manual coverage artifacts.
	ArtifactCount int `json:"artifactCount"`

	// BySource breaks down artifacts by origin system.
	BySource map[string]int `json:"bySource"`

	// ByCriticality breaks down artifacts by criticality level.
	ByCriticality map[string]int `json:"byCriticality"`

	// Areas lists the code areas covered by manual validation.
	Areas []string `json:"areas"`

	// StaleCount is the number of artifacts with no recent execution date.
	StaleCount int `json:"staleCount"`
}

// BuildInput contains everything needed to build an analyze report.
type BuildInput struct {
	Snapshot  *models.TestSuiteSnapshot
	HasPolicy bool
}

// Build constructs an AnalyzeReport from a pipeline snapshot.
// It runs depgraph analysis internally to produce coverage, redundancy,
// and fanout findings.
func Build(input *BuildInput) *Report {
	snap := input.Snapshot

	r := &Report{SchemaVersion: AnalyzeReportSchemaVersion}

	// Repository info.
	r.Repository = buildRepositoryInfo(snap)

	// Data completeness.
	r.DataCompleteness = buildDataCompleteness(snap, input.HasPolicy)

	// Tests detected.
	r.TestsDetected = buildTestSummary(snap)

	// Run depgraph analysis for profile, coverage, duplicates, fanout.
	dg := depgraph.Build(snap)
	dgCov := depgraph.AnalyzeCoverage(dg)
	dgDupes := depgraph.DetectDuplicates(dg)
	dgFanout := depgraph.AnalyzeFanout(dg, depgraph.DefaultFanoutThreshold)
	profileInsights := depgraph.ProfileInsights{
		Coverage:   &dgCov,
		Duplicates: &dgDupes,
		Fanout:     &dgFanout,
		Snapshot:   BuildSnapshotProfileData(snap),
	}
	dgProfile := depgraph.AnalyzeProfile(dg, profileInsights)

	// Edge cases and policy.
	dgEdgeCases := depgraph.DetectEdgeCases(dgProfile, dg, profileInsights)
	dgPolicy := depgraph.ApplyEdgeCasePolicy(dgEdgeCases, dgProfile)
	if len(dgEdgeCases) > 0 {
		r.EdgeCases = dgEdgeCases
		r.Policy = &dgPolicy
	}

	// Repo profile.
	r.RepoProfile = ProfileSummary{
		TestVolume:             dgProfile.TestVolume,
		CIPressure:             dgProfile.CIPressure,
		CoverageConfidence:     dgProfile.CoverageConfidence,
		RedundancyLevel:        dgProfile.RedundancyLevel,
		FanoutBurden:           dgProfile.FanoutBurden,
		SkipBurden:             dgProfile.SkipBurden,
		FlakeBurden:            dgProfile.FlakeBurden,
		ManualCoveragePresence: dgProfile.ManualCoveragePresence,
		GraphDensity:           dgProfile.GraphDensity,
	}

	// Coverage confidence.
	r.CoverageConfidence = buildCoverageSummary(&dgCov)

	// Duplicate clusters.
	r.DuplicateClusters = buildDuplicateSummary(&dgDupes)

	// High-fanout nodes.
	r.HighFanout = buildFanoutSummary(&dgFanout)

	// Skipped test burden.
	r.SkippedTestBurden = buildSkipSummary(snap)

	// Weak coverage areas.
	r.WeakCoverageAreas = buildWeakAreas(&dgCov)

	// CI optimization.
	r.CIOptimization = buildCIOptimization(&dgDupes, &dgFanout, snap)

	// Behavior redundancy.
	dgRedundancy := depgraph.AnalyzeRedundancy(dg)
	if len(dgRedundancy.Clusters) > 0 {
		r.BehaviorRedundancy = &dgRedundancy
	}

	// Stability clusters.
	clusters := stability.DetectClusters(dg, snap.Signals)
	if len(clusters.Clusters) > 0 {
		r.StabilityClusters = clusters
	}

	// Matrix coverage.
	matrixResult := matrix.Analyze(dg)
	if len(matrixResult.Classes) > 0 {
		r.MatrixCoverage = matrixResult
	}

	// Manual coverage overlay.
	if len(snap.ManualCoverage) > 0 {
		r.ManualCoverage = buildManualCoverageSummary(snap)
	}

	// Risk posture.
	r.RiskPosture = buildRiskPosture(snap)

	// Signal summary.
	r.SignalSummary = buildSignalSummary(snap)

	// Top insight (backward compat).
	r.TopInsight = deriveTopInsight(r, &dgFanout, &dgDupes, &dgCov)

	// Key findings — top 3 prioritized issues from the same data.
	r.KeyFindings, r.TotalFindingCount = deriveKeyFindings(r, &dgFanout, &dgDupes, &dgCov, snap)

	// Limitations.
	r.Limitations = buildLimitations(snap, input.HasPolicy)

	// Enrich CI optimization with depgraph stats.
	dgStats := dg.Stats()
	if dgStats.NodeCount > 0 && r.CIOptimization.Recommendation == "" {
		r.CIOptimization.Recommendation = fmt.Sprintf(
			"Graph has %d nodes and %d edges — confidence-based test selection available via `terrain impact`.",
			dgStats.NodeCount, dgStats.EdgeCount)
	}

	return r
}

func buildRepositoryInfo(snap *models.TestSuiteSnapshot) RepositoryInfo {
	ri := RepositoryInfo{
		Name:            snap.Repository.Name,
		Branch:          snap.Repository.Branch,
		Languages:       snap.Repository.Languages,
		PackageManagers: snap.Repository.PackageManagers,
		CISystems:       snap.Repository.CISystems,
	}
	if sha := snap.Repository.CommitSHA; len(sha) > 10 {
		ri.CommitSHA = sha[:10]
	} else {
		ri.CommitSHA = sha
	}
	return ri
}

func buildDataCompleteness(snap *models.TestSuiteSnapshot, hasPolicy bool) []DataSource {
	sourceAvailable := len(snap.TestFiles) > 0 || len(snap.CodeUnits) > 0
	sources := []DataSource{
		{Name: "source", Available: sourceAvailable},
		{Name: "coverage", Available: dsAvailable(snap, "coverage")},
		{Name: "runtime", Available: dsAvailable(snap, "runtime")},
		{Name: "policy", Available: hasPolicy || dsAvailable(snap, "policy")},
	}
	return sources
}

func dsAvailable(snap *models.TestSuiteSnapshot, name string) bool {
	for _, ds := range snap.DataSources {
		if ds.Name == name {
			return ds.Status == models.DataSourceAvailable
		}
	}
	return false
}

func buildTestSummary(snap *models.TestSuiteSnapshot) TestSummary {
	ts := TestSummary{
		TestFileCount:    len(snap.TestFiles),
		TestCaseCount:    len(snap.TestCases),
		CodeUnitCount:    len(snap.CodeUnits),
		ScenarioCount:    len(snap.Scenarios),
		CodeSurfaceCount: len(snap.CodeSurfaces),
	}
	for _, cs := range snap.CodeSurfaces {
		switch cs.Kind {
		case models.SurfacePrompt:
			ts.PromptCount++
		case models.SurfaceDataset:
			ts.DatasetCount++
		}
	}
	for _, fw := range snap.Frameworks {
		ts.Frameworks = append(ts.Frameworks, FrameworkCount{
			Name:      fw.Name,
			FileCount: fw.FileCount,
			Type:      string(fw.Type),
		})
	}
	return ts
}

func buildManualCoverageSummary(snap *models.TestSuiteSnapshot) *ManualCoverageSummary {
	if len(snap.ManualCoverage) == 0 {
		return nil
	}

	summary := &ManualCoverageSummary{
		ArtifactCount: len(snap.ManualCoverage),
		BySource:      map[string]int{},
		ByCriticality: map[string]int{},
	}

	seen := map[string]bool{}
	for _, mc := range snap.ManualCoverage {
		src := mc.Source
		if src == "" {
			src = "manual"
		}
		summary.BySource[src]++

		crit := mc.Criticality
		if crit == "" {
			crit = "medium"
		}
		summary.ByCriticality[crit]++

		if mc.Area != "" && !seen[mc.Area] {
			seen[mc.Area] = true
			summary.Areas = append(summary.Areas, mc.Area)
		}

		if mc.LastExecuted == "" {
			summary.StaleCount++
		}
	}

	sort.Strings(summary.Areas)
	return summary
}

func buildCoverageSummary(cov *depgraph.CoverageResult) CoverageSummary {
	cs := CoverageSummary{
		TotalFiles: cov.SourceCount,
	}
	cs.HighCount = cov.BandCounts[depgraph.CoverageBandHigh]
	cs.MediumCount = cov.BandCounts[depgraph.CoverageBandMedium]
	cs.LowCount = cov.BandCounts[depgraph.CoverageBandLow]
	return cs
}

func buildDuplicateSummary(dupes *depgraph.DuplicateResult) DuplicateSummary {
	ds := DuplicateSummary{
		ClusterCount:       len(dupes.Clusters),
		RedundantTestCount: dupes.DuplicateCount,
	}
	for _, c := range dupes.Clusters {
		if c.Similarity > ds.HighestSimilarity {
			ds.HighestSimilarity = c.Similarity
		}
	}
	return ds
}

func buildFanoutSummary(fanout *depgraph.FanoutResult) FanoutSummary {
	fs := FanoutSummary{
		FlaggedCount: fanout.FlaggedCount,
		Threshold:    fanout.Threshold,
	}

	// Top 5 high-fanout nodes.
	limit := 5
	if len(fanout.Entries) < limit {
		limit = len(fanout.Entries)
	}
	for _, e := range fanout.Entries[:limit] {
		if !e.Flagged {
			break
		}
		// Use path if available, otherwise extract label from node ID.
		displayPath := e.Path
		if displayPath == "" {
			displayPath = fanoutNodeLabel(e.NodeID, e.NodeType)
		}
		fs.TopNodes = append(fs.TopNodes, FanoutNode{
			Path:             displayPath,
			NodeType:         e.NodeType,
			TransitiveFanout: e.TransitiveFanout,
		})
	}

	return fs
}

func buildSkipSummary(snap *models.TestSuiteSnapshot) SkipSummary {
	skipped := 0
	total := 0
	for _, sig := range snap.Signals {
		if sig.Type == "skippedTest" {
			skipped++
		}
	}
	for _, tf := range snap.TestFiles {
		total += tf.TestCount
	}
	if total == 0 {
		total = len(snap.TestCases)
	}

	ratio := 0.0
	if total > 0 {
		ratio = float64(skipped) / float64(total)
	}
	return SkipSummary{
		SkippedCount: skipped,
		TotalTests:   total,
		SkipRatio:    ratio,
	}
}

func buildWeakAreas(cov *depgraph.CoverageResult) []WeakArea {
	var areas []WeakArea
	for _, src := range cov.Sources {
		if src.Band == depgraph.CoverageBandLow {
			areas = append(areas, WeakArea{
				Path:      src.Path,
				TestCount: src.TestCount,
				Band:      string(src.Band),
			})
		}
	}
	// Limit to top 10 weakest areas.
	if len(areas) > 10 {
		areas = areas[:10]
	}
	return areas
}

func buildCIOptimization(dupes *depgraph.DuplicateResult, fanout *depgraph.FanoutResult, snap *models.TestSuiteSnapshot) CIOptimizationSummary {
	ci := CIOptimizationSummary{
		DuplicateTestsRemovable: dupes.DuplicateCount,
		HighFanoutNodes:         fanout.FlaggedCount,
	}
	// Count skipped tests as reviewable.
	for _, sig := range snap.Signals {
		if sig.Type == "skippedTest" {
			ci.SkippedTestsReviewable++
		}
	}

	// Build recommendation.
	parts := []string{}
	if dupes.DuplicateCount > 0 {
		parts = append(parts, fmt.Sprintf("%d duplicate tests could be consolidated", dupes.DuplicateCount))
	}
	if fanout.FlaggedCount > 0 {
		parts = append(parts, fmt.Sprintf("%d high-fanout nodes could be refactored to reduce blast radius", fanout.FlaggedCount))
	}
	if ci.SkippedTestsReviewable > 0 {
		parts = append(parts, fmt.Sprintf("%d skipped tests should be reviewed or removed", ci.SkippedTestsReviewable))
	}
	if len(parts) > 0 {
		ci.Recommendation = strings.Join(parts, "; ") + "."
	}

	return ci
}

func buildRiskPosture(snap *models.TestSuiteSnapshot) []RiskDimension {
	if snap.Measurements == nil || len(snap.Measurements.Posture) == 0 {
		return nil
	}
	var dims []RiskDimension
	for _, p := range snap.Measurements.Posture {
		dims = append(dims, RiskDimension{
			Dimension: p.Dimension,
			Band:      p.Band,
		})
	}
	return dims
}

func buildSignalSummary(snap *models.TestSuiteSnapshot) SignalBreakdown {
	sb := SignalBreakdown{
		Total:      len(snap.Signals),
		ByCategory: map[string]int{},
	}
	for _, s := range snap.Signals {
		switch s.Severity {
		case models.SeverityCritical:
			sb.Critical++
		case models.SeverityHigh:
			sb.High++
		case models.SeverityMedium:
			sb.Medium++
		case models.SeverityLow:
			sb.Low++
		}
		sb.ByCategory[string(s.Category)]++
	}
	return sb
}

// fanoutNodeLabel extracts a human-readable label from a graph node ID.
// Node IDs use prefixes like "behavior:module:src/auth.ts" or "file:src/db.ts".
func fanoutNodeLabel(nodeID, nodeType string) string {
	// Try stripping the type prefix: "behavior:module:src/auth.ts" → "src/auth.ts"
	parts := strings.SplitN(nodeID, ":", 3)
	if len(parts) >= 3 {
		return parts[2]
	}
	if len(parts) >= 2 {
		return parts[1]
	}
	if nodeType != "" {
		return nodeType
	}
	return nodeID
}

func deriveTopInsight(r *Report, fanout *depgraph.FanoutResult, dupes *depgraph.DuplicateResult, cov *depgraph.CoverageResult) string {
	// Priority: high-fanout > duplicates > weak coverage > skip burden > generic.
	if fanout.FlaggedCount > 0 && len(fanout.Entries) > 0 {
		top := fanout.Entries[0]
		if top.Flagged {
			label := top.Path
			if label == "" {
				label = fanoutNodeLabel(top.NodeID, top.NodeType)
			}
			return fmt.Sprintf("%s fans out to %d transitive dependents — changes here trigger wide test impact. Consider splitting or isolating.",
				label, top.TransitiveFanout)
		}
	}

	if dupes.DuplicateCount > 0 {
		return fmt.Sprintf("%d tests across %d clusters are structurally similar — consolidation could reduce CI runtime and maintenance burden.",
			dupes.DuplicateCount, len(dupes.Clusters))
	}

	lowCount := cov.BandCounts[depgraph.CoverageBandLow]
	if lowCount > 0 && cov.SourceCount > 0 {
		pct := 100 * lowCount / cov.SourceCount
		return fmt.Sprintf("%d source files (%d%%) have low structural coverage — these are blind spots for change-scoped test selection.",
			lowCount, pct)
	}

	if r.SkippedTestBurden.SkippedCount > 0 {
		return fmt.Sprintf("%d skipped tests are consuming CI resources without providing value — review or remove them.",
			r.SkippedTestBurden.SkippedCount)
	}

	return "No major issues detected. Consider adding coverage or runtime data to unlock deeper analysis."
}

// deriveKeyFindings produces a prioritized list of findings from the same
// data used by deriveTopInsight. Returns the top 3 and the total count.
// Severity assignment mirrors the insights package: the ranking is
// severity (critical > high > medium > low), then category priority
// (reliability > architecture > coverage > optimization).
func deriveKeyFindings(r *Report, fanout *depgraph.FanoutResult, dupes *depgraph.DuplicateResult, cov *depgraph.CoverageResult, snap *models.TestSuiteSnapshot) ([]KeyFinding, int) {
	type candidate struct {
		finding       KeyFinding
		severityOrder int // lower = more severe
		categoryOrder int // lower = higher priority
	}

	var candidates []candidate

	// High-fanout nodes.
	if fanout.FlaggedCount > 0 {
		sev := "medium"
		sevOrd := 3
		if fanout.FlaggedCount > 5 {
			sev = "high"
			sevOrd = 2
		}
		candidates = append(candidates, candidate{
			finding: KeyFinding{
				Title:    fmt.Sprintf("%d high-fanout fixture(s) — changes trigger wide test impact", fanout.FlaggedCount),
				Severity: sev,
				Category: "architecture_debt",
				Metric:   fmt.Sprintf("%d flagged", fanout.FlaggedCount),
			},
			severityOrder: sevOrd,
			categoryOrder: 2,
		})
	}

	// Duplicate clusters.
	if dupes.DuplicateCount > 0 {
		sev := "medium"
		sevOrd := 3
		if dupes.DuplicateCount > 100 {
			sev = "high"
			sevOrd = 2
		}
		candidates = append(candidates, candidate{
			finding: KeyFinding{
				Title:    fmt.Sprintf("%d duplicate tests across %d clusters — consolidation reduces CI time", dupes.DuplicateCount, len(dupes.Clusters)),
				Severity: sev,
				Category: "optimization",
				Metric:   fmt.Sprintf("%d duplicates", dupes.DuplicateCount),
			},
			severityOrder: sevOrd,
			categoryOrder: 4,
		})
	}

	// Behavior redundancy (wasteful overlaps).
	if r.BehaviorRedundancy != nil {
		wastefulCount := 0
		for _, c := range r.BehaviorRedundancy.Clusters {
			if c.OverlapKind == depgraph.OverlapWasteful {
				wastefulCount++
			}
		}
		if wastefulCount > 0 {
			sev := "medium"
			sevOrd := 3
			if wastefulCount > 5 {
				sev = "high"
				sevOrd = 2
			}
			candidates = append(candidates, candidate{
				finding: KeyFinding{
					Title:    fmt.Sprintf("%d wasteful overlap clusters — tests exercise identical behavior surfaces", wastefulCount),
					Severity: sev,
					Category: "optimization",
					Metric:   fmt.Sprintf("%d clusters", wastefulCount),
				},
				severityOrder: sevOrd,
				categoryOrder: 4,
			})
		}
	}

	// Weak coverage.
	lowCount := cov.BandCounts[depgraph.CoverageBandLow]
	if lowCount > 0 && cov.SourceCount > 0 {
		pct := 100 * lowCount / cov.SourceCount
		sev := "medium"
		sevOrd := 3
		if pct > 75 {
			sev = "critical"
			sevOrd = 1
		} else if pct > 50 {
			sev = "high"
			sevOrd = 2
		}
		candidates = append(candidates, candidate{
			finding: KeyFinding{
				Title:    fmt.Sprintf("%d source files (%d%%) have low structural coverage — blind spots for test selection", lowCount, pct),
				Severity: sev,
				Category: "coverage_debt",
				Metric:   fmt.Sprintf("%d files", lowCount),
			},
			severityOrder: sevOrd,
			categoryOrder: 3,
		})
	}

	// Skip burden.
	if r.SkippedTestBurden.SkippedCount > 0 {
		ratio := float64(r.SkippedTestBurden.SkippedCount) / float64(max(r.SkippedTestBurden.TotalTests, 1)) * 100
		sev := "low"
		sevOrd := 4
		if ratio > 10 {
			sev = "high"
			sevOrd = 2
		} else if ratio > 3 {
			sev = "medium"
			sevOrd = 3
		}
		candidates = append(candidates, candidate{
			finding: KeyFinding{
				Title:    fmt.Sprintf("%d skipped tests (%.0f%%) — consuming CI resources without providing value", r.SkippedTestBurden.SkippedCount, ratio),
				Severity: sev,
				Category: "reliability",
				Metric:   fmt.Sprintf("%d skipped", r.SkippedTestBurden.SkippedCount),
			},
			severityOrder: sevOrd,
			categoryOrder: 1,
		})
	}

	// Stability clusters.
	if r.StabilityClusters != nil && r.StabilityClusters.ClusteredTestCount > 0 {
		sev := "medium"
		sevOrd := 3
		if r.StabilityClusters.ClusteredTestCount > 50 {
			sev = "critical"
			sevOrd = 1
		} else if r.StabilityClusters.ClusteredTestCount > 10 {
			sev = "high"
			sevOrd = 2
		}
		candidates = append(candidates, candidate{
			finding: KeyFinding{
				Title:    fmt.Sprintf("%d unstable tests cluster around %d shared root causes", r.StabilityClusters.ClusteredTestCount, len(r.StabilityClusters.Clusters)),
				Severity: sev,
				Category: "reliability",
				Metric:   fmt.Sprintf("%d unstable", r.StabilityClusters.ClusteredTestCount),
			},
			severityOrder: sevOrd,
			categoryOrder: 1,
		})
	}

	// Critical signals.
	if r.SignalSummary.Critical > 0 {
		candidates = append(candidates, candidate{
			finding: KeyFinding{
				Title:    fmt.Sprintf("%d critical signal(s) detected — immediate attention required", r.SignalSummary.Critical),
				Severity: "critical",
				Category: "reliability",
				Metric:   fmt.Sprintf("%d critical", r.SignalSummary.Critical),
			},
			severityOrder: 1,
			categoryOrder: 1,
		})
	}

	// Sort: severity first (ascending = most severe first), then category.
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].severityOrder != candidates[j].severityOrder {
			return candidates[i].severityOrder < candidates[j].severityOrder
		}
		return candidates[i].categoryOrder < candidates[j].categoryOrder
	})

	total := len(candidates)
	top := candidates
	if len(top) > 3 {
		top = top[:3]
	}

	findings := make([]KeyFinding, len(top))
	for i, c := range top {
		findings[i] = c.finding
	}
	return findings, total
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// BuildSnapshotProfileData extracts aggregates from the snapshot for
// the depgraph profiler. Exported for use by CLI commands that need
// to build a profile outside the analyze package.
func BuildSnapshotProfileData(snap *models.TestSuiteSnapshot) depgraph.SnapshotProfileData {
	spd := depgraph.SnapshotProfileData{
		FrameworkCount: len(snap.Frameworks),
	}
	for _, fw := range snap.Frameworks {
		spd.FrameworkTypes = append(spd.FrameworkTypes, string(fw.Type))
	}
	for _, tf := range snap.TestFiles {
		spd.SnapshotAssertionCount += tf.SnapshotCount
		spd.TotalAssertionCount += tf.AssertionCount
	}
	for _, s := range snap.Signals {
		if s.Category == models.CategoryMigration {
			spd.MigrationSignalCount++
		}
		if s.Type == "legacyFrameworkUsage" {
			spd.LegacyFrameworkSignalCount++
		}
	}
	spd.ManualCoverageCount = len(snap.ManualCoverage)
	return spd
}

func buildLimitations(snap *models.TestSuiteSnapshot, hasPolicy bool) []string {
	var lims []string

	if !dsAvailable(snap, "coverage") {
		lims = append(lims, "No coverage data provided; coverage confidence is structural (import-based) only.")
	}
	if !dsAvailable(snap, "runtime") {
		lims = append(lims, "No runtime data provided; skip/flaky/slow test detection unavailable.")
	}
	if !hasPolicy && !dsAvailable(snap, "policy") {
		lims = append(lims, "No policy file found; governance checks skipped.")
	}
	if snap.Ownership == nil || len(snap.Ownership) == 0 {
		lims = append(lims, "No ownership data available; per-owner risk breakdown unavailable.")
	}

	// Sort for determinism.
	sort.Strings(lims)

	return lims
}
