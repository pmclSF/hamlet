package analysis

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/pmclSF/terrain/internal/models"
)

// DetectorID constants — aliases to the canonical constants in models.
const (
	DetectorRAGRetriever     = models.DetectorRAGRetriever
	DetectorRAGEmbedding     = models.DetectorRAGEmbedding
	DetectorRAGChunking      = models.DetectorRAGChunking
	DetectorRAGVectorStore   = models.DetectorRAGVectorStore
	DetectorRAGReranker      = models.DetectorRAGReranker
	DetectorRAGQueryBuilder  = models.DetectorRAGQueryBuilder
	DetectorRAGDocLoader     = models.DetectorRAGDocLoader
	DetectorRAGCitation      = models.DetectorRAGCitation
	DetectorRAGContextWindow = models.DetectorRAGContextWindow
)

// ParseRAGStructured performs deep structural detection of RAG pipeline
// components with configuration extraction. This supplements the existing
// ParseRAGPipeline with:
//   - Concrete parameter extraction (chunk_size=500, top_k=5, model="...")
//   - New component types (query builders, document loaders, citation assembly)
//   - Cross-reference to CodeSurface IDs for graph linkage
//   - RAGPipelineSurface nodes with structured config metadata
func ParseRAGStructured(relPath, src, lang string) []models.RAGPipelineSurface {
	switch lang {
	case "js":
		return parseRAGStructuredJS(relPath, src)
	case "python":
		return parseRAGStructuredPython(relPath, src)
	default:
		return nil
	}
}

// --- JS/TS structured RAG detection ---

var (
	// Document loaders: new PDFLoader, new TextLoader, new DirectoryLoader, etc.
	jsDocLoaderPattern = regexp.MustCompile(`new\s+(PDFLoader|TextLoader|DirectoryLoader|CSVLoader|JSONLoader|CheerioWebBaseLoader|PuppeteerWebBaseLoader|NotionLoader|GithubRepoLoader|UnstructuredLoader|S3Loader)\s*\(`)

	// Query builder patterns: MultiQueryRetriever, QueryTransformationChain
	jsQueryBuilderPattern = regexp.MustCompile(`(?:new\s+)?(MultiQueryRetriever|QueryTransformationChain|HydeRetriever|StepBackPromptRetriever|ContextualCompressionRetriever)(?:\.from\w+)?\s*\(`)

	// Citation patterns: formatDocumentsAsString, createStuffDocumentsChain, sources
	jsCitationPattern = regexp.MustCompile(`\b(formatDocumentsAsString|createStuffDocumentsChain|createRetrievalChain|createCitationChain)\s*\(`)

	// Context window assembly: ContextualCompressionRetriever, maxTokens, contextWindow
	jsContextWindowPattern = regexp.MustCompile(`\b(ContextualCompressionRetriever|DocumentCompressor|LLMChainExtractor)\s*\(`)

	// Config value extractors
	jsChunkSizePattern   = regexp.MustCompile(`chunkSize\s*:\s*(\d+)`)
	jsChunkOverlapPattern = regexp.MustCompile(`chunkOverlap\s*:\s*(\d+)`)
	jsTopKPattern        = regexp.MustCompile(`(?:k|topK|topN|top_k|top_n)\s*:\s*(\d+)`)
	jsModelNamePattern   = regexp.MustCompile(`model(?:Name)?\s*:\s*["']([^"']+)["']`)
	jsSearchTypePattern  = regexp.MustCompile(`(?:searchType|search_type)\s*:\s*["']([^"']+)["']`)
	jsPersistDirPattern  = regexp.MustCompile(`(?:persistDir|persist_directory|directory)\s*:\s*["']([^"']+)["']`)
)

