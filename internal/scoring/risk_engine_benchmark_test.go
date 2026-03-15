package scoring

import (
	"testing"

	"github.com/pmclSF/terrain/internal/testdata"
)

func BenchmarkRiskScore(b *testing.B) {
	snap := testdata.LargeScaleSnapshot()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ComputeRisk(snap)
	}
}
