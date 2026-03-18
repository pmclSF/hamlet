package analysis

import (
	"strings"
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

// --- JS/TS Structural Tests ---

func TestStructuralJS_MessageArray(t *testing.T) {
	t.Parallel()
	src := `
const messages = [
  { role: "system", content: "You are a helpful assistant." },
  { role: "user", content: userInput },
  { role: "assistant", content: "How can I help?" },
];
`
	surfaces := ParseStructural("src/chat.ts", src, "js")
	found := false
	for _, s := range surfaces {
		if strings.Contains(s.Name, "message_array") {
			found = true
			if s.DetectionTier != models.TierStructural {
				t.Errorf("tier = %q, want structural", s.DetectionTier)
			}
			if s.Confidence < 0.93 {
				t.Errorf("confidence = %.2f, want >= 0.93", s.Confidence)
			}
			if !strings.Contains(s.Reason, DetectorStructuralMessageArray) {
				t.Errorf("reason should include detector ID, got: %s", s.Reason)
			}
			if !strings.Contains(s.Name, "messages") {
				t.Errorf("expected variable name 'messages' in surface name, got: %s", s.Name)
			}
		}
	}
	if !found {
		t.Errorf("expected structural message array, got %v", surfaceNames(surfaces))
	}
}

func TestStructuralJS_NestedMessageBuilder(t *testing.T) {
	t.Parallel()
	src := `
function buildPrompt(query, context) {
  const messages = [
    { role: "system", content: "You are a QA assistant." },
    { role: "user", content: query },
  ];
  return messages;
}
`
	surfaces := ParseStructural("src/builder.ts", src, "js")
	// Should find both: message_array AND prompt_builder
	arrayFound := false
	builderFound := false
	for _, s := range surfaces {
		if strings.Contains(s.Name, "message_array") {
			arrayFound = true
		}
		if strings.Contains(s.Name, "prompt_builder") {
			builderFound = true
		}
	}
	if !arrayFound {
		t.Error("expected message_array inside function")
	}
	if !builderFound {
		t.Errorf("expected prompt_builder detection, got %v", surfaceNames(surfaces))
	}
}

func TestStructuralJS_FewShotArray(t *testing.T) {
	t.Parallel()
	src := `
const fewShotExamples = [
  { input: "What is the weather?", output: "I can check weather for you." },
  { input: "Book a flight", output: "I'll help you book a flight." },
  { input: "Cancel order", output: "I can help cancel your order." },
];
`
	surfaces := ParseStructural("src/examples.ts", src, "js")
	found := false
	for _, s := range surfaces {
		if strings.Contains(s.Name, "few_shot") {
			found = true
			if s.DetectionTier != models.TierStructural {
				t.Errorf("tier = %q, want structural", s.DetectionTier)
			}
		}
	}
	if !found {
		t.Errorf("expected few-shot array, got %v", surfaceNames(surfaces))
	}
}

func TestStructuralJS_ExportedPromptConstant(t *testing.T) {
	t.Parallel()
	src := `const systemInstructions = "You are a helpful AI assistant. Your role is to answer questions based on the provided documentation. Do not make up information. Always respond with cited data.";`
	surfaces := ParseStructural("src/config.ts", src, "js")
	found := false
	for _, s := range surfaces {
		if strings.Contains(s.Name, "prompt_const") {
			found = true
			if !strings.Contains(s.Name, "systemInstructions") {
				t.Errorf("expected variable name in surface, got: %s", s.Name)
			}
		}
	}
	if !found {
		t.Errorf("expected prompt constant, got %v", surfaceNames(surfaces))
	}
}

func TestStructuralJS_HelperFunctionReturningMessages(t *testing.T) {
	t.Parallel()
	src := `
function createMessageContext(user, history) {
  return [
    { role: "system", content: "You are a chat assistant." },
    ...history,
    { role: "user", content: user },
  ];
}
`
	surfaces := ParseStructural("src/helpers.ts", src, "js")
	found := false
	for _, s := range surfaces {
		if strings.Contains(s.Name, "prompt_builder") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected prompt_builder for helper function, got %v", surfaceNames(surfaces))
	}
}

func TestStructuralJS_NonAI_NoDetection(t *testing.T) {
	t.Parallel()
	src := `
const users = [
  { name: "Alice", role: "admin" },
  { name: "Bob", role: "viewer" },
];
const config = { database: "postgres", host: "localhost" };
`
	surfaces := ParseStructural("src/config.ts", src, "js")
	// "role: admin" should NOT be detected — it's not system/user/assistant.
	for _, s := range surfaces {
		if strings.Contains(s.Name, "message_array") {
			t.Errorf("non-AI role values should NOT trigger detection: %s", s.Name)
		}
	}
}

// --- Python Structural Tests ---

func TestStructuralPython_MessageList(t *testing.T) {
	t.Parallel()
	src := `
messages = [
    {"role": "system", "content": "You are a financial advisor."},
    {"role": "user", "content": user_query},
]
`
	surfaces := ParseStructural("chat.py", src, "python")
	found := false
	for _, s := range surfaces {
		if strings.Contains(s.Name, "message_array") {
			found = true
			if s.DetectionTier != models.TierStructural {
				t.Errorf("tier = %q", s.DetectionTier)
			}
			if !strings.Contains(s.Name, "messages") {
				t.Errorf("expected var name, got: %s", s.Name)
			}
		}
	}
	if !found {
		t.Error("expected message_array for Python")
	}
}

func TestStructuralPython_FewShotList(t *testing.T) {
	t.Parallel()
	src := `
examples = [
    {"input": "How do I return?", "output": "Visit returns.example.com"},
    {"input": "Track my order", "output": "Check your tracking page"},
]
`
	surfaces := ParseStructural("examples.py", src, "python")
	found := false
	for _, s := range surfaces {
		if strings.Contains(s.Name, "few_shot") {
			found = true
		}
	}
	if !found {
		t.Error("expected few-shot for Python")
	}
}

func TestStructuralPython_PromptBuilderFunc(t *testing.T) {
	t.Parallel()
	src := `
def build_prompt_messages(query, context):
    return [
        {"role": "system", "content": "You are a QA bot."},
        {"role": "user", "content": query},
    ]
`
	surfaces := ParseStructural("builder.py", src, "python")
	found := false
	for _, s := range surfaces {
		if strings.Contains(s.Name, "prompt_builder") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected prompt_builder for Python, got %v", surfaceNames(surfaces))
	}
}

func TestStructuralPython_NonAI(t *testing.T) {
	t.Parallel()
	src := `
users = [
    {"name": "Alice", "role": "admin"},
    {"name": "Bob", "role": "viewer"},
]
`
	surfaces := ParseStructural("users.py", src, "python")
	for _, s := range surfaces {
		if strings.Contains(s.Name, "message_array") {
			t.Errorf("non-AI roles should NOT trigger: %s", s.Name)
		}
	}
}

// --- Cross-language ---

func TestStructural_StableIDs(t *testing.T) {
	t.Parallel()
	src := `const msgs = [{ role: "system", content: "Hi" }, { role: "user", content: "Q" }];`
	s1 := ParseStructural("src/x.ts", src, "js")
	s2 := ParseStructural("src/x.ts", src, "js")
	if len(s1) != len(s2) {
		t.Fatalf("non-deterministic: %d vs %d", len(s1), len(s2))
	}
	for i := range s1 {
		if s1[i].SurfaceID != s2[i].SurfaceID {
			t.Errorf("ID differs: %s vs %s", s1[i].SurfaceID, s2[i].SurfaceID)
		}
	}
}

func TestStructural_EvidenceMetadata(t *testing.T) {
	t.Parallel()
	src := `const messages = [{ role: "system", content: "Hello" }, { role: "user", content: "Q" }];`
	surfaces := ParseStructural("src/chat.ts", src, "js")
	for _, s := range surfaces {
		if s.DetectionTier == "" {
			t.Errorf("surface %q missing tier", s.Name)
		}
		if s.Confidence <= 0 {
			t.Errorf("surface %q missing confidence", s.Name)
		}
		if s.Reason == "" {
			t.Errorf("surface %q missing reason", s.Name)
		}
		if !strings.Contains(s.Reason, "[") {
			t.Errorf("surface %q reason should contain detector ID, got: %s", s.Name, s.Reason)
		}
	}
}

func TestStructural_Unsupported(t *testing.T) {
	t.Parallel()
	if ParseStructural("x.rb", "", "ruby") != nil {
		t.Error("expected nil")
	}
}
