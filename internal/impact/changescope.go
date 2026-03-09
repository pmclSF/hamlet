package impact

import (
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pmclSF/hamlet/internal/models"
)

// ChangeScopeFromGitDiff creates a ChangeScope from git diff against a base ref.
func ChangeScopeFromGitDiff(repoRoot, baseRef string) (*ChangeScope, error) {
	if baseRef == "" {
		baseRef = "HEAD~1"
	}

	cmd := exec.Command("git", "diff", "--name-status", baseRef)
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseGitDiffOutput(string(out), repoRoot), nil
}

// ChangeScopeFromPaths creates a ChangeScope from explicit file paths.
func ChangeScopeFromPaths(paths []string, changeKind ChangeKind) *ChangeScope {
	scope := &ChangeScope{
		Source: "explicit",
	}
	for _, p := range paths {
		scope.ChangedFiles = append(scope.ChangedFiles, ChangedFile{
			Path:       p,
			ChangeKind: changeKind,
			IsTestFile: isTestFilePath(p),
		})
	}
	return scope
}

func parseGitDiffOutput(output, repoRoot string) *ChangeScope {
	scope := &ChangeScope{
		Source: "git-diff",
	}

	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		status := parts[0]
		path := parts[1]

		// Normalize to repo-relative path.
		if filepath.IsAbs(path) {
			rel, err := filepath.Rel(repoRoot, path)
			if err == nil {
				path = rel
			}
		}

		cf := ChangedFile{
			Path:       path,
			IsTestFile: isTestFilePath(path),
		}

		switch {
		case status == "A":
			cf.ChangeKind = ChangeAdded
		case status == "D":
			cf.ChangeKind = ChangeDeleted
		case status == "M":
			cf.ChangeKind = ChangeModified
		case strings.HasPrefix(status, "R"):
			cf.ChangeKind = ChangeRenamed
			if len(parts) >= 3 {
				cf.OldPath = path
				cf.Path = parts[2]
			}
		default:
			cf.ChangeKind = ChangeModified
		}

		scope.ChangedFiles = append(scope.ChangedFiles, cf)
	}

	return scope
}

func isTestFilePath(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	testPatterns := []string{
		".test.", ".spec.", "_test.", "_spec.",
		"test_", "spec_",
	}
	for _, p := range testPatterns {
		if strings.Contains(base, p) {
			return true
		}
	}
	// Go test files.
	if strings.HasSuffix(base, "_test.go") {
		return true
	}
	// Test directories.
	dir := strings.ToLower(path)
	testDirs := []string{"/test/", "/tests/", "/__tests__/", "/e2e/", "/cypress/", "/spec/"}
	for _, d := range testDirs {
		if strings.Contains(dir, d) {
			return true
		}
	}
	return false
}

// ChangeScopeFromCIList creates a ChangeScope from a newline-separated list
// of changed file paths, as typically provided by CI systems.
// All files are treated as modified unless they match known patterns.
func ChangeScopeFromCIList(fileList string, repoRoot string) *ChangeScope {
	scope := &ChangeScope{
		Source: "ci-changed-files",
	}

	for _, line := range strings.Split(strings.TrimSpace(fileList), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Normalize to repo-relative path.
		path := line
		if filepath.IsAbs(path) && repoRoot != "" {
			rel, err := filepath.Rel(repoRoot, path)
			if err == nil {
				path = rel
			}
		}
		// Clean path separators.
		path = filepath.ToSlash(filepath.Clean(path))

		scope.ChangedFiles = append(scope.ChangedFiles, ChangedFile{
			Path:       path,
			ChangeKind: ChangeModified,
			IsTestFile: isTestFilePath(path),
		})
	}

	return scope
}

// ChangeScopeFromComparison creates a ChangeScope by comparing two snapshots.
// It identifies files that were added, removed, or modified between snapshots
// by comparing test file and code unit inventories.
func ChangeScopeFromComparison(from, to *models.TestSuiteSnapshot) *ChangeScope {
	scope := &ChangeScope{
		Source: "snapshot-compare",
	}

	if from == nil || to == nil {
		return scope
	}

	// Collect all file paths from both snapshots.
	fromFiles := collectSnapshotFiles(from)
	toFiles := collectSnapshotFiles(to)

	// Find added files (in to but not from).
	for path := range toFiles {
		if !fromFiles[path] {
			scope.ChangedFiles = append(scope.ChangedFiles, ChangedFile{
				Path:       path,
				ChangeKind: ChangeAdded,
				IsTestFile: isTestFilePath(path),
			})
		}
	}

	// Find deleted files (in from but not to).
	for path := range fromFiles {
		if !toFiles[path] {
			scope.ChangedFiles = append(scope.ChangedFiles, ChangedFile{
				Path:       path,
				ChangeKind: ChangeDeleted,
				IsTestFile: isTestFilePath(path),
			})
		}
	}

	// Files in both are assumed modified (we can't know without content diff).
	for path := range toFiles {
		if fromFiles[path] {
			scope.ChangedFiles = append(scope.ChangedFiles, ChangedFile{
				Path:       path,
				ChangeKind: ChangeModified,
				IsTestFile: isTestFilePath(path),
			})
		}
	}

	// Sort for determinism.
	sort.Slice(scope.ChangedFiles, func(i, j int) bool {
		return scope.ChangedFiles[i].Path < scope.ChangedFiles[j].Path
	})

	return scope
}

// collectSnapshotFiles collects all unique file paths from a snapshot's
// test files and code units.
func collectSnapshotFiles(snap *models.TestSuiteSnapshot) map[string]bool {
	files := map[string]bool{}
	for _, tf := range snap.TestFiles {
		files[tf.Path] = true
	}
	for _, cu := range snap.CodeUnits {
		files[cu.Path] = true
	}
	return files
}
