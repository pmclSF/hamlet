package depgraph

import (
	"fmt"
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

// buildBenchGraph creates a graph of the specified size for benchmarking.
func buildBenchGraph(testFiles, sourcesPerTest, testsPerFile int) *Graph {
	snap := &models.TestSuiteSnapshot{}

	for i := 0; i < testFiles; i++ {
		path := fmt.Sprintf("test/file_%d.test.js", i)
		tf := models.TestFile{
			Path:      path,
			Framework: "jest",
			TestCount: testsPerFile,
		}
		for j := 0; j < sourcesPerTest; j++ {
			tf.LinkedCodeUnits = append(tf.LinkedCodeUnits, fmt.Sprintf("src/mod_%d.js:fn_%d", j, j))
		}
		snap.TestFiles = append(snap.TestFiles, tf)

		for k := 0; k < testsPerFile; k++ {
			snap.TestCases = append(snap.TestCases, models.TestCase{
				TestID:   fmt.Sprintf("t:%s:%d:test_%d", path, k+1, k),
				TestName: fmt.Sprintf("test_%d", k),
				FilePath: path,
				Line:     k + 1,
			})
		}
	}

	for j := 0; j < sourcesPerTest; j++ {
		snap.CodeUnits = append(snap.CodeUnits, models.CodeUnit{
			UnitID:   fmt.Sprintf("src/mod_%d.js:fn_%d", j, j),
			Name:     fmt.Sprintf("fn_%d", j),
			Path:     fmt.Sprintf("src/mod_%d.js", j),
			Kind:     models.CodeUnitKindFunction,
			Exported: true,
		})
	}

	snap.Frameworks = []models.Framework{{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: testFiles}}

	return Build(snap)
}

func BenchmarkBuild_Small(b *testing.B) {
	snap := &models.TestSuiteSnapshot{
		Frameworks: []models.Framework{{Name: "jest", FileCount: 10}},
	}
	for i := 0; i < 10; i++ {
		snap.TestFiles = append(snap.TestFiles, models.TestFile{
			Path: fmt.Sprintf("test/%d.test.js", i), Framework: "jest", TestCount: 5,
		})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Build(snap)
	}
}

func BenchmarkBuild_Medium(b *testing.B) {
	snap := &models.TestSuiteSnapshot{
		Frameworks: []models.Framework{{Name: "jest", FileCount: 100}},
	}
	for i := 0; i < 100; i++ {
		snap.TestFiles = append(snap.TestFiles, models.TestFile{
			Path: fmt.Sprintf("test/%d.test.js", i), Framework: "jest", TestCount: 10,
		})
		for j := 0; j < 10; j++ {
			snap.TestCases = append(snap.TestCases, models.TestCase{
				TestID: fmt.Sprintf("t:%d:%d", i, j), FilePath: fmt.Sprintf("test/%d.test.js", i),
			})
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Build(snap)
	}
}

func BenchmarkNodesByType(b *testing.B) {
	g := buildBenchGraph(50, 10, 5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = g.NodesByType(NodeTestFile)
	}
}

func BenchmarkNodesByFamily(b *testing.B) {
	g := buildBenchGraph(50, 10, 5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = g.NodesByFamily(FamilyValidation)
	}
}

func BenchmarkNodes(b *testing.B) {
	g := buildBenchGraph(50, 10, 5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = g.Nodes()
	}
}

func BenchmarkNeighbors(b *testing.B) {
	g := buildBenchGraph(50, 10, 5)
	nodes := g.Nodes()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = g.Neighbors(nodes[i%len(nodes)].ID)
	}
}

func BenchmarkAnalyzeCoverage(b *testing.B) {
	g := buildBenchGraph(50, 10, 5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AnalyzeCoverage(g)
	}
}

func BenchmarkAnalyzeImpact(b *testing.B) {
	g := buildBenchGraph(50, 10, 5)
	changed := []string{"src/mod_0.js"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AnalyzeImpact(g, changed)
	}
}

func BenchmarkDetectDuplicates(b *testing.B) {
	g := buildBenchGraph(50, 10, 5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DetectDuplicates(g)
	}
}

func BenchmarkAnalyzeFanout(b *testing.B) {
	g := buildBenchGraph(50, 10, 5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AnalyzeFanout(g, DefaultFanoutThreshold)
	}
}

func BenchmarkStats(b *testing.B) {
	g := buildBenchGraph(50, 10, 5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Stats()
	}
}

func BenchmarkValidationTargets(b *testing.B) {
	g := buildBenchGraph(50, 10, 5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = g.ValidationTargets()
	}
}
