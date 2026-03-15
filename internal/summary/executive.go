// Package summary builds executive-level summaries from Terrain analysis data.
//
// The ExecutiveSummary model synthesizes risk, trends, hotspots, and
// benchmark readiness into a single artifact suitable for:
//   - engineering managers reviewing test system health
//   - tech leads planning remediation
//   - directors/VPEs tracking risk posture
//   - leadership updates and technical debt reviews
//
// The summary is derived entirely from local data. It does not claim
// comparison against external peers. When benchmark readiness is reported,
// it describes what dimensions are measurable — not how they rank.
//
// This model is designed to be reusable by future hosted product UIs
// without schema changes.
package summary

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pmclSF/terrain/internal/benchmark"
	"github.com/pmclSF/terrain/internal/comparison"
	"github.com/pmclSF/terrain/internal/heatmap"
	"github.com/pmclSF/terrain/internal/metrics"
	"github.com/pmclSF/terrain/internal/models"
)

// ExecutiveSummary is the top-level leadership summary artifact.
//
// It is intentionally compact, serializable, and reusable by both
// CLI renderers and future hosted product UIs.
type ExecutiveSummary struct {
	// Posture describes the overall risk posture by dimension.
	Posture PostureSummary `json:"posture"`

	// TopRiskAreas are the highest-concentration risk hotspots.
	TopRiskAreas []FocusArea `json:"topRiskAreas,omitempty"`

	// TrendHighlights summarize notable changes since the previous snapshot.
	// Empty if no prior snapshot exists.
	TrendHighlights []TrendCallout `json:"trendHighlights,omitempty"`

	// HasTrendData indicates whether trend information was available.
	HasTrendData bool `json:"hasTrendData"`

	// DominantDrivers are the most frequent signal types across the repo.
	DominantDrivers []string `json:"dominantDrivers,omitempty"`

	// RecommendedFocus is a concise, evidence-based prioritization note.
	RecommendedFocus string `json:"recommendedFocus"`

	// Recommendations provides structured, prioritized recommendations
	// with what/why/where/evidence-strength context.
	Recommendations []Recommendation `json:"recommendations,omitempty"`

	// BlindSpots identifies areas where data is missing or evidence is weak,
	// limiting confidence in the analysis.
	BlindSpots []BlindSpot `json:"blindSpots,omitempty"`

	// BenchmarkReadiness describes which dimensions are ready for future
	// cross-repo comparison and where data gaps exist.
	BenchmarkReadiness BenchmarkReadinessSummary `json:"benchmarkReadiness"`

	// KeyNumbers provides the essential aggregate counts.
	KeyNumbers KeyNumbers `json:"keyNumbers"`
}

// PostureSummary captures the overall risk posture by dimension.
type PostureSummary struct {
	// OverallBand is the highest risk band across all dimensions.
	OverallBand models.RiskBand `json:"overallBand"`

	// OverallStatement is a one-line summary of overall posture.
	OverallStatement string `json:"overallStatement"`

	// Dimensions lists per-dimension posture.
	Dimensions []DimensionPosture `json:"dimensions,omitempty"`
}

// DimensionPosture is the posture for a single risk dimension.
type DimensionPosture struct {
	Dimension string          `json:"dimension"`
	Band      models.RiskBand `json:"band"`
}

// FocusArea identifies a concentrated risk area.
type FocusArea struct {
	Name        string          `json:"name"`
	Scope       string          `json:"scope"` // "directory" or "owner"
	Band        models.RiskBand `json:"band"`
	RiskType    string          `json:"riskType"`
	SignalCount int             `json:"signalCount"`
}

// TrendCallout is a notable trend change worth surfacing.
type TrendCallout struct {
	// Description is a human-readable callout.
	Description string `json:"description"`

	// Direction is "improved", "worsened", or "unchanged".
	Direction string `json:"direction"`

	// Dimension is the affected area (e.g., "reliability", "quality").
	Dimension string `json:"dimension,omitempty"`
}

