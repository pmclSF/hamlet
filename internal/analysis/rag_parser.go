package analysis

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/pmclSF/terrain/internal/models"
)

// ParseRAGPipeline performs structured detection of RAG pipeline components
// in source code. Unlike flat regex matching on export names, this parser
// identifies framework-specific construction patterns with configuration.
//
// Detections:
//
//   JS/TS:
//   - Vector store constructors (new PineconeClient, new ChromaClient, etc.)
//   - Text splitter instantiation with config (chunkSize, chunkOverlap)
//   - Retriever construction (.asRetriever, createRetriever)
//   - Reranker setup (CohereRerank, cross-encoder config)
//   - Embedding model configuration
//   - Citation/context assembly functions
//
//   Python:
//   - Vector store factory methods (Chroma.from_documents, FAISS.from_texts)
//   - Text splitter with parameters (chunk_size=, chunk_overlap=)
//   - .as_retriever(search_kwargs={...}) with top-k extraction
//   - Reranker instantiation (CohereRerank, CrossEncoder)
//   - Embedding model construction (OpenAIEmbeddings, HuggingFaceEmbeddings)
//
// Each detection carries TierSemantic with evidence metadata.
func ParseRAGPipeline(relPath, src, lang string) []models.CodeSurface {
	switch lang {
	case "js":
		return parseJSRAG(relPath, src)
	case "python":
		return parsePythonRAG(relPath, src)
	default:
		return nil
	}
}

// --- JS/TS RAG detection ---

var (
	// Vector store constructors: new PineconeClient(...), new ChromaClient(...)
	jsVectorStoreConstructor = regexp.MustCompile(`new\s+(Pinecone(?:Client)?|Chroma(?:Client)?|Qdrant(?:Client)?|Weaviate(?:Client)?|Milvus(?:Client)?)\s*\(`)

	// Text splitter with config: new RecursiveCharacterTextSplitter({ chunkSize: ... })
	jsTextSplitter = regexp.MustCompile(`new\s+(RecursiveCharacterTextSplitter|CharacterTextSplitter|TokenTextSplitter|MarkdownTextSplitter)\s*\(`)

	// Retriever construction: .asRetriever(, createRetriever(
	jsRetrieverConstruction = regexp.MustCompile(`\.(asRetriever|as_retriever)\s*\(|createRetriever\s*\(`)

	// Reranker: new CohereRerank, new CrossEncoderReranker
	jsReranker = regexp.MustCompile(`new\s+(CohereRerank(?:er)?|CrossEncoder(?:Reranker)?|FlashrankRerank)\s*\(`)

	// Embedding model: new OpenAIEmbeddings, new HuggingFaceEmbeddings
	jsEmbeddingModel = regexp.MustCompile(`new\s+(OpenAIEmbeddings|HuggingFaceEmbeddings|CohereEmbeddings|VoyageEmbeddings)\s*\(`)

	// Similarity search / query execution
	jsSimilaritySearch = regexp.MustCompile(`\.(similaritySearch|similarity_search|maxMarginalRelevanceSearch)\s*\(`)
)

func parseJSRAG(relPath, src string) []models.CodeSurface {
	var surfaces []models.CodeSurface
	pkg := inferSurfacePackage(relPath)
	seen := map[string]bool{}

	add := func(name, reason string, line int, confidence float64) {
		sid := models.BuildSurfaceID(relPath, name+"_L"+strconv.Itoa(line), "")
		if seen[sid] {
			return
		}
		seen[sid] = true
		surfaces = append(surfaces, models.CodeSurface{
			SurfaceID:     sid,
			Name:          name,
			Path:          relPath,
			Kind:          models.SurfaceRetrieval,
			Language:      "js",
			Package:       pkg,
			Line:          line,
			Exported:      false,
			DetectionTier: models.TierSemantic,
			Confidence:    confidence,
			Reason:        reason,
		})
	}

	lines := strings.Split(src, "\n")

	for i, l := range lines {
		// Vector store constructor.
		if m := jsVectorStoreConstructor.FindStringSubmatch(l); m != nil {
			add("vector_store_"+strings.ToLower(m[1]), m[1]+" vector store constructor", i+1, 0.93)
		}

		// Text splitter with config (look at a 5-line window for config params).
		if m := jsTextSplitter.FindStringSubmatch(l); m != nil {
			reason := m[1] + " instantiation"
			window := l
			for w := 1; w <= 4 && i+w < len(lines); w++ {
				window += lines[i+w]
			}
			if strings.Contains(window, "chunkSize") || strings.Contains(window, "chunk_size") {
				reason += " (with chunk size config)"
			}
			add("text_splitter", reason, i+1, 0.92)
		}

		// Retriever construction.
		if jsRetrieverConstruction.MatchString(l) {
			reason := "retriever construction"
			if strings.Contains(l, "k:") || strings.Contains(l, "k=") {
				reason += " (with top-k config)"
			}
			add("retriever_config", reason, i+1, 0.90)
		}

		// Reranker.
		if m := jsReranker.FindStringSubmatch(l); m != nil {
			add("reranker_config", m[1]+" reranker setup", i+1, 0.92)
		}

		// Embedding model.
		if m := jsEmbeddingModel.FindStringSubmatch(l); m != nil {
			add("embedding_model", m[1]+" embedding model", i+1, 0.90)
		}

		// Similarity search (retrieval query).
		if jsSimilaritySearch.MatchString(l) {
			add("retrieval_query", "similarity search / retrieval query execution", i+1, 0.85)
		}
	}

	return surfaces
}

