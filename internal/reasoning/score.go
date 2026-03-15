package reasoning

import (
	"math"
	"sort"
)

// Confidence band thresholds.
const (
	HighConfidence   = 0.7
	MediumConfidence = 0.4
	MinConfidence    = 0.1

	// DefaultLengthDecay is the per-hop confidence decay factor.
	DefaultLengthDecay = 0.85

	// DefaultFanoutThreshold is the out-degree above which fanout penalty applies.
	DefaultFanoutThreshold = 5
)

// ConfidenceBand classifies a confidence score into a discrete band.
type ConfidenceBand string

const (
	BandHigh   ConfidenceBand = "high"
	BandMedium ConfidenceBand = "medium"
	BandLow    ConfidenceBand = "low"
)

// ClassifyBand maps a confidence score to a discrete band.
func ClassifyBand(confidence float64) ConfidenceBand {
	if confidence >= HighConfidence {
		return BandHigh
	}
	if confidence >= MediumConfidence {
		return BandMedium
	}
	return BandLow
}

// ScoreHop computes the confidence after traversing one edge.
//
// Formula: confidence × edgeConfidence × lengthDecay × fanoutPenalty
//
// Where fanoutPenalty = 1 / log2(outDegree + 1) when outDegree > fanoutThreshold,
// otherwise 1.0.
//
// This matches the scoring used in depgraph.AnalyzeImpact but is extracted
// as a reusable primitive.
func ScoreHop(currentConfidence, edgeConfidence float64, depth, outDegree int, lengthDecay float64, fanoutThreshold int) float64 {
	score := currentConfidence * edgeConfidence * lengthDecay

	if outDegree > fanoutThreshold {
		score /= math.Log2(float64(outDegree) + 1)
	}

	return score
}

// FanoutPenalty computes the fanout penalty for a node with the given out-degree.
// Returns 1.0 (no penalty) when outDegree <= threshold.
func FanoutPenalty(outDegree, threshold int) float64 {
	if outDegree <= threshold {
		return 1.0
	}
	return 1.0 / math.Log2(float64(outDegree)+1)
}

// CompoundConfidence computes the confidence along a multi-hop path
// by applying decay and edge confidences at each step.
func CompoundConfidence(edgeConfidences []float64, lengthDecay float64) float64 {
	if len(edgeConfidences) == 0 {
		return 1.0
	}
	conf := 1.0
	for i, ec := range edgeConfidences {
		_ = i // depth is implicit in iteration
		conf *= ec * lengthDecay
	}
	return conf
}

// sortResults sorts ReachResults by confidence descending, then by NodeID
// for deterministic output.
func sortResults(results []ReachResult) {
	sort.Slice(results, func(i, j int) bool {
		if math.Abs(results[i].Confidence-results[j].Confidence) > 1e-9 {
			return results[i].Confidence > results[j].Confidence
		}
		return results[i].NodeID < results[j].NodeID
	})
}

// TopN returns the top n results by confidence. If n <= 0 or n >= len(results),
// returns all results.
func TopN(results []ReachResult, n int) []ReachResult {
	if n <= 0 || n >= len(results) {
		return results
	}
	// Results are already sorted by sortResults.
	return results[:n]
}

// FilterByBand returns only results in the given confidence band.
func FilterByBand(results []ReachResult, band ConfidenceBand) []ReachResult {
	var out []ReachResult
	for _, r := range results {
		if ClassifyBand(r.Confidence) == band {
			out = append(out, r)
		}
	}
	return out
}

// BandCounts returns the number of results in each confidence band.
func BandCounts(results []ReachResult) map[ConfidenceBand]int {
	counts := map[ConfidenceBand]int{}
	for _, r := range results {
		counts[ClassifyBand(r.Confidence)]++
	}
	return counts
}