// BenchmarkReadinessSummary describes what is ready for future comparison.
type BenchmarkReadinessSummary struct {
	// ReadyDimensions are measurable and suitable for comparison.
	ReadyDimensions []string `json:"readyDimensions"`

	// LimitedDimensions have partial data that would limit comparison accuracy.
	LimitedDimensions []BenchmarkLimitation `json:"limitedDimensions,omitempty"`

	// Segment summarizes the repo's benchmark segmentation if available.
	Segment *benchmark.Segment `json:"segment,omitempty"`
}

// BenchmarkLimitation describes a dimension with incomplete data.
type BenchmarkLimitation struct {
	Dimension string `json:"dimension"`
	Reason    string `json:"reason"`
}

// Recommendation is a structured, evidence-aware action item.
type Recommendation struct {
	// What is the recommended action.
	What string `json:"what"`

	// Why explains the rationale.
	Why string `json:"why"`

	// Where identifies the scope (directory, file, owner).
	Where string `json:"where"`

	// EvidenceStrength is the confidence level of the underlying signals.
	EvidenceStrength models.EvidenceStrength `json:"evidenceStrength"`

	// Priority is a computed rank (lower = more urgent).
	Priority int `json:"priority"`
}

// BlindSpot identifies an area where analysis confidence is limited.
type BlindSpot struct {
	// Area describes what is missing or limited.
	Area string `json:"area"`

	// Reason explains why.
	Reason string `json:"reason"`

	// Remediation suggests how to fill the gap.
	Remediation string `json:"remediation,omitempty"`
}

// KeyNumbers provides essential aggregate counts.
type KeyNumbers struct {
	TestFiles        int `json:"testFiles"`
	Frameworks       int `json:"frameworks"`
	TotalSignals     int `json:"totalSignals"`
	CriticalFindings int `json:"criticalFindings"`
	HighRiskAreas    int `json:"highRiskAreas"`
}

// BuildInput collects all the data sources needed to build an executive summary.
type BuildInput struct {
	Snapshot   *models.TestSuiteSnapshot
	Heatmap    *heatmap.Heatmap
	Metrics    *metrics.Snapshot
	Comparison *comparison.SnapshotComparison // nil if no prior snapshot
	Segment    *benchmark.Segment             // nil if not computed
	HasPolicy  bool
}

// Build creates an ExecutiveSummary from the provided inputs.
func Build(in *BuildInput) *ExecutiveSummary {
	es := &ExecutiveSummary{
		HasTrendData: in.Comparison != nil,
	}

	es.Posture = buildPosture(in.Snapshot, in.Heatmap)
	es.TopRiskAreas = buildTopRiskAreas(in.Heatmap)
	es.DominantDrivers = buildDominantDrivers(in.Snapshot)
	es.KeyNumbers = buildKeyNumbers(in.Snapshot, in.Heatmap)
	es.BenchmarkReadiness = buildBenchmarkReadiness(in.Metrics, in.Segment)

	if in.Comparison != nil {
		es.TrendHighlights = buildTrendHighlights(in.Comparison)
	}

	es.Recommendations = buildRecommendations(in.Snapshot, in.Heatmap)
	es.BlindSpots = buildBlindSpots(in.Snapshot, in.Metrics)

	// Enrich recommendations with coverage-by-type and test identity findings.
	es.Recommendations = appendCoverageRecommendations(es.Recommendations, in.Snapshot)

	// Re-sort and re-number after enrichment.
	sort.SliceStable(es.Recommendations, func(i, j int) bool {
		si := evidenceOrder(es.Recommendations[i].EvidenceStrength)
		sj := evidenceOrder(es.Recommendations[j].EvidenceStrength)
		return si > sj
	})
	for i := range es.Recommendations {
		es.Recommendations[i].Priority = i + 1
	}

	es.RecommendedFocus = buildRecommendedFocus(es)

	return es
}

