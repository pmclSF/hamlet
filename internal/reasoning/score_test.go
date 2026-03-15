package reasoning

import (
	"math"
	"testing"
)

func TestClassifyBand(t *testing.T) {
	tests := []struct {
		confidence float64
		want       ConfidenceBand
	}{
		{1.0, BandHigh},
		{0.7, BandHigh},
		{0.69, BandMedium},
		{0.4, BandMedium},
		{0.39, BandLow},
		{0.1, BandLow},
		{0.0, BandLow},
	}
	for _, tt := range tests {
		got := ClassifyBand(tt.confidence)
		if got != tt.want {
			t.Errorf("ClassifyBand(%v) = %v, want %v", tt.confidence, got, tt.want)
		}
	}
}

func TestScoreHop(t *testing.T) {
	// Basic case: no fanout penalty.
	got := ScoreHop(1.0, 0.9, 0, 3, 0.85, 5)
	want := 1.0 * 0.9 * 0.85
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("ScoreHop(no fanout) = %v, want %v", got, want)
	}

	// With fanout penalty: outDegree=10 > threshold=5.
	got = ScoreHop(1.0, 0.9, 0, 10, 0.85, 5)
	want = 1.0 * 0.9 * 0.85 / math.Log2(11)
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("ScoreHop(with fanout) = %v, want %v", got, want)
	}
}

func TestScoreHop_FanoutAtThreshold(t *testing.T) {
	// At threshold: no penalty.
	got := ScoreHop(0.8, 1.0, 0, 5, 0.85, 5)
	want := 0.8 * 1.0 * 0.85
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("ScoreHop(at threshold) = %v, want %v", got, want)
	}

	// Just above threshold: penalty applies.
	got = ScoreHop(0.8, 1.0, 0, 6, 0.85, 5)
	want = 0.8 * 1.0 * 0.85 / math.Log2(7)
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("ScoreHop(above threshold) = %v, want %v", got, want)
	}
}

func TestFanoutPenalty(t *testing.T) {
	if got := FanoutPenalty(3, 5); got != 1.0 {
		t.Errorf("FanoutPenalty(3, 5) = %v, want 1.0", got)
	}
	if got := FanoutPenalty(5, 5); got != 1.0 {
		t.Errorf("FanoutPenalty(5, 5) = %v, want 1.0", got)
	}
	got := FanoutPenalty(10, 5)
	want := 1.0 / math.Log2(11)
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("FanoutPenalty(10, 5) = %v, want %v", got, want)
	}
}

func TestCompoundConfidence(t *testing.T) {
	// No edges: identity.
	if got := CompoundConfidence(nil, 0.85); got != 1.0 {
		t.Errorf("CompoundConfidence(nil) = %v, want 1.0", got)
	}

	// Two hops with perfect edge confidence.
	got := CompoundConfidence([]float64{1.0, 1.0}, 0.85)
	want := 0.85 * 0.85
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("CompoundConfidence([1,1]) = %v, want %v", got, want)
	}

	// Two hops with varying edge confidence.
	got = CompoundConfidence([]float64{0.9, 0.8}, 0.85)
	want = 0.9 * 0.85 * 0.8 * 0.85
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("CompoundConfidence([0.9,0.8]) = %v, want %v", got, want)
	}
}

func TestSortResults(t *testing.T) {
	results := []ReachResult{
		{NodeID: "b", Confidence: 0.5},
		{NodeID: "a", Confidence: 0.9},
		{NodeID: "c", Confidence: 0.5},
	}
	sortResults(results)

	if results[0].NodeID != "a" {
		t.Errorf("expected first result to be 'a', got %s", results[0].NodeID)
	}
	if results[1].NodeID != "b" {
		t.Errorf("expected second result to be 'b', got %s", results[1].NodeID)
	}
	if results[2].NodeID != "c" {
		t.Errorf("expected third result to be 'c', got %s", results[2].NodeID)
	}
}

func TestTopN(t *testing.T) {
	results := []ReachResult{
		{NodeID: "a", Confidence: 0.9},
		{NodeID: "b", Confidence: 0.5},
		{NodeID: "c", Confidence: 0.3},
	}

	top := TopN(results, 2)
	if len(top) != 2 {
		t.Fatalf("TopN(2) returned %d results, want 2", len(top))
	}

	// n >= len: return all.
	all := TopN(results, 10)
	if len(all) != 3 {
		t.Fatalf("TopN(10) returned %d results, want 3", len(all))
	}

	// n <= 0: return all.
	all = TopN(results, 0)
	if len(all) != 3 {
		t.Fatalf("TopN(0) returned %d results, want 3", len(all))
	}
}

func TestFilterByBand(t *testing.T) {
	results := []ReachResult{
		{NodeID: "a", Confidence: 0.9},
		{NodeID: "b", Confidence: 0.5},
		{NodeID: "c", Confidence: 0.3},
		{NodeID: "d", Confidence: 0.05},
	}

	high := FilterByBand(results, BandHigh)
	if len(high) != 1 || high[0].NodeID != "a" {
		t.Errorf("FilterByBand(high) = %v, want [a]", high)
	}

	medium := FilterByBand(results, BandMedium)
	if len(medium) != 1 || medium[0].NodeID != "b" {
		t.Errorf("FilterByBand(medium) = %v, want [b]", medium)
	}

	low := FilterByBand(results, BandLow)
	if len(low) != 2 {
		t.Errorf("FilterByBand(low) = %d results, want 2", len(low))
	}
}

func TestBandCounts(t *testing.T) {
	results := []ReachResult{
		{Confidence: 0.9},
		{Confidence: 0.8},
		{Confidence: 0.5},
		{Confidence: 0.1},
	}
	counts := BandCounts(results)
	if counts[BandHigh] != 2 {
		t.Errorf("BandCounts[high] = %d, want 2", counts[BandHigh])
	}
	if counts[BandMedium] != 1 {
		t.Errorf("BandCounts[medium] = %d, want 1", counts[BandMedium])
	}
	if counts[BandLow] != 1 {
		t.Errorf("BandCounts[low] = %d, want 1", counts[BandLow])
	}
}
