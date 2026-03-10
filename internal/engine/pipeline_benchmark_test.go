package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkRunPipeline(b *testing.B) {
	repos := []struct {
		name string
		root string
	}{
		{name: "small", root: "../analysis/testdata/sample-repo"},
		{name: "medium", root: makeSyntheticRepo(b, 20)},
		{name: "large", root: makeSyntheticRepo(b, 60)},
	}

	for _, repo := range repos {
		repo := repo
		b.Run(repo.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := RunPipeline(repo.root, PipelineOptions{EngineVersion: "bench"}); err != nil {
					b.Fatalf("RunPipeline(%s) failed: %v", repo.name, err)
				}
			}
		})
	}
}

func makeSyntheticRepo(b *testing.B, files int) string {
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

	for i := 0; i < files; i++ {
		src := fmt.Sprintf("export function fn%d() { return %d }\n", i, i)
		if err := os.WriteFile(filepath.Join(srcDir, fmt.Sprintf("m%d.js", i)), []byte(src), 0o644); err != nil {
			b.Fatalf("write src file: %v", err)
		}
		test := fmt.Sprintf("import { fn%d } from '../src/m%d'\n\ntest('fn%d', () => { expect(fn%d()).toBe(%d) })\n", i, i, i, i, i)
		if err := os.WriteFile(filepath.Join(testDir, fmt.Sprintf("m%d.test.js", i)), []byte(test), 0o644); err != nil {
			b.Fatalf("write test file: %v", err)
		}
	}

	return dir
}
