package analysis

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/pmclSF/terrain/internal/models"
)

// InferAIContextSurfaces performs content-based detection of AI context surfaces
// that may not be detectable by exported symbol names alone.
//
// Detection targets:
//   - String literals used as system messages or prompts (message array patterns)
//   - Template files (.hbs, .j2, .tmpl, .mustache, .prompt) containing AI markers
//   - LangChain/LlamaIndex message builder patterns
//   - YAML/JSON configs containing model instructions
//
// Each detection requires two or more corroborating signals to avoid false positives.
// Every detected surface includes a Reason explaining the classification.
func InferAIContextSurfaces(root string, testFiles []models.TestFile, existing []models.CodeSurface) []models.CodeSurface {
	return InferAIContextSurfacesFromList(root, testFiles, existing, collectSourceFiles(root))
}

// InferAIContextSurfacesFromList is like InferAIContextSurfaces but uses a
// pre-collected file list to avoid redundant directory walks.
func InferAIContextSurfacesFromList(root string, testFiles []models.TestFile, existing []models.CodeSurface, sourceFiles []string) []models.CodeSurface {
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

		content, err := os.ReadFile(filepath.Join(root, relPath))
		if err != nil {
			continue
		}
		src := string(content)

		// Pass 0: AST-level prompt detection (highest priority).
		// Uses real AST for Go, deep structural analysis for JS/Python.
		// Detects message arrays, framework constructors, template factories,
		// system prompt assignments, and few-shot arrays.
		astSurfaces := ParsePromptAST(relPath, src, lang)
		for _, s := range astSurfaces {
			if !existingIDs[s.SurfaceID] {
				existingIDs[s.SurfaceID] = true
				surfaces = append(surfaces, s)
			}
		}

		// Pass 1a: Bracket-aware structural parsing.
		structural := ParseStructural(relPath, src, lang)
		for _, s := range structural {
			if !existingIDs[s.SurfaceID] {
				existingIDs[s.SurfaceID] = true
				surfaces = append(surfaces, s)
			}
		}

		// Pass 1a2: Regex-based prompt parsing (supplements structural).
		parsed := ParseEmbeddedPrompts(relPath, src, lang)
		for _, s := range parsed {
			if !existingIDs[s.SurfaceID] {
				existingIDs[s.SurfaceID] = true
				surfaces = append(surfaces, s)
			}
		}

		// Pass 1b: Structured RAG pipeline parsing (vector stores, splitters, etc.).
		ragSurfaces := ParseRAGPipeline(relPath, src, lang)
		for _, s := range ragSurfaces {
			if !existingIDs[s.SurfaceID] {
				existingIDs[s.SurfaceID] = true
				surfaces = append(surfaces, s)
			}
		}

		// Pass 1c: Structured schema/contract parsing (Zod, Pydantic, OpenAI tools).
		schemas := ParseToolSchemas(relPath, src, lang)
		for _, s := range schemas {
			if !existingIDs[s.SurfaceID] {
				existingIDs[s.SurfaceID] = true
				surfaces = append(surfaces, s)
			}
		}

		// Pass 1c: Framework-aware content inference (LangChain, LlamaIndex, etc.).
		inferred := inferFromContent(relPath, src, lang)
		for _, s := range inferred {
			if !existingIDs[s.SurfaceID] {
				existingIDs[s.SurfaceID] = true
				surfaces = append(surfaces, s)
			}
		}
	}

	// Pass 2: Detect AI template files by extension + content.
	surfaces = append(surfaces, detectTemplateFiles(root, existingIDs)...)

	// Pass 3: Detect RAG config files (YAML/JSON with retrieval settings).
	surfaces = append(surfaces, detectRAGConfigFiles(root, existingIDs)...)

	return surfaces
}

// --- Content-based inference ---

