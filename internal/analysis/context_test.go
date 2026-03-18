package analysis

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

// TestParallelForEachIndexCtx_CancelledBeforeStart verifies that a
// pre-cancelled context causes no work items to execute.
func TestParallelForEachIndexCtx_CancelledBeforeStart(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	var count int64
	parallelForEachIndexCtx(ctx, 100, func(i int) {
		atomic.AddInt64(&count, 1)
	})

	if count != 0 {
		t.Errorf("expected 0 items processed with pre-cancelled context, got %d", count)
	}
}

// TestParallelForEachIndexCtx_CancelMidway verifies that cancellation
// during processing stops further work items from being dispatched.
func TestParallelForEachIndexCtx_CancelMidway(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())

	var count int64
	parallelForEachIndexCtx(ctx, 1000, func(i int) {
		n := atomic.AddInt64(&count, 1)
		if n == 10 {
			cancel()
		}
	})

	processed := atomic.LoadInt64(&count)
	// With cancellation after 10 items, we should process far fewer than 1000.
	// Allow some slack for in-flight items.
	if processed >= 100 {
		t.Errorf("expected significantly fewer than 1000 items, got %d", processed)
	}
	if processed < 10 {
		t.Errorf("expected at least 10 items (cancel fires at 10), got %d", processed)
	}
}

// TestParallelForEachIndexCtx_CompletesNormally verifies normal execution
// when context is not cancelled.
func TestParallelForEachIndexCtx_CompletesNormally(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	var count int64
	parallelForEachIndexCtx(ctx, 50, func(i int) {
		atomic.AddInt64(&count, 1)
	})

	if count != 50 {
		t.Errorf("expected 50 items processed, got %d", count)
	}
}

