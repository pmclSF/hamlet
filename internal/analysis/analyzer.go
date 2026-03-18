// Package analysis implements Terrain's static analysis engine.
//
// This package is responsible for scanning a repository, discovering test
// files, detecting frameworks, and producing the initial snapshot foundation.
//
// Limitations (current detector stage):
//   - Framework detection uses file naming heuristics and simple content
//     patterns rather than full AST analysis.
//   - Code unit extraction uses regex patterns, not full AST analysis.
//   - Runtime artifact ingestion is not yet implemented.
package analysis

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/pmclSF/terrain/internal/models"
	"github.com/pmclSF/terrain/internal/ownership"
	"github.com/pmclSF/terrain/internal/testcase"
	"github.com/pmclSF/terrain/internal/testtype"
)

// Analyzer performs static analysis on a repository root to produce
// the initial TestSuiteSnapshot foundation.
type Analyzer struct {
	root string

	// Cache is a shared file content and AST cache. When non-nil,
	// all analysis stages read files through the cache to eliminate
	// redundant I/O. Populated automatically during Analyze().
	Cache *FileCache
}

// New creates an Analyzer for the given repository root path.
func New(root string) *Analyzer {
	return &Analyzer{root: root}
}

// Analyze scans the repository and returns a populated TestSuiteSnapshot.
// This is a convenience wrapper that uses context.Background().
// For cancellation support, use AnalyzeContext.
func (a *Analyzer) Analyze() (*models.TestSuiteSnapshot, error) {
	return a.AnalyzeContext(context.Background())
}

