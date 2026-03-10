// Package heatmap builds a risk concentration model from a TestSuiteSnapshot.
//
// The heatmap identifies where risk concentrates — by directory, by owner,
// and across risk dimensions — so leadership can understand the overall
// test suite posture at a glance.
package heatmap

import (
	"sort"

	"github.com/pmclSF/hamlet/internal/graph"
	"github.com/pmclSF/hamlet/internal/models"
)

// HotSpot represents a single area of concentrated risk.
type HotSpot struct {
	// Scope is "directory" or "owner".
	Scope string `json:"scope"`

	// Name is the directory path or owner name.
	Name string `json:"name"`

	// Band is the qualitative risk band for this area.
	Band models.RiskBand `json:"band"`

	// Score is the weighted risk score.
	Score float64 `json:"score"`

	// SignalCount is the number of signals contributing to this hotspot.
	SignalCount int `json:"signalCount"`

	// TopSignalTypes lists the most frequent signal types in this area.
	TopSignalTypes []string `json:"topSignalTypes"`

	// FileCount is the number of test files in this hotspot scope.
	// For owner hotspots, this is owner-owned test files.
	FileCount int `json:"fileCount,omitempty"`

	// SignalDensity is the average signals per test file for this hotspot.
	SignalDensity float64 `json:"signalDensity,omitempty"`

	// UncoveredExported is the count of exported code units with no coverage.
	// Populated when the graph is available.
	UncoveredExported int `json:"uncoveredExported,omitempty"`

	// E2EOnlyUnits is the count of code units covered only by e2e tests.
	// Populated when the graph is available.
	E2EOnlyUnits int `json:"e2eOnlyUnits,omitempty"`

	// HealthSignals is the count of health-category signals (flaky, slow, skipped).
	// Populated when the graph is available.
	HealthSignals int `json:"healthSignals,omitempty"`
}

// Heatmap is the full risk concentration model.
type Heatmap struct {
	// DirectoryHotSpots are risk-sorted directory concentrations.
	DirectoryHotSpots []HotSpot `json:"directoryHotSpots"`

	// OwnerHotSpots are risk-sorted owner concentrations.
	OwnerHotSpots []HotSpot `json:"ownerHotSpots"`

	// PostureSummary is a one-line summary of overall risk posture.
	PostureSummary string `json:"postureSummary"`

	// PostureBand is the highest risk band across all surfaces.
	PostureBand models.RiskBand `json:"postureBand"`

	// TotalSignals is the total number of actionable signals.
	TotalSignals int `json:"totalSignals"`

	// CriticalCount is the number of critical-severity signals.
	CriticalCount int `json:"criticalCount"`

	// HighRiskAreaCount is the number of areas with high or critical risk.
	HighRiskAreaCount int `json:"highRiskAreaCount"`
}

// Build creates a Heatmap from a TestSuiteSnapshot.
// For richer owner hotspots with coverage data, use BuildWithGraph.
func Build(snap *models.TestSuiteSnapshot) *Heatmap {
	return buildHeatmap(snap, nil)
}

// BuildWithGraph creates a Heatmap enriched with graph-derived coverage
// and health data for owner hotspots.
func BuildWithGraph(snap *models.TestSuiteSnapshot, g *graph.Graph) *Heatmap {
	return buildHeatmap(snap, g)
}

func buildHeatmap(snap *models.TestSuiteSnapshot, g *graph.Graph) *Heatmap {
	h := &Heatmap{}
	dirFileCounts := countFilesByDirectory(snap.TestFiles)
	ownerFileCounts := countFilesByOwner(snap.TestFiles)

	// Extract directory hotspots from risk surfaces.
	for _, rs := range snap.Risk {
		if rs.Scope == "directory" {
			fileCount := dirFileCounts[rs.ScopeName]
			score := rs.Score
			if fileCount > 0 && len(rs.ContributingSignals) > 0 {
				score = normalizedRiskScore(weightSignals(rs.ContributingSignals), fileCount)
			}
			h.DirectoryHotSpots = append(h.DirectoryHotSpots, HotSpot{
				Scope:          "directory",
				Name:           rs.ScopeName,
				Band:           rs.Band,
				Score:          score,
				SignalCount:    len(rs.ContributingSignals),
				TopSignalTypes: topTypes(rs.ContributingSignals, 3),
				FileCount:      fileCount,
				SignalDensity:  signalDensity(len(rs.ContributingSignals), fileCount),
			})
		}
	}
	sort.Slice(h.DirectoryHotSpots, func(i, j int) bool {
		if h.DirectoryHotSpots[i].Score != h.DirectoryHotSpots[j].Score {
			return h.DirectoryHotSpots[i].Score > h.DirectoryHotSpots[j].Score
		}
		return h.DirectoryHotSpots[i].Name < h.DirectoryHotSpots[j].Name
	})

	// Build owner hotspots from signals.
	ownerSignals := map[string][]models.Signal{}
	for _, s := range snap.Signals {
		owner := s.Owner
		if owner == "" {
			owner = "unknown"
		}
		ownerSignals[owner] = append(ownerSignals[owner], s)
	}
	for owner, sigs := range ownerSignals {
		fileCount := ownerFileCounts[owner]
		if fileCount == 0 {
			fileCount = uniqueFileCountForSignals(sigs)
		}
		score := normalizedRiskScore(weightSignals(sigs), fileCount)
		h.OwnerHotSpots = append(h.OwnerHotSpots, HotSpot{
			Scope:          "owner",
			Name:           owner,
			Band:           scoreToBand(score),
			Score:          score,
			SignalCount:    len(sigs),
			TopSignalTypes: topTypes(sigs, 3),
			FileCount:      fileCount,
			SignalDensity:  signalDensity(len(sigs), fileCount),
		})
	}
	// Enrich owner hotspots with graph data when available.
	if g != nil {
		ownerSummaries := g.OwnerRiskSummaries()
		summaryMap := make(map[string]*graph.OwnerRiskSummary, len(ownerSummaries))
		for i := range ownerSummaries {
			summaryMap[ownerSummaries[i].Owner] = &ownerSummaries[i]
		}
		for i := range h.OwnerHotSpots {
			if s, ok := summaryMap[h.OwnerHotSpots[i].Name]; ok {
				h.OwnerHotSpots[i].UncoveredExported = s.UncoveredExported
				h.OwnerHotSpots[i].E2EOnlyUnits = s.E2EOnlyUnits
				h.OwnerHotSpots[i].HealthSignals = s.HealthSignals
			}
		}
	}

	sort.Slice(h.OwnerHotSpots, func(i, j int) bool {
		if h.OwnerHotSpots[i].Score != h.OwnerHotSpots[j].Score {
			return h.OwnerHotSpots[i].Score > h.OwnerHotSpots[j].Score
		}
		return h.OwnerHotSpots[i].Name < h.OwnerHotSpots[j].Name
	})

	// Compute aggregate posture.
	h.TotalSignals = len(snap.Signals)
	for _, s := range snap.Signals {
		if s.Severity == models.SeverityCritical {
			h.CriticalCount++
		}
	}

	h.PostureBand = highestBand(snap.Risk)
	h.HighRiskAreaCount = countHighRisk(h.DirectoryHotSpots, h.OwnerHotSpots)
	h.PostureSummary = buildPostureSummary(h)

	return h
}

