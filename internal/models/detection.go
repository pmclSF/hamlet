package models

import "strings"

// DetectionEvidence records how and why Terrain inferred something.
// This is the unified evidence model for all detection paths.
//
// Every inferred entity (surface, framework, test classification, scenario)
// should carry DetectionEvidence so users and CI pipelines can assess trust.
//
// Detection tiers (priority order):
//
//	Tier 1 — structural: AST parsing, bracket-aware scanning, Go
//	         capitalization rules, assignment-target extraction,
//	         function-boundary detection.
//	         Highest confidence. False positive rate < 1%.
//
//	Tier 2 — semantic: framework-specific constructor/factory patterns
//	         (LangChain, Zod, Pydantic, OpenAI tools, RAG frameworks).
//	         High confidence. Requires framework knowledge.
//
//	Tier 3 — pattern: regex on export names, path conventions, file
//	         extensions, naming conventions. Standard confidence.
//
//	Tier 4 — content: file content scanning for AI instruction markers,
//	         template file analysis, config key counting.
//	         Lower confidence. Requires corroboration (2+ signals).
type DetectionEvidence struct {
	// DetectorID identifies the specific detector that produced this result.
	// Format: "tier:detector-name" e.g., "structural:ast-message-array".
	DetectorID string `json:"detectorId"`

	// Tier is the detection method tier.
	Tier string `json:"tier"`

	// Confidence is the detection confidence (0.0–1.0).
	Confidence float64 `json:"confidence"`

	// FilePath is the file where the detection was made.
	FilePath string `json:"filePath"`

	// Symbol is the detected symbol name (variable, function, class) if known.
	Symbol string `json:"symbol,omitempty"`

	// Line is the source line number of the detection.
	Line int `json:"line,omitempty"`

	// Reason is a human-readable explanation of why this was detected.
	Reason string `json:"reason"`
}

// ---------------------------------------------------------------------------
// DetectorID constants — the canonical registry of all detection paths.
//
// Naming convention: "tier:detector-name"
//   - The prefix before ":" MUST be a valid tier (structural/semantic/pattern/content).
//   - The suffix after ":" is the detector-specific identifier.
//
// Every parser that sets a DetectionTier on a CodeSurface should use one of
// these constants (or a derivative) so that detection provenance is traceable
// end-to-end from surface → detector → tier.
// ---------------------------------------------------------------------------
const (
	// ── Structural detectors (Tier 1) ──────────────────────────────────
	// AST-verified or bracket-matched detections with <1% FP rate.

	// Go capitalization-rule exports.
	DetectorGoExport = "structural:go-export"

	// Bracket-aware structural parsers (structural_parser.go).
	DetectorBracketMessageArray = "structural:message-array"
	DetectorBracketFewShot      = "structural:few-shot-array"
	DetectorBracketPromptAssign = "structural:prompt-assignment"

	// AST-level prompt/context detectors (prompt_ast_parser.go).
	DetectorASTMessageArray  = "structural:ast-message-array"
	DetectorASTSystemPrompt  = "structural:ast-system-prompt"
	DetectorASTFewShot       = "structural:ast-few-shot-array"
	DetectorASTPromptBuilder = "structural:ast-prompt-builder"
	DetectorASTTemplateCall  = "structural:ast-template-prompt"

	// Fixture detectors that use language-level structure (fixture_parser.go).
	DetectorFixtureGoTestMain  = "structural:go-testmain"
	DetectorFixtureGoHelper    = "structural:go-test-helper"
	DetectorFixturePyFixture   = "structural:pytest-fixture"
	DetectorFixtureJavaLifecycle = "structural:java-lifecycle"

	// ── Semantic detectors (Tier 2) ────────────────────────────────────
	// Framework-specific constructor/factory patterns.

	// AI framework constructors.
	DetectorLangChainConstructor  = "semantic:langchain-constructor"
	DetectorLlamaIndexConstructor = "semantic:llamaindex-constructor"

	// Prompt builder functions (structural_parser.go prompt-builder detection).
	DetectorPromptBuilderFunc = "semantic:prompt-builder-func"
	DetectorTemplatePrompt    = "semantic:template-prompt"

	// Schema/tool definition detectors (schema_parser.go).
	DetectorZodSchema     = "semantic:zod-schema"
	DetectorPydanticModel = "semantic:pydantic-model"
	DetectorOpenAITools   = "semantic:openai-tools"
	DetectorToolDecorator = "semantic:tool-decorator"

	// RAG pipeline component detectors (rag_structured_parser.go).
	DetectorRAGConstructor   = "semantic:rag-constructor"
	DetectorRAGRetriever     = "semantic:rag-retriever"
	DetectorRAGEmbedding     = "semantic:rag-embedding"
	DetectorRAGChunking      = "semantic:rag-chunking"
	DetectorRAGVectorStore   = "semantic:rag-vector-store"
	DetectorRAGReranker      = "semantic:rag-reranker"
	DetectorRAGQueryBuilder  = "semantic:rag-query-builder"
	DetectorRAGDocLoader     = "semantic:rag-document-loader"
	DetectorRAGCitation      = "semantic:rag-citation"
	DetectorRAGContextWindow = "semantic:rag-context-assembly"

	// ── Pattern detectors (Tier 3) ─────────────────────────────────────
	// Regex/naming convention matching.

	DetectorExportName       = "pattern:export-name"
	DetectorRouteRegistration = "pattern:route-registration"
	DetectorHandlerName      = "pattern:handler-name"
	DetectorPathConvention   = "pattern:path-convention"
	DetectorImportMatch      = "pattern:import-match"

	// Fixture detectors using naming conventions (fixture_parser.go).
	DetectorFixtureLifecycleHook = "pattern:fixture-lifecycle-hook"
	DetectorFixtureBuilder       = "pattern:fixture-builder"
	DetectorFixtureMockProvider  = "pattern:fixture-mock-provider"
	DetectorFixtureDataLoader    = "pattern:fixture-data-loader"

	// ── Content detectors (Tier 4) ─────────────────────────────────────
	// Content scanning requiring corroboration.

	DetectorContentMarkers = "content:ai-instruction-markers"
	DetectorContentString  = "content:string-prompt"
	DetectorTemplateFile   = "content:template-file"
	DetectorRAGConfigFile  = "content:rag-config-file"
)

