package testdata

import (
	"encoding/json"
	"testing"

	"github.com/pmclSF/terrain/internal/comparison"
	"github.com/pmclSF/terrain/internal/heatmap"
	"github.com/pmclSF/terrain/internal/impact"
	"github.com/pmclSF/terrain/internal/measurement"
	"github.com/pmclSF/terrain/internal/metrics"
	"github.com/pmclSF/terrain/internal/portfolio"
	"github.com/pmclSF/terrain/internal/scoring"
)

// TestDeterminism_MetricsIdentical verifies that metrics.Derive produces
// identical output across multiple runs with the same input.
func TestDeterminism_MetricsIdentical(t *testing.T) {
	t.Parallel()
	snap := HealthyBalancedSnapshot()

	results := make([]string, 10)
	for i := 0; i < 10; i++ {
		ms := metrics.Derive(snap)
		ms.GeneratedAt = FixedTime // normalize time
		data, _ := json.Marshal(ms)
		results[i] = string(data)
	}

	for i := 1; i < 10; i++ {
		if results[i] != results[0] {
			t.Errorf("metrics run %d differs from run 0", i)
		}
	}
}

// TestDeterminism_MeasurementsIdentical verifies measurement computation
// produces identical posture bands across multiple runs.
func TestDeterminism_MeasurementsIdentical(t *testing.T) {
	t.Parallel()
	snap := HealthyBalancedSnapshot()

	results := make([]string, 10)
	for i := 0; i < 10; i++ {
		reg, mErr := measurement.DefaultRegistry(); if mErr != nil { t.Fatal(mErr) }
		ms := reg.ComputeSnapshot(snap)
		model := ms.ToModel()
		data, _ := json.Marshal(model)
		results[i] = string(data)
	}

	for i := 1; i < 10; i++ {
		if results[i] != results[0] {
			t.Errorf("measurement run %d differs from run 0", i)
		}
	}
}

// TestDeterminism_HeatmapIdentical verifies heatmap computation is deterministic.
func TestDeterminism_HeatmapIdentical(t *testing.T) {
	t.Parallel()
	snap := HealthyBalancedSnapshot()
	snap.Risk = scoring.ComputeRisk(snap)

	results := make([]string, 10)
	for i := 0; i < 10; i++ {
		h := heatmap.Build(snap)
		data, _ := json.Marshal(h)
		results[i] = string(data)
	}

	for i := 1; i < 10; i++ {
		if results[i] != results[0] {
			t.Errorf("heatmap run %d differs from run 0", i)
		}
	}
}

// TestDeterminism_RiskScoringIdentical verifies risk scoring is deterministic.
func TestDeterminism_RiskScoringIdentical(t *testing.T) {
	t.Parallel()
	snap := HealthyBalancedSnapshot()

	results := make([]string, 10)
	for i := 0; i < 10; i++ {
		risks := scoring.ComputeRisk(snap)
		data, _ := json.Marshal(risks)
		results[i] = string(data)
	}

	for i := 1; i < 10; i++ {
		if results[i] != results[0] {
			t.Errorf("risk run %d differs from run 0", i)
		}
	}
}

// TestDeterminism_LargeScaleStable verifies determinism at scale.
func TestDeterminism_LargeScaleStable(t *testing.T) {
	t.Parallel()
	snap := LargeScaleSnapshot()

	results := make([]string, 5)
	for i := 0; i < 5; i++ {
		reg, mErr := measurement.DefaultRegistry(); if mErr != nil { t.Fatal(mErr) }
		ms := reg.ComputeSnapshot(snap)
		model := ms.ToModel()
		data, _ := json.Marshal(model)
		results[i] = string(data)
	}

	for i := 1; i < 5; i++ {
		if results[i] != results[0] {
			t.Errorf("large-scale run %d differs from run 0", i)
		}
	}
}

// TestDeterminism_ImpactIdentical verifies impact analysis is deterministic.
func TestDeterminism_ImpactIdentical(t *testing.T) {
	t.Parallel()
	snap := HealthyBalancedSnapshot()
	scope := impact.ChangeScopeFromPaths(
		[]string{"src/auth.js", "src/payment.js", "src/__tests__/auth.test.js"},
		impact.ChangeModified,
	)

	results := make([]string, 5)
	for i := 0; i < 5; i++ {
		result := impact.Analyze(scope, snap)
		data, _ := json.Marshal(result)
		results[i] = string(data)
	}

	for i := 1; i < 5; i++ {
		if results[i] != results[0] {
			t.Errorf("impact run %d differs from run 0", i)
		}
	}
}

// TestDeterminism_ComparisonIdentical verifies comparison is deterministic.
func TestDeterminism_ComparisonIdentical(t *testing.T) {
	t.Parallel()
	from := FlakyConcentratedSnapshot()
	to := HealthyBalancedSnapshot()

	results := make([]string, 5)
	for i := 0; i < 5; i++ {
		comp := comparison.Compare(from, to)
		data, _ := json.Marshal(comp)
		results[i] = string(data)
	}

	for i := 1; i < 5; i++ {
		if results[i] != results[0] {
			t.Errorf("comparison run %d differs from run 0", i)
		}
	}
}

// TestDeterminism_PortfolioIdentical verifies portfolio analysis is deterministic.
func TestDeterminism_PortfolioIdentical(t *testing.T) {
	t.Parallel()
	snap := FlakyConcentratedSnapshot()
	snap.Risk = scoring.ComputeRisk(snap)

	results := make([]string, 5)
	for i := 0; i < 5; i++ {
		ps := portfolio.Analyze(snap)
		data, _ := json.Marshal(ps.ToModel())
		results[i] = string(data)
	}

	for i := 1; i < 5; i++ {
		if results[i] != results[0] {
			t.Errorf("portfolio run %d differs from run 0", i)
		}
	}
}

// TestDeterminism_ImpactAggregateIdentical verifies impact aggregate is deterministic.
func TestDeterminism_ImpactAggregateIdentical(t *testing.T) {
	t.Parallel()
	snap := HealthyBalancedSnapshot()
	scope := impact.ChangeScopeFromPaths(
		[]string{"src/auth.js", "src/user.js"},
		impact.ChangeModified,
	)

	results := make([]string, 5)
	for i := 0; i < 5; i++ {
		result := impact.Analyze(scope, snap)
		agg := impact.BuildAggregate(result)
		data, _ := json.Marshal(agg)
		results[i] = string(data)
	}

	for i := 1; i < 5; i++ {
		if results[i] != results[0] {
			t.Errorf("impact aggregate run %d differs from run 0", i)
		}
	}
}
