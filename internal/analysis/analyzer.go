// Package analysis implements Hamlet's static analysis nucleus.
//
// This package is responsible for scanning a repository, discovering test
// files, detecting frameworks, and producing the initial snapshot foundation.
//
// Limitations (V3 nucleus):
//   - Framework detection uses file naming heuristics and simple content
//     patterns rather than full AST analysis.
//   - Code unit extraction uses regex patterns, not full AST analysis.
//   - Runtime artifact ingestion is not yet implemented.
package analysis

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/pmclSF/hamlet/internal/models"
	"github.com/pmclSF/hamlet/internal/ownership"
)

// Analyzer performs static analysis on a repository root to produce
// the initial TestSuiteSnapshot foundation.
type Analyzer struct {
	root string
}

// New creates an Analyzer for the given repository root path.
func New(root string) *Analyzer {
	return &Analyzer{root: root}
}

// Analyze scans the repository and returns a populated TestSuiteSnapshot.
func (a *Analyzer) Analyze() (*models.TestSuiteSnapshot, error) {
	absRoot, err := filepath.Abs(a.root)
	if err != nil {
		return nil, err
	}

	testFiles, err := discoverTestFiles(absRoot)
	if err != nil {
		return nil, err
	}

	// Analyze content of each test file (counts for tests, assertions, mocks).
	for i := range testFiles {
		analyzeTestFileContent(&testFiles[i], absRoot)
	}

	frameworks := buildFrameworkInventory(testFiles)
	languages := detectLanguages(testFiles)
	packageManagers := detectPackageManagers(absRoot)
	ciSystems := detectCISystems(absRoot)
	commitSHA, branch := gitInfo(absRoot)

	// Extract exported code units for untested-export detection.
	codeUnits := extractExportedCodeUnits(absRoot, testFiles)

	// Resolve ownership for test files.
	resolver := ownership.NewResolver(absRoot)
	for i := range testFiles {
		testFiles[i].Owner = resolver.Resolve(testFiles[i].Path)
	}

	snapshot := &models.TestSuiteSnapshot{
		Repository: models.RepositoryMetadata{
			Name:              filepath.Base(absRoot),
			RootPath:          absRoot,
			Languages:         languages,
			PackageManagers:   packageManagers,
			CISystems:         ciSystems,
			SnapshotTimestamp: time.Now().UTC(),
			CommitSHA:         commitSHA,
			Branch:            branch,
		},
		Frameworks: frameworks,
		TestFiles:  testFiles,
		CodeUnits:  codeUnits,
		// Signals: populated by detectors after snapshot creation.
		// Risk: populated by risk engine after signal generation.
		GeneratedAt: time.Now().UTC(),
	}

	return snapshot, nil
}

// detectLanguages infers languages from the discovered test files.
func detectLanguages(testFiles []models.TestFile) []string {
	seen := map[string]bool{}
	for _, tf := range testFiles {
		ext := strings.ToLower(filepath.Ext(tf.Path))
		switch ext {
		case ".js", ".jsx", ".mjs", ".cjs":
			seen["javascript"] = true
		case ".ts", ".tsx", ".mts", ".cts":
			seen["typescript"] = true
		case ".go":
			seen["go"] = true
		case ".py":
			seen["python"] = true
		case ".java":
			seen["java"] = true
		case ".rb":
			seen["ruby"] = true
		}
	}
	langs := make([]string, 0, len(seen))
	for l := range seen {
		langs = append(langs, l)
	}
	return langs
}

// detectPackageManagers checks for known package manager lock files.
func detectPackageManagers(root string) []string {
	indicators := map[string]string{
		"package-lock.json": "npm",
		"yarn.lock":         "yarn",
		"pnpm-lock.yaml":   "pnpm",
		"bun.lockb":        "bun",
		"go.mod":           "go-modules",
		"requirements.txt": "pip",
		"Pipfile.lock":     "pipenv",
		"poetry.lock":      "poetry",
		"pom.xml":          "maven",
		"build.gradle":     "gradle",
		"Gemfile.lock":     "bundler",
	}
	var result []string
	for file, name := range indicators {
		if _, err := os.Stat(filepath.Join(root, file)); err == nil {
			result = append(result, name)
		}
	}
	return result
}

// detectCISystems checks for known CI configuration files/directories.
func detectCISystems(root string) []string {
	indicators := map[string]string{
		".github/workflows": "github-actions",
		".circleci":         "circleci",
		".travis.yml":       "travis",
		"Jenkinsfile":       "jenkins",
		".gitlab-ci.yml":    "gitlab-ci",
		"bitbucket-pipelines.yml": "bitbucket",
		".buildkite":        "buildkite",
	}
	var result []string
	for path, name := range indicators {
		full := filepath.Join(root, path)
		if _, err := os.Stat(full); err == nil {
			result = append(result, name)
		}
	}
	return result
}

// gitInfo attempts to read the current commit SHA and branch.
func gitInfo(root string) (sha, branch string) {
	if cmd := exec.Command("git", "-C", root, "rev-parse", "HEAD"); cmd != nil {
		if out, err := cmd.Output(); err == nil {
			sha = strings.TrimSpace(string(out))
		}
	}
	if cmd := exec.Command("git", "-C", root, "rev-parse", "--abbrev-ref", "HEAD"); cmd != nil {
		if out, err := cmd.Output(); err == nil {
			branch = strings.TrimSpace(string(out))
		}
	}
	return
}
