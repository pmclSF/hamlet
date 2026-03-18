package models

// AICapability is a canonical AI system capability inferred from code
// structure. Capabilities describe what an AI system CAN DO, independent
// of which framework implements it. This enables capability-level impact
// analysis: "this change affects the retrieval_augmented_generation
// capability" rather than "this change affects a Chroma vector store."
type AICapability string

const (
	// CapabilityRAG is retrieval-augmented generation — the system retrieves
	// context from external sources to ground LLM responses.
	// Inferred from: SurfaceRetrieval, RAG pipeline components, vector stores,
	// embeddings, chunking, retriever construction.
	CapabilityRAG AICapability = "retrieval_augmented_generation"

	// CapabilityToolUse is tool/function calling — the system can invoke
	// external tools based on LLM decisions.
	// Inferred from: SurfaceToolDef, function schemas, tool decorators,
	// structured output with tool routing.
	CapabilityToolUse AICapability = "tool_use"

	// CapabilityPromptGeneration is prompt construction — the system builds
	// or manages prompt templates, system messages, or few-shot examples.
	// Inferred from: SurfacePrompt, SurfaceContext, template factories,
	// message array construction, prompt builder functions.
	CapabilityPromptGeneration AICapability = "prompt_generation"

	// CapabilityStructuredOutput is structured/constrained generation — the
	// system constrains LLM output to a schema (JSON, Pydantic, Zod).
	// Inferred from: SurfaceToolDef with schema patterns, response_model,
	// JSON mode, output parsers.
	CapabilityStructuredOutput AICapability = "structured_output"

	// CapabilityCitation is source attribution — the system tracks which
	// retrieved documents contributed to a response.
	// Inferred from: citation assembly patterns, source attribution logic,
	// retrieval chain with source tracking.
	CapabilityCitation AICapability = "citation_assembly"

	// CapabilitySafety is safety guardrailing — the system enforces content
	// safety, input validation, or output filtering.
	// Inferred from: safety overlay surfaces, guardrail configurations,
	// content filtering patterns, moderation API calls.
	CapabilitySafety AICapability = "safety_guardrailing"

	// CapabilityMemory is conversational memory — the system maintains
	// state across conversation turns.
	// Inferred from: memory window configurations, conversation history
	// management, session state patterns.
	CapabilityMemory AICapability = "conversational_memory"

	// CapabilityAgentOrchestration is multi-step agent orchestration —
	// the system coordinates multiple LLM calls or tools in a workflow.
	// Inferred from: SurfaceAgent, agent routers, step budgets, planners,
	// handoff logic, multi-step execution.
	CapabilityAgentOrchestration AICapability = "agent_orchestration"

	// CapabilityEvaluation is AI evaluation — the system measures AI output
	// quality against rubrics or baselines.
	// Inferred from: SurfaceEvalDef, eval metrics, scoring functions,
	// baseline schemas, grading criteria.
	CapabilityEvaluation AICapability = "evaluation"
)

// AllAICapabilities returns all canonical capability values.
func AllAICapabilities() []AICapability {
	return []AICapability{
		CapabilityRAG,
		CapabilityToolUse,
		CapabilityPromptGeneration,
		CapabilityStructuredOutput,
		CapabilityCitation,
		CapabilitySafety,
		CapabilityMemory,
		CapabilityAgentOrchestration,
		CapabilityEvaluation,
	}
}

// CapabilityLabel returns a user-friendly label for a capability.
func CapabilityLabel(cap AICapability) string {
	switch cap {
	case CapabilityRAG:
		return "Retrieval-Augmented Generation"
	case CapabilityToolUse:
		return "Tool / Function Calling"
	case CapabilityPromptGeneration:
		return "Prompt Generation"
	case CapabilityStructuredOutput:
		return "Structured Output"
	case CapabilityCitation:
		return "Citation / Source Attribution"
	case CapabilitySafety:
		return "Safety Guardrailing"
	case CapabilityMemory:
		return "Conversational Memory"
	case CapabilityAgentOrchestration:
		return "Agent Orchestration"
	case CapabilityEvaluation:
		return "Evaluation"
	default:
		return string(cap)
	}
}

// InferredCapability represents a capability detected in the codebase
// with evidence linking it to specific surfaces.
type InferredCapability struct {
	// Capability is the canonical capability identifier.
	Capability AICapability `json:"capability"`

	// Label is the user-friendly name.
	Label string `json:"label"`

	// SurfaceIDs are the CodeSurface IDs that evidence this capability.
	SurfaceIDs []string `json:"surfaceIds"`

	// ScenarioIDs are the Scenario IDs that validate this capability.
	ScenarioIDs []string `json:"scenarioIds,omitempty"`

	// Confidence is the inference confidence (0.0–1.0).
	Confidence float64 `json:"confidence"`

	// Covered is true if at least one scenario validates this capability.
	Covered bool `json:"covered"`
}
