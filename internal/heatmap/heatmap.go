// Package heatmap builds a risk concentration model from a TestSuiteSnapshot.
//
// The heatmap identifies where risk concentrates — by directory, by owner,
// and across risk dimensions — so leadership can understand the overall
// test suite posture at a glance.
package heatmap

import (
	"sort"

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
func Build(snap *models.TestSuiteSnapshot) *Heatmap {
	h := &Heatmap{}

	// Extract directory hotspots from risk surfaces.
	for _, rs := range snap.Risk {
		if rs.Scope == "directory" {
			h.DirectoryHotSpots = append(h.DirectoryHotSpots, HotSpot{
				Scope:          "directory",
				Name:           rs.ScopeName,
				Band:           rs.Band,
				Score:          rs.Score,
				SignalCount:    len(rs.ContributingSignals),
				TopSignalTypes: topTypes(rs.ContributingSignals, 3),
			})
		}
	}
	sort.Slice(h.DirectoryHotSpots, func(i, j int) bool {
		return h.DirectoryHotSpots[i].Score > h.DirectoryHotSpots[j].Score
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
		score := weightSignals(sigs)
		h.OwnerHotSpots = append(h.OwnerHotSpots, HotSpot{
			Scope:          "owner",
			Name:           owner,
			Band:           scoreToBand(score),
			Score:          score,
			SignalCount:    len(sigs),
			TopSignalTypes: topTypes(sigs, 3),
		})
	}
	sort.Slice(h.OwnerHotSpots, func(i, j int) bool {
		return h.OwnerHotSpots[i].Score > h.OwnerHotSpots[j].Score
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
		return pairs[i].count > pairs[j].count
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