func parseRAGStructuredJS(relPath, src string) []models.RAGPipelineSurface {
	var components []models.RAGPipelineSurface
	lines := strings.Split(src, "\n")
	seen := map[string]bool{}

	add := func(c models.RAGPipelineSurface) {
		if seen[c.ComponentID] {
			return
		}
		seen[c.ComponentID] = true
		components = append(components, c)
	}

	for i, line := range lines {
		window := buildWindow(lines, i, 5)

		// Vector store constructors (enhanced with config extraction).
		if m := jsVectorStoreConstructor.FindStringSubmatch(line); m != nil {
			config := extractJSConfig(window)
			config.Provider = strings.ToLower(m[1])
			add(models.RAGPipelineSurface{
				ComponentID:   models.BuildRAGComponentID(relPath, models.RAGVectorStore, strings.ToLower(m[1])),
				Name:          "vector_store_" + strings.ToLower(m[1]),
				Path:          relPath,
				Kind:          models.RAGVectorStore,
				Framework:     inferJSRAGFramework(line),
				ClassName:     m[1],
				Language:      "js",
				Line:          i + 1,
				Config:        config,
				DetectionTier: models.TierSemantic,
				Confidence:    0.95,
				Reason:        "[" + DetectorRAGVectorStore + "] " + m[1] + " constructor with provider " + config.Provider,
			})
		}

		// Text splitter with config extraction.
		if m := jsTextSplitter.FindStringSubmatch(line); m != nil {
			config := extractJSConfig(window)
			reason := "[" + DetectorRAGChunking + "] " + m[1]
			if config.ChunkSize > 0 {
				reason += " (chunkSize=" + strconv.Itoa(config.ChunkSize) + ")"
			}
			if config.ChunkOverlap > 0 {
				reason += " (chunkOverlap=" + strconv.Itoa(config.ChunkOverlap) + ")"
			}
			add(models.RAGPipelineSurface{
				ComponentID:   models.BuildRAGComponentID(relPath, models.RAGChunking, "text_splitter"),
				Name:          "text_splitter",
				Path:          relPath,
				Kind:          models.RAGChunking,
				Framework:     inferJSRAGFramework(line),
				ClassName:     m[1],
				Language:      "js",
				Line:          i + 1,
				Config:        config,
				DetectionTier: models.TierSemantic,
				Confidence:    0.93,
				Reason:        reason,
			})
		}

		// Retriever construction with config.
		if jsRetrieverConstruction.MatchString(line) {
			config := extractJSConfig(window)
			reason := "[" + DetectorRAGRetriever + "] retriever construction"
			if config.TopK > 0 {
				reason += " (topK=" + strconv.Itoa(config.TopK) + ")"
			}
			if config.SearchType != "" {
				reason += " (searchType=" + config.SearchType + ")"
			}
			add(models.RAGPipelineSurface{
				ComponentID:   models.BuildRAGComponentID(relPath, models.RAGRetriever, "retriever"),
				Name:          "retriever_config",
				Path:          relPath,
				Kind:          models.RAGRetriever,
				Framework:     inferJSRAGFramework(line),
				Language:      "js",
				Line:          i + 1,
				Config:        config,
				DetectionTier: models.TierSemantic,
				Confidence:    0.92,
				Reason:        reason,
			})
		}

		// Reranker with config.
		if m := jsReranker.FindStringSubmatch(line); m != nil {
			config := extractJSConfig(window)
			reason := "[" + DetectorRAGReranker + "] " + m[1] + " reranker"
			if config.TopK > 0 {
				reason += " (topN=" + strconv.Itoa(config.TopK) + ")"
			}
			add(models.RAGPipelineSurface{
				ComponentID:   models.BuildRAGComponentID(relPath, models.RAGReranker, "reranker"),
				Name:          "reranker_config",
				Path:          relPath,
				Kind:          models.RAGReranker,
				Framework:     inferJSRAGFramework(line),
				ClassName:     m[1],
				Language:      "js",
				Line:          i + 1,
				Config:        config,
				DetectionTier: models.TierSemantic,
				Confidence:    0.93,
				Reason:        reason,
			})
		}

		// Embedding model with config.
		if m := jsEmbeddingModel.FindStringSubmatch(line); m != nil {
			config := extractJSConfig(window)
			reason := "[" + DetectorRAGEmbedding + "] " + m[1]
			if config.ModelName != "" {
				reason += " (model=" + config.ModelName + ")"
			}
			add(models.RAGPipelineSurface{
				ComponentID:   models.BuildRAGComponentID(relPath, models.RAGEmbedding, "embedding"),
				Name:          "embedding_model",
				Path:          relPath,
				Kind:          models.RAGEmbedding,
				Framework:     inferJSRAGFramework(line),
				ClassName:     m[1],
				Language:      "js",
				Line:          i + 1,
				Config:        config,
				DetectionTier: models.TierSemantic,
				Confidence:    0.92,
				Reason:        reason,
			})
		}

		// Document loaders.
		if m := jsDocLoaderPattern.FindStringSubmatch(line); m != nil {
			add(models.RAGPipelineSurface{
				ComponentID:   models.BuildRAGComponentID(relPath, models.RAGDocumentLoader, strings.ToLower(m[1])),
				Name:          "doc_loader_" + strings.ToLower(m[1]),
				Path:          relPath,
				Kind:          models.RAGDocumentLoader,
				Framework:     inferJSRAGFramework(line),
				ClassName:     m[1],
				Language:      "js",
				Line:          i + 1,
				DetectionTier: models.TierSemantic,
				Confidence:    0.90,
				Reason:        "[" + DetectorRAGDocLoader + "] " + m[1] + " document loader",
			})
		}

		// Query builders.
		if m := jsQueryBuilderPattern.FindStringSubmatch(line); m != nil {
			add(models.RAGPipelineSurface{
				ComponentID:   models.BuildRAGComponentID(relPath, models.RAGQueryBuilder, strings.ToLower(m[1])),
				Name:          "query_builder_" + strings.ToLower(m[1]),
				Path:          relPath,
				Kind:          models.RAGQueryBuilder,
				Framework:     inferJSRAGFramework(line),
				ClassName:     m[1],
				Language:      "js",
				Line:          i + 1,
				DetectionTier: models.TierSemantic,
				Confidence:    0.90,
				Reason:        "[" + DetectorRAGQueryBuilder + "] " + m[1] + " query transformation",
			})
		}

		// Citation assembly.
		if m := jsCitationPattern.FindStringSubmatch(line); m != nil {
			add(models.RAGPipelineSurface{
				ComponentID:   models.BuildRAGComponentID(relPath, models.RAGCitationAssembly, strings.ToLower(m[1])),
				Name:          "citation_" + strings.ToLower(m[1]),
				Path:          relPath,
				Kind:          models.RAGCitationAssembly,
				Framework:     inferJSRAGFramework(line),
				ClassName:     m[1],
				Language:      "js",
				Line:          i + 1,
				DetectionTier: models.TierSemantic,
				Confidence:    0.88,
				Reason:        "[" + DetectorRAGCitation + "] " + m[1] + " citation/source attribution",
			})
		}

		// Context window assembly.
		if m := jsContextWindowPattern.FindStringSubmatch(line); m != nil {
			add(models.RAGPipelineSurface{
				ComponentID:   models.BuildRAGComponentID(relPath, models.RAGContextAssembly, strings.ToLower(m[1])),
				Name:          "context_assembly_" + strings.ToLower(m[1]),
				Path:          relPath,
				Kind:          models.RAGContextAssembly,
				Framework:     inferJSRAGFramework(line),
				ClassName:     m[1],
				Language:      "js",
				Line:          i + 1,
				DetectionTier: models.TierSemantic,
				Confidence:    0.88,
				Reason:        "[" + DetectorRAGContextWindow + "] " + m[1] + " context window assembly",
			})
		}
	}

	return components
}

