package engine

import (
	"testing"

	"github.com/pmclSF/terrain/internal/testdata"
)

func BenchmarkSignalDetection(b *testing.B) {
	base := testdata.LargeScaleSnapshot()
	registry := DefaultRegistry(Config{RepoRoot: "."})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		snap := *base
		snap.Signals = nil
		registry.Run(&snap)
	}
}
