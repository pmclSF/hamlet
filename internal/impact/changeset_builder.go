package impact

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/pmclSF/terrain/internal/models"
)

// ChangeSetFromGitDiff creates a ChangeSet from git diff against a base ref.
//
// It resolves SHAs, detects shallow clones, infers packages and services,
// and degrades gracefully when history is limited.
func ChangeSetFromGitDiff(repoRoot, baseRef string) (*models.ChangeSet, error) {
	cs := &models.ChangeSet{
		Repository: repoRoot,
		BaseRef:    baseRef,
		CreatedAt:  time.Now().UTC(),
	}

	// Detect shallow clone.
	if isShallowClone(repoRoot) {
		cs.IsShallow = true
		cs.Limitations = append(cs.Limitations, "shallow clone: commit history may be incomplete")
	}

	// Resolve base ref.
	if baseRef == "" {
		if refExists(repoRoot, "HEAD~1") {
			baseRef = "HEAD~1"
		}
	}

	// Resolve SHAs when possible.
	if headSHA, err := resolveRef(repoRoot, "HEAD"); err == nil {
		cs.HeadSHA = headSHA
	}
	if baseRef != "" {
		if baseSHA, err := resolveRef(repoRoot, baseRef); err == nil {
			cs.BaseSHA = baseSHA
		}
	}

	// Run git diff.
	out, err := gitDiffNameStatus(repoRoot, baseRef)
	if err != nil {
		// If the base ref doesn't exist in a shallow clone, fall back to
		// working tree diff.
		if cs.IsShallow && baseRef != "" {
			cs.Limitations = append(cs.Limitations,
				fmt.Sprintf("base ref %q not reachable in shallow clone: falling back to working tree diff", baseRef))
			cs.BaseSHA = ""
			out, err = gitDiffNameStatus(repoRoot, "")
			if err != nil {
				return nil, err
			}
			cs.Source = "shallow-clone"
		} else {
			return nil, err
		}
	}

	// Determine source.
	if cs.Source == "" {
		if baseRef == "" {
			cs.Source = "git-diff-working-tree"
		} else {
			cs.Source = "git-diff"
		}
	}

	// Parse diff output into ChangedFiles.
	cs.ChangedFiles = parseGitDiffToChangedFiles(string(out), repoRoot)

	// Derive packages, services, and configs.
	cs.ChangedPackages = inferChangedPackages(cs.ChangedFiles)
	cs.ChangedServices = inferChangedServices(cs.ChangedFiles)
	cs.ChangedConfigs = collectChangedConfigs(cs.ChangedFiles)

	return cs, nil
}

// ChangeSetFromPaths creates a ChangeSet from explicit file paths.
func ChangeSetFromPaths(paths []string, changeKind models.ChangeKind) *models.ChangeSet {
	cs := &models.ChangeSet{
		Source:    "explicit",
		CreatedAt: time.Now().UTC(),
	}
	for _, p := range paths {
		cs.ChangedFiles = append(cs.ChangedFiles, models.ChangedFile{
			Path:       p,
			ChangeKind: changeKind,
			IsTestFile: isTestFilePath(p),
		})
	}
	cs.ChangedPackages = inferChangedPackages(cs.ChangedFiles)
	cs.ChangedServices = inferChangedServices(cs.ChangedFiles)
	cs.ChangedConfigs = collectChangedConfigs(cs.ChangedFiles)
	return cs
}

// ChangeSetFromCIList creates a ChangeSet from a newline-separated list
// of changed file paths, as typically provided by CI systems.
func ChangeSetFromCIList(fileList string, repoRoot string) *models.ChangeSet {
	cs := &models.ChangeSet{
		Source:     "ci-changed-files",
		Repository: repoRoot,
		CreatedAt:  time.Now().UTC(),
	}

	for _, line := range strings.Split(strings.TrimSpace(fileList), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		path := line
		if filepath.IsAbs(path) && repoRoot != "" {
			rel, err := filepath.Rel(repoRoot, path)
			if err == nil {
				path = rel
			}
		}
		path = filepath.ToSlash(filepath.Clean(path))

		cs.ChangedFiles = append(cs.ChangedFiles, models.ChangedFile{
			Path:       path,
			ChangeKind: models.ChangeModified,
			IsTestFile: isTestFilePath(path),
		})
	}

	cs.ChangedPackages = inferChangedPackages(cs.ChangedFiles)
	cs.ChangedServices = inferChangedServices(cs.ChangedFiles)
	cs.ChangedConfigs = collectChangedConfigs(cs.ChangedFiles)
	return cs
}