// --- Python structured RAG detection ---

var (
	// Document loaders
	pyDocLoaderPattern = regexp.MustCompile(`(PyPDFLoader|TextLoader|DirectoryLoader|CSVLoader|JSONLoader|WebBaseLoader|NotionDirectoryLoader|UnstructuredFileLoader|S3FileLoader|GCSFileLoader)\s*\(`)

	// Query builders
	pyQueryBuilderPattern = regexp.MustCompile(`(MultiQueryRetriever|SelfQueryRetriever|ContextualCompressionRetriever|EnsembleRetriever|ParentDocumentRetriever)(?:\.from_\w+)?\s*\(`)

	// Citation patterns
	pyCitationPattern = regexp.MustCompile(`\b(create_stuff_documents_chain|create_retrieval_chain|create_citation_chain|format_docs|StuffDocumentsChain)\s*\(`)

	// Context window / compression
	pyContextWindowPattern = regexp.MustCompile(`(ContextualCompressionRetriever|DocumentCompressorPipeline|LLMChainExtractor|EmbeddingsFilter)\s*\(`)

	// Config value extractors (Python)
	pyChunkSizePattern    = regexp.MustCompile(`chunk_size\s*=\s*(\d+)`)
	pyChunkOverlapPattern = regexp.MustCompile(`chunk_overlap\s*=\s*(\d+)`)
	pyTopKPattern         = regexp.MustCompile(`(?:["']?k["']?|top_k|top_n)\s*[=:]\s*(\d+)`)
	pyModelNamePattern    = regexp.MustCompile(`model(?:_name)?\s*=\s*["']([^"']+)["']`)
	pySearchTypePattern   = regexp.MustCompile(`search_type\s*=\s*["']([^"']+)["']`)
	pyPersistDirPattern   = regexp.MustCompile(`persist_directory\s*=\s*["']([^"']+)["']`)
)

