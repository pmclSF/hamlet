package analysis

import (
	"strings"
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

// --- JS/TS AST tests ---

func TestASTJS_MessageArray(t *testing.T) {
	t.Parallel()
	src := `
const messages = [
  { role: "system", content: "You are a helpful coding assistant." },
  { role: "user", content: userQuery },
  { role: "assistant", content: "I can help with that." },
];
`
	surfaces := ParsePromptAST("src/chat.ts", src, "js")

	found := findSurfaceByPrefix(surfaces, "message_array")
	if found == nil {
		t.Fatalf("expected message_array surface, got %v", surfaceNames(surfaces))
	}
	if found.Kind != models.SurfaceContext {
		t.Errorf("kind: want context (has system role), got %s", found.Kind)
	}
	if found.DetectionTier != models.TierStructural {
		t.Errorf("tier: want structural, got %s", found.DetectionTier)
	}
	if found.Confidence < 0.95 {
		t.Errorf("confidence: want >= 0.95, got %.2f", found.Confidence)
	}
	if !strings.Contains(found.Reason, DetectorASTMessageArray) {
		t.Errorf("reason should contain detector ID, got: %s", found.Reason)
	}
	if !strings.Contains(found.Reason, "3 objects") {
		t.Errorf("reason should mention object count, got: %s", found.Reason)
	}
}

func TestASTJS_MessageArrayNoSystemRole(t *testing.T) {
	t.Parallel()
	src := `
const conversation = [
  { role: "user", content: "Hello" },
  { role: "assistant", content: "Hi there!" },
];
`
	surfaces := ParsePromptAST("src/convo.ts", src, "js")

	found := findSurfaceByPrefix(surfaces, "message_array")
	if found == nil {
		t.Fatalf("expected message_array surface, got %v", surfaceNames(surfaces))
	}
	// No system role → SurfacePrompt, not SurfaceContext.
	if found.Kind != models.SurfacePrompt {
		t.Errorf("kind: want prompt (no system role), got %s", found.Kind)
	}
}

func TestASTJS_FrameworkConstructors(t *testing.T) {
	t.Parallel()
	src := `
import { SystemMessage, HumanMessage } from "@langchain/core/messages";

const messages = [
  new SystemMessage("You are a helpful assistant."),
  new HumanMessage(userQuery),
];
`
	surfaces := ParsePromptAST("src/chain.ts", src, "js")

	var systemFound, humanFound bool
	for _, s := range surfaces {
		if strings.Contains(s.Name, "SystemMessage") {
			systemFound = true
			if s.Kind != models.SurfaceContext {
				t.Errorf("SystemMessage kind: want context, got %s", s.Kind)
			}
		}
		if strings.Contains(s.Name, "HumanMessage") {
			humanFound = true
			if s.Kind != models.SurfacePrompt {
				t.Errorf("HumanMessage kind: want prompt, got %s", s.Kind)
			}
		}
	}
	if !systemFound {
		t.Error("expected SystemMessage constructor to be detected")
	}
	if !humanFound {
		t.Error("expected HumanMessage constructor to be detected")
	}
}

func TestASTJS_TemplateFactory(t *testing.T) {
	t.Parallel()
	src := `
const prompt = ChatPromptTemplate.fromMessages([
  ["system", "You are a {role} assistant."],
  ["human", "{input}"],
]);
`
	surfaces := ParsePromptAST("src/template.ts", src, "js")

	found := findSurfaceByPrefix(surfaces, "template_prompt")
	if found == nil {
		t.Fatalf("expected template_prompt surface, got %v", surfaceNames(surfaces))
	}
	if found.Kind != models.SurfacePrompt {
		t.Errorf("kind: want prompt, got %s", found.Kind)
	}
	if found.Confidence < 0.93 {
		t.Errorf("confidence: want >= 0.93, got %.2f", found.Confidence)
	}
}

func TestASTJS_SystemPromptAssignment(t *testing.T) {
	t.Parallel()
	src := `
const systemPrompt = "You are a helpful assistant. Your role is to answer questions accurately.";
`
	surfaces := ParsePromptAST("src/config.ts", src, "js")

	found := findSurfaceByPrefix(surfaces, "system_prompt")
	if found == nil {
		t.Fatalf("expected system_prompt surface, got %v", surfaceNames(surfaces))
	}
	if found.Kind != models.SurfaceContext {
		t.Errorf("kind: want context, got %s", found.Kind)
	}
	if !strings.Contains(found.Reason, DetectorASTSystemPrompt) {
		t.Errorf("reason should contain detector ID, got: %s", found.Reason)
	}
}

func TestASTJS_TemplateLiteralPrompt(t *testing.T) {
	t.Parallel()
	src := "const chatPrompt = `You are a helpful coding assistant. Your task is to help the user write clean code.\nGiven the context, always respond with clear explanations.`;\n"

	surfaces := ParsePromptAST("src/prompts.ts", src, "js")

	found := findSurfaceByPrefix(surfaces, "template_prompt")
	if found == nil {
		t.Fatalf("expected template_prompt surface, got %v", surfaceNames(surfaces))
	}
	if found.Kind != models.SurfacePrompt {
		t.Errorf("kind: want prompt, got %s", found.Kind)
	}
}

func TestASTJS_OpenAICall(t *testing.T) {
	t.Parallel()
	src := `
const response = await openai.chat.completions.create({
  model: "gpt-4",
  messages: [
    { role: "system", content: systemPrompt },
    { role: "user", content: userMessage },
  ],
});
`
	surfaces := ParsePromptAST("src/api.ts", src, "js")

	found := findSurfaceByPrefix(surfaces, "api_prompt")
	if found == nil {
		t.Fatalf("expected api_prompt surface, got %v", surfaceNames(surfaces))
	}
	if found.Kind != models.SurfacePrompt {
		t.Errorf("kind: want prompt, got %s", found.Kind)
	}
}

func TestASTJS_RejectsNonAIStrings(t *testing.T) {
	t.Parallel()
	src := `
const emailTemplate = "Hello {name}, your order {orderId} has been shipped.";
const htmlTemplate = "<div class='container'><h1>Welcome</h1></div>";
const sqlQuery = "SELECT * FROM users WHERE id = ?";
const config = { apiKey: "sk-12345", baseUrl: "https://api.example.com" };
`
	surfaces := ParsePromptAST("src/templates.ts", src, "js")

	if len(surfaces) != 0 {
		t.Errorf("expected 0 surfaces for non-AI strings, got %d: %v",
			len(surfaces), surfaceNames(surfaces))
	}
}

func TestASTJS_RejectsShortStrings(t *testing.T) {
	t.Parallel()
	src := `const systemPrompt = "Hi";`

	surfaces := ParsePromptAST("src/short.ts", src, "js")
	if len(surfaces) != 0 {
		t.Errorf("expected 0 surfaces for short strings, got %d", len(surfaces))
	}
}

// --- Python AST tests ---

func TestASTPython_MessageList(t *testing.T) {
	t.Parallel()
	src := `
messages = [
    {"role": "system", "content": "You are a helpful assistant."},
    {"role": "user", "content": user_input},
]
`
	surfaces := ParsePromptAST("src/chat.py", src, "python")

	found := findSurfaceByPrefix(surfaces, "message_list")
	if found == nil {
		t.Fatalf("expected message_list surface, got %v", surfaceNames(surfaces))
	}
	if found.Kind != models.SurfaceContext {
		t.Errorf("kind: want context, got %s", found.Kind)
	}
	if found.Confidence < 0.95 {
		t.Errorf("confidence: want >= 0.95, got %.2f", found.Confidence)
	}
}

func TestASTPython_FrameworkConstructors(t *testing.T) {
	t.Parallel()
	src := `
from langchain.schema import SystemMessage, HumanMessage

messages = [
    SystemMessage(content="You are a helpful assistant."),
    HumanMessage(content=user_query),
]
`
	surfaces := ParsePromptAST("src/chain.py", src, "python")

	var systemFound bool
	for _, s := range surfaces {
		if strings.Contains(s.Name, "SystemMessage") {
			systemFound = true
			if s.Kind != models.SurfaceContext {
				t.Errorf("SystemMessage kind: want context, got %s", s.Kind)
			}
		}
	}
	if !systemFound {
		t.Error("expected SystemMessage constructor to be detected")
	}
}

func TestASTPython_TemplateFactory(t *testing.T) {
	t.Parallel()
	src := `
from langchain.prompts import ChatPromptTemplate

prompt = ChatPromptTemplate.from_messages([
    ("system", "You are a {role} assistant."),
    ("human", "{input}"),
])
`
	surfaces := ParsePromptAST("src/template.py", src, "python")

	found := findSurfaceByPrefix(surfaces, "template_prompt")
	if found == nil {
		t.Fatalf("expected template_prompt surface, got %v", surfaceNames(surfaces))
	}
	if found.Kind != models.SurfacePrompt {
		t.Errorf("kind: want prompt, got %s", found.Kind)
	}
	if found.Confidence < 0.93 {
		t.Errorf("confidence: want >= 0.93, got %.2f", found.Confidence)
	}
}

func TestASTPython_TripleQuotePrompt(t *testing.T) {
	t.Parallel()
	src := `
system_prompt = """You are a helpful assistant.
Your role is to answer questions accurately.
Always respond with clear explanations.
Do not make up information."""
`
	surfaces := ParsePromptAST("src/prompts.py", src, "python")

	found := findSurfaceByPrefix(surfaces, "system_prompt")
	if found == nil {
		// Could also be template_prompt.
		found = findSurfaceByPrefix(surfaces, "template_prompt")
	}
	if found == nil {
		t.Fatalf("expected system_prompt or template_prompt surface, got %v", surfaceNames(surfaces))
	}
	if found.Kind != models.SurfaceContext {
		t.Errorf("kind: want context, got %s", found.Kind)
	}
}

func TestASTPython_FewShotArray(t *testing.T) {
	t.Parallel()
	src := `
examples = [
    {"input": "What is 2+2?", "output": "4"},
    {"input": "What is the capital of France?", "output": "Paris"},
    {"input": "Who wrote Hamlet?", "output": "Shakespeare"},
]
`
	surfaces := ParsePromptAST("src/examples.py", src, "python")

	found := findSurfaceByPrefix(surfaces, "few_shot")
	if found == nil {
		t.Fatalf("expected few_shot surface, got %v", surfaceNames(surfaces))
	}
	if found.Confidence < 0.90 {
		t.Errorf("confidence: want >= 0.90, got %.2f", found.Confidence)
	}
	if !strings.Contains(found.Reason, DetectorASTFewShot) {
		t.Errorf("reason should contain detector ID, got: %s", found.Reason)
	}
}

func TestASTPython_FewShotQuestionAnswer(t *testing.T) {
	t.Parallel()
	src := `
qa_examples = [
    {"question": "What is gravity?", "answer": "A fundamental force of nature."},
    {"question": "What is DNA?", "answer": "Deoxyribonucleic acid."},
]
`
	surfaces := ParsePromptAST("src/qa.py", src, "python")

	found := findSurfaceByPrefix(surfaces, "few_shot")
	if found == nil {
		t.Fatalf("expected few_shot surface, got %v", surfaceNames(surfaces))
	}
}

func TestASTPython_RejectsNonAILists(t *testing.T) {
	t.Parallel()
	src := `
users = [
    {"name": "Alice", "email": "alice@example.com"},
    {"name": "Bob", "email": "bob@example.com"},
]

config = {"database": "postgres", "port": 5432}
items = [1, 2, 3, 4, 5]
`
	surfaces := ParsePromptAST("src/data.py", src, "python")

	if len(surfaces) != 0 {
		t.Errorf("expected 0 surfaces for non-AI data, got %d: %v",
			len(surfaces), surfaceNames(surfaces))
	}
}

func TestASTPython_OpenAICall(t *testing.T) {
	t.Parallel()
	src := `
response = client.chat.completions.create(
    model="gpt-4",
    messages=[
        {"role": "system", "content": system_prompt},
        {"role": "user", "content": user_message},
    ],
)
`
	surfaces := ParsePromptAST("src/api.py", src, "python")

	found := findSurfaceByPrefix(surfaces, "api_prompt")
	if found == nil {
		t.Fatalf("expected api_prompt surface, got %v", surfaceNames(surfaces))
	}
}

// --- Go AST tests ---

func TestASTGo_MessageSlice(t *testing.T) {
	t.Parallel()
	src := `package chat

type Message struct {
	Role    string
	Content string
}

var messages = []Message{
	{Role: "system", Content: "You are a helpful assistant."},
	{Role: "user", Content: userQuery},
}
`
	surfaces := ParsePromptAST("src/chat.go", src, "go")

	found := findSurfaceByPrefix(surfaces, "message_slice")
	if found == nil {
		t.Fatalf("expected message_slice surface, got %v", surfaceNames(surfaces))
	}
	if found.Kind != models.SurfaceContext {
		t.Errorf("kind: want context, got %s", found.Kind)
	}
	if found.Confidence < 0.95 {
		t.Errorf("confidence: want >= 0.95, got %.2f", found.Confidence)
	}
}

func TestASTGo_PromptConst(t *testing.T) {
	t.Parallel()
	src := "package ai\n\nconst SystemPrompt = `You are a helpful assistant. Your role is to answer questions. Always respond clearly.`\n"

	surfaces := ParsePromptAST("src/ai.go", src, "go")

	found := findSurfaceByPrefix(surfaces, "prompt_const")
	if found == nil {
		t.Fatalf("expected prompt_const surface, got %v", surfaceNames(surfaces))
	}
	if found.Kind != models.SurfaceContext {
		t.Errorf("kind: want context, got %s", found.Kind)
	}
}

func TestASTGo_RejectsNonAI(t *testing.T) {
	t.Parallel()
	src := `package main

var greeting = "Hello, World!"
var config = map[string]string{"key": "value"}
`
	surfaces := ParsePromptAST("src/main.go", src, "go")

	if len(surfaces) != 0 {
		t.Errorf("expected 0 surfaces for non-AI code, got %d: %v",
			len(surfaces), surfaceNames(surfaces))
	}
}

// --- Evidence metadata tests ---

func TestASTPrompt_EvidenceMetadata(t *testing.T) {
	t.Parallel()
	src := `
const messages = [
  { role: "system", content: "You are helpful." },
  { role: "user", content: input },
];
`
	surfaces := ParsePromptAST("src/chat.ts", src, "js")

	for _, s := range surfaces {
		if s.DetectionTier == "" {
			t.Errorf("surface %q missing DetectionTier", s.Name)
		}
		if s.Confidence == 0 {
			t.Errorf("surface %q has zero Confidence", s.Name)
		}
		if s.Reason == "" {
			t.Errorf("surface %q missing Reason", s.Name)
		}
		if s.SurfaceID == "" {
			t.Errorf("surface %q missing SurfaceID", s.Name)
		}
		if s.Path == "" {
			t.Errorf("surface %q missing Path", s.Name)
		}
		if s.Line == 0 {
			t.Errorf("surface %q missing Line", s.Name)
		}
	}
}

func TestASTPrompt_KindAssignment(t *testing.T) {
	t.Parallel()

	// System message → SurfaceContext.
	src1 := `
messages = [
    {"role": "system", "content": "You are an assistant."},
    {"role": "user", "content": query},
]
`
	surfaces1 := ParsePromptAST("src/a.py", src1, "python")
	for _, s := range surfaces1 {
		if strings.Contains(s.Name, "message_list") {
			if s.Kind != models.SurfaceContext {
				t.Errorf("message array with system role: want context, got %s", s.Kind)
			}
		}
	}

	// User-only message → SurfacePrompt.
	src2 := `
const msgs = [
  { role: "user", content: "Hello" },
  { role: "assistant", content: "Hi" },
];
`
	surfaces2 := ParsePromptAST("src/b.ts", src2, "js")
	for _, s := range surfaces2 {
		if strings.Contains(s.Name, "message_array") {
			if s.Kind != models.SurfacePrompt {
				t.Errorf("message array without system role: want prompt, got %s", s.Kind)
			}
		}
	}
}

// --- Helpers ---

func findSurfaceByPrefix(surfaces []models.CodeSurface, prefix string) *models.CodeSurface {
	for i, s := range surfaces {
		if strings.HasPrefix(s.Name, prefix) {
			return &surfaces[i]
		}
	}
	return nil
}
