package reasoning

import "testing"

func TestAggregateStability_NoSignals(t *testing.T) {
	s := AggregateStability("test:1", nil, DefaultStabilityConfig())
	if s.Band != StabilityUnknown {
		t.Errorf("expected unknown band with no signals, got %v", s.Band)
	}
	if s.Score != 0.5 {
		t.Errorf("expected 0.5 score with no signals, got %v", s.Score)
	}
}

func TestAggregateStability_Stable(t *testing.T) {
	signals := []StabilitySignal{
		{NodeID: "test:1", SignalType: SignalFailureRate, Value: 0.01},
		{NodeID: "test:1", SignalType: SignalFlakyRate, Value: 0.02},
	}
	s := AggregateStability("test:1", signals, DefaultStabilityConfig())
	if s.Band != StabilityStable {
		t.Errorf("expected stable, got %v", s.Band)
	}
	if s.Score < 0.8 {
		t.Errorf("stable score should be >= 0.8, got %v", s.Score)
	}
}

func TestAggregateStability_Unstable(t *testing.T) {
	signals := []StabilitySignal{
		{NodeID: "test:1", SignalType: SignalFailureRate, Value: 0.25},
	}
	s := AggregateStability("test:1", signals, DefaultStabilityConfig())
	if s.Band != StabilityUnstable {
		t.Errorf("expected unstable, got %v", s.Band)
	}
}

func TestAggregateStability_Critical(t *testing.T) {
	signals := []StabilitySignal{
		{NodeID: "test:1", SignalType: SignalFailureRate, Value: 0.60},
	}
	s := AggregateStability("test:1", signals, DefaultStabilityConfig())
	if s.Band != StabilityCritical {
		t.Errorf("expected critical, got %v", s.Band)
	}
	if s.Score > 0.2 {
		t.Errorf("critical score should be <= 0.2, got %v", s.Score)
	}
}

func TestAggregateStability_FlakyTriggersUnstable(t *testing.T) {
	signals := []StabilitySignal{
		{NodeID: "test:1", SignalType: SignalFailureRate, Value: 0.01},
		{NodeID: "test:1", SignalType: SignalFlakyRate, Value: 0.10},
	}
	s := AggregateStability("test:1", signals, DefaultStabilityConfig())
	if s.Band != StabilityUnstable {
		t.Errorf("expected unstable from flaky rate, got %v", s.Band)
	}
}

func TestStabilityAdjustedConfidence(t *testing.T) {
	tests := []struct {
		band ConfidenceBand
		stab StabilityBand
		want float64
	}{
		{BandHigh, StabilityStable, 0.9},
		{BandHigh, StabilityUnknown, 0.9 * 0.9},
		{BandHigh, StabilityUnstable, 0.9 * 0.75},
		{BandHigh, StabilityCritical, 0.9 * 0.5},
	}

	for _, tt := range tests {
		summary := StabilitySummary{Band: tt.stab}
		got := StabilityAdjustedConfidence(0.9, summary)
		if got < tt.want-0.001 || got > tt.want+0.001 {
			t.Errorf("StabilityAdjustedConfidence(0.9, %v) = %v, want %v", tt.stab, got, tt.want)
		}
	}
}