func parseRAGStructuredPython(relPath, src string) []models.RAGPipelineSurface {
	var components []models.RAGPipelineSurface
	lines := strings.Split(src, "\n")
	seen := map[string]bool{}

	add := func(c models.RAGPipelineSurface) {
		if seen[c.ComponentID] {
			return
		}
		seen[c.ComponentID] = true
		components = append(components, c)
	}

	for i, line := range lines {
		window := buildWindow(lines, i, 5)

		// Vector store factory methods.
		if m := pyVectorStoreFactory.FindStringSubmatch(line); m != nil {
			config := extractPyConfig(window)
			config.Provider = strings.ToLower(m[1])
			reason := "[" + DetectorRAGVectorStore + "] " + m[1] + ".from_* construction"
			if config.PersistDir != "" {
				reason += " (persist_directory=" + config.PersistDir + ")"
			}
			add(models.RAGPipelineSurface{
				ComponentID:   models.BuildRAGComponentID(relPath, models.RAGVectorStore, strings.ToLower(m[1])),
				Name:          "vector_store_" + strings.ToLower(m[1]),
				Path:          relPath,
				Kind:          models.RAGVectorStore,
				Framework:     inferPyRAGFramework(line),
				ClassName:     m[1],
				Language:      "python",
				Line:          i + 1,
				Config:        config,
				DetectionTier: models.TierSemantic,
				Confidence:    0.95,
				Reason:        reason,
			})
			continue
		}

		// Text splitter with config extraction.
		if m := pyTextSplitter.FindStringSubmatch(line); m != nil {
			config := extractPyConfig(window)
			reason := "[" + DetectorRAGChunking + "] " + m[1]
			if config.ChunkSize > 0 {
				reason += " (chunk_size=" + strconv.Itoa(config.ChunkSize) + ")"
			}
			if config.ChunkOverlap > 0 {
				reason += " (chunk_overlap=" + strconv.Itoa(config.ChunkOverlap) + ")"
			}
			add(models.RAGPipelineSurface{
				ComponentID:   models.BuildRAGComponentID(relPath, models.RAGChunking, "text_splitter"),
				Name:          "text_splitter",
				Path:          relPath,
				Kind:          models.RAGChunking,
				Framework:     inferPyRAGFramework(line),
				ClassName:     m[1],
				Language:      "python",
				Line:          i + 1,
				Config:        config,
				DetectionTier: models.TierSemantic,
				Confidence:    0.93,
				Reason:        reason,
			})
		}

		// Retriever.
		if pyRetrieverConstruction.MatchString(line) {
			config := extractPyConfig(window)
			reason := "[" + DetectorRAGRetriever + "] .as_retriever() construction"
			if config.TopK > 0 {
				reason += " (top_k=" + strconv.Itoa(config.TopK) + ")"
			}
			if config.SearchType != "" {
				reason += " (search_type=" + config.SearchType + ")"
			}
			add(models.RAGPipelineSurface{
				ComponentID:   models.BuildRAGComponentID(relPath, models.RAGRetriever, "retriever"),
				Name:          "retriever_config",
				Path:          relPath,
				Kind:          models.RAGRetriever,
				Framework:     inferPyRAGFramework(line),
				Language:      "python",
				Line:          i + 1,
				Config:        config,
				DetectionTier: models.TierSemantic,
				Confidence:    0.93,
				Reason:        reason,
			})
		}

		// Reranker.
		if m := pyReranker.FindStringSubmatch(line); m != nil {
			config := extractPyConfig(window)
			reason := "[" + DetectorRAGReranker + "] " + m[1]
			if config.TopK > 0 {
				reason += " (top_n=" + strconv.Itoa(config.TopK) + ")"
			}
			if config.ModelName != "" {
				reason += " (model=" + config.ModelName + ")"
			}
			add(models.RAGPipelineSurface{
				ComponentID:   models.BuildRAGComponentID(relPath, models.RAGReranker, "reranker"),
				Name:          "reranker_config",
				Path:          relPath,
				Kind:          models.RAGReranker,
				Framework:     inferPyRAGFramework(line),
				ClassName:     m[1],
				Language:      "python",
				Line:          i + 1,
				Config:        config,
				DetectionTier: models.TierSemantic,
				Confidence:    0.93,
				Reason:        reason,
			})
		}

		// Embedding model.
		if m := pyEmbeddingModel.FindStringSubmatch(line); m != nil {
			config := extractPyConfig(window)
			reason := "[" + DetectorRAGEmbedding + "] " + m[1]
			if config.ModelName != "" {
				reason += " (model=" + config.ModelName + ")"
			}
			add(models.RAGPipelineSurface{
				ComponentID:   models.BuildRAGComponentID(relPath, models.RAGEmbedding, "embedding"),
				Name:          "embedding_model",
				Path:          relPath,
				Kind:          models.RAGEmbedding,
				Framework:     inferPyRAGFramework(line),
				ClassName:     m[1],
				Language:      "python",
				Line:          i + 1,
				Config:        config,
				DetectionTier: models.TierSemantic,
				Confidence:    0.92,
				Reason:        reason,
			})
		}

		// Document loaders.
		if m := pyDocLoaderPattern.FindStringSubmatch(line); m != nil {
			add(models.RAGPipelineSurface{
				ComponentID:   models.BuildRAGComponentID(relPath, models.RAGDocumentLoader, strings.ToLower(m[1])),
				Name:          "doc_loader_" + strings.ToLower(m[1]),
				Path:          relPath,
				Kind:          models.RAGDocumentLoader,
				Framework:     inferPyRAGFramework(line),
				ClassName:     m[1],
				Language:      "python",
				Line:          i + 1,
				DetectionTier: models.TierSemantic,
				Confidence:    0.90,
				Reason:        "[" + DetectorRAGDocLoader + "] " + m[1] + " document loader",
			})
		}

		// Query builders.
		if m := pyQueryBuilderPattern.FindStringSubmatch(line); m != nil {
			add(models.RAGPipelineSurface{
				ComponentID:   models.BuildRAGComponentID(relPath, models.RAGQueryBuilder, strings.ToLower(m[1])),
				Name:          "query_builder_" + strings.ToLower(m[1]),
				Path:          relPath,
				Kind:          models.RAGQueryBuilder,
				Framework:     inferPyRAGFramework(line),
				ClassName:     m[1],
				Language:      "python",
				Line:          i + 1,
				DetectionTier: models.TierSemantic,
				Confidence:    0.90,
				Reason:        "[" + DetectorRAGQueryBuilder + "] " + m[1] + " query transformation",
			})
		}

		// Citation assembly.
		if m := pyCitationPattern.FindStringSubmatch(line); m != nil {
			add(models.RAGPipelineSurface{
				ComponentID:   models.BuildRAGComponentID(relPath, models.RAGCitationAssembly, strings.ToLower(m[1])),
				Name:          "citation_" + strings.ToLower(m[1]),
				Path:          relPath,
				Kind:          models.RAGCitationAssembly,
				Framework:     inferPyRAGFramework(line),
				ClassName:     m[1],
				Language:      "python",
				Line:          i + 1,
				DetectionTier: models.TierSemantic,
				Confidence:    0.88,
				Reason:        "[" + DetectorRAGCitation + "] " + m[1] + " citation/source attribution",
			})
		}

		// Context window assembly.
		if m := pyContextWindowPattern.FindStringSubmatch(line); m != nil {
			add(models.RAGPipelineSurface{
				ComponentID:   models.BuildRAGComponentID(relPath, models.RAGContextAssembly, strings.ToLower(m[1])),
				Name:          "context_assembly_" + strings.ToLower(m[1]),
				Path:          relPath,
				Kind:          models.RAGContextAssembly,
				Framework:     inferPyRAGFramework(line),
				ClassName:     m[1],
				Language:      "python",
				Line:          i + 1,
				DetectionTier: models.TierSemantic,
				Confidence:    0.88,
				Reason:        "[" + DetectorRAGContextWindow + "] " + m[1] + " context compression",
			})
		}
	}

	return components
}