func buildPosture(snap *models.TestSuiteSnapshot, h *heatmap.Heatmap) PostureSummary {
	ps := PostureSummary{
		OverallBand:      h.PostureBand,
		OverallStatement: h.PostureSummary,
	}

	// Include measurement-layer posture if available (preferred).
	if snap.Measurements != nil && len(snap.Measurements.Posture) > 0 {
		for _, p := range snap.Measurements.Posture {
			ps.Dimensions = append(ps.Dimensions, DimensionPosture{
				Dimension: p.Dimension,
				Band:      models.RiskBand(p.Band),
			})
		}
		return ps
	}

	// Fallback to risk surface posture.
	for _, r := range snap.Risk {
		if r.Scope == "repository" {
			ps.Dimensions = append(ps.Dimensions, DimensionPosture{
				Dimension: r.Type,
				Band:      r.Band,
			})
		}
	}

	return ps
}

func buildTopRiskAreas(h *heatmap.Heatmap) []FocusArea {
	var areas []FocusArea

	limit := 5
	if len(h.DirectoryHotSpots) < limit {
		limit = len(h.DirectoryHotSpots)
	}
	for _, hs := range h.DirectoryHotSpots[:limit] {
		if hs.Band == models.RiskBandLow {
			continue
		}
		riskType := "quality"
		if len(hs.TopSignalTypes) > 0 {
			riskType = categorizeSignalType(hs.TopSignalTypes[0])
		}
		areas = append(areas, FocusArea{
			Name:        hs.Name,
			Scope:       "directory",
			Band:        hs.Band,
			RiskType:    riskType,
			SignalCount: hs.SignalCount,
		})
	}

	return areas
}

func buildDominantDrivers(snap *models.TestSuiteSnapshot) []string {
	counts := map[string]int{}
	for _, s := range snap.Signals {
		counts[string(s.Type)]++
	}

	type kv struct {
		key   string
		count int
	}
	var pairs []kv
	for k, v := range counts {
		pairs = append(pairs, kv{k, v})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].count != pairs[j].count {
			return pairs[i].count > pairs[j].count
		}
		return pairs[i].key < pairs[j].key
	})

	limit := 5
	if len(pairs) < limit {
		limit = len(pairs)
	}
	result := make([]string, limit)
	for i := 0; i < limit; i++ {
		result[i] = pairs[i].key
	}
	return result
}

func buildKeyNumbers(snap *models.TestSuiteSnapshot, h *heatmap.Heatmap) KeyNumbers {
	return KeyNumbers{
		TestFiles:        len(snap.TestFiles),
		Frameworks:       len(snap.Frameworks),
		TotalSignals:     h.TotalSignals,
		CriticalFindings: h.CriticalCount,
		HighRiskAreas:    h.HighRiskAreaCount,
	}
}

func buildTrendHighlights(comp *comparison.SnapshotComparison) []TrendCallout {
	var callouts []TrendCallout

	// Risk band changes (repo-level only)
	for _, rd := range comp.RiskDeltas {
		if rd.Scope != "repository" || !rd.Changed {
			continue
		}
		dir := "worsened"
		if bandOrder(rd.After) < bandOrder(rd.Before) {
			dir = "improved"
		}
		callouts = append(callouts, TrendCallout{
			Description: fmt.Sprintf("%s risk %s (%s → %s)", rd.Type, dir, rd.Before, rd.After),
			Direction:   dir,
			Dimension:   rd.Type,
		})
	}

	// Top signal deltas (limit to most significant)
	limit := 5
	if len(comp.SignalDeltas) < limit {
		limit = len(comp.SignalDeltas)
	}
	for _, sd := range comp.SignalDeltas[:limit] {
		if sd.Delta == 0 {
			continue
		}
		dir := "worsened"
		verb := "increased"
		if sd.Delta < 0 {
			dir = "improved"
			verb = "decreased"
		}
		callouts = append(callouts, TrendCallout{
			Description: fmt.Sprintf("%s findings %s (%+d)", sd.Type, verb, sd.Delta),
			Direction:   dir,
			Dimension:   string(sd.Category),
		})
	}

	// Test file count change
	if comp.TestFileCountDelta != 0 {
		dir := "improved"
		if comp.TestFileCountDelta < 0 {
			dir = "worsened"
		}
		callouts = append(callouts, TrendCallout{
			Description: fmt.Sprintf("test file count changed (%+d)", comp.TestFileCountDelta),
			Direction:   dir,
		})
	}

	return callouts
}

