package analysis

import (
	"strings"
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

// --- JS/TS Structured RAG Tests ---

func TestRAGStructuredJS_RetrieverWithConfig(t *testing.T) {
	t.Parallel()
	src := `
const retriever = vectorStore.asRetriever({
  k: 5,
  searchType: "mmr",
});
`
	components := ParseRAGStructured("src/search.ts", src, "js")

	found := findRAGByKind(components, models.RAGRetriever)
	if found == nil {
		t.Fatalf("expected retriever component, got %v", ragNames(components))
	}
	if found.Config.TopK != 5 {
		t.Errorf("topK: want 5, got %d", found.Config.TopK)
	}
	if found.Config.SearchType != "mmr" {
		t.Errorf("searchType: want mmr, got %s", found.Config.SearchType)
	}
	if !strings.Contains(found.Reason, DetectorRAGRetriever) {
		t.Errorf("reason should contain detector ID, got: %s", found.Reason)
	}
}

func TestRAGStructuredJS_ChunkingConfig(t *testing.T) {
	t.Parallel()
	src := `
const splitter = new RecursiveCharacterTextSplitter({
  chunkSize: 500,
  chunkOverlap: 50,
});
`
	components := ParseRAGStructured("src/ingest.ts", src, "js")

	found := findRAGByKind(components, models.RAGChunking)
	if found == nil {
		t.Fatalf("expected chunking component, got %v", ragNames(components))
	}
	if found.Config.ChunkSize != 500 {
		t.Errorf("chunkSize: want 500, got %d", found.Config.ChunkSize)
	}
	if found.Config.ChunkOverlap != 50 {
		t.Errorf("chunkOverlap: want 50, got %d", found.Config.ChunkOverlap)
	}
	if found.ClassName != "RecursiveCharacterTextSplitter" {
		t.Errorf("className: want RecursiveCharacterTextSplitter, got %s", found.ClassName)
	}
	if !strings.Contains(found.Reason, "chunkSize=500") {
		t.Errorf("reason should contain extracted config, got: %s", found.Reason)
	}
}

func TestRAGStructuredJS_EmbeddingModel(t *testing.T) {
	t.Parallel()
	src := `
const embeddings = new OpenAIEmbeddings({
  model: "text-embedding-3-small",
});
`
	components := ParseRAGStructured("src/embed.ts", src, "js")

	found := findRAGByKind(components, models.RAGEmbedding)
	if found == nil {
		t.Fatalf("expected embedding component, got %v", ragNames(components))
	}
	if found.Config.ModelName != "text-embedding-3-small" {
		t.Errorf("modelName: want text-embedding-3-small, got %s", found.Config.ModelName)
	}
	if found.ClassName != "OpenAIEmbeddings" {
		t.Errorf("className: want OpenAIEmbeddings, got %s", found.ClassName)
	}
}

func TestRAGStructuredJS_VectorStore(t *testing.T) {
	t.Parallel()
	src := `
const client = new PineconeClient({
  apiKey: process.env.PINECONE_API_KEY,
});
`
	components := ParseRAGStructured("src/store.ts", src, "js")

	found := findRAGByKind(components, models.RAGVectorStore)
	if found == nil {
		t.Fatalf("expected vector store component, got %v", ragNames(components))
	}
	if found.Config.Provider != "pineconeclient" {
		t.Errorf("provider: want pineconeclient, got %s", found.Config.Provider)
	}
	if found.Confidence < 0.93 {
		t.Errorf("confidence: want >= 0.93, got %.2f", found.Confidence)
	}
}

func TestRAGStructuredJS_Reranker(t *testing.T) {
	t.Parallel()
	src := `
const reranker = new CohereRerank({
  model: "rerank-english-v3.0",
  topN: 3,
});
`
	components := ParseRAGStructured("src/rerank.ts", src, "js")

	found := findRAGByKind(components, models.RAGReranker)
	if found == nil {
		t.Fatalf("expected reranker component, got %v", ragNames(components))
	}
	if found.Config.TopK != 3 {
		t.Errorf("topK: want 3, got %d", found.Config.TopK)
	}
	if found.Config.ModelName != "rerank-english-v3.0" {
		t.Errorf("modelName: want rerank-english-v3.0, got %s", found.Config.ModelName)
	}
}

func TestRAGStructuredJS_DocumentLoader(t *testing.T) {
	t.Parallel()
	src := `
const loader = new PDFLoader("./data/report.pdf");
const docs = await loader.load();
`
	components := ParseRAGStructured("src/load.ts", src, "js")

	found := findRAGByKind(components, models.RAGDocumentLoader)
	if found == nil {
		t.Fatalf("expected document loader, got %v", ragNames(components))
	}
	if found.ClassName != "PDFLoader" {
		t.Errorf("className: want PDFLoader, got %s", found.ClassName)
	}
}

func TestRAGStructuredJS_QueryBuilder(t *testing.T) {
	t.Parallel()
	src := `
const retriever = MultiQueryRetriever.fromLLM({
  llm: model,
  retriever: baseRetriever,
});
`
	components := ParseRAGStructured("src/query.ts", src, "js")

	found := findRAGByKind(components, models.RAGQueryBuilder)
	if found == nil {
		t.Fatalf("expected query builder, got %v", ragNames(components))
	}
	if found.ClassName != "MultiQueryRetriever" {
		t.Errorf("className: want MultiQueryRetriever, got %s", found.ClassName)
	}
}

func TestRAGStructuredJS_CitationAssembly(t *testing.T) {
	t.Parallel()
	src := `
const chain = createRetrievalChain({
  retriever,
  combineDocsChain: stuffChain,
});
`
	components := ParseRAGStructured("src/cite.ts", src, "js")

	found := findRAGByKind(components, models.RAGCitationAssembly)
	if found == nil {
		t.Fatalf("expected citation assembly, got %v", ragNames(components))
	}
}

// --- Python Structured RAG Tests ---

func TestRAGStructuredPy_RetrieverWithConfig(t *testing.T) {
	t.Parallel()
	src := `
retriever = vectorstore.as_retriever(
    search_type="mmr",
    search_kwargs={"k": 10}
)
`
	components := ParseRAGStructured("src/search.py", src, "python")

	found := findRAGByKind(components, models.RAGRetriever)
	if found == nil {
		t.Fatalf("expected retriever component, got %v", ragNames(components))
	}
	if found.Config.TopK != 10 {
		t.Errorf("topK: want 10, got %d", found.Config.TopK)
	}
	if found.Config.SearchType != "mmr" {
		t.Errorf("searchType: want mmr, got %s", found.Config.SearchType)
	}
}

func TestRAGStructuredPy_ChunkingConfig(t *testing.T) {
	t.Parallel()
	src := `
text_splitter = RecursiveCharacterTextSplitter(
    chunk_size=1000,
    chunk_overlap=200,
)
`
	components := ParseRAGStructured("src/chunk.py", src, "python")

	found := findRAGByKind(components, models.RAGChunking)
	if found == nil {
		t.Fatalf("expected chunking component, got %v", ragNames(components))
	}
	if found.Config.ChunkSize != 1000 {
		t.Errorf("chunkSize: want 1000, got %d", found.Config.ChunkSize)
	}
	if found.Config.ChunkOverlap != 200 {
		t.Errorf("chunkOverlap: want 200, got %d", found.Config.ChunkOverlap)
	}
	if !strings.Contains(found.Reason, "chunk_size=1000") {
		t.Errorf("reason should contain extracted config, got: %s", found.Reason)
	}
}

func TestRAGStructuredPy_EmbeddingModel(t *testing.T) {
	t.Parallel()
	src := `
embeddings = OpenAIEmbeddings(model="text-embedding-ada-002")
`
	components := ParseRAGStructured("src/embed.py", src, "python")

	found := findRAGByKind(components, models.RAGEmbedding)
	if found == nil {
		t.Fatalf("expected embedding component, got %v", ragNames(components))
	}
	if found.Config.ModelName != "text-embedding-ada-002" {
		t.Errorf("modelName: want text-embedding-ada-002, got %s", found.Config.ModelName)
	}
	if !strings.Contains(found.Reason, "model=text-embedding-ada-002") {
		t.Errorf("reason should contain model name, got: %s", found.Reason)
	}
}

func TestRAGStructuredPy_VectorStoreFactory(t *testing.T) {
	t.Parallel()
	src := `
vectorstore = Chroma.from_documents(
    documents,
    embeddings,
    persist_directory="./chroma_db"
)
`
	components := ParseRAGStructured("src/store.py", src, "python")

	found := findRAGByKind(components, models.RAGVectorStore)
	if found == nil {
		t.Fatalf("expected vector store, got %v", ragNames(components))
	}
	if found.Config.Provider != "chroma" {
		t.Errorf("provider: want chroma, got %s", found.Config.Provider)
	}
	if found.Config.PersistDir != "./chroma_db" {
		t.Errorf("persistDir: want ./chroma_db, got %s", found.Config.PersistDir)
	}
}

func TestRAGStructuredPy_RerankerWithTopN(t *testing.T) {
	t.Parallel()
	src := `
reranker = CohereRerank(
    model="rerank-english-v3.0",
    top_n=5
)
`
	components := ParseRAGStructured("src/rerank.py", src, "python")

	found := findRAGByKind(components, models.RAGReranker)
	if found == nil {
		t.Fatalf("expected reranker, got %v", ragNames(components))
	}
	if found.Config.TopK != 5 {
		t.Errorf("topN: want 5, got %d", found.Config.TopK)
	}
	if found.Config.ModelName != "rerank-english-v3.0" {
		t.Errorf("modelName: want rerank-english-v3.0, got %s", found.Config.ModelName)
	}
}

func TestRAGStructuredPy_DocumentLoader(t *testing.T) {
	t.Parallel()
	src := `
loader = PyPDFLoader("data/manual.pdf")
docs = loader.load()
`
	components := ParseRAGStructured("src/load.py", src, "python")

	found := findRAGByKind(components, models.RAGDocumentLoader)
	if found == nil {
		t.Fatalf("expected document loader, got %v", ragNames(components))
	}
	if found.ClassName != "PyPDFLoader" {
		t.Errorf("className: want PyPDFLoader, got %s", found.ClassName)
	}
}

func TestRAGStructuredPy_QueryBuilder(t *testing.T) {
	t.Parallel()
	src := `
retriever = MultiQueryRetriever.from_llm(
    retriever=base_retriever,
    llm=llm,
)
`
	components := ParseRAGStructured("src/query.py", src, "python")

	found := findRAGByKind(components, models.RAGQueryBuilder)
	if found == nil {
		t.Fatalf("expected query builder, got %v", ragNames(components))
	}
	if found.ClassName != "MultiQueryRetriever" {
		t.Errorf("className: want MultiQueryRetriever, got %s", found.ClassName)
	}
}

func TestRAGStructuredPy_CitationAssembly(t *testing.T) {
	t.Parallel()
	src := `
chain = create_retrieval_chain(retriever, combine_docs_chain)
`
	components := ParseRAGStructured("src/cite.py", src, "python")

	found := findRAGByKind(components, models.RAGCitationAssembly)
	if found == nil {
		t.Fatalf("expected citation assembly, got %v", ragNames(components))
	}
}

func TestRAGStructuredPy_ContextCompression(t *testing.T) {
	t.Parallel()
	src := `
compressor = LLMChainExtractor.from_llm(llm)
compression_retriever = ContextualCompressionRetriever(
    base_compressor=compressor,
    base_retriever=retriever,
)
`
	components := ParseRAGStructured("src/compress.py", src, "python")

	found := findRAGByKind(components, models.RAGContextAssembly)
	if found == nil {
		t.Fatalf("expected context assembly, got %v", ragNames(components))
	}
}

// --- Cross-cutting tests ---

func TestRAGStructured_NonRAGCode(t *testing.T) {
	t.Parallel()
	src := `
const db = new PostgresClient({ host: "localhost" });
const results = await db.query("SELECT * FROM users");
`
	components := ParseRAGStructured("src/db.ts", src, "js")
	if len(components) != 0 {
		t.Errorf("expected 0 for non-RAG code, got %d: %v", len(components), ragNames(components))
	}
}

func TestRAGStructured_UnsupportedLanguage(t *testing.T) {
	t.Parallel()
	components := ParseRAGStructured("file.rb", "Redis.new", "ruby")
	if components != nil {
		t.Error("expected nil for unsupported language")
	}
}

func TestRAGStructured_StableIDs(t *testing.T) {
	t.Parallel()
	src := `
retriever = vectorstore.as_retriever(search_kwargs={"k": 5})
`
	c1 := ParseRAGStructured("src/s.py", src, "python")
	c2 := ParseRAGStructured("src/s.py", src, "python")
	if len(c1) != len(c2) {
		t.Fatalf("non-deterministic: %d vs %d", len(c1), len(c2))
	}
	for i := range c1 {
		if c1[i].ComponentID != c2[i].ComponentID {
			t.Errorf("ID differs: %s vs %s", c1[i].ComponentID, c2[i].ComponentID)
		}
	}
}

func TestRAGStructured_EvidenceMetadata(t *testing.T) {
	t.Parallel()
	src := `
text_splitter = RecursiveCharacterTextSplitter(chunk_size=500)
`
	components := ParseRAGStructured("src/chunk.py", src, "python")

	for _, c := range components {
		if c.DetectionTier == "" {
			t.Errorf("component %q missing DetectionTier", c.Name)
		}
		if c.Confidence == 0 {
			t.Errorf("component %q has zero Confidence", c.Name)
		}
		if c.Reason == "" {
			t.Errorf("component %q missing Reason", c.Name)
		}
		if c.ComponentID == "" {
			t.Errorf("component %q missing ComponentID", c.Name)
		}
		if c.Path == "" {
			t.Errorf("component %q missing Path", c.Name)
		}
	}
}

func TestRAGStructured_FullPipeline(t *testing.T) {
	t.Parallel()
	src := `
from langchain.document_loaders import PyPDFLoader
from langchain.text_splitter import RecursiveCharacterTextSplitter
from langchain.embeddings import OpenAIEmbeddings
from langchain.vectorstores import Chroma

loader = PyPDFLoader("docs/manual.pdf")
documents = loader.load()

text_splitter = RecursiveCharacterTextSplitter(chunk_size=500, chunk_overlap=50)
chunks = text_splitter.split_documents(documents)

embeddings = OpenAIEmbeddings(model="text-embedding-3-small")
vectorstore = Chroma.from_documents(chunks, embeddings, persist_directory="./db")

retriever = vectorstore.as_retriever(search_kwargs={"k": 5})
`
	components := ParseRAGStructured("src/pipeline.py", src, "python")

	kinds := map[models.RAGComponentKind]bool{}
	for _, c := range components {
		kinds[c.Kind] = true
	}

	expected := []models.RAGComponentKind{
		models.RAGDocumentLoader,
		models.RAGChunking,
		models.RAGEmbedding,
		models.RAGVectorStore,
		models.RAGRetriever,
	}
	for _, k := range expected {
		if !kinds[k] {
			t.Errorf("expected %s component in full pipeline", k)
		}
	}

	if len(components) < 5 {
		t.Errorf("expected at least 5 components in full pipeline, got %d", len(components))
	}
}

func TestLinkRAGSurfacesToCodeSurfaces(t *testing.T) {
	t.Parallel()

	ragComponents := []models.RAGPipelineSurface{
		{ComponentID: "rag:src/search.py:retriever:retriever", Path: "src/search.py", Kind: models.RAGRetriever, Line: 10},
		{ComponentID: "rag:src/embed.py:embedding:embedding", Path: "src/embed.py", Kind: models.RAGEmbedding, Line: 5},
	}

	codeSurfaces := []models.CodeSurface{
		{SurfaceID: "surface:src/search.py:retrieverConfig", Path: "src/search.py", Kind: models.SurfaceRetrieval, Line: 10},
		{SurfaceID: "surface:src/other.py:unrelated", Path: "src/other.py", Kind: models.SurfaceRetrieval, Line: 1},
	}

	LinkRAGSurfacesToCodeSurfaces(ragComponents, codeSurfaces)

	if ragComponents[0].LinkedSurfaceID != "surface:src/search.py:retrieverConfig" {
		t.Errorf("expected retriever linked to search.py surface, got %s", ragComponents[0].LinkedSurfaceID)
	}
	if ragComponents[1].LinkedSurfaceID != "" {
		t.Errorf("expected embedding not linked (no retrieval surface in embed.py), got %s", ragComponents[1].LinkedSurfaceID)
	}
}

// --- Helpers ---

func findRAGByKind(components []models.RAGPipelineSurface, kind models.RAGComponentKind) *models.RAGPipelineSurface {
	for i, c := range components {
		if c.Kind == kind {
			return &components[i]
		}
	}
	return nil
}

func ragNames(components []models.RAGPipelineSurface) []string {
	var names []string
	for _, c := range components {
		names = append(names, c.Name)
	}
	return names
}