// --- Config extraction ---

func extractJSConfig(window string) models.RAGComponentConfig {
	config := models.RAGComponentConfig{}
	if m := jsChunkSizePattern.FindStringSubmatch(window); m != nil {
		config.ChunkSize, _ = strconv.Atoi(m[1])
	}
	if m := jsChunkOverlapPattern.FindStringSubmatch(window); m != nil {
		config.ChunkOverlap, _ = strconv.Atoi(m[1])
	}
	if m := jsTopKPattern.FindStringSubmatch(window); m != nil {
		config.TopK, _ = strconv.Atoi(m[1])
	}
	if m := jsModelNamePattern.FindStringSubmatch(window); m != nil {
		config.ModelName = m[1]
	}
	if m := jsSearchTypePattern.FindStringSubmatch(window); m != nil {
		config.SearchType = m[1]
	}
	if m := jsPersistDirPattern.FindStringSubmatch(window); m != nil {
		config.PersistDir = m[1]
	}
	return config
}

func extractPyConfig(window string) models.RAGComponentConfig {
	config := models.RAGComponentConfig{}
	if m := pyChunkSizePattern.FindStringSubmatch(window); m != nil {
		config.ChunkSize, _ = strconv.Atoi(m[1])
	}
	if m := pyChunkOverlapPattern.FindStringSubmatch(window); m != nil {
		config.ChunkOverlap, _ = strconv.Atoi(m[1])
	}
	if m := pyTopKPattern.FindStringSubmatch(window); m != nil {
		config.TopK, _ = strconv.Atoi(m[1])
	}
	if m := pyModelNamePattern.FindStringSubmatch(window); m != nil {
		config.ModelName = m[1]
	}
	if m := pySearchTypePattern.FindStringSubmatch(window); m != nil {
		config.SearchType = m[1]
	}
	if m := pyPersistDirPattern.FindStringSubmatch(window); m != nil {
		config.PersistDir = m[1]
	}
	return config
}