// --- Python RAG detection ---

var (
	// Vector store factory: Chroma.from_documents(...), FAISS.from_texts(...)
	pyVectorStoreFactory = regexp.MustCompile(`(Chroma|FAISS|Pinecone|Qdrant|Weaviate|Milvus)\.from_(?:documents|texts|embeddings)\s*\(`)

	// Vector store constructor: Chroma(...), FAISS(...)
	pyVectorStoreConstructor = regexp.MustCompile(`(?:^|\s)(Chroma|FAISS|Pinecone|Qdrant|Weaviate|Milvus)\s*\(`)

	// Text splitter with params: RecursiveCharacterTextSplitter(chunk_size=500)
	pyTextSplitter = regexp.MustCompile(`(RecursiveCharacterTextSplitter|CharacterTextSplitter|TokenTextSplitter|MarkdownTextSplitter|SpacyTextSplitter)\s*\(`)

	// Retriever: .as_retriever(search_kwargs=...)
	pyRetrieverConstruction = regexp.MustCompile(`\.as_retriever\s*\(`)

	// Reranker: CohereRerank, CrossEncoderReranker
	pyReranker = regexp.MustCompile(`(CohereRerank|CrossEncoderReranker|FlashrankRerank|SentenceTransformerRerank)\s*\(`)

	// Embedding model: OpenAIEmbeddings(...), HuggingFaceEmbeddings(...)
	pyEmbeddingModel = regexp.MustCompile(`(OpenAIEmbeddings|HuggingFaceEmbeddings|CohereEmbeddings|VoyageEmbeddings|SentenceTransformerEmbeddings)\s*\(`)

	// similarity_search call
	pySimilaritySearch = regexp.MustCompile(`\.similarity_search\s*\(|\.max_marginal_relevance_search\s*\(`)
)

func parsePythonRAG(relPath, src string) []models.CodeSurface {
	var surfaces []models.CodeSurface
	pkg := inferSurfacePackage(relPath)
	seen := map[string]bool{}

	add := func(name, reason string, line int, confidence float64) {
		sid := models.BuildSurfaceID(relPath, name+"_L"+strconv.Itoa(line), "")
		if seen[sid] {
			return
		}
		seen[sid] = true
		surfaces = append(surfaces, models.CodeSurface{
			SurfaceID:     sid,
			Name:          name,
			Path:          relPath,
			Kind:          models.SurfaceRetrieval,
			Language:      "python",
			Package:       pkg,
			Line:          line,
			Exported:      false,
			DetectionTier: models.TierSemantic,
			Confidence:    confidence,
			Reason:        reason,
		})
	}

	lines := strings.Split(src, "\n")

	for i, l := range lines {
		// Vector store factory.
		if m := pyVectorStoreFactory.FindStringSubmatch(l); m != nil {
			add("vector_store_"+strings.ToLower(m[1]), m[1]+".from_* vector store construction", i+1, 0.95)
			continue // Don't double-detect with constructor pattern.
		}

		// Text splitter with config extraction.
		if m := pyTextSplitter.FindStringSubmatch(l); m != nil {
			reason := m[1] + " instantiation"
			if strings.Contains(l, "chunk_size") {
				reason += " (with chunk_size config)"
			}
			if strings.Contains(l, "chunk_overlap") {
				reason += " (with chunk_overlap config)"
			}
			add("text_splitter", reason, i+1, 0.93)
		}

		// Retriever with config.
		if pyRetrieverConstruction.MatchString(l) {
			reason := ".as_retriever() construction"
			if strings.Contains(l, "search_kwargs") || strings.Contains(l, "'k'") || strings.Contains(l, "\"k\"") {
				reason += " (with top-k search config)"
			}
			add("retriever_config", reason, i+1, 0.92)
		}

		// Reranker.
		if m := pyReranker.FindStringSubmatch(l); m != nil {
			reason := m[1] + " reranker instantiation"
			if strings.Contains(l, "top_n") || strings.Contains(l, "top_k") {
				reason += " (with top-n config)"
			}
			add("reranker_config", reason, i+1, 0.92)
		}

		// Embedding model.
		if m := pyEmbeddingModel.FindStringSubmatch(l); m != nil {
			reason := m[1] + " embedding model"
			if strings.Contains(l, "model=") || strings.Contains(l, "model_name=") {
				reason += " (with model config)"
			}
			add("embedding_model", reason, i+1, 0.90)
		}

		// Similarity search.
		if pySimilaritySearch.MatchString(l) {
			add("retrieval_query", "similarity search execution", i+1, 0.85)
		}
	}

	return surfaces
}