// Message array patterns: { role: "system", content: "..." }
// These appear in OpenAI, Anthropic, LangChain, and custom message builders.
var (
	// JS/TS: { role: "system", content: "..." } or new SystemMessage("...")
	jsMessageArrayPattern = regexp.MustCompile(`\{\s*role\s*:\s*["']system["']`)
	jsLangChainSystem     = regexp.MustCompile(`new\s+(?:SystemMessage|HumanMessage|AIMessage|SystemMessagePromptTemplate)\s*\(`)
	jsLlamaIndex          = regexp.MustCompile(`\b(?:ChatMessage|MessageRole)\b.*(?:SYSTEM|system)`)

	// Python: {"role": "system", "content": "..."} or SystemMessage(content="...")
	pyMessageArrayPattern = regexp.MustCompile(`["']role["']\s*:\s*["']system["']`)
	pyLangChainSystem     = regexp.MustCompile(`\b(?:SystemMessage|HumanMessage|AIMessage|SystemMessagePromptTemplate)\s*\(`)
	pyLlamaIndex          = regexp.MustCompile(`\b(?:ChatMessage|MessageRole)\b.*(?:SYSTEM|system)`)

	// RAG framework patterns in source code.
	jsRAGFrameworkPattern = regexp.MustCompile(`\b(?:VectorStoreRetriever|RetrievalQAChain|ConversationalRetrievalChain|PineconeClient|WeaviateClient|ChromaClient|QdrantClient|createRetriever|similaritySearch|asRetriever|RecursiveCharacterTextSplitter|CharacterTextSplitter|TokenTextSplitter)\b`)
	pyRAGFrameworkPattern = regexp.MustCompile(`\b(?:VectorStoreRetriever|RetrievalQA|ConversationalRetrievalChain|Pinecone|Weaviate|Chroma|Qdrant|FAISS|as_retriever|similarity_search|RecursiveCharacterTextSplitter|CharacterTextSplitter|TokenTextSplitter|from\s+langchain\.retrievers|from\s+llama_index\.retrievers)\b`)

	// Format string / f-string prompt assembly: f"You are {role}..." or `You are ${role}...`
	// Requires AI-related content to distinguish from non-AI format strings.
	jsTemplateLiteralAI = regexp.MustCompile("(?s)`[^`]{20,}`")
	pyFStringAI         = regexp.MustCompile(`(?s)f["'][^"']{20,}["']`)

	// AI instruction indicators in string content (used as secondary signal).
	aiInstructionMarkers = regexp.MustCompile(`(?i)\b(you are a|you are an|as an ai|as a helpful|respond with|do not|always respond|your (role|task|job) is|instructions?:|system:)\b`)
)