// ChangeSetFromComparison creates a ChangeSet by comparing two snapshots.
func ChangeSetFromComparison(from, to *models.TestSuiteSnapshot) *models.ChangeSet {
	cs := &models.ChangeSet{
		Source:    "snapshot-compare",
		CreatedAt: time.Now().UTC(),
	}

	if from == nil || to == nil {
		return cs
	}

	fromFiles := collectSnapshotFiles(from)
	toFiles := collectSnapshotFiles(to)

	for path := range toFiles {
		if !fromFiles[path] {
			cs.ChangedFiles = append(cs.ChangedFiles, models.ChangedFile{
				Path:       path,
				ChangeKind: models.ChangeAdded,
				IsTestFile: isTestFilePath(path),
			})
		}
	}
	for path := range fromFiles {
		if !toFiles[path] {
			cs.ChangedFiles = append(cs.ChangedFiles, models.ChangedFile{
				Path:       path,
				ChangeKind: models.ChangeDeleted,
				IsTestFile: isTestFilePath(path),
			})
		}
	}
	for path := range toFiles {
		if fromFiles[path] {
			cs.ChangedFiles = append(cs.ChangedFiles, models.ChangedFile{
				Path:       path,
				ChangeKind: models.ChangeModified,
				IsTestFile: isTestFilePath(path),
			})
		}
	}

	sort.Slice(cs.ChangedFiles, func(i, j int) bool {
		return cs.ChangedFiles[i].Path < cs.ChangedFiles[j].Path
	})

	cs.ChangedPackages = inferChangedPackages(cs.ChangedFiles)
	cs.ChangedServices = inferChangedServices(cs.ChangedFiles)
	cs.ChangedConfigs = collectChangedConfigs(cs.ChangedFiles)
	return cs
}

// ChangeSetToScope converts a ChangeSet to a ChangeScope for backward
// compatibility with existing impact analysis code. This is a bridge during
// the migration — new code should consume ChangeSet directly.
func ChangeSetToScope(cs *models.ChangeSet) *ChangeScope {
	scope := &ChangeScope{
		BaselineRef: cs.BaseRef,
		CurrentRef:  cs.HeadSHA,
		Source:      cs.Source,
	}

	for _, f := range cs.ChangedFiles {
		scope.ChangedFiles = append(scope.ChangedFiles, ChangedFile{
			Path:       f.Path,
			ChangeKind: ChangeKind(f.ChangeKind),
			OldPath:    f.OldPath,
			IsTestFile: f.IsTestFile,
		})
	}

	return scope
}

// parseGitDiffToChangedFiles parses git diff --name-status output into
// models.ChangedFile values.
func parseGitDiffToChangedFiles(output, repoRoot string) []models.ChangedFile {
	var files []models.ChangedFile

	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			parts = strings.Fields(line)
		}
		if len(parts) < 2 {
			continue
		}

		status := parts[0]
		path := parts[1]

		if filepath.IsAbs(path) {
			rel, err := filepath.Rel(repoRoot, path)
			if err == nil {
				path = rel
			}
		}

		cf := models.ChangedFile{
			Path:       path,
			IsTestFile: isTestFilePath(path),
		}

		switch {
		case status == "A":
			cf.ChangeKind = models.ChangeAdded
		case status == "D":
			cf.ChangeKind = models.ChangeDeleted
		case status == "M":
			cf.ChangeKind = models.ChangeModified
		case strings.HasPrefix(status, "R"):
			cf.ChangeKind = models.ChangeRenamed
			if len(parts) >= 3 {
				cf.OldPath = path
				cf.Path = parts[2]
				cf.IsTestFile = isTestFilePath(cf.Path)
			}
		default:
			cf.ChangeKind = models.ChangeModified
		}

		files = append(files, cf)
	}

	return files
}

