// Package analysis implements Hamlet's static analysis engine.
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
	goruntime "runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/pmclSF/hamlet/internal/models"
	"github.com/pmclSF/hamlet/internal/ownership"
	"github.com/pmclSF/hamlet/internal/testcase"
	"github.com/pmclSF/hamlet/internal/testtype"
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
	analyzedAt := time.Now().UTC()

	// Layer 1: Detect project-level frameworks from config files and dependencies.
	projectCtx := DetectProjectFrameworks(absRoot)

	testFiles, err := discoverTestFiles(absRoot, projectCtx)
	if err != nil {
		return nil, err
	}

	// Analyze content and extract test cases in one pass per file.
	// Merging these eliminates a redundant os.ReadFile per test file.
	rawByFile := make([][]models.TestCase, len(testFiles))
	parallelForEachIndex(len(testFiles), func(i int) {
		src := analyzeTestFileContentCached(&testFiles[i], absRoot)
		if src != "" {
			cases := testcase.ExtractFromContent(src, testFiles[i].Path, testFiles[i].Framework)
			rawByFile[i] = testcase.ToModels(cases)
		}
	})

	frameworks := buildFrameworkInventory(testFiles)
	languages := detectLanguages(testFiles)
	packageManagers := detectPackageManagers(absRoot)
	ciSystems := detectCISystems(absRoot)
	commitSHA, branch := gitInfo(absRoot)

	// Run two independent I/O stages concurrently:
	//   1. Extract exported code units (reads source files)
	//   2. Build import graph (reads test file imports)
	var (
		codeUnits   []models.CodeUnit
		importGraph *ImportGraph
		testCases   []models.TestCase
		stageWG     sync.WaitGroup
	)

	stageWG.Add(2)
	go func() {
		defer stageWG.Done()
		codeUnits = extractExportedCodeUnits(absRoot, testFiles)
	}()
	go func() {
		defer stageWG.Done()
		importGraph = BuildImportGraph(absRoot, testFiles)
	}()
	stageWG.Wait()

	// Populate per-test linked code units from the import graph.
	populateLinkedCodeUnits(testFiles, codeUnits, importGraph)

	// Flatten test cases and resolve collisions.
	rawTestCases := make([]models.TestCase, 0, len(testFiles))
	for i := range rawByFile {
		rawTestCases = append(rawTestCases, rawByFile[i]...)
	}
	testCases, _ = testcase.DetectAndResolveCollisions(rawTestCases)

	// Infer test types (unit, integration, e2e, etc.) with evidence.
	testCases = testtype.InferAll(testCases)

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
			SnapshotTimestamp: analyzedAt,
			CommitSHA:         commitSHA,
			Branch:            branch,
		},
		Frameworks:  frameworks,
		TestFiles:   testFiles,
		TestCases:   testCases,
		CodeUnits:   codeUnits,
		ImportGraph: importGraph.TestImports,
		// Signals: populated by detectors after snapshot creation.
		// Risk: populated by risk engine after signal generation.
		GeneratedAt: analyzedAt,
	}

	return snapshot, nil
}

func parallelForEachIndex(n int, fn func(i int)) {
	if n <= 1 {
		for i := 0; i < n; i++ {
			fn(i)
		}
		return
	}

	workers := goruntime.GOMAXPROCS(0)
	if workers < 1 {
		workers = 1
	}
	if workers > n {
		workers = n
	}

	indexCh := make(chan int, n)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range indexCh {
				fn(idx)
			}
		}()
	}
	for i := 0; i < n; i++ {
		indexCh <- i
	}
	close(indexCh)
	wg.Wait()
}

func populateLinkedCodeUnits(testFiles []models.TestFile, codeUnits []models.CodeUnit, graph *ImportGraph) {
	if graph == nil || len(graph.TestImports) == 0 || len(testFiles) == 0 || len(codeUnits) == 0 {
		return
	}

	unitsByPath := make(map[string][]models.CodeUnit, len(codeUnits))
	for _, cu := range codeUnits {
		unitsByPath[cu.Path] = append(unitsByPath[cu.Path], cu)
	}

	for i := range testFiles {
		imports := graph.TestImports[testFiles[i].Path]
		if len(imports) == 0 {
			continue
		}

		seen := map[string]bool{}
		linked := make([]string, 0, 8)
		for src := range imports {
			for _, cu := range unitsByPath[src] {
				id := cu.UnitID
				if id == "" {
					id = buildUnitID(cu.Path, cu.Name, cu.ParentName)
				}
				if id == "" || seen[id] {
					continue
				}
				seen[id] = true
				linked = append(linked, id)
			}
		}
		if len(linked) == 0 {
			continue
		}
		sort.Strings(linked)
		testFiles[i].LinkedCodeUnits = linked
	}
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
	sort.Strings(langs)
	return langs
}

// detectPackageManagers checks for known package manager lock files.
func detectPackageManagers(root string) []string {
	indicators := map[string]string{
		"package-lock.json": "npm",
		"yarn.lock":         "yarn",
		"pnpm-lock.yaml":    "pnpm",
		"bun.lockb":         "bun",
		"go.mod":            "go-modules",
		"requirements.txt":  "pip",
		"Pipfile.lock":      "pipenv",
		"poetry.lock":       "poetry",
		"pom.xml":           "maven",
		"build.gradle":      "gradle",
		"Gemfile.lock":      "bundler",
	}
	var result []string
	for file, name := range indicators {
		if _, err := os.Stat(filepath.Join(root, file)); err == nil {
			result = append(result, name)
		}
	}
	sort.Strings(result)
	return result
}

// detectCISystems checks for known CI configuration files/directories.
func detectCISystems(root string) []string {
	indicators := map[string]string{
		".github/workflows":       "github-actions",
		".circleci":               "circleci",
		".travis.yml":             "travis",
		"Jenkinsfile":             "jenkins",
		".gitlab-ci.yml":          "gitlab-ci",
		"bitbucket-pipelines.yml": "bitbucket",
		".buildkite":              "buildkite",
	}
	var result []string
	for path, name := range indicators {
		full := filepath.Join(root, path)
		if _, err := os.Stat(full); err == nil {
			result = append(result, name)
		}
	}
	sort.Strings(result)
	return result
}

// gitInfo attempts to read the current commit SHA and branch.
// Returns empty strings if git is unavailable or the directory is not a repo.
func gitInfo(root string) (sha, branch string) {
	if out, err := exec.Command("git", "-C", root, "rev-parse", "HEAD").Output(); err == nil {
		sha = strings.TrimSpace(string(out))
	}
	if out, err := exec.Command("git", "-C", root, "rev-parse", "--abbrev-ref", "HEAD").Output(); err == nil {
		branch = strings.TrimSpace(string(out))
	}
	return
}
