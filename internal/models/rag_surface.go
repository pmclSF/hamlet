package models

// RAGComponentKind classifies the type of RAG pipeline component.
type RAGComponentKind string

const (
	// RAGRetriever is a retriever that fetches documents from a vector store.
	RAGRetriever RAGComponentKind = "retriever"

	// RAGEmbedding is an embedding model that converts text to vectors.
	RAGEmbedding RAGComponentKind = "embedding"

	// RAGChunking is a text splitter/chunker that segments documents.
	RAGChunking RAGComponentKind = "chunking"

	// RAGVectorStore is a vector database or index.
	RAGVectorStore RAGComponentKind = "vector_store"

	// RAGReranker is a reranker that reorders retrieved documents.
	RAGReranker RAGComponentKind = "reranker"

	// RAGQueryBuilder is a query construction/rewriting component.
	RAGQueryBuilder RAGComponentKind = "query_builder"

	// RAGDocumentLoader loads raw documents from external sources.
	RAGDocumentLoader RAGComponentKind = "document_loader"

	// RAGCitationAssembly is citation/source attribution logic.
	RAGCitationAssembly RAGComponentKind = "citation_assembly"

	// RAGContextAssembly is context window assembly/truncation logic.
	RAGContextAssembly RAGComponentKind = "context_assembly"
)

// RAGPipelineSurface represents a detected RAG pipeline component with
// extracted configuration metadata. This enables structured reasoning
// about RAG pipelines: what chunking strategy is used, what top-k value
// is configured, which embedding model is selected.
type RAGPipelineSurface struct {
	// ComponentID is a deterministic stable identifier.
	// Format: "rag:<path>:<kind>:<name>".
	ComponentID string `json:"componentId"`

	// Name is the detected component name or variable name.
	Name string `json:"name"`

	// Path is the repository-relative file path.
	Path string `json:"path"`

	// Kind classifies the RAG component.
	Kind RAGComponentKind `json:"kind"`

	// Framework is the detected framework (langchain, llamaindex, custom).
	Framework string `json:"framework,omitempty"`

	// ClassName is the constructor/class name used (e.g., "RecursiveCharacterTextSplitter").
	ClassName string `json:"className,omitempty"`

	// Language is the programming language.
	Language string `json:"language"`

	// Line is the source line where this component is defined.
	Line int `json:"line,omitempty"`

	// Config holds extracted configuration parameters.
	Config RAGComponentConfig `json:"config,omitempty"`

	// LinkedSurfaceID links this component to its CodeSurface (SurfaceRetrieval).
	LinkedSurfaceID string `json:"linkedSurfaceId,omitempty"`

	// DetectionTier records the inference method.
	DetectionTier string `json:"detectionTier,omitempty"`

	// Confidence is the detection confidence (0.0–1.0).
	Confidence float64 `json:"confidence,omitempty"`

	// Reason explains why this component was detected.
	Reason string `json:"reason,omitempty"`
}

// RAGComponentConfig holds extracted configuration parameters for a RAG
// component. Fields are populated when the parser can extract concrete
// values from the source code.
type RAGComponentConfig struct {
	// ChunkSize is the configured chunk size for text splitters.
	ChunkSize int `json:"chunkSize,omitempty"`

	// ChunkOverlap is the configured chunk overlap for text splitters.
	ChunkOverlap int `json:"chunkOverlap,omitempty"`

	// TopK is the configured top-k value for retrievers and rerankers.
	TopK int `json:"topK,omitempty"`

	// ModelName is the configured model name for embeddings or rerankers.
	ModelName string `json:"modelName,omitempty"`

	// SearchType is the configured search type (similarity, mmr, hybrid).
	SearchType string `json:"searchType,omitempty"`

	// Provider is the vector store or embedding provider name.
	Provider string `json:"provider,omitempty"`

	// PersistDir is the configured persistence directory for vector stores.
	PersistDir string `json:"persistDir,omitempty"`
}

// Evidence returns a unified DetectionEvidence view.
func (rs *RAGPipelineSurface) Evidence() DetectionEvidence {
	return DetectionEvidence{
		Tier:       rs.DetectionTier,
		Confidence: rs.Confidence,
		FilePath:   rs.Path,
		Symbol:     rs.Name,
		Line:       rs.Line,
		Reason:     rs.Reason,
	}
}

// BuildRAGComponentID constructs a deterministic RAG component ID.
func BuildRAGComponentID(path string, kind RAGComponentKind, name string) string {
	return "rag:" + path + ":" + string(kind) + ":" + name
}