func inferFromContent(relPath, src, lang string) []models.CodeSurface {
	var surfaces []models.CodeSurface
	pkg := inferSurfacePackage(relPath)
	lines := strings.Split(src, "\n")

	// Heuristic 1: Message array with role:"system" (high confidence).
	var msgPattern, lcPattern, llPattern *regexp.Regexp
	switch lang {
	case "js":
		msgPattern = jsMessageArrayPattern
		lcPattern = jsLangChainSystem
		llPattern = jsLlamaIndex
	case "python":
		msgPattern = pyMessageArrayPattern
		lcPattern = pyLangChainSystem
		llPattern = pyLlamaIndex
	default:
		return nil
	}

	// Track what we've found to require corroboration.
	hasMessageArray := msgPattern.MatchString(src)
	hasLangChain := lcPattern.MatchString(src)
	hasLlamaIndex := llPattern.MatchString(src)
	hasAIMarkers := aiInstructionMarkers.MatchString(src)

	// Heuristic 2: RAG framework usage patterns in source code.
	hasRAGFramework := ragFrameworkPattern(lang).MatchString(src)

	// Only proceed if we have at least one structural AI pattern.
	if !hasMessageArray && !hasLangChain && !hasLlamaIndex && !hasRAGFramework {
		return nil
	}

	// Find specific locations for message array patterns.
	if hasMessageArray {
		for i, line := range lines {
			if msgPattern.MatchString(line) {
				reason := "[" + models.DetectorContentMarkers + "] message array with role:\"system\" detected"
				if hasAIMarkers {
					reason += "; contains AI instruction markers"
				}
				sid := models.BuildSurfaceID(relPath, "system_message_L"+itoa(i+1), "")
				surfaces = append(surfaces, models.CodeSurface{
					SurfaceID: sid,
					Name:      "system_message",
					Path:      relPath,
					Kind:      models.SurfaceContext,
					Language:  lang,
					Package:   pkg,
					Line:      i + 1,
					Exported:  false,
					DetectionTier: models.TierContent, Confidence: 0.75, Reason: reason,
				})
				break // One per file to avoid noise.
			}
		}
	}

	// LangChain/LlamaIndex message builders.
	if hasLangChain {
		for i, line := range lines {
			if lcPattern.MatchString(line) {
				surfaces = append(surfaces, models.CodeSurface{
					SurfaceID: models.BuildSurfaceID(relPath, "langchain_message_L"+itoa(i+1), ""),
					Name:      "langchain_message",
					Path:      relPath,
					Kind:      models.SurfaceContext,
					Language:  lang,
					Package:   pkg,
					Line:      i + 1,
					Exported:  false,
					DetectionTier: models.TierSemantic, Confidence: 0.9, Reason: "[" + models.DetectorLangChainConstructor + "] LangChain message constructor (SystemMessage/HumanMessage)",
				})
				break
			}
		}
	}

	if hasLlamaIndex && !hasLangChain {
		for i, line := range lines {
			if llPattern.MatchString(line) {
				surfaces = append(surfaces, models.CodeSurface{
					SurfaceID: models.BuildSurfaceID(relPath, "llamaindex_message_L"+itoa(i+1), ""),
					Name:      "llamaindex_message",
					Path:      relPath,
					Kind:      models.SurfaceContext,
					Language:  lang,
					Package:   pkg,
					Line:      i + 1,
					Exported:  false,
					DetectionTier: models.TierSemantic, Confidence: 0.85, Reason: "[" + models.DetectorLlamaIndexConstructor + "] LlamaIndex ChatMessage with SYSTEM role",
				})
				break
			}
		}
	}

	// RAG framework usage in source — detect retriever/splitter/vector store patterns.
	if hasRAGFramework {
		ragPat := ragFrameworkPattern(lang)
		for i, line := range lines {
			if ragPat.MatchString(line) {
				// Classify the specific RAG component.
				name, reason := classifyRAGLine(line)
				sid := models.BuildSurfaceID(relPath, name+"_L"+itoa(i+1), "")
				surfaces = append(surfaces, models.CodeSurface{
					SurfaceID: sid,
					Name:      name,
					Path:      relPath,
					Kind:      models.SurfaceRetrieval,
					Language:  lang,
					Package:   pkg,
					Line:      i + 1,
					Exported:  false,
					DetectionTier: models.TierContent, Confidence: 0.75, Reason: reason,
				})
				break // One per file to avoid noise.
			}
		}
	}

	return surfaces
}

func ragFrameworkPattern(lang string) *regexp.Regexp {
	switch lang {
	case "js":
		return jsRAGFrameworkPattern
	case "python":
		return pyRAGFrameworkPattern
	default:
		return regexp.MustCompile(`$^`) // never matches
	}
}

// classifyRAGLine determines the specific RAG component from a source line.
func classifyRAGLine(line string) (name, reason string) {
	lower := strings.ToLower(line)
	switch {
	case strings.Contains(lower, "textsplitter") || strings.Contains(lower, "chunk"):
		return "chunking_config", "text splitter / chunking strategy detected"
	case strings.Contains(lower, "reranker") || strings.Contains(lower, "rerank"):
		return "reranker_config", "reranker configuration detected"
	case strings.Contains(lower, "retriever") || strings.Contains(lower, "as_retriever") || strings.Contains(lower, "asretriever"):
		return "retriever_config", "retriever instantiation or configuration detected"
	case strings.Contains(lower, "pinecone") || strings.Contains(lower, "weaviate") || strings.Contains(lower, "chroma") || strings.Contains(lower, "qdrant") || strings.Contains(lower, "faiss"):
		return "vector_store_config", "vector store client or configuration detected"
	case strings.Contains(lower, "similaritysearch") || strings.Contains(lower, "similarity_search"):
		return "retrieval_query", "similarity search / retrieval query detected"
	default:
		return "rag_component", "RAG framework component detected"
	}
}

// --- Template file detection ---

// AI template file extensions.
var templateExts = map[string]bool{
	".hbs":      true, // Handlebars
	".j2":       true, // Jinja2
	".jinja":    true, // Jinja2
	".jinja2":   true, // Jinja2
	".tmpl":     true, // Go templates
	".mustache": true, // Mustache
	".prompt":   true, // Prompt-specific files
}