func buildBenchmarkReadiness(ms *metrics.Snapshot, seg *benchmark.Segment) BenchmarkReadinessSummary {
	br := BenchmarkReadinessSummary{
		Segment: seg,
	}

	// Always-ready dimensions based on static analysis
	br.ReadyDimensions = []string{
		"test structure",
		"quality metrics",
		"migration blocker counts",
	}

	// Check for runtime data availability.
	runtimeLimited := false
	for _, note := range ms.Notes {
		if strings.Contains(note, "runtime") {
			runtimeLimited = true
		}
	}

	if !runtimeLimited {
		br.ReadyDimensions = append(br.ReadyDimensions, "health metrics (runtime-backed)")
	} else {
		br.LimitedDimensions = append(br.LimitedDimensions, BenchmarkLimitation{
			Dimension: "speed comparison",
			Reason:    "runtime data is partial or absent",
		})
	}

	// Governance readiness depends on policy presence.
	if ms.Governance.PolicyViolationCount > 0 || (seg != nil && seg.HasPolicy) {
		br.ReadyDimensions = append(br.ReadyDimensions, "governance metrics")
	}

	return br
}

func buildRecommendations(snap *models.TestSuiteSnapshot, h *heatmap.Heatmap) []Recommendation {
	var recs []Recommendation

	// Group signals by directory and evidence strength.
	type dirInfo struct {
		signalCount int
		strongCount int
		modCount    int
		weakCount   int
		topType     string
		topCount    int
	}
	dirs := map[string]*dirInfo{}
	for _, s := range snap.Signals {
		dir := dirFromFile(s.Location.File)
		d, ok := dirs[dir]
		if !ok {
			d = &dirInfo{}
			dirs[dir] = d
		}
		d.signalCount++
		switch s.EvidenceStrength {
		case models.EvidenceStrong:
			d.strongCount++
		case models.EvidenceModerate:
			d.modCount++
		default:
			d.weakCount++
		}
		tkey := string(s.Type)
		if d.topType == "" {
			d.topType = tkey
			d.topCount = 1
		} else if tkey == d.topType {
			d.topCount++
		}
	}

	// Build recommendations from hotspots, prioritized by evidence strength then concentration.
	limit := 5
	if len(h.DirectoryHotSpots) < limit {
		limit = len(h.DirectoryHotSpots)
	}
	for _, hs := range h.DirectoryHotSpots[:limit] {
		if hs.Band == models.RiskBandLow {
			continue
		}
		di := dirs[hs.Name]
		strength := models.EvidenceWeak
		if di != nil {
			if di.strongCount > 0 {
				strength = models.EvidenceStrong
			} else if di.modCount > 0 {
				strength = models.EvidenceModerate
			}
		}

		riskType := "quality"
		if len(hs.TopSignalTypes) > 0 {
			riskType = categorizeSignalType(hs.TopSignalTypes[0])
		}

		rec := Recommendation{
			What:             fmt.Sprintf("Reduce %s findings in %s (%d signals)", riskType, hs.Name, hs.SignalCount),
			Why:              fmt.Sprintf("%s risk band with %s-confidence evidence", strings.ToUpper(string(hs.Band)[:1])+string(hs.Band)[1:], string(strength)),
			Where:            hs.Name,
			EvidenceStrength: strength,
		}
		recs = append(recs, rec)
	}

	// Prioritize: strong evidence first, then by signal count descending.
	sort.SliceStable(recs, func(i, j int) bool {
		si := evidenceOrder(recs[i].EvidenceStrength)
		sj := evidenceOrder(recs[j].EvidenceStrength)
		return si > sj
	})
	for i := range recs {
		recs[i].Priority = i + 1
	}

	return recs
}