// TestWalkDirCtx_CancelledBeforeStart verifies that walkDirCtx returns
// immediately with a pre-cancelled context.
func TestWalkDirCtx_CancelledBeforeStart(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	// Create some files.
	for i := 0; i < 10; i++ {
		writeTempFile(t, root, fmt.Sprintf("dir%d/file.ts", i), "content")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var count int
	err := walkDirCtx(ctx, root, func(relPath string, isDir bool) bool {
		if !isDir {
			count++
		}
		return false
	})

	if err == nil {
		t.Error("expected context error from walkDirCtx")
	}
	if count != 0 {
		t.Errorf("expected 0 files visited, got %d", count)
	}
}

// TestWalkDirCtx_CompletesNormally verifies walkDirCtx works without cancellation.
func TestWalkDirCtx_CompletesNormally(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for i := 0; i < 5; i++ {
		writeTempFile(t, root, fmt.Sprintf("file%d.ts", i), "content")
	}

	var count int
	err := walkDirCtx(context.Background(), root, func(relPath string, isDir bool) bool {
		if !isDir {
			count++
		}
		return false
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if count != 5 {
		t.Errorf("expected 5 files, got %d", count)
	}
}

// TestCollectSourceFilesCtx_CancelledReturnsError verifies that
// collectSourceFilesCtx returns an error when cancelled.
func TestCollectSourceFilesCtx_CancelledReturnsError(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for i := 0; i < 10; i++ {
		writeTempFile(t, root, fmt.Sprintf("pkg%d/file.ts", i), "export function f() {}")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := collectSourceFilesCtx(ctx, root)
	if err == nil {
		t.Error("expected error from cancelled context")
	}
}

// TestCollectSourceFilesCtx_CompletesNormally verifies it works without cancellation.
func TestCollectSourceFilesCtx_CompletesNormally(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for i := 0; i < 5; i++ {
		writeTempFile(t, root, fmt.Sprintf("src/mod%d.ts", i), "export function f() {}")
	}

	files, err := collectSourceFilesCtx(context.Background(), root)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(files) != 5 {
		t.Errorf("expected 5 files, got %d", len(files))
	}
}

// TestAnalyzeContext_CancelledReturnsError verifies that AnalyzeContext
// returns a context error when cancelled before analysis starts.
func TestAnalyzeContext_CancelledReturnsError(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTempFile(t, root, "src/app.ts", "export function hello() {}")
	writeTempFile(t, root, "test/app.test.ts", `
import { hello } from '../src/app';
describe('app', () => { it('works', () => { hello(); }); });
`)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	a := New(root)
	_, err := a.AnalyzeContext(ctx)
	if err == nil {
		t.Error("expected error from cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got: %v", err)
	}
}

// TestAnalyzeContext_CompletesNormally verifies that AnalyzeContext produces
// the same results as Analyze() when context is not cancelled.
func TestAnalyzeContext_CompletesNormally(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTempFile(t, root, "src/utils.ts", `
export function add(a, b) { return a + b; }
export function sub(a, b) { return a - b; }
`)
	writeTempFile(t, root, "test/utils.test.ts", `
import { add } from '../src/utils';
describe('utils', () => { it('adds', () => { expect(add(1,2)).toBe(3); }); });
`)

	// Analyze without context.
	a1 := New(root)
	snap1, err := a1.Analyze()
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// AnalyzeContext with background context should produce identical results.
	a2 := New(root)
	snap2, err := a2.AnalyzeContext(context.Background())
	if err != nil {
		t.Fatalf("AnalyzeContext failed: %v", err)
	}

	// Compare key counts for determinism.
	if len(snap1.TestFiles) != len(snap2.TestFiles) {
		t.Errorf("TestFile count: %d vs %d", len(snap1.TestFiles), len(snap2.TestFiles))
	}
	if len(snap1.CodeUnits) != len(snap2.CodeUnits) {
		t.Errorf("CodeUnit count: %d vs %d", len(snap1.CodeUnits), len(snap2.CodeUnits))
	}
	if len(snap1.CodeSurfaces) != len(snap2.CodeSurfaces) {
		t.Errorf("CodeSurface count: %d vs %d", len(snap1.CodeSurfaces), len(snap2.CodeSurfaces))
	}
	if len(snap1.TestCases) != len(snap2.TestCases) {
		t.Errorf("TestCase count: %d vs %d", len(snap1.TestCases), len(snap2.TestCases))
	}
}

// TestExtractFixturesCtx_CancelledSkipsWork verifies that ExtractFixturesCtx
// produces partial or empty results when cancelled.
func TestExtractFixturesCtx_CancelledSkipsWork(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for i := 0; i < 20; i++ {
		writeTempFile(t, root, fmt.Sprintf("test/t%d.test.ts", i), `
beforeEach(() => { setup(); });
it('test', () => {});
`)
	}

	testFiles := make([]models.TestFile, 20)
	for i := range testFiles {
		testFiles[i] = models.TestFile{
			Path:      fmt.Sprintf("test/t%d.test.ts", i),
			Framework: "jest",
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	fixtures := ExtractFixturesCtx(ctx, root, testFiles)
	// With pre-cancelled context, should get 0 or very few fixtures.
	if len(fixtures) > 5 {
		t.Errorf("expected few/no fixtures with cancelled context, got %d", len(fixtures))
	}
}

// TestPrewarmSourceFilesCtx_CancelledSkipsWork verifies that prewarming
// respects cancellation.
func TestPrewarmSourceFilesCtx_CancelledSkipsWork(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	var files []string
	for i := 0; i < 50; i++ {
		relPath := fmt.Sprintf("src/mod%d.ts", i)
		writeTempFile(t, root, relPath, "export function f() {}")
		files = append(files, relPath)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	fc := NewFileCache(root)
	fc.PrewarmSourceFilesCtx(ctx, files)

	stats := fc.Stats()
	// With cancelled context, should have cached 0 or very few files.
	if stats.CachedFiles > 5 {
		t.Errorf("expected few cached files with cancelled context, got %d", stats.CachedFiles)
	}
}

// Helper to create test fixture files.
func createFixtureDir(t *testing.T, root string, count int) {
	t.Helper()
	for i := 0; i < count; i++ {
		dir := filepath.Join(root, fmt.Sprintf("pkg%d", i/10))
		os.MkdirAll(dir, 0o755)
		os.WriteFile(filepath.Join(dir, fmt.Sprintf("mod%d.ts", i%10)),
			[]byte(fmt.Sprintf("export function fn%d() { return %d; }", i, i)), 0o644)
	}
}
