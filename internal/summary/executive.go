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
// This model is a compact, serializable summary artifact reused across
// CLI renderers without schema changes.
package summary

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/pmclSF/terrain/internal/benchmark"
	"github.com/pmclSF/terrain/internal/comparison"
	"github.com/pmclSF/terrain/internal/heatmap"
	"github.com/pmclSF/terrain/internal/measurement"
	"github.com/pmclSF/terrain/internal/metrics"
	"github.com/pmclSF/terrain/internal/models"
	"github.com/pmclSF/terrain/internal/signals"
)

// ExecutiveSummary is the top-level leadership summary artifact.
//
// It is intentionally compact, serializable, and reusable across
// CLI renderers.
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

// DimensionPosture is the posture for a single risk dimension. 0.2.0:
// gained `KeyMeasurements` so the executive renderer can surface
// concrete numbers ("0.7% uncovered exports · 7.6% weak assertions")
// instead of just band labels ("Strong"). Bands are categorical
// compression of these numbers; the renderer's job is to give the
// reader the actual measurements alongside (or instead of) the band.
type DimensionPosture struct {
	Dimension       string           `json:"dimension"`
	Band            models.RiskBand  `json:"band"`
	KeyMeasurements []KeyMeasurement `json:"keyMeasurements,omitempty"`
	// NeedsInput names the missing input ("test files", "runtime
	// data", "coverage data") when Band is unknown and the gap is
	// identifiable from the dimension's measurements. Empty when the
	// dimension was measured or the gap could not be attributed.
	NeedsInput string `json:"needsInput,omitempty"`
}

