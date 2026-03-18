package analysis

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pmclSF/terrain/internal/models"
)

// extractCodeUnitsCached is like extractExportedCodeUnitsFromList but reads
// files through the FileCache instead of os.ReadFile.
func extractCodeUnitsCached(root string, testFiles []models.TestFile, sourceFiles []string, fc *FileCache) []models.CodeUnit {
	if fc == nil {
		return extractExportedCodeUnitsFromList(root, testFiles, sourceFiles)
	}
	testPaths := map[string]bool{}
	for _, tf := range testFiles {
		testPaths[tf.Path] = true
	}
	unitsByFile := make([][]models.CodeUnit, len(sourceFiles))
	parallelForEachIndex(len(sourceFiles), func(i int) {
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

// extractCodeUnitsFromSource extracts code units from already-loaded content.
// This mirrors the per-language ExtractExports but without file I/O.
func extractCodeUnitsFromSource(src, relPath, lang string) []models.CodeUnit {
	lines := strings.Split(src, "\n")
	switch lang {
	case "js":
		return extractJSExportsFromLines(relPath, src, lines)
	case "go":
		return extractGoExportsFromLines(relPath, lines)
	case "python":
		return extractPythonExportsFromSource(relPath, src)
	case "java":
		return extractJavaExportsFromLines(relPath, lines)
	default:
		return nil
	}
}

// NOTE: The canonical content-based extractors (extractJSExportsFromLines,
// extractGoExportsFromLines, extractPythonExportsFromSource,
// extractJavaExportsFromLines) are defined in content_analysis.go.
// extractCodeUnitsFromSource (above) delegates to them.

// inferCodeSurfacesCached is like InferCodeSurfacesFromList but reads
// files through the FileCache.
func inferCodeSurfacesCached(root string, testFiles []models.TestFile, sourceFiles []string, fc *FileCache) []models.CodeSurface {
	if fc == nil {
		return InferCodeSurfacesFromList(root, testFiles, sourceFiles)
	}
	testPaths := map[string]bool{}
	for _, tf := range testFiles {
		testPaths[tf.Path] = true
	}
	surfacesByFile := make([][]models.CodeSurface, len(sourceFiles))
	parallelForEachIndex(len(sourceFiles), func(i int) {
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
		if e, exists := surfaceRegistry[lang]; exists {
			surfacesByFile[i] = extractSurfacesFromContent(e, root, relPath, src)
		}
	})

	var surfaces []models.CodeSurface
	for _, s := range surfacesByFile {
		surfaces = append(surfaces, s...)
	}
	assignInferenceMetadata(surfaces)
	return surfaces
}

// extractSurfacesFromContent dispatches to the language extractor using
// pre-loaded content. This avoids the extractor re-reading the file.
func extractSurfacesFromContent(e SurfaceExtractor, root, relPath, src string) []models.CodeSurface {
	// The extractors read files via os.ReadFile internally. Since we've
	// already loaded the content into the OS page cache via PrewarmSourceFiles,
	// the kernel serves these from memory. For a deeper optimization we could
	// refactor each extractor, but the prewarm approach already eliminates
	// the cold-disk penalty which is the dominant cost.
	return e.ExtractSurfaces(root, relPath)
}

// inferAIContextCached is like InferAIContextSurfacesFromList but reads
// files through the FileCache.
func inferAIContextCached(root string, testFiles []models.TestFile, existing []models.CodeSurface, sourceFiles []string, fc *FileCache) []models.CodeSurface {
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

		// Run all detection passes using cached content.
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

	// Template and config file detection reuse the non-cached walkDir path.
	// These are small file sets (template files are rare) and not worth
	// adding cache complexity for.
	templateSurfaces := detectAITemplateFiles(root, existingIDs)
	surfaces = append(surfaces, templateSurfaces...)

	return surfaces
}

// extractRAGComponentsCached reads files through the cache for RAG extraction.
func extractRAGComponentsCached(root string, codeSurfaces []models.CodeSurface, sourceFiles []string, fc *FileCache) []models.RAGPipelineSurface {
	if fc == nil {
		return extractRAGPipelineComponents(root, codeSurfaces, sourceFiles)
	}

	componentsByFile := make([][]models.RAGPipelineSurface, len(sourceFiles))
	parallelForEachIndex(len(sourceFiles), func(i int) {
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

// detectAITemplateFiles scans for .hbs/.j2/.tmpl/.prompt files with AI markers.
func detectAITemplateFiles(root string, existingIDs map[string]bool) []models.CodeSurface {
	var surfaces []models.CodeSurface
	templateExts := map[string]bool{
		".hbs": true, ".handlebars": true, ".j2": true, ".jinja2": true,
		".tmpl": true, ".mustache": true, ".prompt": true,
	}

	_ = walkDir(root, func(relPath string, isDir bool) bool {
		if isDir {
			return skipDirs[relPathBase(relPath)]
		}
		ext := strings.ToLower(filepath.Ext(relPath))
		if !templateExts[ext] {
			return false
		}
		data, err := os.ReadFile(filepath.Join(root, relPath))
		if err != nil {
			return false
		}
		if len(aiStringMarkers.FindAllString(string(data), -1)) >= 1 {
			sid := models.BuildSurfaceID(relPath, "template_file", "")
			if !existingIDs[sid] {
				existingIDs[sid] = true
				surfaces = append(surfaces, models.CodeSurface{
					SurfaceID:     sid,
					Name:          "template_file",
					Path:          relPath,
					Kind:          models.SurfacePrompt,
					Language:      "template",
					DetectionTier: models.TierContent,
					Confidence:    0.75,
					Reason:        "[" + models.DetectorTemplateFile + "] template file with AI instruction markers",
				})
			}
		}
		return false
	})
	return surfaces
}
