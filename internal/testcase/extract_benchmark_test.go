package testcase

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func BenchmarkExtractTestCases(b *testing.B) {
	var src strings.Builder
	src.WriteString("describe('BenchmarkSuite', () => {\n")
	for i := 0; i < 300; i++ {
		src.WriteString("  describe('feature ")
		src.WriteString("x")
		src.WriteString("', () => {\n")
		for j := 0; j < 5; j++ {
			src.WriteString("    it('case ")
			src.WriteString("y")
			src.WriteString("', () => { expect(true).toBe(true) })\n")
		}
		src.WriteString("  })\n")
	}
	src.WriteString("})\n")

	dir := b.TempDir()
	path := filepath.Join(dir, "bench2.test.js")
	if err := os.WriteFile(path, []byte(src.String()), 0o644); err != nil {
		b.Fatalf("write bench test file: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Extract(dir, "bench2.test.js", "jest")
	}
}