func countFilesByDirectory(testFiles []models.TestFile) map[string]int {
	counts := map[string]int{}
	for _, tf := range testFiles {
		dir := filepathDir(tf.Path)
		if dir == "" || dir == "." {
			continue
		}
		counts[dir]++
	}
	return counts
}

func countFilesByOwner(testFiles []models.TestFile) map[string]int {
	counts := map[string]int{}
	for _, tf := range testFiles {
		owner := tf.Owner
		if owner == "" {
			owner = "unknown"
		}
		counts[owner]++
	}
	return counts
}

func uniqueFileCountForSignals(signals []models.Signal) int {
	seen := map[string]bool{}
	for _, s := range signals {
		if s.Location.File == "" {
			continue
		}
		seen[s.Location.File] = true
	}
	return len(seen)
}

func signalDensity(signalCount, fileCount int) float64 {
	if signalCount <= 0 || fileCount <= 0 {
		return 0
	}
	return float64(signalCount) / float64(fileCount)
}

func normalizedRiskScore(weight float64, fileCount int) float64 {
	if fileCount <= 0 {
		return weight
	}
	return (weight / float64(fileCount)) * 10.0
}

func filepathDir(path string) string {
	// Lightweight directory extraction without importing filepath.
	lastSlash := -1
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			lastSlash = i
		}
	}
	if lastSlash <= 0 {
		return "."
	}
	return path[:lastSlash]
}

var severityWeight = map[models.SignalSeverity]float64{
	models.SeverityCritical: 4.0,
	models.SeverityHigh:     3.0,
	models.SeverityMedium:   2.0,
	models.SeverityLow:      1.0,
	models.SeverityInfo:     0.5,
}

func weightSignals(signals []models.Signal) float64 {
	var total float64
	for _, s := range signals {
		total += severityWeight[s.Severity]
	}
	return total
}

func scoreToBand(score float64) models.RiskBand {
	switch {
	case score >= 16:
		return models.RiskBandCritical
	case score >= 9:
		return models.RiskBandHigh
	case score >= 4:
		return models.RiskBandMedium
	default:
		return models.RiskBandLow
	}
}

func topTypes(signals []models.Signal, n int) []string {
	counts := map[string]int{}
	for _, s := range signals {
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

	result := make([]string, 0, n)
	for i, p := range pairs {
		if i >= n {
			break
		}
		result = append(result, p.key)
	}
	return result
}

func highestBand(risks []models.RiskSurface) models.RiskBand {
	bandOrder := map[models.RiskBand]int{
		models.RiskBandLow:      0,
		models.RiskBandMedium:   1,
		models.RiskBandHigh:     2,
		models.RiskBandCritical: 3,
	}

	highest := models.RiskBandLow
	for _, r := range risks {
		if bandOrder[r.Band] > bandOrder[highest] {
			highest = r.Band
		}
	}
	return highest
}

func countHighRisk(dirSpots, ownerSpots []HotSpot) int {
	count := 0
	for _, h := range dirSpots {
		if h.Band == models.RiskBandHigh || h.Band == models.RiskBandCritical {
			count++
		}
	}
	for _, h := range ownerSpots {
		if h.Band == models.RiskBandHigh || h.Band == models.RiskBandCritical {
			count++
		}
	}
	return count
}

func buildPostureSummary(h *Heatmap) string {
	if h.TotalSignals == 0 {
		return "No actionable signals detected. Test suite posture is clean."
	}

	switch h.PostureBand {
	case models.RiskBandCritical:
		return "Critical risk detected. Immediate attention required on high-risk areas."
	case models.RiskBandHigh:
		return "High risk detected. Prioritize remediation of concentrated risk areas."
	case models.RiskBandMedium:
		return "Moderate risk. Address top findings to improve test suite health."
	default:
		return "Low risk. Test suite is in good shape with minor improvement opportunities."
	}
}
