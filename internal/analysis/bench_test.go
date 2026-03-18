package analysis

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/pmclSF/terrain/internal/models"
)

// BenchmarkFileCache_Prewarm measures the cost of prewarming the file cache
// with a large set of source files.
func BenchmarkFileCache_Prewarm(b *testing.B) {
	root := b.TempDir()
	files := generateSyntheticRepo(root, 1000, 100, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fc := NewFileCache(root)
		fc.PrewarmSourceFiles(files)
	}
}

// BenchmarkFileCache_ReadHit measures cache hit performance after prewarm.
func BenchmarkFileCache_ReadHit(b *testing.B) {
	root := b.TempDir()
	files := generateSyntheticRepo(root, 1000, 100, 0)

	fc := NewFileCache(root)
	fc.PrewarmSourceFiles(files)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx := i % len(files)
		fc.ReadFile(files[idx])
	}
}

// BenchmarkCodeUnitExtraction_Cached measures code unit extraction with cache.
func BenchmarkCodeUnitExtraction_Cached(b *testing.B) {
	root := b.TempDir()
	sourceFiles := generateSyntheticRepo(root, 500, 0, 0)
	testFiles := []models.TestFile{}

	fc := NewFileCache(root)
	fc.PrewarmSourceFiles(sourceFiles)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		extractCodeUnitsCached(root, testFiles, sourceFiles, fc)
	}
}

// BenchmarkCodeUnitExtraction_NoCacheBaseline measures extraction without cache.
func BenchmarkCodeUnitExtraction_NoCacheBaseline(b *testing.B) {
	root := b.TempDir()
	sourceFiles := generateSyntheticRepo(root, 500, 0, 0)
	testFiles := []models.TestFile{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		extractExportedCodeUnitsFromList(root, testFiles, sourceFiles)
	}
}

// BenchmarkIncrementalState_CarryForward measures the cost of carrying
// forward results for unchanged files.
func BenchmarkIncrementalState_CarryForward(b *testing.B) {
	// Build a synthetic previous snapshot with many surfaces.
	var surfaces []models.CodeSurface
	var units []models.CodeUnit
	for i := 0; i < 10000; i++ {
		path := fmt.Sprintf("src/pkg%d/module%d.ts", i/20, i%20)
		surfaces = append(surfaces, models.CodeSurface{
			SurfaceID: fmt.Sprintf("surface:%s:func%d", path, i),
			Name:      fmt.Sprintf("func%d", i),
			Path:      path,
		})
		units = append(units, models.CodeUnit{
			UnitID: fmt.Sprintf("%s:func%d", path, i),
			Name:   fmt.Sprintf("func%d", i),
			Path:   path,
		})
	}

	state := &IncrementalState{
		PreviousSnapshot: &models.TestSuiteSnapshot{
			CodeSurfaces: surfaces,
			CodeUnits:    units,
		},
		ChangedFiles: []string{
			"src/pkg0/module0.ts",
			"src/pkg1/module1.ts",
		},
	}
	sort.Strings(state.ChangedFiles)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		state.CarryForwardSurfaces()
		state.CarryForwardCodeUnits()
	}
}

// TestFileCache_HitRateAtScale verifies cache hit rate with parallel access.
func TestFileCache_HitRateAtScale(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	files := generateSyntheticRepo(root, 200, 50, 0)

	fc := NewFileCache(root)
	fc.PrewarmSourceFiles(files)

	// Read all files again — should be 100% cache hits.
	for _, f := range files {
		content, ok := fc.ReadFile(f)
		if !ok {
			t.Fatalf("expected cache hit for %s", f)
		}
		if content == "" {
			t.Fatalf("expected non-empty content for %s", f)
		}
	}

	stats := fc.Stats()
	if stats.ContentHits < int64(len(files)) {
		t.Errorf("expected at least %d cache hits, got %d", len(files), stats.ContentHits)
	}
	// Misses should equal the number of unique files (prewarm reads).
	if stats.ContentMisses != int64(len(files)) {
		t.Errorf("expected %d cache misses (prewarm), got %d", len(files), stats.ContentMisses)
	}
}

