package reasoning

// StabilitySignal represents a single stability observation for a test or node.
// Stability signals come from execution history (flaky tests, consistent failures,
// skip records) and are aggregated to produce a stability band.
type StabilitySignal struct {
	// NodeID is the node this signal applies to.
	NodeID string

	// SignalType classifies the observation.
	SignalType StabilitySignalType

	// Value is a numeric metric (e.g., failure rate, skip count).
	Value float64

	// Source describes where this signal came from (e.g., "ci-run-123").
	Source string
}

// StabilitySignalType classifies stability observations.
type StabilitySignalType string

const (
	SignalFailureRate   StabilitySignalType = "failure_rate"
	SignalFlakyRate     StabilitySignalType = "flaky_rate"
	SignalSkipCount     StabilitySignalType = "skip_count"
	SignalConsecutiveFail StabilitySignalType = "consecutive_fail"
	SignalMeanDuration  StabilitySignalType = "mean_duration"
)

// StabilityBand classifies overall stability.
type StabilityBand string

const (
	StabilityStable   StabilityBand = "stable"
	StabilityUnstable StabilityBand = "unstable"
	StabilityCritical StabilityBand = "critical"
	StabilityUnknown  StabilityBand = "unknown"
)

// StabilitySummary is the aggregated stability assessment for a node.
type StabilitySummary struct {
	NodeID  string
	Band    StabilityBand
	Signals []StabilitySignal
	Score   float64 // 0 = worst, 1 = best
}

// StabilityConfig controls stability aggregation thresholds.
type StabilityConfig struct {
	// FlakyThreshold: flaky rate above this → unstable. Default: 0.05.
	FlakyThreshold float64

	// FailureThreshold: failure rate above this → critical. Default: 0.20.
	FailureThreshold float64

	// CriticalFailureThreshold: failure rate above this → critical. Default: 0.50.
	CriticalFailureThreshold float64
}

// DefaultStabilityConfig returns standard stability thresholds.
func DefaultStabilityConfig() StabilityConfig {
	return StabilityConfig{
		FlakyThreshold:           0.05,
		FailureThreshold:         0.20,
		CriticalFailureThreshold: 0.50,
	}
}

// AggregateStability computes a stability summary from a set of signals.
//
// The stability score is derived from the worst signal observed:
//   - failure_rate > CriticalFailureThreshold → critical (score 0.0–0.2)
//   - failure_rate > FailureThreshold → unstable (score 0.2–0.5)
//   - flaky_rate > FlakyThreshold → unstable (score 0.3–0.6)
//   - otherwise → stable (score 0.8–1.0)
func AggregateStability(nodeID string, signals []StabilitySignal, cfg StabilityConfig) StabilitySummary {
	if len(signals) == 0 {
		return StabilitySummary{
			NodeID: nodeID,
			Band:   StabilityUnknown,
			Score:  0.5, // neutral when no data
		}
	}

	if cfg.FlakyThreshold <= 0 {
		cfg.FlakyThreshold = 0.05
	}
	if cfg.FailureThreshold <= 0 {
		cfg.FailureThreshold = 0.20
	}
	if cfg.CriticalFailureThreshold <= 0 {
		cfg.CriticalFailureThreshold = 0.50
	}

	var failureRate, flakyRate float64
	for _, s := range signals {
		switch s.SignalType {
		case SignalFailureRate:
			if s.Value > failureRate {
				failureRate = s.Value
			}
		case SignalFlakyRate:
			if s.Value > flakyRate {
				flakyRate = s.Value
			}
		}
	}

	band := StabilityStable
	score := 1.0

	if failureRate > cfg.CriticalFailureThreshold {
		band = StabilityCritical
		score = 0.1
	} else if failureRate > cfg.FailureThreshold {
		band = StabilityUnstable
		score = 0.4
	} else if flakyRate > cfg.FlakyThreshold {
		band = StabilityUnstable
		score = 0.5
	} else {
		// Stable — score degrades slightly with any failure/flakiness.
		score = 1.0 - failureRate - flakyRate
		if score < 0.8 {
			score = 0.8
		}
	}

	return StabilitySummary{
		NodeID:  nodeID,
		Band:    band,
		Signals: signals,
		Score:   score,
	}
}

// StabilityAdjustedConfidence adjusts a traversal confidence by the
// stability score of the target node. Unstable nodes receive a penalty.
func StabilityAdjustedConfidence(confidence float64, stability StabilitySummary) float64 {
	// Stable nodes: no adjustment. Unknown: slight penalty. Unstable/Critical: larger penalty.
	switch stability.Band {
	case StabilityCritical:
		return confidence * 0.5
	case StabilityUnstable:
		return confidence * 0.75
	case StabilityUnknown:
		return confidence * 0.9
	default:
		return confidence
	}
}
