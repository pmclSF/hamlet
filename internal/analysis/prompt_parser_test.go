package analysis

import (
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

// --- JS/TS Tests ---

func TestParseJSPrompts_MessageArray(t *testing.T) {
	t.Parallel()
	src := `
const messages = [
  { role: "system", content: "You are a helpful assistant." },
  { role: "user", content: userInput },
];
`
	surfaces := ParseEmbeddedPrompts("src/chat.ts", src, "js")
	found := false
	for _, s := range surfaces {
		if s.Name == "message_array" {
			found = true
			if s.DetectionTier != models.TierSemantic {
				t.Errorf("message array tier = %q, want semantic", s.DetectionTier)
			}
			if s.Confidence < 0.9 {
				t.Errorf("message array confidence = %.2f, want >= 0.9", s.Confidence)
			}
			if s.Reason == "" {
				t.Error("expected non-empty reason")
			}
		}
	}
	if !found {
		t.Errorf("expected message_array detection, got %v", surfaceNames(surfaces))
	}
}

func TestParseJSPrompts_TemplateLiteral(t *testing.T) {
	t.Parallel()
	src := "const prompt = `You are a helpful assistant. Your role is to answer questions based on the provided context. Do not make up information. Always respond with accurate data.`;"
	surfaces := ParseEmbeddedPrompts("src/prompt.ts", src, "js")
	found := false
	for _, s := range surfaces {
		if s.Name == "template_prompt" {
			found = true
			if s.DetectionTier != models.TierContent {
				t.Errorf("template prompt tier = %q, want content", s.DetectionTier)
			}
		}
	}
	if !found {
		t.Errorf("expected template_prompt detection, got %v", surfaceNames(surfaces))
	}
}

func TestParseJSPrompts_FewShotArray(t *testing.T) {
	t.Parallel()
	src := `
const examples = [
  { input: "What is the weather?", output: "I can help with weather queries." },
  { input: "Book a flight", output: "I'll help you book a flight." },
];
`
	surfaces := ParseEmbeddedPrompts("src/examples.ts", src, "js")
	found := false
	for _, s := range surfaces {
		if s.Name == "few_shot_examples" {
			found = true
			if s.DetectionTier != models.TierSemantic {
				t.Errorf("few-shot tier = %q, want semantic", s.DetectionTier)
			}
		}
	}
	if !found {
		t.Errorf("expected few_shot_examples detection, got %v", surfaceNames(surfaces))
	}
}

func TestParseJSPrompts_AssignedString(t *testing.T) {
	t.Parallel()
	src := `const systemMessage = "You are a helpful customer service assistant. Your task is to answer questions based on the provided documentation. Do not make up information.";`
	surfaces := ParseEmbeddedPrompts("src/config.ts", src, "js")
	found := false
	for _, s := range surfaces {
		if s.Kind == models.SurfaceContext {
			found = true
		}
	}
	if !found {
		t.Errorf("expected inline prompt detection, got %v", surfaceNames(surfaces))
	}
}

func TestParseJSPrompts_NonAIStrings(t *testing.T) {
	t.Parallel()
	src := `
const greeting = "Hello world! Welcome to our platform. We hope you enjoy your experience here.";
const error = "An error occurred while processing your request. Please try again later or contact support.";
const template = '<div class="container"><h1>{{title}}</h1><p>{{content}}</p></div>';
`
	surfaces := ParseEmbeddedPrompts("src/utils.ts", src, "js")
	if len(surfaces) != 0 {
		t.Errorf("expected 0 surfaces for non-AI strings, got %d: %v", len(surfaces), surfaceNames(surfaces))
	}
}

func TestParseJSPrompts_SingleRoleNotDetected(t *testing.T) {
	t.Parallel()
	// A single role entry is not enough — could be config, not a message array.
	src := `const config = { role: "admin", permissions: ["read", "write"] };`
	surfaces := ParseEmbeddedPrompts("src/config.ts", src, "js")
	for _, s := range surfaces {
		if s.Name == "message_array" {
			t.Error("single role entry should NOT trigger message array detection")
		}
	}
}

// --- Python Tests ---

func TestParsePythonPrompts_MessageArray(t *testing.T) {
	t.Parallel()
	src := `
messages = [
    {"role": "system", "content": "You are a financial advisor."},
    {"role": "user", "content": user_query},
]
`
	surfaces := ParseEmbeddedPrompts("chat.py", src, "python")
	found := false
	for _, s := range surfaces {
		if s.Name == "message_array" {
			found = true
			if s.Confidence < 0.9 {
				t.Errorf("confidence = %.2f", s.Confidence)
			}
		}
	}
	if !found {
		t.Error("expected message_array for Python")
	}
}

func TestParsePythonPrompts_TripleQuote(t *testing.T) {
	t.Parallel()
	src := `
SYSTEM_PROMPT = """
You are a helpful AI assistant. Your role is to answer questions
based on the provided context. Do not make up information.
Always respond with accurate, cited data.
"""
`
	surfaces := ParseEmbeddedPrompts("prompts.py", src, "python")
	found := false
	for _, s := range surfaces {
		if s.Name == "docstring_prompt" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected docstring_prompt for triple-quote, got %v", surfaceNames(surfaces))
	}
}

func TestParsePythonPrompts_FewShot(t *testing.T) {
	t.Parallel()
	src := `
examples = [
    {"input": "How do I return an item?", "output": "Visit returns.example.com"},
    {"input": "What is my balance?", "output": "Check your account page"},
]
`
	surfaces := ParseEmbeddedPrompts("examples.py", src, "python")
	found := false
	for _, s := range surfaces {
		if s.Name == "few_shot_examples" {
			found = true
		}
	}
	if !found {
		t.Error("expected few_shot_examples for Python")
	}
}

func TestParsePythonPrompts_NonAI(t *testing.T) {
	t.Parallel()
	src := `
ERROR_MSG = "An error occurred. Please try again."
config = {"database": "postgres", "host": "localhost"}
`
	surfaces := ParseEmbeddedPrompts("config.py", src, "python")
	if len(surfaces) != 0 {
		t.Errorf("expected 0 for non-AI Python, got %d", len(surfaces))
	}
}

// --- Go Tests ---

func TestParseGoPrompts_BacktickString(t *testing.T) {
	t.Parallel()
	src := "var systemPrompt = `You are a helpful coding assistant. Your task is to review code and suggest improvements. Do not generate harmful or offensive content. Always respond with clear explanations.`"
	surfaces := ParseEmbeddedPrompts("prompt.go", src, "go")
	found := false
	for _, s := range surfaces {
		if s.Name == "backtick_prompt" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected backtick_prompt for Go, got %v", surfaceNames(surfaces))
	}
}

func TestParseGoPrompts_MessageStruct(t *testing.T) {
	t.Parallel()
	src := `
var messages = []Message{
	{Role: "system", Content: "You are a helpful assistant."},
	{Role: "user", Content: userInput},
}
`
	// Go role pattern uses lowercase "role": — struct fields are capitalized.
	// This should NOT match because the pattern looks for "role": not Role:.
	surfaces := ParseEmbeddedPrompts("chat.go", src, "go")
	// This is a known limitation: Go struct fields are capitalized, pattern
	// looks for JSON-style lowercase. Accepted tradeoff for false positive control.
	t.Logf("Go struct message detection: %d surfaces", len(surfaces))
}

func TestParseGoPrompts_NonAI(t *testing.T) {
	t.Parallel()
	src := "var config = `{\"host\": \"localhost\", \"port\": 8080, \"database\": \"mydb\"}`"
	surfaces := ParseEmbeddedPrompts("config.go", src, "go")
	if len(surfaces) != 0 {
		t.Errorf("expected 0 for non-AI Go, got %d", len(surfaces))
	}
}

// --- Cross-language ---

func TestParseEmbeddedPrompts_UnsupportedLanguage(t *testing.T) {
	t.Parallel()
	surfaces := ParseEmbeddedPrompts("file.rb", "puts 'hello'", "ruby")
	if surfaces != nil {
		t.Errorf("expected nil for unsupported language, got %d", len(surfaces))
	}
}

func TestParseEmbeddedPrompts_StableIDs(t *testing.T) {
	t.Parallel()
	src := `const messages = [{ role: "system", content: "You are helpful." }, { role: "user", content: q }];`
	s1 := ParseEmbeddedPrompts("src/chat.ts", src, "js")
	s2 := ParseEmbeddedPrompts("src/chat.ts", src, "js")
	if len(s1) != len(s2) {
		t.Fatalf("non-deterministic: %d vs %d", len(s1), len(s2))
	}
	for i := range s1 {
		if s1[i].SurfaceID != s2[i].SurfaceID {
			t.Errorf("ID differs: %s vs %s", s1[i].SurfaceID, s2[i].SurfaceID)
		}
	}
}