func buildBlindSpots(snap *models.TestSuiteSnapshot, ms *metrics.Snapshot) []BlindSpot {
	var spots []BlindSpot

	// Check for missing coverage data.
	if snap.CoverageSummary == nil {
		spots = append(spots, BlindSpot{
			Area:        "Coverage data",
			Reason:      "No coverage artifacts were ingested",
			Remediation: "Run with --coverage <path> to include coverage analysis",
		})
	}

	// Check for missing runtime data.
	hasRuntimeNote := false
	for _, note := range ms.Notes {
		if strings.Contains(note, "runtime") || strings.Contains(note, "Runtime") {
			hasRuntimeNote = true
			break
		}
	}
	if hasRuntimeNote {
		spots = append(spots, BlindSpot{
			Area:        "Runtime metrics",
			Reason:      "No CI runtime data available",
			Remediation: "Provide JUnit XML or CI artifacts for runtime analysis",
		})
	}

	// Check for weak-evidence-only signals.
	weakOnly := 0
	total := 0
	for _, s := range snap.Signals {
		total++
		if s.EvidenceStrength == models.EvidenceWeak || s.EvidenceStrength == "" {
			weakOnly++
		}
	}
	if total > 0 && weakOnly > total/2 {
		spots = append(spots, BlindSpot{
			Area:   "Signal confidence",
			Reason: fmt.Sprintf("%d of %d signals rely on weak evidence (path/name heuristics)", weakOnly, total),
		})
	}

	// Check for missing or weak ownership data.
	if len(snap.Ownership) == 0 {
		spots = append(spots, BlindSpot{
			Area:        "Ownership attribution",
			Reason:      "No ownership data available",
			Remediation: "Add a CODEOWNERS file for per-team risk attribution",
		})
	} else {
		// Check if ownership is sparse.
		allFiles := map[string]bool{}
		for _, tf := range snap.TestFiles {
			allFiles[tf.Path] = true
		}
		for _, cu := range snap.CodeUnits {
			allFiles[cu.Path] = true
		}
		if len(allFiles) > 0 {
			ownedCount := 0
			for path := range allFiles {
				if _, ok := snap.Ownership[path]; ok {
					ownedCount++
				}
			}
			ratio := float64(ownedCount) / float64(len(allFiles))
			if ratio < 0.50 {
				spots = append(spots, BlindSpot{
					Area:        "Ownership coverage",
					Reason:      fmt.Sprintf("Only %d of %d files have ownership (%0.f%%)", ownedCount, len(allFiles), ratio*100),
					Remediation: "Expand CODEOWNERS or .terrain/ownership.yaml to cover more files",
				})
			}
		}
	}

	return spots
}

func dirFromFile(file string) string {
	if idx := strings.LastIndex(file, "/"); idx >= 0 {
		return file[:idx]
	}
	return "."
}

func evidenceOrder(s models.EvidenceStrength) int {
	switch s {
	case models.EvidenceStrong:
		return 3
	case models.EvidenceModerate:
		return 2
	case models.EvidenceWeak:
		return 1
	default:
		return 0
	}
}

func buildRecommendedFocus(es *ExecutiveSummary) string {
	if len(es.TopRiskAreas) == 0 && len(es.DominantDrivers) == 0 {
		return "No significant risk areas identified. Continue monitoring."
	}

	var parts []string

	// Focus on highest-risk area
	if len(es.TopRiskAreas) > 0 {
		top := es.TopRiskAreas[0]
		parts = append(parts, fmt.Sprintf("address %s risk in %s", top.RiskType, top.Name))
	}

	// Include dominant driver if different from risk area focus
	if len(es.DominantDrivers) > 0 {
		parts = append(parts, fmt.Sprintf("reduce %s findings", es.DominantDrivers[0]))
	}

	// Add trend-based focus
	for _, t := range es.TrendHighlights {
		if t.Direction == "worsened" {
			parts = append(parts, fmt.Sprintf("investigate %s trend", t.Dimension))
			break
		}
	}

	if len(parts) == 0 {
		return "Continue monitoring test suite health."
	}

	// Capitalize first part
	parts[0] = strings.ToUpper(parts[0][:1]) + parts[0][1:]
	return strings.Join(parts, "; ") + "."
}