// Confidence basis constants.
const (
	// ConfidenceBasisCalibrated indicates the score was measured against
	// the ground truth fixture suite with observed precision data.
	ConfidenceBasisCalibrated = "calibrated"

	// ConfidenceBasisHeuristic indicates the score is an expert estimate
	// that has not been validated against ground truth.
	ConfidenceBasisHeuristic = "heuristic"
)

// FormatReason produces a standardized reason string with embedded detector ID.
// All detectors should use this format: "[detectorID] description".
func FormatReason(detectorID, description string) string {
	if detectorID == "" {
		return description
	}
	return "[" + detectorID + "] " + description
}

// AllTiers returns the valid tier values in priority order.
func AllTiers() []string {
	return []string{TierStructural, TierSemantic, TierPattern, TierContent}
}

// IsValidTier returns true if tier is one of the four canonical tiers.
func IsValidTier(tier string) bool {
	switch tier {
	case TierStructural, TierSemantic, TierPattern, TierContent:
		return true
	default:
		return false
	}
}

// DetectionTierOrder returns a numeric priority for sorting (lower = higher priority).
func DetectionTierOrder(tier string) int {
	switch tier {
	case TierStructural:
		return 1
	case TierSemantic:
		return 2
	case TierPattern:
		return 3
	case TierContent:
		return 4
	default:
		return 5
	}
}

// TierFromDetectorID extracts the tier from a DetectorID string.
// DetectorIDs follow the format "tier:detector-name".
func TierFromDetectorID(detectorID string) string {
	if idx := strings.Index(detectorID, ":"); idx > 0 {
		return detectorID[:idx]
	}
	return ""
}

// TierLabel returns a user-friendly label for a tier.
func TierLabel(tier string) string {
	switch tier {
	case TierStructural:
		return "Structural (AST/parser-verified)"
	case TierSemantic:
		return "Semantic (framework-aware)"
	case TierPattern:
		return "Pattern (naming convention)"
	case TierContent:
		return "Content (instruction markers)"
	default:
		return tier
	}
}

// TierConfidenceRange returns the expected confidence range for a tier.
// Returns (min, max). Used for validation and calibration.
func TierConfidenceRange(tier string) (float64, float64) {
	switch tier {
	case TierStructural:
		return 0.85, 1.0
	case TierSemantic:
		return 0.80, 0.96
	case TierPattern:
		return 0.70, 0.96
	case TierContent:
		return 0.60, 0.85
	default:
		return 0.0, 1.0
	}
}

// ValidateSurfaceTiers checks that every CodeSurface carries a valid
// DetectionTier and a non-zero Confidence. Returns a list of violations.
// This is intended for testing and debug assertions, not production gates.
func ValidateSurfaceTiers(surfaces []CodeSurface) []string {
	var violations []string
	for _, s := range surfaces {
		if s.DetectionTier == "" {
			violations = append(violations, s.SurfaceID+": missing DetectionTier")
			continue
		}
		if !IsValidTier(s.DetectionTier) {
			violations = append(violations, s.SurfaceID+": invalid tier "+s.DetectionTier)
		}
		if s.Confidence == 0 {
			violations = append(violations, s.SurfaceID+": zero confidence")
		}
	}
	return violations
}
