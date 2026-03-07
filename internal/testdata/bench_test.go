package testdata

import (
	"testing"

	"github.com/pmclSF/hamlet/internal/heatmap"
	"github.com/pmclSF/hamlet/internal/impact"
	"github.com/pmclSF/hamlet/internal/measurement"
	"github.com/pmclSF/hamlet/internal/metrics"
	"github.com/pmclSF/hamlet/internal/scoring"
)

func BenchmarkMetrics_Minimal(b *testing.B) {
	snap := MinimalSnapshot()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.Derive(snap)
	}
}

func BenchmarkMetrics_LargeScale(b *testing.B) {
	snap := LargeScaleSnapshot()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.Derive(snap)
	}
}

func BenchmarkMeasurements_Minimal(b *testing.B) {
	snap := MinimalSnapshot()
	reg := measurement.DefaultRegistry()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reg.ComputeSnapshot(snap)
	}
}

func BenchmarkMeasurements_LargeScale(b *testing.B) {
	snap := LargeScaleSnapshot()
	reg := measurement.DefaultRegistry()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reg.ComputeSnapshot(snap)
	}
}

func BenchmarkHeatmap_LargeScale(b *testing.B) {
	snap := LargeScaleSnapshot()
	snap.Risk = scoring.ComputeRisk(snap)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		heatmap.Build(snap)
	}
}

func BenchmarkRiskScoring_LargeScale(b *testing.B) {
	snap := LargeScaleSnapshot()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scoring.ComputeRisk(snap)
	}
}

func BenchmarkImpactAnalysis(b *testing.B) {
	snap := HealthyBalancedSnapshot()
	scope := &impact.ChangeScope{
		ChangedFiles: []impact.ChangedFile{
			{Path: "src/auth.js", ChangeKind: impact.ChangeModified},
			{Path: "src/user.js", ChangeKind: impact.ChangeModified},
			{Path: "src/payment.js", ChangeKind: impact.ChangeModified},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		impact.Analyze(scope, snap)
	}
}