func categorizeSignalType(signalType string) string {
	switch signalType {
	case "flakyTest", "skippedTest", "deadTest", "unstableSuite":
		return "reliability"
	case "slowTest", "runtimeBudgetExceeded":
		return "speed"
	case "weakAssertion", "mockHeavyTest", "untestedExport", "coverageThresholdBreak":
		return "quality"
	case "migrationBlocker", "deprecatedTestPattern", "customMatcherRisk", "dynamicTestGeneration":
		return "migration"
	case "policyViolation", "legacyFrameworkUsage", "skippedTestsInCI":
		return "governance"
	default:
		return "quality"
	}
}

func bandOrder(b models.RiskBand) int {
	switch b {
	case models.RiskBandCritical:
		return 3
	case models.RiskBandHigh:
		return 2
	case models.RiskBandMedium:
		return 1
	default:
		return 0
	}
}

// appendCoverageRecommendations adds recommendations derived from
// coverage-by-type data and test identity when available.
func appendCoverageRecommendations(recs []Recommendation, snap *models.TestSuiteSnapshot) []Recommendation {
	cs := snap.CoverageSummary
	if cs == nil {
		return recs
	}

	// Recommend adding unit tests for e2e-only code units.
	if cs.CoveredOnlyByE2E > 0 {
		recs = append(recs, Recommendation{
			What:             fmt.Sprintf("Add unit tests for %d code unit(s) covered only by e2e tests", cs.CoveredOnlyByE2E),
			Why:              "Code covered only by e2e tests has no fast feedback loop. Failures are expensive to diagnose.",
			Where:            "see coverage insights for specific functions",
			EvidenceStrength: models.EvidenceStrong,
		})
	}

	// Recommend covering uncovered exported functions.
	if cs.UncoveredExported > 0 {
		recs = append(recs, Recommendation{
			What:             fmt.Sprintf("Add test coverage for %d uncovered exported function(s)", cs.UncoveredExported),
			Why:              "Public API surface without tests risks silent regressions.",
			Where:            "see untestedExport signals for specific functions",
			EvidenceStrength: models.EvidenceStrong,
		})
	}

	// Recommend improving coverage diversity when unit test coverage is low.
	if cs.TotalCodeUnits > 0 && cs.CoveredByUnitTests > 0 {
		unitRatio := float64(cs.CoveredByUnitTests) / float64(cs.TotalCodeUnits)
		if unitRatio < 0.40 {
			recs = append(recs, Recommendation{
				What:             fmt.Sprintf("Increase unit test coverage (currently %.0f%% of code units)", unitRatio*100),
				Why:              "Low unit test coverage means most validation depends on slower, broader tests.",
				Where:            "prioritize exported functions and high-complexity modules",
				EvidenceStrength: models.EvidenceStrong,
			})
		}
	}

	// Surface concentrated health instability from test identity data.
	healthByTest := map[string]int{}
	for _, s := range snap.Signals {
		if s.Category != models.CategoryHealth {
			continue
		}
		if testID, ok := s.Metadata["testId"].(string); ok && testID != "" {
			healthByTest[testID]++
		}
	}
	if len(healthByTest) > 0 {
		// Find the test with most health signals.
		maxCount := 0
		for _, c := range healthByTest {
			if c > maxCount {
				maxCount = c
			}
		}
		if maxCount >= 2 {
			recs = append(recs, Recommendation{
				What:             fmt.Sprintf("Investigate %d test(s) with concentrated instability", len(healthByTest)),
				Why:              "Health signals (slow, flaky, skipped) cluster around specific persistent tests.",
				Where:            "see health signals with testId metadata for specific tests",
				EvidenceStrength: models.EvidenceStrong,
			})
		}
	}

	return recs
}