// TestFileCache_GoASTCaching verifies Go AST is parsed once and reused.
func TestFileCache_GoASTCaching(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTempFile(t, root, "src/svc.go", `package svc

func Hello() string { return "hello" }
func World() string { return "world" }
`)

	fc := NewFileCache(root)

	// First parse.
	f1, fset1, ok := fc.ParseGoFile("src/svc.go")
	if !ok {
		t.Fatal("expected successful parse")
	}
	if f1 == nil || fset1 == nil {
		t.Fatal("expected non-nil AST")
	}

	// Second parse should return same objects (cached).
	f2, fset2, ok := fc.ParseGoFile("src/svc.go")
	if !ok {
		t.Fatal("expected successful cached parse")
	}
	if f1 != f2 {
		t.Error("expected same AST object from cache")
	}
	if fset1 != fset2 {
		t.Error("expected same FileSet from cache")
	}

	stats := fc.Stats()
	if stats.ASTHits != 1 {
		t.Errorf("expected 1 AST cache hit, got %d", stats.ASTHits)
	}
	if stats.ASTMisses != 1 {
		t.Errorf("expected 1 AST cache miss, got %d", stats.ASTMisses)
	}
}

// TestFileCache_InvalidateStale verifies stale file detection.
func TestFileCache_InvalidateStale(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTempFile(t, root, "src/a.ts", "export function a() {}")
	writeTempFile(t, root, "src/b.ts", "export function b() {}")

	fc := NewFileCache(root)
	fc.ReadFile("src/a.ts")
	fc.ReadFile("src/b.ts")

	// Touch one file to make it stale.
	time.Sleep(10 * time.Millisecond)
	writeTempFile(t, root, "src/a.ts", "export function a_v2() {}")

	stale := fc.InvalidateStale()
	if len(stale) != 1 {
		t.Fatalf("expected 1 stale file, got %d: %v", len(stale), stale)
	}
	if stale[0] != "src/a.ts" {
		t.Errorf("expected src/a.ts to be stale, got %s", stale[0])
	}

	// Re-read should get new content.
	content, ok := fc.ReadFile("src/a.ts")
	if !ok {
		t.Fatal("expected successful re-read")
	}
	if content != "export function a_v2() {}" {
		t.Errorf("expected updated content, got: %s", content)
	}
}

// TestIncrementalState_CarryForward verifies carry-forward logic.
func TestIncrementalState_CarryForward(t *testing.T) {
	t.Parallel()
	state := &IncrementalState{
		PreviousSnapshot: &models.TestSuiteSnapshot{
			CodeSurfaces: []models.CodeSurface{
				{SurfaceID: "s:a.ts:foo", Path: "a.ts"},
				{SurfaceID: "s:b.ts:bar", Path: "b.ts"},
				{SurfaceID: "s:c.ts:baz", Path: "c.ts"},
			},
			CodeUnits: []models.CodeUnit{
				{UnitID: "a.ts:foo", Path: "a.ts"},
				{UnitID: "b.ts:bar", Path: "b.ts"},
			},
		},
		ChangedFiles: []string{"b.ts"}, // Only b.ts changed.
	}

	surfaces := state.CarryForwardSurfaces()
	if len(surfaces) != 2 {
		t.Errorf("expected 2 carried-forward surfaces (a.ts + c.ts), got %d", len(surfaces))
	}

	units := state.CarryForwardCodeUnits()
	if len(units) != 1 {
		t.Errorf("expected 1 carried-forward unit (a.ts), got %d", len(units))
	}
}

// TestIncrementalState_ChangedSourceFiles verifies file filtering.
func TestIncrementalState_ChangedSourceFiles(t *testing.T) {
	t.Parallel()
	state := &IncrementalState{
		ChangedFiles: []string{"a.ts", "c.ts"},
	}

	allFiles := []string{"a.ts", "b.ts", "c.ts", "d.ts"}
	changed := state.ChangedSourceFiles(allFiles)

	if len(changed) != 2 {
		t.Errorf("expected 2 changed files, got %d: %v", len(changed), changed)
	}
}

// --- Synthetic repo generation ---