// KeyMeasurement is a single measurement surfaced in the executive
// summary, compact-formatted for one-line display.
type KeyMeasurement struct {
	// ID is the stable measurement identifier (e.g. "health.flaky_share").
	// Carried for stability + JSON roundtrip; renderers don't display it
	// directly.
	ID string `json:"id"`
	// ShortLabel is the human-readable noun the renderer pairs with
	// the formatted value (e.g. "flaky", "uncovered exports",
	// "high-fanout fixtures"). Derived from the measurement ID via
	// `measurementShortLabel` in this package.
	ShortLabel string `json:"shortLabel"`
	// FormattedValue is the value pre-rendered for display
	// ("3 / 850", "891", "low"). Lets renderers stay agnostic
	// about Units; storage carries both for tooling that wants to
	// re-render.
	FormattedValue string `json:"formattedValue"`
	// Value + Units carry the raw measurement so JSON consumers can
	// re-render in their own format.
	Value float64 `json:"value"`
	Units string  `json:"units,omitempty"`
	// Numerator/Denominator carry the totals parsed from the
	// measurement explanation (e.g. 28 / 772). Present when the
	// explanation exposes counts; zero when not applicable.
	Numerator   int `json:"numerator,omitempty"`
	Denominator int `json:"denominator,omitempty"`
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

// plural returns the singular form when n == 1, otherwise singular +
// "s". Local helper used in recommendation titles to avoid awkward
// `n thing(s)` notation in user-visible text.
func plural(n int, singular string) string {
	if n == 1 {
		return singular
	}
	return singular + "s"
}

// measurementShortLabels maps the canonical measurement IDs to a
// short noun the executive renderer can pair with the formatted
// value (e.g. "0.0% flaky", "3.6% skipped", "891 high-fanout
// fixtures"). Only shipping measurements appear here; unknown IDs
// fall through to the bare measurement-id suffix.
var measurementShortLabels = map[string]string{
	// Health.
	"health.flaky_share":     "flaky",
	"health.skip_density":    "skipped",
	"health.dead_test_share": "dead",
	"health.slow_test_share": "slow",
	// Coverage depth.
	"coverage_depth.uncovered_exports":     "uncovered exports",
	"coverage_depth.weak_assertion_share":  "weak assertions",
	"coverage_depth.coverage_breach_share": "coverage breaches",
	// Coverage diversity.
	"coverage_diversity.mock_heavy_share":        "mock-heavy",
	"coverage_diversity.framework_fragmentation": "frameworks",
	"coverage_diversity.e2e_concentration":       "e2e-concentrated",
	"coverage_diversity.e2e_only_units":          "e2e-only units",
	"coverage_diversity.unit_test_coverage":      "unit-test covered",
	// Structural risk.
	"structural_risk.migration_blocker_density": "migration blockers",
	"structural_risk.deprecated_pattern_share":  "deprecated patterns",
	"structural_risk.dynamic_generation_share":  "dynamic generation",
	// Operational risk.
	"operational_risk.policy_violation_density": "policy violations",
	"operational_risk.legacy_framework_share":   "legacy frameworks",
	"operational_risk.runtime_budget_breach":    "runtime-budget breaches",
}

// numeratorDenominatorRe extracts the first two integers from a
// measurement explanation. Every shipping measurement formats its
// explanation as "%d of %d ..." or "%d ... out of %d ..." or "%d
// ... across %d ..." so a leading-pair extractor reliably surfaces
// numerator and denominator without needing each measurement to
// carry separate fields.
var numeratorDenominatorRe = regexp.MustCompile(`(\d+)\D+(\d+)`)

// toKeyMeasurement converts a models.MeasurementResult into the
// compact KeyMeasurement shape used by the executive summary. The
// preferred display form is "N / D label" (e.g. "28 / 772 skipped"),
// extracted from the measurement explanation. When the explanation
// does not expose counts (band-typed measurements, count-only
// totals), formats fall back to:
//
//	count   → "891"    (no unit suffix; renderer pairs with label)
//	band    → "low"    (the band word itself; pairs with label like "structural risk")
//	ratio   → "3.6%"   (× 100, one decimal — fallback when N/D parse fails)
//	percent → "3.6%"   (already a percent — same fallback)
//
// Unknown units fall back to the raw value formatted with %v.
func toKeyMeasurement(m models.MeasurementResult) KeyMeasurement {
	label := measurementShortLabels[m.ID]
	if label == "" {
		// Fallback: use the part after the last dot ("uncovered_exports"
		// → "uncovered exports") with underscores → spaces.
		if dot := strings.LastIndex(m.ID, "."); dot >= 0 && dot < len(m.ID)-1 {
			label = strings.ReplaceAll(m.ID[dot+1:], "_", " ")
		} else {
			label = m.ID
		}
	}

	num, den := parseNumeratorDenominator(m.Explanation)

	return KeyMeasurement{
		ID:             m.ID,
		ShortLabel:     label,
		FormattedValue: formatMeasurementValue(m, num, den),
		Value:          m.Value,
		Units:          m.Units,
		Numerator:      num,
		Denominator:    den,
	}
}

// parseNumeratorDenominator extracts the leading two integers from a
// measurement explanation. Returns (0, 0) if no pair was found.
func parseNumeratorDenominator(explanation string) (int, int) {
	m := numeratorDenominatorRe.FindStringSubmatch(explanation)
	if len(m) != 3 {
		return 0, 0
	}
	num, err1 := strconv.Atoi(m[1])
	den, err2 := strconv.Atoi(m[2])
	if err1 != nil || err2 != nil {
		return 0, 0
	}
	return num, den
}

// formatMeasurementValue renders the numeric value of a measurement
// in the units convention the user expects. Used by toKeyMeasurement.
// When numerator/denominator were parsed from the explanation, the
// "N / D" form is preferred over a percentage — concrete totals are
// more legible than a percentage that hides scale ("28 / 772 skipped"
// versus "3.6% skipped").
func formatMeasurementValue(m models.MeasurementResult, num, den int) string {
	if den > 0 {
		return fmt.Sprintf("%d / %d", num, den)
	}
	switch measurement.Units(m.Units) {
	case measurement.UnitsRatio:
		return fmt.Sprintf("%.1f%%", m.Value*100)
	case measurement.UnitsPercent:
		return fmt.Sprintf("%.1f%%", m.Value)
	case measurement.UnitsCount:
		return fmt.Sprintf("%d", int(m.Value))
	case measurement.UnitsBand:
		return m.Band
	default:
		return fmt.Sprintf("%v", m.Value)
	}
}

// Build creates an ExecutiveSummary from the provided inputs.
func Build(in *BuildInput) *ExecutiveSummary {
	es := &ExecutiveSummary{
		HasTrendData: in.Comparison != nil,
	}

	es.Posture = buildPosture(in.Snapshot, in.Heatmap)
	es.TopRiskAreas = buildTopRiskAreas(in.Heatmap)
	var driverCounts map[string]int
	es.DominantDrivers, driverCounts = buildDominantDrivers(in.Snapshot)
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

	es.RecommendedFocus = buildRecommendedFocus(es, driverCounts)

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
			dp := DimensionPosture{
				Dimension: p.Dimension,
				Band:      models.RiskBand(p.Band),
			}
			// Surface up to 4 measurements per dimension so the
			// executive renderer can show concrete numbers alongside
			// (or instead of) the band. Driving measurements first
			// (by ID match), then any remaining measurement from the
			// dimension to fill the slot. Measurements with a zero
			// numerator are dropped — "0 / 772 flaky" is noise, not
			// signal, and crowds out the measurements that actually
			// changed.
			driving := map[string]bool{}
			for _, id := range p.DrivingMeasurements {
				driving[id] = true
			}
			const maxKeyMeasurements = 4
			appendKM := func(m models.MeasurementResult) {
				km := toKeyMeasurement(m)
				// Hide zero-valued measurements. Two cases share the
				// same skip rule:
				//   1. counted form parsed: "0 / 772 flaky" — true
				//      zero, redundant given the band already says
				//      "Strong"
				//   2. no-data fallback: "0.0% slow" — measurement
				//      was bypassed because evidence was weak/none,
				//      so the value is structurally zero, not
				//      empirically zero
				// Both are noise; they crowd out the measurements
				// that actually moved the band.
				if km.Numerator == 0 && km.Value == 0 {
					return
				}
				dp.KeyMeasurements = append(dp.KeyMeasurements, km)
			}
			// Pass 1: driving measurements (most informative).
			for _, m := range p.Measurements {
				if len(dp.KeyMeasurements) >= maxKeyMeasurements {
					break
				}
				if !driving[m.ID] {
					continue
				}
				appendKM(m)
			}
			// Pass 2: fill remaining slots with non-driving measurements.
			for _, m := range p.Measurements {
				if len(dp.KeyMeasurements) >= maxKeyMeasurements {
					break
				}
				if driving[m.ID] {
					continue
				}
				appendKM(m)
			}
			if dp.Band == models.RiskBand(measurement.PostureUnknown) {
				dp.NeedsInput = unmeasuredInput(p.Measurements)
			}
			ps.Dimensions = append(ps.Dimensions, dp)
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

// unmeasuredInput derives which input an unmeasured dimension is
// waiting on. When a dimension's band is unknown, its measurements
// carry explanations naming the missing data ("No test files
// detected.", "No runtime data available; ...", "No coverage data
// available."). The most fundamental gap wins: without test files
// nothing downstream can be measured. Returns "" when the gap cannot
// be attributed, so renderers fall back to a plain "not yet measured".
func unmeasuredInput(measurements []models.MeasurementResult) string {
	var needsTests, needsRuntime, needsCoverage bool
	for _, m := range measurements {
		if m.Band != "" && m.Band != string(measurement.PostureUnknown) {
			continue
		}
		switch {
		case strings.Contains(m.Explanation, "No test files"):
			needsTests = true
		case strings.Contains(m.Explanation, "No runtime data"):
			needsRuntime = true
		case strings.Contains(m.Explanation, "No coverage data"):
			needsCoverage = true
		}
	}
	switch {
	case needsTests:
		return "test files"
	case needsRuntime:
		return "runtime data"
	case needsCoverage:
		return "coverage data"
	}
	return ""
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

// buildDominantDrivers returns the most frequent signal types (raw
// type ids, for JSON stability) plus a type → count map so the
// recommended-focus sentence can cite concrete finding counts.
func buildDominantDrivers(snap *models.TestSuiteSnapshot) ([]string, map[string]int) {
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
	return result, counts
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
			Description: fmt.Sprintf("%s findings %s (%+d)", signals.TitleForType(sd.Type), verb, sd.Delta),
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
		strongCount int
		modCount    int
	}
	dirs := map[string]*dirInfo{}
	for _, s := range snap.Signals {
		dir := dirFromFile(s.Location.File)
		d, ok := dirs[dir]
		if !ok {
			d = &dirInfo{}
			dirs[dir] = d
		}
		switch s.EvidenceStrength {
		case models.EvidenceStrong:
			d.strongCount++
		case models.EvidenceModerate:
			d.modCount++
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

// buildRecommendedFocus produces the single action line for the
// Recommended Focus section. Priority: dominant driver first — it
// names a concrete rule with a runnable next step — then the top risk
// area, then a worsened trend, then the calm default. The report's
// Top Risk Areas and Trend Highlights sections carry the detail this
// line does not repeat.
func buildRecommendedFocus(es *ExecutiveSummary, driverCounts map[string]int) string {
	if len(es.DominantDrivers) > 0 {
		d := es.DominantDrivers[0]
		title := signals.TitleForType(models.SignalType(d))
		if n := driverCounts[d]; n > 0 {
			return fmt.Sprintf("Start with: %s — %d %s › terrain explain %s", title, n, plural(n, "finding"), d)
		}
		return fmt.Sprintf("Start with: %s › terrain explain %s", title, d)
	}

	if len(es.TopRiskAreas) > 0 {
		top := es.TopRiskAreas[0]
		return fmt.Sprintf("Start with: %s — %d %s %s", top.Name, top.SignalCount, top.RiskType, plural(top.SignalCount, "finding"))
	}

	for _, t := range es.TrendHighlights {
		if t.Direction == "worsened" {
			return fmt.Sprintf("Investigate the worsened %s trend.", t.Dimension)
		}
	}

	return "No significant risk areas identified. Continue monitoring."
}

func categorizeSignalType(signalType string) string {
	st := models.SignalType(signalType)
	switch st {
	case signals.SignalFlakyTest, signals.SignalSkippedTest, signals.SignalDeadTest, signals.SignalUnstableSuite:
		return "reliability"
	case signals.SignalSlowTest, signals.SignalRuntimeBudgetExceeded:
		return "speed"
	case signals.SignalWeakAssertion, signals.SignalMockHeavyTest, signals.SignalUntestedExport, signals.SignalCoverageThresholdBreak:
		return "quality"
	case signals.SignalMigrationBlocker, signals.SignalDeprecatedTestPattern, signals.SignalCustomMatcherRisk, signals.SignalDynamicTestGeneration:
		return "migration"
	case signals.SignalPolicyViolation, signals.SignalLegacyFrameworkUsage, signals.SignalSkippedTestsInCI:
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
			What:             fmt.Sprintf("Add unit tests for %d code %s covered only by e2e tests", cs.CoveredOnlyByE2E, plural(cs.CoveredOnlyByE2E, "unit")),
			Why:              "Code covered only by e2e tests has no fast feedback loop. Failures are expensive to diagnose.",
			Where:            "see coverage insights for specific functions",
			EvidenceStrength: models.EvidenceStrong,
		})
	}

	// Recommend covering uncovered exported functions.
	if cs.UncoveredExported > 0 {
		recs = append(recs, Recommendation{
			What:             fmt.Sprintf("Add test coverage for %d uncovered exported %s", cs.UncoveredExported, plural(cs.UncoveredExported, "function")),
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
				What:             fmt.Sprintf("Investigate %d %s with concentrated instability", len(healthByTest), plural(len(healthByTest), "test")),
				Why:              "Health signals (slow, flaky, skipped) cluster around specific persistent tests.",
				Where:            "see health signals with testId metadata for specific tests",
				EvidenceStrength: models.EvidenceStrong,
			})
		}
	}

	return recs
}
