package analysis

import (
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

// --- Surface-to-capability inference ---

func TestInferAICapabilities_RAGFromRetrieval(t *testing.T) {
	t.Parallel()
	surfaces := []models.CodeSurface{
		{SurfaceID: "s:retriever", Kind: models.SurfaceRetrieval, Name: "retrieverConfig", Confidence: 0.92},
		{SurfaceID: "s:embedding", Kind: models.SurfaceRetrieval, Name: "embeddingModel", Confidence: 0.90},
	}
	caps := InferAICapabilities(surfaces, nil)

	found := findCapability(caps, models.CapabilityRAG)
	if found == nil {
		t.Fatal("expected RAG capability from SurfaceRetrieval")
	}
	if len(found.SurfaceIDs) != 2 {
		t.Errorf("expected 2 supporting surfaces, got %d", len(found.SurfaceIDs))
	}
	if found.Confidence < 0.90 {
		t.Errorf("confidence: want >= 0.90, got %.2f", found.Confidence)
	}
}

func TestInferAICapabilities_ToolUseFromToolDef(t *testing.T) {
	t.Parallel()
	surfaces := []models.CodeSurface{
		{SurfaceID: "s:tool1", Kind: models.SurfaceToolDef, Name: "weatherTool", Confidence: 0.88},
	}
	caps := InferAICapabilities(surfaces, nil)

	found := findCapability(caps, models.CapabilityToolUse)
	if found == nil {
		t.Fatal("expected tool_use capability from SurfaceToolDef")
	}
}

func TestInferAICapabilities_PromptFromPromptSurface(t *testing.T) {
	t.Parallel()
	surfaces := []models.CodeSurface{
		{SurfaceID: "s:prompt", Kind: models.SurfacePrompt, Name: "buildUserPrompt", Confidence: 0.85},
	}
	caps := InferAICapabilities(surfaces, nil)

	found := findCapability(caps, models.CapabilityPromptGeneration)
	if found == nil {
		t.Fatal("expected prompt_generation from SurfacePrompt")
	}
}

func TestInferAICapabilities_PromptFromContextSurface(t *testing.T) {
	t.Parallel()
	surfaces := []models.CodeSurface{
		{SurfaceID: "s:ctx", Kind: models.SurfaceContext, Name: "systemMessage", Confidence: 0.95},
	}
	caps := InferAICapabilities(surfaces, nil)

	found := findCapability(caps, models.CapabilityPromptGeneration)
	if found == nil {
		t.Fatal("expected prompt_generation from SurfaceContext")
	}
}

func TestInferAICapabilities_SafetyFromContextName(t *testing.T) {
	t.Parallel()
	surfaces := []models.CodeSurface{
		{SurfaceID: "s:safety", Kind: models.SurfaceContext, Name: "safetyOverlay", Confidence: 0.85},
	}
	caps := InferAICapabilities(surfaces, nil)

	found := findCapability(caps, models.CapabilitySafety)
	if found == nil {
		t.Fatal("expected safety_guardrailing from 'safetyOverlay' context surface")
	}
}

func TestInferAICapabilities_MemoryFromContextName(t *testing.T) {
	t.Parallel()
	surfaces := []models.CodeSurface{
		{SurfaceID: "s:mem", Kind: models.SurfaceContext, Name: "conversationHistory", Confidence: 0.80},
	}
	caps := InferAICapabilities(surfaces, nil)

	found := findCapability(caps, models.CapabilityMemory)
	if found == nil {
		t.Fatal("expected conversational_memory from 'conversationHistory' context")
	}
}

func TestInferAICapabilities_AgentOrchestration(t *testing.T) {
	t.Parallel()
	surfaces := []models.CodeSurface{
		{SurfaceID: "s:agent", Kind: models.SurfaceAgent, Name: "agentRouter", Confidence: 0.80},
	}
	caps := InferAICapabilities(surfaces, nil)

	found := findCapability(caps, models.CapabilityAgentOrchestration)
	if found == nil {
		t.Fatal("expected agent_orchestration from SurfaceAgent")
	}
}

func TestInferAICapabilities_EvalFromEvalDef(t *testing.T) {
	t.Parallel()
	surfaces := []models.CodeSurface{
		{SurfaceID: "s:eval", Kind: models.SurfaceEvalDef, Name: "accuracyRubric", Confidence: 0.80},
	}
	caps := InferAICapabilities(surfaces, nil)

	found := findCapability(caps, models.CapabilityEvaluation)
	if found == nil {
		t.Fatal("expected evaluation from SurfaceEvalDef")
	}
}

func TestInferAICapabilities_CitationFromRetrievalName(t *testing.T) {
	t.Parallel()
	surfaces := []models.CodeSurface{
		{SurfaceID: "s:cite", Kind: models.SurfaceRetrieval, Name: "citationAssembly", Confidence: 0.88},
	}
	caps := InferAICapabilities(surfaces, nil)

	ragCap := findCapability(caps, models.CapabilityRAG)
	citeCap := findCapability(caps, models.CapabilityCitation)

	if ragCap == nil {
		t.Error("expected RAG capability (parent of citation)")
	}
	if citeCap == nil {
		t.Error("expected citation_assembly sub-capability from 'citationAssembly' name")
	}
}

func TestInferAICapabilities_StructuredOutputFromToolSchema(t *testing.T) {
	t.Parallel()
	surfaces := []models.CodeSurface{
		{SurfaceID: "s:schema", Kind: models.SurfaceToolDef, Name: "outputSchema", Reason: "[semantic:zod-schema] structured output", Confidence: 0.88},
	}
	caps := InferAICapabilities(surfaces, nil)

	toolCap := findCapability(caps, models.CapabilityToolUse)
	structCap := findCapability(caps, models.CapabilityStructuredOutput)

	if toolCap == nil {
		t.Error("expected tool_use from SurfaceToolDef")
	}
	if structCap == nil {
		t.Error("expected structured_output from schema-named tool def")
	}
}

// --- Cross-framework capability inference ---

func TestInferAICapabilities_VercelAISDK(t *testing.T) {
	t.Parallel()
	// Simulate surfaces detected from a Vercel AI SDK codebase.
	surfaces := []models.CodeSurface{
		{SurfaceID: "s:gen", Kind: models.SurfacePrompt, Name: "llm_call_result", Reason: "[structural:ast-message-array] AST: LLM generation/streaming call", Confidence: 0.90},
		{SurfaceID: "s:tool", Kind: models.SurfaceToolDef, Name: "structured_output", Reason: "[structural:ast-template-prompt] AST: structured output", Confidence: 0.88},
	}
	caps := InferAICapabilities(surfaces, nil)

	if findCapability(caps, models.CapabilityPromptGeneration) == nil {
		t.Error("Vercel AI SDK: expected prompt_generation")
	}
	if findCapability(caps, models.CapabilityToolUse) == nil {
		t.Error("Vercel AI SDK: expected tool_use")
	}
}

func TestInferAICapabilities_DSPy(t *testing.T) {
	t.Parallel()
	surfaces := []models.CodeSurface{
		{SurfaceID: "s:sig", Kind: models.SurfacePrompt, Name: "dspy_Signature", Confidence: 0.90},
		{SurfaceID: "s:ret", Kind: models.SurfaceRetrieval, Name: "dspy_Retrieve", Confidence: 0.90},
		{SurfaceID: "s:mod", Kind: models.SurfaceAgent, Name: "dspy_Module", Confidence: 0.90},
	}
	caps := InferAICapabilities(surfaces, nil)

	if findCapability(caps, models.CapabilityPromptGeneration) == nil {
		t.Error("DSPy: expected prompt_generation from Signature")
	}
	if findCapability(caps, models.CapabilityRAG) == nil {
		t.Error("DSPy: expected RAG from Retrieve")
	}
	if findCapability(caps, models.CapabilityAgentOrchestration) == nil {
		t.Error("DSPy: expected agent_orchestration from Module")
	}
}

func TestInferAICapabilities_LangChain(t *testing.T) {
	t.Parallel()
	surfaces := []models.CodeSurface{
		{SurfaceID: "s:sys", Kind: models.SurfaceContext, Name: "systemMessage", Confidence: 0.95},
		{SurfaceID: "s:ret", Kind: models.SurfaceRetrieval, Name: "vectorStore", Confidence: 0.92},
		{SurfaceID: "s:tool", Kind: models.SurfaceToolDef, Name: "toolSchema", Confidence: 0.85},
	}
	caps := InferAICapabilities(surfaces, nil)

	if findCapability(caps, models.CapabilityPromptGeneration) == nil {
		t.Error("LangChain: expected prompt_generation")
	}
	if findCapability(caps, models.CapabilityRAG) == nil {
		t.Error("LangChain: expected RAG")
	}
	if findCapability(caps, models.CapabilityToolUse) == nil {
		t.Error("LangChain: expected tool_use")
	}
}

// --- Scenario-to-canonical capability matching ---

func TestMatchCanonicalCapability(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input string
		want  string
	}{
		{"retrieval-quality", string(models.CapabilityRAG)},
		{"rag-accuracy", string(models.CapabilityRAG)},
		{"tool-routing", string(models.CapabilityToolUse)},
		{"prompt-safety", string(models.CapabilitySafety)},
		{"agent-workflow", string(models.CapabilityAgentOrchestration)},
		{"extraction-quality", string(models.CapabilityStructuredOutput)},
		{"eval-accuracy", string(models.CapabilityEvaluation)},
		{"billing-lookup", ""}, // domain-specific, no canonical match
	}
	for _, tc := range cases {
		got := matchCanonicalCapability(tc.input)
		if got != tc.want {
			t.Errorf("matchCanonicalCapability(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

// --- Scenario coverage tracking ---

func TestInferAICapabilities_ScenarioCoverage(t *testing.T) {
	t.Parallel()
	surfaces := []models.CodeSurface{
		{SurfaceID: "s:ret", Kind: models.SurfaceRetrieval, Name: "retriever", Confidence: 0.92},
	}
	scenarios := []models.Scenario{
		{ScenarioID: "sc:1", Capability: "retrieval-quality"},
	}

	caps := InferAICapabilities(surfaces, scenarios)

	ragCap := findCapability(caps, models.CapabilityRAG)
	if ragCap == nil {
		t.Fatal("expected RAG capability")
	}
	if !ragCap.Covered {
		t.Error("RAG capability should be covered (scenario with retrieval capability exists)")
	}
	if len(ragCap.ScenarioIDs) != 1 {
		t.Errorf("expected 1 scenario, got %d", len(ragCap.ScenarioIDs))
	}
}

func TestInferAICapabilities_UncoveredCapability(t *testing.T) {
	t.Parallel()
	surfaces := []models.CodeSurface{
		{SurfaceID: "s:tool", Kind: models.SurfaceToolDef, Name: "searchTool", Confidence: 0.85},
	}
	// No scenarios validate tool_use.
	caps := InferAICapabilities(surfaces, nil)

	toolCap := findCapability(caps, models.CapabilityToolUse)
	if toolCap == nil {
		t.Fatal("expected tool_use capability")
	}
	if toolCap.Covered {
		t.Error("tool_use should be uncovered (no scenarios)")
	}
}

// --- Empty input ---

func TestInferAICapabilities_NoSurfaces(t *testing.T) {
	t.Parallel()
	caps := InferAICapabilities(nil, nil)
	if len(caps) != 0 {
		t.Errorf("expected 0 capabilities from empty input, got %d", len(caps))
	}
}

func TestInferAICapabilities_NonAISurfaces(t *testing.T) {
	t.Parallel()
	surfaces := []models.CodeSurface{
		{SurfaceID: "s:fn", Kind: models.SurfaceFunction, Name: "processPayment", Confidence: 0.90},
		{SurfaceID: "s:route", Kind: models.SurfaceRoute, Name: "GET /api/users", Confidence: 0.95},
	}
	caps := InferAICapabilities(surfaces, nil)
	if len(caps) != 0 {
		t.Errorf("expected 0 AI capabilities from non-AI surfaces, got %d", len(caps))
	}
}

// --- Determinism ---

func TestInferAICapabilities_Deterministic(t *testing.T) {
	t.Parallel()
	surfaces := []models.CodeSurface{
		{SurfaceID: "s:ret", Kind: models.SurfaceRetrieval, Name: "retriever", Confidence: 0.92},
		{SurfaceID: "s:prompt", Kind: models.SurfacePrompt, Name: "buildPrompt", Confidence: 0.85},
		{SurfaceID: "s:tool", Kind: models.SurfaceToolDef, Name: "searchTool", Confidence: 0.88},
	}

	c1 := InferAICapabilities(surfaces, nil)
	c2 := InferAICapabilities(surfaces, nil)

	if len(c1) != len(c2) {
		t.Fatalf("non-deterministic count: %d vs %d", len(c1), len(c2))
	}
	for i := range c1 {
		if c1[i].Capability != c2[i].Capability {
			t.Errorf("non-deterministic at %d: %s vs %s", i, c1[i].Capability, c2[i].Capability)
		}
	}
}

// --- Helpers ---

func findCapability(caps []models.InferredCapability, target models.AICapability) *models.InferredCapability {
	for i, c := range caps {
		if c.Capability == target {
			return &caps[i]
		}
	}
	return nil
}
