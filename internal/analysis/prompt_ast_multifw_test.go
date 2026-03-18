package analysis

import (
	"strings"
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

// --- Vercel AI SDK ---

func TestASTJS_VercelAISDK_GenerateText(t *testing.T) {
	t.Parallel()
	src := `
import { generateText } from 'ai';
import { openai } from '@ai-sdk/openai';

const result = await generateText({
  model: openai('gpt-4'),
  system: 'You are a helpful assistant.',
  prompt: userInput,
});
`
	surfaces := ParsePromptAST("src/chat.ts", src, "js")

	found := false
	for _, s := range surfaces {
		if strings.Contains(s.Name, "llm_call") || strings.Contains(s.Name, "generation") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected Vercel AI SDK generateText to be detected, got %v", surfaceNames(surfaces))
	}
}

func TestASTJS_VercelAISDK_StreamText(t *testing.T) {
	t.Parallel()
	src := `
import { streamText } from 'ai';
const stream = await streamText({
  model: openai('gpt-4'),
  messages: [{ role: 'system', content: 'You are helpful.' }],
});
`
	surfaces := ParsePromptAST("src/stream.ts", src, "js")

	found := false
	for _, s := range surfaces {
		if strings.Contains(s.Name, "llm_call") {
			found = true
			if s.Kind != models.SurfaceContext {
				t.Errorf("streamText with system message: want context, got %s", s.Kind)
			}
		}
	}
	if !found {
		t.Errorf("expected streamText to be detected, got %v", surfaceNames(surfaces))
	}
}

func TestASTJS_VercelAISDK_Tool(t *testing.T) {
	t.Parallel()
	src := `
import { tool } from 'ai';
const weatherTool = tool({
  description: 'Get the weather for a location',
  parameters: z.object({ city: z.string() }),
  execute: async ({ city }) => getWeather(city),
});
`
	surfaces := ParsePromptAST("src/tools.ts", src, "js")

	found := false
	for _, s := range surfaces {
		if s.Kind == models.SurfaceToolDef {
			found = true
		}
	}
	if !found {
		t.Errorf("expected Vercel AI SDK tool() to be detected as tool_definition, got %v", surfaceNames(surfaces))
	}
}

func TestASTJS_VercelAISDK_UseChat(t *testing.T) {
	t.Parallel()
	src := `
import { useChat } from 'ai/react';
const { messages, input, handleSubmit } = useChat({
  api: '/api/chat',
});
`
	surfaces := ParsePromptAST("src/ChatUI.tsx", src, "js")

	found := false
	for _, s := range surfaces {
		if strings.Contains(s.Name, "llm_call") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected useChat hook to be detected, got %v", surfaceNames(surfaces))
	}
}

// --- Direct SDK constructors ---

func TestASTJS_OpenAIConstructor(t *testing.T) {
	t.Parallel()
	src := `
import OpenAI from 'openai';
const client = new OpenAI({ apiKey: process.env.OPENAI_API_KEY });
`
	surfaces := ParsePromptAST("src/client.ts", src, "js")

	found := false
	for _, s := range surfaces {
		if strings.Contains(s.Name, "sdk_client") && strings.Contains(s.Name, "OpenAI") {
			found = true
			if s.Kind != models.SurfaceAgent {
				t.Errorf("SDK constructor: want agent, got %s", s.Kind)
			}
		}
	}
	if !found {
		t.Errorf("expected new OpenAI() to be detected, got %v", surfaceNames(surfaces))
	}
}

func TestASTJS_AnthropicConstructor(t *testing.T) {
	t.Parallel()
	src := `
import Anthropic from '@anthropic-ai/sdk';
const anthropic = new Anthropic();
`
	surfaces := ParsePromptAST("src/client.ts", src, "js")

	found := false
	for _, s := range surfaces {
		if strings.Contains(s.Name, "sdk_client") && strings.Contains(s.Name, "Anthropic") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected new Anthropic() to be detected, got %v", surfaceNames(surfaces))
	}
}

// --- Python: Mirascope ---

func TestASTPython_Mirascope_PromptTemplate(t *testing.T) {
	t.Parallel()
	src := `
from mirascope.core import prompt_template

@prompt_template()
def recommend_book(genre: str) -> str:
    return f"Recommend a {genre} book"
`
	surfaces := ParsePromptAST("src/rec.py", src, "python")

	found := false
	for _, s := range surfaces {
		if strings.Contains(s.Name, "ai_decorator") {
			found = true
			if s.Kind != models.SurfacePrompt {
				t.Errorf("@prompt_template: want prompt, got %s", s.Kind)
			}
		}
	}
	if !found {
		t.Errorf("expected @prompt_template to be detected, got %v", surfaceNames(surfaces))
	}
}

func TestASTPython_Mirascope_OpenAICall(t *testing.T) {
	t.Parallel()
	src := `
from mirascope.core import openai

@openai.call("gpt-4")
def recommend(genre: str) -> str:
    return f"Recommend a {genre} book"
`
	surfaces := ParsePromptAST("src/rec.py", src, "python")

	found := false
	for _, s := range surfaces {
		if strings.Contains(s.Name, "ai_decorator") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected @openai.call to be detected, got %v", surfaceNames(surfaces))
	}
}

// --- Python: Marvin ---

func TestASTPython_Marvin_Fn(t *testing.T) {
	t.Parallel()
	src := `
import marvin

@marvin.fn
def generate_recipe(ingredients: list[str]) -> str:
    """Generate a recipe from ingredients."""
`
	surfaces := ParsePromptAST("src/recipe.py", src, "python")

	found := false
	for _, s := range surfaces {
		if strings.Contains(s.Name, "ai_decorator") {
			found = true
			if s.Kind != models.SurfaceToolDef {
				t.Errorf("@marvin.fn: want tool_definition, got %s", s.Kind)
			}
		}
	}
	if !found {
		t.Errorf("expected @marvin.fn to be detected, got %v", surfaceNames(surfaces))
	}
}

func TestASTPython_Marvin_Extract(t *testing.T) {
	t.Parallel()
	src := `
import marvin

result = marvin.extract("text about colors", target=str, instructions="extract color names")
`
	surfaces := ParsePromptAST("src/extract.py", src, "python")

	found := false
	for _, s := range surfaces {
		if strings.Contains(s.Name, "structured_output") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected marvin.extract to be detected, got %v", surfaceNames(surfaces))
	}
}

// --- Python: DSPy ---

func TestASTPython_DSPy_Signature(t *testing.T) {
	t.Parallel()
	src := `
import dspy

class BasicQA(dspy.Signature):
    """Answer questions given context."""
    context = dspy.InputField()
    question = dspy.InputField()
    answer = dspy.OutputField()
`
	surfaces := ParsePromptAST("src/qa.py", src, "python")

	found := false
	for _, s := range surfaces {
		if strings.Contains(s.Name, "dspy") && strings.Contains(s.Name, "Signature") {
			found = true
			if s.Kind != models.SurfacePrompt {
				t.Errorf("dspy.Signature: want prompt, got %s", s.Kind)
			}
		}
	}
	if !found {
		t.Errorf("expected dspy.Signature to be detected, got %v", surfaceNames(surfaces))
	}
}

func TestASTPython_DSPy_ChainOfThought(t *testing.T) {
	t.Parallel()
	src := `
import dspy
cot = dspy.ChainOfThought("question -> answer")
`
	surfaces := ParsePromptAST("src/cot.py", src, "python")

	found := false
	for _, s := range surfaces {
		if strings.Contains(s.Name, "dspy") && strings.Contains(s.Name, "ChainOfThought") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected dspy.ChainOfThought to be detected, got %v", surfaceNames(surfaces))
	}
}

func TestASTPython_DSPy_Retrieve(t *testing.T) {
	t.Parallel()
	src := `
import dspy
retriever = dspy.Retrieve(k=5)
`
	surfaces := ParsePromptAST("src/rag.py", src, "python")

	found := false
	for _, s := range surfaces {
		if strings.Contains(s.Name, "dspy") && strings.Contains(s.Name, "Retrieve") {
			found = true
			if s.Kind != models.SurfaceRetrieval {
				t.Errorf("dspy.Retrieve: want retrieval, got %s", s.Kind)
			}
		}
	}
	if !found {
		t.Errorf("expected dspy.Retrieve to be detected as retrieval, got %v", surfaceNames(surfaces))
	}
}

// --- Python: Instructor ---

func TestASTPython_Instructor_ResponseModel(t *testing.T) {
	t.Parallel()
	src := `
import instructor
from openai import OpenAI

client = instructor.patch(OpenAI())
result = client.chat.completions.create(
    model="gpt-4",
    response_model=UserProfile,
    messages=[{"role": "user", "content": "Extract user info"}],
)
`
	surfaces := ParsePromptAST("src/extract.py", src, "python")

	// Should detect: SDK constructor, generation call with response_model, structured output
	kinds := map[models.CodeSurfaceKind]bool{}
	for _, s := range surfaces {
		kinds[s.Kind] = true
	}
	if !kinds[models.SurfaceToolDef] {
		t.Error("expected structured output (response_model) to produce tool_definition surface")
	}
}

// --- LangChain backward compatibility ---

func TestASTJS_LangChainStillWorks(t *testing.T) {
	t.Parallel()
	src := `
import { SystemMessage, HumanMessage } from "@langchain/core/messages";
import { ChatPromptTemplate } from "@langchain/core/prompts";

const messages = [
  new SystemMessage("You are a helpful assistant."),
  new HumanMessage(userQuery),
];

const prompt = ChatPromptTemplate.fromMessages([
  ["system", "You are a {role} assistant."],
  ["human", "{input}"],
]);
`
	surfaces := ParsePromptAST("src/chain.ts", src, "js")

	var systemFound, templateFound bool
	for _, s := range surfaces {
		if strings.Contains(s.Name, "SystemMessage") {
			systemFound = true
		}
		if strings.Contains(s.Name, "template_prompt") && strings.Contains(s.Name, "ChatPromptTemplate") {
			templateFound = true
		}
	}
	if !systemFound {
		t.Error("LangChain SystemMessage regression: not detected")
	}
	if !templateFound {
		t.Error("LangChain ChatPromptTemplate regression: not detected")
	}
}

func TestASTPython_LangChainStillWorks(t *testing.T) {
	t.Parallel()
	src := `
from langchain.schema import SystemMessage, HumanMessage
from langchain.prompts import ChatPromptTemplate

messages = [
    SystemMessage(content="You are a helpful assistant."),
    HumanMessage(content=user_query),
]

prompt = ChatPromptTemplate.from_messages([
    ("system", "You are a {role} assistant."),
    ("human", "{input}"),
])
`
	surfaces := ParsePromptAST("src/chain.py", src, "python")

	var systemFound, templateFound bool
	for _, s := range surfaces {
		if strings.Contains(s.Name, "SystemMessage") {
			systemFound = true
		}
		if strings.Contains(s.Name, "template_prompt") {
			templateFound = true
		}
	}
	if !systemFound {
		t.Error("LangChain SystemMessage regression: not detected (Python)")
	}
	if !templateFound {
		t.Error("LangChain ChatPromptTemplate regression: not detected (Python)")
	}
}

// --- False positive rejection ---

func TestASTJS_RejectsNonAITool(t *testing.T) {
	t.Parallel()
	src := `
const tools = ['hammer', 'screwdriver', 'wrench'];
const tool = { name: 'pliers' };
`
	surfaces := ParsePromptAST("src/hardware.ts", src, "js")

	for _, s := range surfaces {
		if s.Kind == models.SurfaceToolDef {
			t.Errorf("should not detect non-AI tool arrays, got %s: %s", s.Name, s.Reason)
		}
	}
}

func TestASTPython_RejectsNonAIDecorator(t *testing.T) {
	t.Parallel()
	src := `
from flask import Flask
app = Flask(__name__)

@app.route("/api/users")
def get_users():
    return []
`
	surfaces := ParsePromptAST("src/api.py", src, "python")

	for _, s := range surfaces {
		if strings.Contains(s.Name, "ai_decorator") {
			t.Errorf("should not detect @app.route as AI decorator, got %s", s.Name)
		}
	}
}
