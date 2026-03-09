package analysis

import (
	"path/filepath"
	"runtime"
	"sort"
	"testing"
)

func fixtureRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to determine test file path")
	}
	return filepath.Join(filepath.Dir(thisFile), "testdata", "sample-repo")
}

func TestDiscoverTestFiles(t *testing.T) {
	root := fixtureRoot(t)
	files, err := discoverTestFiles(root)
	if err != nil {
		t.Fatalf("discoverTestFiles failed: %v", err)
	}

	// We expect: auth.test.js, mocked.test.js, utils.spec.ts, weak.test.js,
	// server.test.js, helpers_test.go, test_login.py
	// We do NOT expect: node_modules/pkg/index.test.js
	if len(files) != 7 {
		paths := make([]string, len(files))
		for i, f := range files {
			paths[i] = f.Path
		}
		t.Fatalf("expected 7 test files, got %d: %v", len(files), paths)
	}

	// Verify node_modules was skipped
	for _, f := range files {
		if filepath.Base(f.Path) == "index.test.js" {
			t.Errorf("node_modules file should have been skipped: %s", f.Path)
		}
	}
}

func TestDiscoverTestFiles_SkipsDirs(t *testing.T) {
	root := fixtureRoot(t)
	files, err := discoverTestFiles(root)
	if err != nil {
		t.Fatalf("discoverTestFiles failed: %v", err)
	}

	for _, f := range files {
		if filepath.Base(filepath.Dir(f.Path)) == "node_modules" {
			t.Errorf("should skip node_modules, found: %s", f.Path)
		}
	}
}

func TestFrameworkDetection(t *testing.T) {
	root := fixtureRoot(t)
	files, err := discoverTestFiles(root)
	if err != nil {
		t.Fatalf("discoverTestFiles failed: %v", err)
	}

	fwByFile := map[string]string{}
	for _, f := range files {
		fwByFile[filepath.Base(f.Path)] = f.Framework
	}

	tests := []struct {
		file     string
		wantFW   string
	}{
		{"auth.test.js", "jest"},
		{"utils.spec.ts", "vitest"},
		{"helpers_test.go", "go-testing"},
		{"test_login.py", "pytest"},
	}

	for _, tc := range tests {
		t.Run(tc.file, func(t *testing.T) {
			got := fwByFile[tc.file]
			if got != tc.wantFW {
				t.Errorf("framework for %s = %q, want %q", tc.file, got, tc.wantFW)
			}
		})
	}
}

func TestBuildFrameworkInventory(t *testing.T) {
	root := fixtureRoot(t)
	files, err := discoverTestFiles(root)
	if err != nil {
		t.Fatalf("discoverTestFiles failed: %v", err)
	}

	frameworks := buildFrameworkInventory(files)

	if len(frameworks) == 0 {
		t.Fatal("expected at least one framework, got 0")
	}

	// Verify sorted by file count descending
	counts := make([]int, len(frameworks))
	for i, fw := range frameworks {
		counts[i] = fw.FileCount
	}
	if !sort.SliceIsSorted(counts, func(i, j int) bool {
		return counts[i] > counts[j]
	}) {
		t.Error("frameworks should be sorted by file count descending")
	}

	// Verify known frameworks are present
	names := map[string]bool{}
	for _, fw := range frameworks {
		names[fw.Name] = true
	}
	for _, want := range []string{"jest", "vitest", "go-testing", "pytest"} {
		if !names[want] {
			t.Errorf("expected framework %q in inventory", want)
		}
	}
}

func TestAnalyzerProducesSnapshot(t *testing.T) {
	root := fixtureRoot(t)
	a := New(root)
	snap, err := a.Analyze()
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if snap.Repository.Name != "sample-repo" {
		t.Errorf("Repository.Name = %q, want %q", snap.Repository.Name, "sample-repo")
	}
	if len(snap.Frameworks) == 0 {
		t.Error("expected frameworks in snapshot")
	}
	if len(snap.TestFiles) == 0 {
		t.Error("expected test files in snapshot")
	}
	if snap.GeneratedAt.IsZero() {
		t.Error("GeneratedAt should not be zero")
	}

	// Package manager detection
	pmFound := false
	for _, pm := range snap.Repository.PackageManagers {
		if pm == "npm" {
			pmFound = true
		}
	}
	if !pmFound {
		t.Errorf("expected npm in package managers, got %v", snap.Repository.PackageManagers)
	}
}
