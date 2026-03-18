package analysis

import (
	"strings"
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

// --- JS/TS Tests ---

func TestParseJSSchemas_ZodToolSchema(t *testing.T) {
	t.Parallel()
	src := `
import { z } from "zod";

const toolInputSchema = z.object({
  query: z.string().describe("The search query"),
  maxResults: z.number().optional(),
});

const toolOutputSchema = z.object({
  results: z.array(z.string()),
  confidence: z.number(),
});
`
	surfaces := ParseToolSchemas("src/schemas.ts", src, "js")
	if len(surfaces) != 2 {
		t.Fatalf("expected 2 Zod tool schemas, got %d: %v", len(surfaces), surfaceNames(surfaces))
	}
	for _, s := range surfaces {
		if s.Kind != models.SurfaceToolDef {
			t.Errorf("%s kind = %s, want tool_definition", s.Name, s.Kind)
		}
		if s.DetectionTier != models.TierSemantic {
			t.Errorf("%s tier = %s, want semantic", s.Name, s.DetectionTier)
		}
		if s.Confidence < 0.85 {
			t.Errorf("%s confidence %.2f too low", s.Name, s.Confidence)
		}
	}
}

func TestParseJSSchemas_ZodNonAI(t *testing.T) {
	t.Parallel()
	// Zod schemas that don't look like AI tools should NOT be detected.
	src := `
const userValidation = z.object({
  email: z.string().email(),
  name: z.string().min(1),
});
const addressForm = z.object({
  street: z.string(),
  city: z.string(),
});
`
	surfaces := ParseToolSchemas("src/validation.ts", src, "js")
	if len(surfaces) != 0 {
		t.Errorf("non-AI Zod schemas should NOT be detected, got %d: %v",
			len(surfaces), surfaceNames(surfaces))
	}
}

func TestParseJSSchemas_OpenAIToolRegistration(t *testing.T) {
	t.Parallel()
	src := `
const tools = [{
  type: "function",
  function: {
    name: "search",
    description: "Search the knowledge base",
    parameters: {
      type: "object",
      properties: { query: { type: "string" } },
    },
  },
}];
`
	surfaces := ParseToolSchemas("src/tools.ts", src, "js")
	found := false
	for _, s := range surfaces {
		if s.Name == "openai_tool_registration" {
			found = true
			if s.Confidence < 0.9 {
				t.Errorf("OpenAI registration confidence %.2f too low", s.Confidence)
			}
		}
	}
	if !found {
		t.Errorf("expected openai_tool_registration, got %v", surfaceNames(surfaces))
	}
}

func TestParseJSSchemas_OutputParser(t *testing.T) {
	t.Parallel()
	src := `
import { StructuredOutputParser } from "langchain/output_parsers";
const parser = StructuredOutputParser.fromZodSchema(responseSchema);
`
	surfaces := ParseToolSchemas("src/parser.ts", src, "js")
	found := false
	for _, s := range surfaces {
		if s.Name == "output_parser" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected output_parser, got %v", surfaceNames(surfaces))
	}
}

func TestParseJSSchemas_ResponseFormat(t *testing.T) {
	t.Parallel()
	src := `
const response = await openai.chat.completions.create({
  model: "gpt-4o",
  messages,
  response_format: { type: "json_schema", json_schema: mySchema },
});
`
	surfaces := ParseToolSchemas("src/structured.ts", src, "js")
	found := false
	for _, s := range surfaces {
		if s.Name == "structured_output_format" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected structured_output_format, got %v", surfaceNames(surfaces))
	}
}

// --- Python Tests ---

func TestParsePythonSchemas_PydanticResponseModel(t *testing.T) {
	t.Parallel()
	src := `
from pydantic import BaseModel

class SearchResponse(BaseModel):
    results: list[str]
    confidence: float
    sources: list[str]

class ToolParams(BaseModel):
    query: str
    max_results: int = 10
`
	surfaces := ParseToolSchemas("models.py", src, "python")
	if len(surfaces) != 2 {
		t.Fatalf("expected 2 Pydantic models, got %d: %v", len(surfaces), surfaceNames(surfaces))
	}
	for _, s := range surfaces {
		if s.DetectionTier != models.TierSemantic {
			t.Errorf("%s tier = %s, want semantic", s.Name, s.DetectionTier)
		}
	}
}

func TestParsePythonSchemas_PydanticNonAI(t *testing.T) {
	t.Parallel()
	// Pydantic models that don't look like AI contracts should NOT be detected.
	src := `
class UserProfile(BaseModel):
    name: str
    email: str

class DatabaseConfig(BaseModel):
    host: str
    port: int
`
	surfaces := ParseToolSchemas("config.py", src, "python")
	if len(surfaces) != 0 {
		t.Errorf("non-AI Pydantic models should NOT be detected, got %d: %v",
			len(surfaces), surfaceNames(surfaces))
	}
}

func TestParsePythonSchemas_InstructorResponseModel(t *testing.T) {
	t.Parallel()
	src := `
import instructor

client = instructor.patch(openai.OpenAI())
response = client.chat.completions.create(
    model="gpt-4",
    messages=messages,
    response_model=SearchResponse,
)
`
	surfaces := ParseToolSchemas("agent.py", src, "python")
	found := false
	for _, s := range surfaces {
		if strings.Contains(s.Name, "response_model") {
			found = true
			if s.Confidence < 0.9 {
				t.Errorf("instructor confidence %.2f too low", s.Confidence)
			}
		}
	}
	if !found {
		t.Errorf("expected response_model detection, got %v", surfaceNames(surfaces))
	}
}

func TestParsePythonSchemas_ToolDecorator(t *testing.T) {
	t.Parallel()
	src := `
from langchain.tools import tool

@tool
def search_documents(query: str) -> list[str]:
    """Search the knowledge base."""
    return search(query)
`
	surfaces := ParseToolSchemas("tools.py", src, "python")
	found := false
	for _, s := range surfaces {
		if strings.Contains(s.Name, "search_documents") {
			found = true
			if s.Confidence < 0.9 {
				t.Errorf("tool decorator confidence %.2f too low", s.Confidence)
			}
		}
	}
	if !found {
		t.Errorf("expected tool_search_documents, got %v", surfaceNames(surfaces))
	}
}

func TestParsePythonSchemas_OutputParser(t *testing.T) {
	t.Parallel()
	src := `
from langchain.output_parsers import PydanticOutputParser
parser = PydanticOutputParser(pydantic_object=SearchResponse)
`
	surfaces := ParseToolSchemas("parser.py", src, "python")
	found := false
	for _, s := range surfaces {
		if s.Name == "output_parser" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected output_parser, got %v", surfaceNames(surfaces))
	}
}

// --- Cross-language ---

func TestParseToolSchemas_StableIDs(t *testing.T) {
	t.Parallel()
	src := `const toolSchema = z.object({ query: z.string() });`
	s1 := ParseToolSchemas("src/tool.ts", src, "js")
	s2 := ParseToolSchemas("src/tool.ts", src, "js")
	if len(s1) != len(s2) {
		t.Fatalf("non-deterministic: %d vs %d", len(s1), len(s2))
	}
	for i := range s1 {
		if s1[i].SurfaceID != s2[i].SurfaceID {
			t.Errorf("ID differs: %s vs %s", s1[i].SurfaceID, s2[i].SurfaceID)
		}
	}
}

func TestParseToolSchemas_UnsupportedLanguage(t *testing.T) {
	t.Parallel()
	surfaces := ParseToolSchemas("file.rb", "schema = {}", "ruby")
	if surfaces != nil {
		t.Error("expected nil for unsupported language")
	}
}