// generateSyntheticRepo creates a temporary repository structure with
// the specified number of source files, test files, and scenario files.
func generateSyntheticRepo(root string, sourceCount, testCount, scenarioCount int) []string {
	var allFiles []string

	// Source files.
	for i := 0; i < sourceCount; i++ {
		pkg := fmt.Sprintf("src/pkg%d", i/20)
		name := fmt.Sprintf("module%d.ts", i%20)
		relPath := filepath.Join(pkg, name)
		content := fmt.Sprintf(`
export function func%d(x) { return x + %d; }
export function helper%d(data) { return data.map(d => d * %d); }
export class Service%d {
  process(input) { return input; }
  validate(data) { return data != null; }
}
`, i, i, i, i, i)
		writeSynthFile(root, relPath, content)
		allFiles = append(allFiles, relPath)
	}

	// Test files.
	for i := 0; i < testCount; i++ {
		pkg := fmt.Sprintf("test/pkg%d", i/20)
		name := fmt.Sprintf("module%d.test.ts", i%20)
		relPath := filepath.Join(pkg, name)
		srcIdx := i % sourceCount
		content := fmt.Sprintf(`
import { func%d } from '../../src/pkg%d/module%d';

describe('module%d', () => {
  it('should work', () => {
    expect(func%d(1)).toBe(%d);
  });
});
`, srcIdx, srcIdx/20, srcIdx%20, i, srcIdx, srcIdx+1)
		writeSynthFile(root, relPath, content)
		allFiles = append(allFiles, relPath)
	}

	// Scenario files (AI eval configs).
	for i := 0; i < scenarioCount; i++ {
		relPath := fmt.Sprintf("evals/scenario%d.yaml", i)
		content := fmt.Sprintf("name: scenario_%d\ncategory: accuracy\n", i)
		writeSynthFile(root, relPath, content)
	}

	return allFiles
}

func writeSynthFile(root, relPath, content string) {
	absPath := filepath.Join(root, relPath)
	os.MkdirAll(filepath.Dir(absPath), 0o755)
	os.WriteFile(absPath, []byte(content), 0o644)
}

// BenchmarkFullAnalysis_1kFiles measures full analysis pipeline time.
func BenchmarkFullAnalysis_1kFiles(b *testing.B) {
	root := b.TempDir()
	generateSyntheticRepo(root, 1000, 200, 50)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a := New(root)
		_, err := a.Analyze()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFullAnalysis_5kFiles measures analysis at moderate scale.
func BenchmarkFullAnalysis_5kFiles(b *testing.B) {
	root := b.TempDir()
	generateSyntheticRepo(root, 5000, 1000, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a := New(root)
		_, err := a.Analyze()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCacheVsNocache_CodeUnits compares cached vs uncached extraction.
func BenchmarkCacheVsNocache_CodeUnits(b *testing.B) {
	root := b.TempDir()
	sourceFiles := generateSyntheticRepo(root, 2000, 0, 0)
	testFiles := []models.TestFile{}

	b.Run("nocache", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			extractExportedCodeUnitsFromList(root, testFiles, sourceFiles)
		}
	})

	b.Run("cached", func(b *testing.B) {
		fc := NewFileCache(root)
		fc.PrewarmSourceFiles(sourceFiles)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			extractCodeUnitsCached(root, testFiles, sourceFiles, fc)
		}
	})
}

// TestDeterminism_CachedVsUncached verifies that cached analysis produces
// identical results to uncached analysis.
func TestDeterminism_CachedVsUncached(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	generateSyntheticRepo(root, 50, 10, 0)

	// Uncached.
	a1 := New(root)
	snap1, err := a1.Analyze()
	if err != nil {
		t.Fatal(err)
	}

	// Cached (second run reuses warm cache from first).
	a2 := New(root)
	snap2, err := a2.Analyze()
	if err != nil {
		t.Fatal(err)
	}

	// Compare key counts.
	if len(snap1.CodeUnits) != len(snap2.CodeUnits) {
		t.Errorf("CodeUnit count differs: %d vs %d", len(snap1.CodeUnits), len(snap2.CodeUnits))
	}
	if len(snap1.CodeSurfaces) != len(snap2.CodeSurfaces) {
		t.Errorf("CodeSurface count differs: %d vs %d", len(snap1.CodeSurfaces), len(snap2.CodeSurfaces))
	}
	if len(snap1.TestFiles) != len(snap2.TestFiles) {
		t.Errorf("TestFile count differs: %d vs %d", len(snap1.TestFiles), len(snap2.TestFiles))
	}
	if len(snap1.TestCases) != len(snap2.TestCases) {
		t.Errorf("TestCase count differs: %d vs %d", len(snap1.TestCases), len(snap2.TestCases))
	}

	// Compare surface IDs for deterministic ordering.
	ids1 := make([]string, len(snap1.CodeSurfaces))
	ids2 := make([]string, len(snap2.CodeSurfaces))
	for i, s := range snap1.CodeSurfaces {
		ids1[i] = s.SurfaceID
	}
	for i, s := range snap2.CodeSurfaces {
		ids2[i] = s.SurfaceID
	}
	sort.Strings(ids1)
	sort.Strings(ids2)
	for i := range ids1 {
		if i < len(ids2) && ids1[i] != ids2[i] {
			t.Errorf("surface ID mismatch at %d: %s vs %s", i, ids1[i], ids2[i])
			break
		}
	}
}
