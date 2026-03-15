package testdata

import (
	"testing"

	"github.com/pmclSF/terrain/internal/comparison"
	"github.com/pmclSF/terrain/internal/heatmap"
	"github.com/pmclSF/terrain/internal/impact"
	"github.com/pmclSF/terrain/internal/measurement"
	"github.com/pmclSF/terrain/internal/metrics"
	"github.com/pmclSF/terrain/internal/portfolio"
	"github.com/pmclSF/terrain/internal/scoring"
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

func BenchmarkComparison(b *testing.B) {
	from := FlakyConcentratedSnapshot()
	to := HealthyBalancedSnapshot()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		comparison.Compare(from, to)
	}
}

func BenchmarkPortfolio(b *testing.B) {
	snap := FlakyConcentratedSnapshot()
	snap.Risk = scoring.ComputeRisk(snap)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		portfolio.Analyze(snap)
	}
}

func BenchmarkFullPipeline_VeryLarge(b *testing.B) {
	snap := VeryLargeSnapshot()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		snap.Risk = scoring.ComputeRisk(snap)
		reg := measurement.DefaultRegistry()
		reg.ComputeSnapshot(snap)
		metrics.Derive(snap)
		heatmap.Build(snap)
	}
}

func BenchmarkImpact_LargeScope(b *testing.B) {
	snap := LargeScaleSnapshot()
	var files []impact.ChangedFile
	for i := 0; i < 50; i++ {
		files = append(files, impact.ChangedFile{
			Path:       "src/auth/module" + string(rune('0'+i%10)) + ".js",
			ChangeKind: impact.ChangeModified,
		})
	}
	scope := &impact.ChangeScope{ChangedFiles: files}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		impact.Analyze(scope, snap)
	}
}
