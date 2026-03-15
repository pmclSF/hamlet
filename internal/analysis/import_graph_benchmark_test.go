package analysis

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

func BenchmarkBuildImportGraph(b *testing.B) {
	root, testFiles := makeImportGraphBenchmarkRepo(b, 200)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BuildImportGraph(root, testFiles)
	}
}

func makeImportGraphBenchmarkRepo(b *testing.B, files int) (string, []models.TestFile) {
	b.Helper()
	dir := b.TempDir()
	srcDir := filepath.Join(dir, "src")
	testDir := filepath.Join(dir, "tests")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		b.Fatalf("mkdir src: %v", err)
	}
	if err := os.MkdirAll(testDir, 0o755); err != nil {
		b.Fatalf("mkdir tests: %v", err)
	}

	testFiles := make([]models.TestFile, 0, files)
	for i := 0; i < files; i++ {
		srcPath := filepath.Join(srcDir, fmt.Sprintf("u%d.ts", i))
		src := fmt.Sprintf("export const U%d = %d\n", i, i)
		if err := os.WriteFile(srcPath, []byte(src), 0o644); err != nil {
			b.Fatalf("write source: %v", err)
		}
		testPath := filepath.Join(testDir, fmt.Sprintf("u%d.test.ts", i))
		test := fmt.Sprintf("import { U%d } from '../src/u%d'\ntest('u%d', () => { expect(U%d).toBe(%d) })\n", i, i, i, i, i)
		if err := os.WriteFile(testPath, []byte(test), 0o644); err != nil {
			b.Fatalf("write test: %v", err)
		}
		testFiles = append(testFiles, models.TestFile{
			Path:      filepath.ToSlash(filepath.Join("tests", fmt.Sprintf("u%d.test.ts", i))),
			Framework: "vitest",
		})
	}

	return dir, testFiles
}
