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

// --- Large-scale benchmarks ---

func buildLargeSnapshot(testFiles, sourcesPerTest, testsPerFile, scenarios int) *models.TestSuiteSnapshot {
	snap := &models.TestSuiteSnapshot{}

	// Build import graph.
	snap.ImportGraph = make(map[string]map[string]bool)

	for i := 0; i < testFiles; i++ {
		path := fmt.Sprintf("test/file_%d.test.js", i)
		tf := models.TestFile{Path: path, Framework: "jest", TestCount: testsPerFile}
		snap.TestFiles = append(snap.TestFiles, tf)

		imports := make(map[string]bool)
		for j := 0; j < sourcesPerTest; j++ {
			srcIdx := (i*sourcesPerTest + j) % (testFiles * 2)
			imports[fmt.Sprintf("src/mod_%d.js", srcIdx)] = true
		}
		snap.ImportGraph[path] = imports

		for k := 0; k < testsPerFile; k++ {
			snap.TestCases = append(snap.TestCases, models.TestCase{
				TestID: fmt.Sprintf("t:%s:%d", path, k), TestName: fmt.Sprintf("test_%d", k),
				FilePath: path, Line: k + 1,
			})
		}
	}

	for j := 0; j < testFiles*2; j++ {
		snap.CodeUnits = append(snap.CodeUnits, models.CodeUnit{
			UnitID: fmt.Sprintf("src/mod_%d.js:fn", j), Name: "fn",
			Path: fmt.Sprintf("src/mod_%d.js", j), Exported: true,
		})
		snap.CodeSurfaces = append(snap.CodeSurfaces, models.CodeSurface{
			SurfaceID: fmt.Sprintf("surface:src/mod_%d.js:fn", j),
			Name: "fn", Path: fmt.Sprintf("src/mod_%d.js", j),
			Kind: models.SurfaceFunction, Exported: true,
		})
	}

	for i := 0; i < scenarios; i++ {
		sc := models.Scenario{
			ScenarioID: fmt.Sprintf("scenario:%d", i), Name: fmt.Sprintf("scenario_%d", i),
			Category: "accuracy", Capability: fmt.Sprintf("cap_%d", i%10),
		}
		for j := 0; j < 3; j++ {
			idx := (i*3 + j) % len(snap.CodeSurfaces)
			sc.CoveredSurfaceIDs = append(sc.CoveredSurfaceIDs, snap.CodeSurfaces[idx].SurfaceID)
		}
		snap.Scenarios = append(snap.Scenarios, sc)
	}

	snap.Frameworks = []models.Framework{{Name: "jest", FileCount: testFiles}}
	return snap
}

// BenchmarkBuild_Large exercises graph construction at 1K test files,
// 5 imports each, 10 tests per file = 10K tests, 2K source files.
func BenchmarkBuild_Large(b *testing.B) {
	snap := buildLargeSnapshot(1000, 5, 10, 100)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Build(snap)
	}
}

// BenchmarkBuild_XLarge exercises 10K test files, 50K tests, 500 scenarios.
func BenchmarkBuild_XLarge(b *testing.B) {
	snap := buildLargeSnapshot(10000, 3, 5, 500)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Build(snap)
	}
}

func BenchmarkAnalyzeImpact_Large(b *testing.B) {
	snap := buildLargeSnapshot(1000, 5, 10, 100)
	g := Build(snap)
	changed := []string{"src/mod_0.js", "src/mod_1.js", "src/mod_2.js"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AnalyzeImpact(g, changed)
	}
}

func BenchmarkDetectDuplicates_Large(b *testing.B) {
	snap := buildLargeSnapshot(1000, 5, 10, 0)
	g := Build(snap)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DetectDuplicates(g)
	}
}

func BenchmarkAnalyzeFanout_Large(b *testing.B) {
	snap := buildLargeSnapshot(1000, 5, 10, 0)
	g := Build(snap)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AnalyzeFanout(g, DefaultFanoutThreshold)
	}
}
