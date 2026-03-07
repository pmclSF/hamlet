package impact

import (
	"os/exec"
	"path/filepath"
	"strings"
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
