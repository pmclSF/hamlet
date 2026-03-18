package models

// CodeSurfaceKind describes the kind of behavior anchor a CodeSurface represents.
type CodeSurfaceKind string

const (
	// SurfaceFunction is a standalone exported function.
	SurfaceFunction CodeSurfaceKind = "function"

	// SurfaceMethod is a method on a type/class.
	SurfaceMethod CodeSurfaceKind = "method"

	// SurfaceHandler is an HTTP/RPC handler or middleware.
	SurfaceHandler CodeSurfaceKind = "handler"

	// SurfaceRoute is a registered route/endpoint.
	SurfaceRoute CodeSurfaceKind = "route"

	// SurfaceClass is a class or struct with public surface area.
	SurfaceClass CodeSurfaceKind = "class"

	// SurfacePrompt is an AI prompt template or prompt-building function.
	SurfacePrompt CodeSurfaceKind = "prompt"

	// SurfaceContext is an AI context surface: system messages, policy blocks,
	// few-shot examples, safety overlays, persona definitions, dynamic instructions,
	// or context assemblers. Context surfaces are behavioral contracts — changes to
	// them alter AI behavior even when the prompt template itself is unchanged.
	SurfaceContext CodeSurfaceKind = "context"

	// SurfaceDataset is a dataset loader, fixture, or data pipeline entry point.
	SurfaceDataset CodeSurfaceKind = "dataset"

	// SurfaceToolDef is a tool/function-calling definition, schema, or description.
	SurfaceToolDef CodeSurfaceKind = "tool_definition"

	// SurfaceRetrieval is retrieval/RAG logic: query builders, chunking, embedding config,
	// vector store setup, reranker config, or context assembly.
	SurfaceRetrieval CodeSurfaceKind = "retrieval"

	// SurfaceAgent is agent/orchestration logic: routers, planners, tool-choice policies,
	// memory config, handoff logic, or fallback strategies.
	SurfaceAgent CodeSurfaceKind = "agent"

	// SurfaceEvalDef is an evaluation definition: rubrics, metrics, baseline schemas,
	// expected output definitions, or eval configuration.
	SurfaceEvalDef CodeSurfaceKind = "eval_definition"

	// SurfaceFixture is a shared test fixture: setup/teardown hooks, helper
	// builders, mock providers, stub services, or dataset loaders that multiple
	// tests depend on. High-fanout fixtures are fragility hotspots.
	SurfaceFixture CodeSurfaceKind = "fixture"
)

// Detection tier constants for the inference architecture.
const (
	// TierStructural is AST or parser-based extraction (highest confidence).
	// Currently used for: Go exported functions (capitalization rule), import
	// graph construction (language-specific import parsing).
	TierStructural = "structural"

	// TierSemantic is framework-aware heuristic detection.
	// Currently used for: LangChain/LlamaIndex message constructors,
	// RAG framework instantiation patterns, eval framework config detection.
	TierSemantic = "semantic"

	// TierPattern is regex/naming convention matching.
	// Currently used for: most surface detection (prompt*, dataset*, tool*),
	// test file discovery, framework detection from imports.
	TierPattern = "pattern"

	// TierContent is content-based inference from file contents.
	// Currently used for: message array detection, template file scanning,
	// RAG config file detection.
	TierContent = "content"
)

// CodeSurface represents an inferred behavior anchor in source code.
//
// Unlike CodeUnit (which tracks individual exported symbols for coverage
// linkage), CodeSurface identifies semantic behavior boundaries: the points
// in code where observable behavior originates. These are the natural
// targets for validation — the things tests should exercise.
//
// CodeSurfaces are inferred automatically from code structure. No manual
// YAML or configuration is required. The inference philosophy: if a function
// is exported, it has surface area. If it registers a route, it has behavior.
// If it's a handler, it transforms input to output. Terrain derives these
// anchors from the code itself.
type CodeSurface struct {
	// SurfaceID is a deterministic stable identifier.
	// Format: "surface:<path>:<name>" or "surface:<path>:<parent>.<name>".
	SurfaceID string `json:"surfaceId"`

	// Name is the local identifier (function name, method name, route path).
	Name string `json:"name"`

	// Path is the repository-relative file path containing this surface.
	Path string `json:"path"`

	// Kind classifies the behavior anchor.
	Kind CodeSurfaceKind `json:"kind"`

	// ParentName is the containing class/struct name for methods.
	ParentName string `json:"parentName,omitempty"`

	// Language is the programming language.
	Language string `json:"language"`

	// Package is the inferred package or module.
	Package string `json:"package,omitempty"`

	// Line is the source line where this surface is defined.
	Line int `json:"line,omitempty"`

	// Receiver is the type receiver for methods (Go-specific: "*Handler").
	Receiver string `json:"receiver,omitempty"`

	// Route is the HTTP route pattern when Kind is SurfaceRoute or SurfaceHandler.
	Route string `json:"route,omitempty"`

	// HTTPMethod is the HTTP method (GET, POST, etc.) when applicable.
	HTTPMethod string `json:"httpMethod,omitempty"`

	// Exported indicates whether this surface is publicly visible.
	Exported bool `json:"exported"`

	// DetectionTier records the inference method that identified this surface.
	//   "structural"  — AST or parser-based extraction (highest confidence)
	//   "semantic"    — framework-aware heuristic (e.g., LangChain constructor detection)
	//   "pattern"     — regex/naming convention matching (most common today)
	//   "content"     — content-based inference from file contents (AI-specific)
	DetectionTier string `json:"detectionTier,omitempty"`

	// Confidence is the detection confidence (0.0–1.0).
	// Reflects how certain Terrain is that this surface is correctly classified.
	Confidence float64 `json:"confidence,omitempty"`

	// ConfidenceBasis indicates whether the confidence score is calibrated
	// against ground truth data or assigned heuristically.
	// Values: "calibrated" (measured against fixture suite), "heuristic" (expert estimate).
	// Empty means heuristic (backward compatible).
	ConfidenceBasis string `json:"confidenceBasis,omitempty"`

	// Reason explains why this surface was classified with this Kind.
	// Empty for standard exports; populated for content-inferred AI surfaces.
	Reason string `json:"reason,omitempty"`

	// LinkedCodeUnit is the CodeUnit.UnitID that corresponds to this surface,
	// if one exists. This links the behavior anchor to the coverage model.
	LinkedCodeUnit string `json:"linkedCodeUnit,omitempty"`
}

// Evidence returns a unified DetectionEvidence view of this surface's
// detection metadata. This bridges the per-field evidence pattern to the
// formal DetectionEvidence struct for rendering and serialization.
func (cs *CodeSurface) Evidence() DetectionEvidence {
	return DetectionEvidence{
		DetectorID: TierFromDetectorID(extractDetectorID(cs.Reason)),
		Tier:       cs.DetectionTier,
		Confidence: cs.Confidence,
		FilePath:   cs.Path,
		Symbol:     cs.Name,
		Line:       cs.Line,
		Reason:     cs.Reason,
	}
}

// extractDetectorID pulls the "[detectorID]" from a reason string.
func extractDetectorID(reason string) string {
	if len(reason) < 3 || reason[0] != '[' {
		return ""
	}
	end := 1
	for end < len(reason) && reason[end] != ']' {
		end++
	}
	if end >= len(reason) {
		return ""
	}
	return reason[1:end]
}

// BuildSurfaceID constructs a deterministic surface ID.
func BuildSurfaceID(path, name, parent string) string {
	if parent != "" {
		return "surface:" + path + ":" + parent + "." + name
	}
	return "surface:" + path + ":" + name
}
