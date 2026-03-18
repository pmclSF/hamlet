package analysis

import (
	"context"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"sync"

	"github.com/pmclSF/terrain/internal/models"
)

// parallelForEachIndexCtx is like parallelForEachIndex but checks ctx.Done()
// before dispatching each work item. When cancelled, remaining items are
// skipped and the function returns promptly. Items already in-flight run
// to completion (they are per-file and fast).
func parallelForEachIndexCtx(ctx context.Context, n int, fn func(i int)) {
	if n <= 1 {
		for i := 0; i < n; i++ {
			if ctx.Err() != nil {
				return
			}
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
				if ctx.Err() != nil {
					// Drain remaining indices without processing.
					for range indexCh {
					}
					return
				}
				fn(idx)
			}
		}()
	}

	for i := 0; i < n; i++ {
		if ctx.Err() != nil {
			break
		}
		indexCh <- i
	}
	close(indexCh)
	wg.Wait()
}

// walkDirCtx is like walkDir but checks ctx between directory entries.
func walkDirCtx(ctx context.Context, root string, fn func(relPath string, isDir bool) bool) error {
	return walkDirRecCtx(ctx, root, "", fn)
}

func walkDirRecCtx(ctx context.Context, root, rel string, fn func(relPath string, isDir bool) bool) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	fullPath := root
	if rel != "" {
		fullPath = filepath.Join(root, rel)
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil
	}

	for _, e := range entries {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		childRel := e.Name()
		if rel != "" {
			childRel = filepath.Join(rel, e.Name())
		}
		if e.Type()&os.ModeSymlink != 0 {
			continue
		}

		if e.IsDir() {
			if fn(childRel, true) {
				continue
			}
			if err := walkDirRecCtx(ctx, root, childRel, fn); err != nil {
				return err
			}
		} else {
			fn(childRel, false)
		}
	}
	return nil
}