// --- Framework inference ---

func inferJSRAGFramework(line string) string {
	lower := strings.ToLower(line)
	switch {
	case strings.Contains(lower, "langchain"):
		return "langchain"
	case strings.Contains(lower, "llamaindex") || strings.Contains(lower, "llama_index"):
		return "llamaindex"
	default:
		return ""
	}
}

func inferPyRAGFramework(line string) string {
	lower := strings.ToLower(line)
	switch {
	case strings.Contains(lower, "langchain"):
		return "langchain"
	case strings.Contains(lower, "llama_index") || strings.Contains(lower, "llamaindex"):
		return "llamaindex"
	default:
		return ""
	}
}

// --- Helpers ---

// buildWindow constructs a multi-line window for config extraction.
func buildWindow(lines []string, startLine, windowSize int) string {
	var b strings.Builder
	for w := 0; w < windowSize && startLine+w < len(lines); w++ {
		b.WriteString(lines[startLine+w])
		b.WriteByte('\n')
	}
	return b.String()
}

// LinkRAGSurfacesToCodeSurfaces links RAGPipelineSurface components to their
// corresponding CodeSurface entries by matching file path and line proximity.
func LinkRAGSurfacesToCodeSurfaces(ragComponents []models.RAGPipelineSurface, codeSurfaces []models.CodeSurface) {
	// Index code surfaces by path.
	surfacesByPath := map[string][]models.CodeSurface{}
	for _, cs := range codeSurfaces {
		if cs.Kind == models.SurfaceRetrieval {
			surfacesByPath[cs.Path] = append(surfacesByPath[cs.Path], cs)
		}
	}

	for i := range ragComponents {
		rc := &ragComponents[i]
		candidates := surfacesByPath[rc.Path]
		if len(candidates) == 0 {
			continue
		}

		// Find closest CodeSurface by line proximity.
		bestDist := 1000
		bestID := ""
		for _, cs := range candidates {
			dist := abs(cs.Line - rc.Line)
			if dist < bestDist {
				bestDist = dist
				bestID = cs.SurfaceID
			}
		}
		if bestDist <= 5 {
			rc.LinkedSurfaceID = bestID
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