// AnalyzeContext scans the repository and returns a populated TestSuiteSnapshot.
// The context is checked at each major stage boundary and propagated into
// parallel file-processing loops, allowing callers to abort analysis cleanly.
func (a *Analyzer) AnalyzeContext(ctx context.Context) (*models.TestSuiteSnapshot, error) {
	absRoot, err := filepath.Abs(a.root)
	if err != nil {
		return nil, err
	}
	analyzedAt := time.Now().UTC()

	// Check context before starting work.
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Layer 1: Detect project-level frameworks from config files and dependencies.
	projectFrameworks := DetectProjectFrameworks(absRoot)

	testFiles, err := discoverTestFiles(absRoot, projectFrameworks)
	if err != nil {
		return nil, err
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Analyze content and extract test cases in one pass per file.
	rawByFile := make([][]models.TestCase, len(testFiles))
	parallelForEachIndexCtx(ctx, len(testFiles), func(i int) {
		src := analyzeTestFileContentCached(&testFiles[i], absRoot)
		if src != "" {
			cases := testcase.ExtractFromContent(src, testFiles[i].Path, testFiles[i].Framework)
			rawByFile[i] = testcase.ToModels(cases)
		}
	})

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	frameworks := buildFrameworkInventory(testFiles)
	languages := detectLanguages(testFiles)
	packageManagers := detectPackageManagers(absRoot)
	ciSystems := detectCISystems(absRoot)
	commitSHA, branch := gitInfo(absRoot)

	// Collect source files once (shared across all stages that need it).
	sourceFileCache, err := collectSourceFilesCtx(ctx, absRoot)
	if err != nil {
		return nil, err
	}

	// Initialize file content cache and prewarm source files.
	if a.Cache == nil {
		a.Cache = NewFileCache(absRoot)
	}
	a.Cache.PrewarmSourceFilesCtx(ctx, sourceFileCache)

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Also prewarm test files.
	testPaths := make([]string, len(testFiles))
	for i, tf := range testFiles {
		testPaths[i] = tf.Path
	}
	a.Cache.PrewarmSourceFilesCtx(ctx, testPaths)

	// Run six independent I/O stages concurrently.
	// Each stage respects ctx for cancellation.
	var (
		codeUnits       []models.CodeUnit
		codeSurfaces    []models.CodeSurface
		fixtureSurfaces []models.FixtureSurface
		importGraph     *ImportGraph
		testCases       []models.TestCase
		ciMatrix        *CIMatrixResult
		fwMatrix        *FrameworkMatrixResult
		stageWG         sync.WaitGroup
	)

	fc := a.Cache

	stageWG.Add(6)
	go func() {
		defer stageWG.Done()
		codeUnits = extractCodeUnitsCachedCtx(ctx, absRoot, testFiles, sourceFileCache, fc)
	}()
	go func() {
		defer stageWG.Done()
		importGraph = BuildImportGraphCtx(ctx, absRoot, testFiles)
	}()
	go func() {
		defer stageWG.Done()
		codeSurfaces = inferCodeSurfacesCachedCtx(ctx, absRoot, testFiles, sourceFileCache, fc)
	}()
	go func() {
		defer stageWG.Done()
		fixtureSurfaces = ExtractFixturesCtx(ctx, absRoot, testFiles)
	}()
	go func() {
		defer stageWG.Done()
		ciMatrix = ParseCIMatrices(absRoot)
	}()
	go func() {
		defer stageWG.Done()
		fwMatrix = ParseFrameworkMatrices(absRoot, testFiles)
	}()
	stageWG.Wait()

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Content-based AI context inference (runs after name-based detection).
	contentSurfaces := inferAIContextCachedCtx(ctx, absRoot, testFiles, codeSurfaces, sourceFileCache, fc)
	codeSurfaces = append(codeSurfaces, contentSurfaces...)

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Populate per-test linked code units using symbol-level resolution.
	PopulateSymbolLinksCtx(ctx, absRoot, testFiles, codeUnits, importGraph)

	// Flatten test cases and resolve collisions.
	rawTestCases := make([]models.TestCase, 0, len(testFiles))
	for i := range rawByFile {
		rawTestCases = append(rawTestCases, rawByFile[i]...)
	}
	testCases, _ = testcase.DetectAndResolveCollisions(rawTestCases)

	// Infer test types (unit, integration, e2e, etc.) with evidence.
	testCases = testtype.InferAll(testCases)

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Resolve ownership for test files.
	resolver := ownership.NewResolver(absRoot)
	for i := range testFiles {
		testFiles[i].Owner = resolver.Resolve(testFiles[i].Path)
	}

	// Derive behavior surfaces from code surfaces (optional layer).
	behaviorSurfaces := DeriveBehaviorSurfaces(codeSurfaces)

	// Extract structured RAG pipeline components from source files.
	ragComponents := extractRAGComponentsCachedCtx(ctx, absRoot, codeSurfaces, sourceFileCache, fc)

	// Merge environment data from CI and framework matrix parsers.
	var environments []models.Environment
	var environmentClasses []models.EnvironmentClass
	var deviceConfigs []models.DeviceConfig
	if ciMatrix != nil {
		environments = append(environments, ciMatrix.Environments...)
		environmentClasses = append(environmentClasses, ciMatrix.EnvironmentClasses...)
	}
	if fwMatrix != nil {
		environments = append(environments, fwMatrix.Environments...)
		environmentClasses = mergeEnvironmentClasses(environmentClasses, fwMatrix.EnvironmentClasses)
		deviceConfigs = append(deviceConfigs, fwMatrix.DeviceConfigs...)
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
		Frameworks:       frameworks,
		TestFiles:        testFiles,
		TestCases:        testCases,
		CodeUnits:        codeUnits,
		CodeSurfaces:     codeSurfaces,
		BehaviorSurfaces: behaviorSurfaces,
		FixtureSurfaces:     fixtureSurfaces,
		RAGPipelineSurfaces: ragComponents,
		Environments:       environments,
		EnvironmentClasses: environmentClasses,
		DeviceConfigs:      deviceConfigs,
		ImportGraph:        importGraph.TestImports,
		SourceImports:      importGraph.SourceImports,
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

// extractRAGPipelineComponents scans source files for RAG pipeline components
// with structured config extraction, then links them to CodeSurfaces.
func extractRAGPipelineComponents(root string, codeSurfaces []models.CodeSurface, sourceFiles []string) []models.RAGPipelineSurface {
	testPaths := map[string]bool{} // no test filtering needed here — sourceFiles already filtered
	_ = testPaths

	componentsByFile := make([][]models.RAGPipelineSurface, len(sourceFiles))
	parallelForEachIndex(len(sourceFiles), func(i int) {
		relPath := sourceFiles[i]
		ext := strings.ToLower(relPathExt(relPath))
		lang, ok := languageForExt[ext]
		if !ok {
			return
		}
		content, err := os.ReadFile(filepath.Join(root, relPath))
		if err != nil {
			return
		}
		componentsByFile[i] = ParseRAGStructured(relPath, string(content), lang)
	})

	var allComponents []models.RAGPipelineSurface
	for _, batch := range componentsByFile {
		allComponents = append(allComponents, batch...)
	}

	// Link RAG components to their corresponding CodeSurfaces.
	LinkRAGSurfacesToCodeSurfaces(allComponents, codeSurfaces)

	return allComponents
}

// mergeEnvironmentClasses appends classes from src into dst, merging member
// IDs for classes that share a ClassID.
func mergeEnvironmentClasses(dst, src []models.EnvironmentClass) []models.EnvironmentClass {
	for _, cls := range src {
		dst = appendClassIfNew(dst, cls)
	}
	return dst
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