func detectTemplateFiles(root string, existingIDs map[string]bool) []models.CodeSurface {
	var surfaces []models.CodeSurface

	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			if info != nil && info.IsDir() {
				base := filepath.Base(path)
				if base == "node_modules" || base == ".git" || base == "vendor" || base == "__pycache__" {
					return filepath.SkipDir
				}
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !templateExts[ext] {
			return nil
		}

		relPath, _ := filepath.Rel(root, path)
		relPath = filepath.ToSlash(relPath)

		content, err := os.ReadFile(path)
		if err != nil || len(content) == 0 {
			return nil
		}
		src := string(content)

		// Template file must contain AI-related content to qualify.
		// Require at least one AI instruction marker.
		if !aiInstructionMarkers.MatchString(src) {
			return nil
		}

		sid := models.BuildSurfaceID(relPath, filepath.Base(relPath), "")
		if existingIDs[sid] {
			return nil
		}

		reason := "template file (" + ext + ") containing AI instruction markers"
		existingIDs[sid] = true
		surfaces = append(surfaces, models.CodeSurface{
			SurfaceID: sid,
			Name:      filepath.Base(relPath),
			Path:      relPath,
			Kind:      models.SurfaceContext,
			Language:  "template",
			Package:   inferSurfacePackage(relPath),
			Line:      1,
			Exported:  true,
			DetectionTier: models.TierContent, Confidence: 0.75, Reason: reason,
		})
		return nil
	})

	return surfaces
}

// --- RAG config file detection ---

// RAG config keys that indicate retrieval pipeline configuration.
// Requires 2+ keys from this set to qualify.
var ragConfigMarkers = regexp.MustCompile(`(?i)\b(chunk_size|chunk_overlap|chunking_strategy|embedding_model|vector_store|vector_db|top_k|num_results|retrieval_filter|reranker|rerank_model|similarity_threshold|context_window|max_tokens|retrieval_mode|search_type|index_type|collection_name)\b`)

// ragConfigExts are file extensions that may contain RAG config.
var ragConfigExts = map[string]bool{
	".yaml": true, ".yml": true, ".json": true, ".toml": true,
}

func detectRAGConfigFiles(root string, existingIDs map[string]bool) []models.CodeSurface {
	var surfaces []models.CodeSurface

	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			if info != nil && info.IsDir() {
				base := filepath.Base(path)
				if base == "node_modules" || base == ".git" || base == "vendor" || base == "__pycache__" || base == ".terrain" {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// Skip large files (configs should be small).
		if info.Size() > 64*1024 {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !ragConfigExts[ext] {
			return nil
		}

		relPath, _ := filepath.Rel(root, path)
		relPath = filepath.ToSlash(relPath)

		// Skip package.json, go.mod, etc. — only detect dedicated config files.
		base := strings.ToLower(filepath.Base(relPath))
		if base == "package.json" || base == "package-lock.json" || base == "go.mod" || base == "go.sum" || base == "tsconfig.json" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil || len(content) == 0 {
			return nil
		}
		src := string(content)

		// Count RAG-related keys. Require 2+ to avoid false positives on
		// generic config files that happen to mention "top_k" or "max_tokens".
		matches := ragConfigMarkers.FindAllString(src, -1)
		if len(matches) < 2 {
			return nil
		}

		// Deduplicate match keys for the reason string.
		seen := map[string]bool{}
		var uniqueKeys []string
		for _, m := range matches {
			lower := strings.ToLower(m)
			if !seen[lower] {
				seen[lower] = true
				uniqueKeys = append(uniqueKeys, lower)
			}
		}

		sid := models.BuildSurfaceID(relPath, filepath.Base(relPath), "")
		if existingIDs[sid] {
			return nil
		}

		reason := "RAG config file with " + itoa(len(uniqueKeys)) + " retrieval key(s): " + strings.Join(uniqueKeys, ", ")
		existingIDs[sid] = true
		surfaces = append(surfaces, models.CodeSurface{
			SurfaceID: sid,
			Name:      filepath.Base(relPath),
			Path:      relPath,
			Kind:      models.SurfaceRetrieval,
			Language:  "config",
			Package:   inferSurfacePackage(relPath),
			Line:      1,
			Exported:  true,
			DetectionTier: models.TierContent, Confidence: 0.75, Reason: reason,
		})
		return nil
	})

	return surfaces
}

func itoa(n int) string {
	return strconv.Itoa(n)
}
