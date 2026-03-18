package analysis

import (
	"strings"
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

// --- JS/TS Tests ---

func TestParseJSRAG_VectorStore(t *testing.T) {
	t.Parallel()
	src := `
import { PineconeClient } from "@pinecone-database/pinecone";
const client = new PineconeClient({ apiKey: process.env.PINECONE_API_KEY });
const index = client.index("knowledge-base");
`
	surfaces := ParseRAGPipeline("src/store.ts", src, "js")
	found := false
	for _, s := range surfaces {
		if strings.Contains(s.Name, "vector_store") {
			found = true
			if s.Kind != models.SurfaceRetrieval {
				t.Errorf("kind = %s, want retrieval", s.Kind)
			}
			if s.Confidence < 0.9 {
				t.Errorf("confidence = %.2f, want >= 0.9", s.Confidence)
			}
			if s.DetectionTier != models.TierSemantic {
				t.Errorf("tier = %s, want semantic", s.DetectionTier)
			}
		}
	}
	if !found {
		t.Errorf("expected vector store detection, got %v", surfaceNames(surfaces))
	}
}

func TestParseJSRAG_TextSplitter(t *testing.T) {
	t.Parallel()
	src := `
import { RecursiveCharacterTextSplitter } from "langchain/text_splitter";
const splitter = new RecursiveCharacterTextSplitter({
  chunkSize: 500,
  chunkOverlap: 50,
});
const chunks = await splitter.splitDocuments(documents);
`
	surfaces := ParseRAGPipeline("src/ingest.ts", src, "js")
	found := false
	for _, s := range surfaces {
		if s.Name == "text_splitter" {
			found = true
			if !strings.Contains(s.Reason, "chunk size") {
				t.Errorf("expected chunk config in reason, got: %s", s.Reason)
			}
		}
	}
	if !found {
		t.Errorf("expected text_splitter, got %v", surfaceNames(surfaces))
	}
}

func TestParseJSRAG_Retriever(t *testing.T) {
	t.Parallel()
	src := `
const retriever = vectorStore.asRetriever({ k: 5 });
const docs = await retriever.getRelevantDocuments(query);
`
	surfaces := ParseRAGPipeline("src/search.ts", src, "js")
	found := false
	for _, s := range surfaces {
		if s.Name == "retriever_config" {
			found = true
			if !strings.Contains(s.Reason, "top-k") {
				t.Errorf("expected top-k in reason, got: %s", s.Reason)
			}
		}
	}
	if !found {
		t.Errorf("expected retriever_config, got %v", surfaceNames(surfaces))
	}
}

func TestParseJSRAG_Reranker(t *testing.T) {
	t.Parallel()
	src := `
import { CohereRerank } from "@langchain/cohere";
const reranker = new CohereRerank({ model: "rerank-english-v3.0", topN: 5 });
`
	surfaces := ParseRAGPipeline("src/rerank.ts", src, "js")
	found := false
	for _, s := range surfaces {
		if s.Name == "reranker_config" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected reranker_config, got %v", surfaceNames(surfaces))
	}
}

func TestParseJSRAG_EmbeddingModel(t *testing.T) {
	t.Parallel()
	src := `
const embeddings = new OpenAIEmbeddings({ model: "text-embedding-3-small" });
`
	surfaces := ParseRAGPipeline("src/embed.ts", src, "js")
	found := false
	for _, s := range surfaces {
		if s.Name == "embedding_model" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected embedding_model, got %v", surfaceNames(surfaces))
	}
}

// --- Python Tests ---

func TestParsePythonRAG_VectorStoreFactory(t *testing.T) {
	t.Parallel()
	src := `
from langchain.vectorstores import Chroma
vectorstore = Chroma.from_documents(documents, embeddings, persist_directory="./db")
`
	surfaces := ParseRAGPipeline("store.py", src, "python")
	found := false
	for _, s := range surfaces {
		if strings.Contains(s.Name, "vector_store") {
			found = true
			if s.Confidence < 0.9 {
				t.Errorf("confidence = %.2f", s.Confidence)
			}
		}
	}
	if !found {
		t.Errorf("expected vector_store, got %v", surfaceNames(surfaces))
	}
}

func TestParsePythonRAG_TextSplitterWithConfig(t *testing.T) {
	t.Parallel()
	src := `
from langchain.text_splitter import RecursiveCharacterTextSplitter
text_splitter = RecursiveCharacterTextSplitter(chunk_size=500, chunk_overlap=50)
chunks = text_splitter.split_documents(documents)
`
	surfaces := ParseRAGPipeline("ingest.py", src, "python")
	found := false
	for _, s := range surfaces {
		if s.Name == "text_splitter" {
			found = true
			if !strings.Contains(s.Reason, "chunk_size") {
				t.Errorf("expected chunk_size in reason, got: %s", s.Reason)
			}
		}
	}
	if !found {
		t.Errorf("expected text_splitter, got %v", surfaceNames(surfaces))
	}
}

func TestParsePythonRAG_RetrieverWithTopK(t *testing.T) {
	t.Parallel()
	src := `
retriever = vectorstore.as_retriever(search_kwargs={"k": 5})
`
	surfaces := ParseRAGPipeline("search.py", src, "python")
	found := false
	for _, s := range surfaces {
		if s.Name == "retriever_config" {
			found = true
			if !strings.Contains(s.Reason, "top-k") {
				t.Errorf("expected top-k in reason, got: %s", s.Reason)
			}
		}
	}
	if !found {
		t.Errorf("expected retriever_config, got %v", surfaceNames(surfaces))
	}
}

func TestParsePythonRAG_Reranker(t *testing.T) {
	t.Parallel()
	src := `
from langchain.retrievers import CohereRerank
reranker = CohereRerank(model="rerank-english-v3.0", top_n=3)
`
	surfaces := ParseRAGPipeline("rerank.py", src, "python")
	found := false
	for _, s := range surfaces {
		if s.Name == "reranker_config" {
			found = true
			if !strings.Contains(s.Reason, "top-n") {
				t.Errorf("expected top-n in reason, got: %s", s.Reason)
			}
		}
	}
	if !found {
		t.Errorf("expected reranker_config, got %v", surfaceNames(surfaces))
	}
}

func TestParsePythonRAG_EmbeddingModel(t *testing.T) {
	t.Parallel()
	src := `
from langchain.embeddings import OpenAIEmbeddings
embeddings = OpenAIEmbeddings(model="text-embedding-3-small")
`
	surfaces := ParseRAGPipeline("embed.py", src, "python")
	found := false
	for _, s := range surfaces {
		if s.Name == "embedding_model" {
			found = true
			if !strings.Contains(s.Reason, "model config") {
				t.Errorf("expected model config in reason, got: %s", s.Reason)
			}
		}
	}
	if !found {
		t.Errorf("expected embedding_model, got %v", surfaceNames(surfaces))
	}
}

// --- Negative tests ---

func TestParseRAG_NonRAGCode(t *testing.T) {
	t.Parallel()
	src := `
const db = new PostgresClient({ host: "localhost" });
const results = await db.query("SELECT * FROM users");
`
	surfaces := ParseRAGPipeline("src/db.ts", src, "js")
	if len(surfaces) != 0 {
		t.Errorf("expected 0 for non-RAG code, got %d: %v", len(surfaces), surfaceNames(surfaces))
	}
}

func TestParseRAG_StableIDs(t *testing.T) {
	t.Parallel()
	src := `const retriever = vectorStore.asRetriever({ k: 5 });`
	s1 := ParseRAGPipeline("src/search.ts", src, "js")
	s2 := ParseRAGPipeline("src/search.ts", src, "js")
	if len(s1) != len(s2) {
		t.Fatalf("non-deterministic: %d vs %d", len(s1), len(s2))
	}
	for i := range s1 {
		if s1[i].SurfaceID != s2[i].SurfaceID {
			t.Errorf("ID differs: %s vs %s", s1[i].SurfaceID, s2[i].SurfaceID)
		}
	}
}

func TestParseRAG_UnsupportedLanguage(t *testing.T) {
	t.Parallel()
	surfaces := ParseRAGPipeline("file.rb", "db = Redis.new", "ruby")
	if surfaces != nil {
		t.Error("expected nil for unsupported language")
	}
}