// resolveRef resolves a git ref to a full SHA.
func resolveRef(repoRoot, ref string) (string, error) {
	cmd := exec.Command("git", "rev-parse", ref)
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// isShallowClone checks if the repository is a shallow clone.
func isShallowClone(repoRoot string) bool {
	cmd := exec.Command("git", "rev-parse", "--is-shallow-repository")
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "true"
}

// inferChangedPackages derives package names from changed file paths.
// Uses language-aware heuristics: Go package = directory, JS/Python = first
// two path segments.
func inferChangedPackages(files []models.ChangedFile) []string {
	seen := map[string]bool{}
	for _, f := range files {
		if f.IsTestFile {
			continue
		}
		pkg := inferPackage(f.Path)
		if pkg != "" {
			seen[pkg] = true
		}
	}

	out := make([]string, 0, len(seen))
	for pkg := range seen {
		out = append(out, pkg)
	}
	sort.Strings(out)
	return out
}

// inferPackage returns the inferred package for a file path.
func inferPackage(path string) string {
	dir := filepath.Dir(filepath.ToSlash(path))
	if dir == "." || dir == "" {
		return ""
	}
	ext := strings.ToLower(filepath.Ext(path))

	// Go: package = directory path.
	if ext == ".go" {
		return dir
	}

	// JS/TS/Python/Java: use first two path segments for a coarser grouping.
	parts := strings.Split(dir, "/")
	if len(parts) >= 2 {
		return parts[0] + "/" + parts[1]
	}
	return parts[0]
}

// inferChangedServices detects service names from well-known directory layouts.
func inferChangedServices(files []models.ChangedFile) []string {
	seen := map[string]bool{}
	serviceDirs := []string{"services/", "cmd/", "apps/", "packages/", "modules/"}

	for _, f := range files {
		path := filepath.ToSlash(f.Path)
		for _, prefix := range serviceDirs {
			if strings.HasPrefix(path, prefix) {
				rest := strings.TrimPrefix(path, prefix)
				if idx := strings.Index(rest, "/"); idx > 0 {
					seen[rest[:idx]] = true
				}
				break
			}
		}
	}

	out := make([]string, 0, len(seen))
	for svc := range seen {
		out = append(out, svc)
	}
	sort.Strings(out)
	return out
}

// collectChangedConfigs identifies config and generated artifact paths.
func collectChangedConfigs(files []models.ChangedFile) []string {
	var configs []string
	for _, f := range files {
		if isConfigFile(f.Path) {
			configs = append(configs, f.Path)
		}
	}
	sort.Strings(configs)
	return configs
}

// isConfigFile returns true if the path looks like a configuration or
// generated artifact.
func isConfigFile(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	ext := strings.ToLower(filepath.Ext(path))

	// Extension-based.
	switch ext {
	case ".yml", ".yaml", ".toml", ".ini", ".cfg":
		return true
	}

	// Name-based.
	configNames := []string{
		"dockerfile", "makefile", "rakefile", "gemfile",
		"package.json", "package-lock.json", "go.mod", "go.sum",
		"cargo.toml", "cargo.lock", "requirements.txt", "setup.py",
		"pyproject.toml", "pom.xml", "build.gradle",
		".eslintrc.json", ".prettierrc", "tsconfig.json",
		".goreleaser.yaml", ".goreleaser.yml",
	}
	for _, name := range configNames {
		if base == name {
			return true
		}
	}

	// Path-based (CI, terraform, etc.).
	pathLower := strings.ToLower(filepath.ToSlash(path))
	configPaths := []string{".github/", ".gitlab-ci", ".circleci/", "terraform/", ".terraform/"}
	for _, p := range configPaths {
		if strings.HasPrefix(pathLower, p) || strings.Contains(pathLower, "/"+p) {
			return true
		}
	}

	return false
}