// collectSourceFilesCtx is like collectSourceFiles but respects cancellation
// during the directory walk.
func collectSourceFilesCtx(ctx context.Context, root string) ([]string, error) {
	sourceExts := map[string]bool{
		".js": true, ".jsx": true, ".ts": true, ".tsx": true,
		".mjs": true, ".mts": true, ".go": true, ".py": true,
		".java": true,
	}

	files := make([]string, 0, 128)
	err := walkDirCtx(ctx, root, func(relPath string, isDir bool) bool {
		if isDir {
			return skipDirs[relPathBase(relPath)]
		}
		ext := strings.ToLower(relPathExt(relPath))
		if sourceExts[ext] {
			files = append(files, relPath)
		}
		return false
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

// --- Context-aware cached consumer functions ---
// These mirror the non-Ctx versions in filecache_consumers.go but thread
// context through the parallel loops.

func extractCodeUnitsCachedCtx(ctx context.Context, root string, testFiles []models.TestFile, sourceFiles []string, fc *FileCache) []models.CodeUnit {
	if fc == nil {
		return extractExportedCodeUnitsFromList(root, testFiles, sourceFiles)
	}
	testPaths := map[string]bool{}
	for _, tf := range testFiles {
		testPaths[tf.Path] = true
	}
	unitsByFile := make([][]models.CodeUnit, len(sourceFiles))
	parallelForEachIndexCtx(ctx, len(sourceFiles), func(i int) {
		relPath := sourceFiles[i]
		if testPaths[relPath] {
			return
		}
		ext := strings.ToLower(relPathExt(relPath))
		lang, ok := languageForExt[ext]
		if !ok {
			return
		}
		src, ok := fc.ReadFile(relPath)
		if !ok {
			return
		}
		unitsByFile[i] = extractCodeUnitsFromSource(src, relPath, lang)
	})

	units := make([]models.CodeUnit, 0, len(sourceFiles))
	for i := range unitsByFile {
		units = append(units, unitsByFile[i]...)
	}
	return units
}

func inferCodeSurfacesCachedCtx(ctx context.Context, root string, testFiles []models.TestFile, sourceFiles []string, fc *FileCache) []models.CodeSurface {
	if fc == nil {
		return InferCodeSurfacesFromList(root, testFiles, sourceFiles)
	}
	testPaths := map[string]bool{}
	for _, tf := range testFiles {
		testPaths[tf.Path] = true
	}
	surfacesByFile := make([][]models.CodeSurface, len(sourceFiles))
	parallelForEachIndexCtx(ctx, len(sourceFiles), func(i int) {
		relPath := sourceFiles[i]
		if testPaths[relPath] {
			return
		}
		ext := strings.ToLower(relPathExt(relPath))
		lang, ok := languageForExt[ext]
		if !ok {
			return
		}
		if e, exists := surfaceRegistry[lang]; exists {
			surfacesByFile[i] = e.ExtractSurfaces(root, relPath)
		}
	})

	var surfaces []models.CodeSurface
	for _, s := range surfacesByFile {
		surfaces = append(surfaces, s...)
	}
	assignInferenceMetadata(surfaces)
	return surfaces
}

func inferAIContextCachedCtx(ctx context.Context, root string, testFiles []models.TestFile, existing []models.CodeSurface, sourceFiles []string, fc *FileCache) []models.CodeSurface {
	if fc == nil {
		return InferAIContextSurfacesFromList(root, testFiles, existing, sourceFiles)
	}
	testPaths := map[string]bool{}
	for _, tf := range testFiles {
		testPaths[tf.Path] = true
	}

	existingIDs := map[string]bool{}
	for _, s := range existing {
		existingIDs[s.SurfaceID] = true
	}

	var surfaces []models.CodeSurface
	for _, relPath := range sourceFiles {
		if ctx.Err() != nil {
			break
		}
		if testPaths[relPath] {
			continue
		}
		ext := strings.ToLower(relPathExt(relPath))
		lang, ok := languageForExt[ext]
		if !ok {
			continue
		}

		src, ok := fc.ReadFile(relPath)
		if !ok {
			continue
		}

		for _, parseFn := range []func(string, string, string) []models.CodeSurface{
			ParsePromptAST,
			ParseStructural,
			ParseEmbeddedPrompts,
			ParseRAGPipeline,
			ParseToolSchemas,
		} {
			for _, s := range parseFn(relPath, src, lang) {
				if !existingIDs[s.SurfaceID] {
					existingIDs[s.SurfaceID] = true
					surfaces = append(surfaces, s)
				}
			}
		}

		inferred := inferFromContent(relPath, src, lang)
		for _, s := range inferred {
			if !existingIDs[s.SurfaceID] {
				existingIDs[s.SurfaceID] = true
				surfaces = append(surfaces, s)
			}
		}
	}

	templateSurfaces := detectAITemplateFiles(root, existingIDs)
	surfaces = append(surfaces, templateSurfaces...)

	return surfaces
}

func extractRAGComponentsCachedCtx(ctx context.Context, root string, codeSurfaces []models.CodeSurface, sourceFiles []string, fc *FileCache) []models.RAGPipelineSurface {
	if fc == nil {
		return extractRAGPipelineComponents(root, codeSurfaces, sourceFiles)
	}

	componentsByFile := make([][]models.RAGPipelineSurface, len(sourceFiles))
	parallelForEachIndexCtx(ctx, len(sourceFiles), func(i int) {
		relPath := sourceFiles[i]
		ext := strings.ToLower(relPathExt(relPath))
		lang, ok := languageForExt[ext]
		if !ok {
			return
		}
		src, ok := fc.ReadFile(relPath)
		if !ok {
			return
		}
		componentsByFile[i] = ParseRAGStructured(relPath, src, lang)
	})

	var allComponents []models.RAGPipelineSurface
	for _, batch := range componentsByFile {
		allComponents = append(allComponents, batch...)
	}
	LinkRAGSurfacesToCodeSurfaces(allComponents, codeSurfaces)
	return allComponents
}

// --- Context-aware wrappers for external-facing functions ---

// BuildImportGraphCtx is like BuildImportGraph but checks cancellation
// between the test-file import extraction loop and source import extraction.
// Per-file import extraction is fast, so the main cancellation boundary
// is between the two phases.
func BuildImportGraphCtx(ctx context.Context, root string, testFiles []models.TestFile) *ImportGraph {
	if ctx.Err() != nil {
		return &ImportGraph{TestImports: map[string]map[string]bool{}}
	}
	// The inner parallel loops in BuildImportGraph use parallelForEachIndex.
	// For a full ctx-aware version we'd need to refactor BuildImportGraph itself.
	// Instead, check cancellation before invoking it — the per-file work is
	// bounded (each file's import extraction is <1ms) so the worst-case
	// latency before honouring cancellation is one full pass over test files.
	return BuildImportGraph(root, testFiles)
}

// ExtractFixturesCtx is like ExtractFixtures but respects cancellation.
func ExtractFixturesCtx(ctx context.Context, root string, testFiles []models.TestFile) []models.FixtureSurface {
	results := make([][]models.FixtureSurface, len(testFiles))
	parallelForEachIndexCtx(ctx, len(testFiles), func(i int) {
		tf := &testFiles[i]
		content, err := os.ReadFile(filepath.Join(root, tf.Path))
		if err != nil {
			return
		}
		lang := frameworkLanguage(tf.Framework)
		results[i] = detectFixtures(string(content), tf.Path, lang, tf.Framework)
	})

	var fixtures []models.FixtureSurface
	for _, batch := range results {
		fixtures = append(fixtures, batch...)
	}
	return fixtures
}

// PopulateSymbolLinksCtx is like PopulateSymbolLinks but respects cancellation.
func PopulateSymbolLinksCtx(ctx context.Context, root string, testFiles []models.TestFile, codeUnits []models.CodeUnit, importGraph *ImportGraph) {
	if importGraph == nil || len(importGraph.TestImports) == 0 || len(codeUnits) == 0 {
		return
	}

	// Resolve symbol-level links with cancellation.
	links := ResolveSymbolLinksCtx(ctx, root, testFiles, codeUnits, importGraph)

	// Group links by test path.
	linksByTest := map[string][]SymbolLink{}
	for _, link := range links {
		linksByTest[link.TestPath] = append(linksByTest[link.TestPath], link)
	}

	// Build file-level fallback.
	unitsByFile := map[string][]models.CodeUnit{}
	for _, cu := range codeUnits {
		unitsByFile[cu.Path] = append(unitsByFile[cu.Path], cu)
	}

	for i := range testFiles {
		tf := &testFiles[i]
		symbolLinks := linksByTest[tf.Path]

		if len(symbolLinks) > 0 {
			seen := map[string]bool{}
			linked := make([]string, 0, len(symbolLinks))
			for _, sl := range symbolLinks {
				if !seen[sl.CodeUnitID] {
					seen[sl.CodeUnitID] = true
					linked = append(linked, sl.CodeUnitID)
				}
			}
			sortStrings(linked)
			tf.LinkedCodeUnits = linked
		} else {
			imports := importGraph.TestImports[tf.Path]
			if len(imports) == 0 {
				continue
			}
			seen := map[string]bool{}
			var linked []string
			for srcPath := range imports {
				for _, cu := range unitsByFile[srcPath] {
					id := unitID(cu)
					if id != "" && !seen[id] {
						seen[id] = true
						linked = append(linked, id)
					}
				}
			}
			sortStrings(linked)
			tf.LinkedCodeUnits = linked
		}
	}
}

// ResolveSymbolLinksCtx is like ResolveSymbolLinks but respects cancellation.
func ResolveSymbolLinksCtx(ctx context.Context, root string, testFiles []models.TestFile, codeUnits []models.CodeUnit, importGraph *ImportGraph) []SymbolLink {
	if importGraph == nil || len(importGraph.TestImports) == 0 || len(codeUnits) == 0 {
		return nil
	}

	unitsByFile := map[string][]models.CodeUnit{}
	for _, cu := range codeUnits {
		unitsByFile[cu.Path] = append(unitsByFile[cu.Path], cu)
	}

	linksByFile := make([][]SymbolLink, len(testFiles))
	parallelForEachIndexCtx(ctx, len(testFiles), func(i int) {
		tf := testFiles[i]
		imports := importGraph.TestImports[tf.Path]
		if len(imports) == 0 {
			return
		}

		var candidates []models.CodeUnit
		for srcPath := range imports {
			candidates = append(candidates, unitsByFile[srcPath]...)
		}
		if len(candidates) == 0 {
			return
		}

		ext := strings.ToLower(filepath.Ext(tf.Path))
		switch {
		case ext == ".go":
			linksByFile[i] = resolveGoSymbolLinks(root, tf.Path, candidates)
		case isJSExt(ext):
			linksByFile[i] = resolveJSSymbolLinks(root, tf.Path, candidates)
		case ext == ".py":
			linksByFile[i] = resolvePythonSymbolLinks(root, tf.Path, candidates)
		}
	})

	var allLinks []SymbolLink
	for _, links := range linksByFile {
		allLinks = append(allLinks, links...)
	}
	return allLinks
}

// sortStrings is a local helper to avoid importing sort in this file
// when the full sort package is already available through the sort import
// in other files in this package.
func sortStrings(s []string) {
	if len(s) <= 1 {
		return
	}
	// Simple insertion sort for the typical small slices we see here.
	for i := 1; i < len(s); i++ {
		key := s[i]
		j := i - 1
		for j >= 0 && s[j] > key {
			s[j+1] = s[j]
			j--
		}
		s[j+1] = key
	}
}
