package models

import (
	"strings"
	"testing"
)

// --- Tier ordering and identity ---

func TestAllTiers_ReturnsCorrectOrder(t *testing.T) {
	t.Parallel()
	tiers := AllTiers()
	if len(tiers) != 4 {
		t.Fatalf("expected 4 tiers, got %d", len(tiers))
	}
	expected := []string{TierStructural, TierSemantic, TierPattern, TierContent}
	for i, tier := range tiers {
		if tier != expected[i] {
			t.Errorf("tier[%d]: want %s, got %s", i, expected[i], tier)
		}
	}
}

func TestIsValidTier(t *testing.T) {
	t.Parallel()
	for _, tier := range AllTiers() {
		if !IsValidTier(tier) {
			t.Errorf("IsValidTier(%q) = false, want true", tier)
		}
	}
	if IsValidTier("") {
		t.Error("empty string should not be valid tier")
	}
	if IsValidTier("ast") {
		t.Error("'ast' is not a valid tier")
	}
	if IsValidTier("rag") {
		t.Error("'rag' is not a valid tier")
	}
}

func TestDetectionTierOrder_Monotonic(t *testing.T) {
	t.Parallel()
	tiers := AllTiers()
	for i := 1; i < len(tiers); i++ {
		if DetectionTierOrder(tiers[i]) <= DetectionTierOrder(tiers[i-1]) {
			t.Errorf("tier order not monotonic: %s (%d) should be > %s (%d)",
				tiers[i], DetectionTierOrder(tiers[i]),
				tiers[i-1], DetectionTierOrder(tiers[i-1]))
		}
	}
}

func TestDetectionTierOrder_UnknownIsLowest(t *testing.T) {
	t.Parallel()
	if DetectionTierOrder("bogus") <= DetectionTierOrder(TierContent) {
		t.Error("unknown tier should have lower priority than any valid tier")
	}
}

// --- TierFromDetectorID ---

func TestTierFromDetectorID(t *testing.T) {
	t.Parallel()
	cases := []struct {
		id   string
		tier string
	}{
		{DetectorGoExport, TierStructural},
		{DetectorASTMessageArray, TierStructural},
		{DetectorLangChainConstructor, TierSemantic},
		{DetectorRAGRetriever, TierSemantic},
		{DetectorExportName, TierPattern},
		{DetectorContentMarkers, TierContent},
		{"", ""},
		{"notier", ""},
	}
	for _, tc := range cases {
		got := TierFromDetectorID(tc.id)
		if got != tc.tier {
			t.Errorf("TierFromDetectorID(%q) = %q, want %q", tc.id, got, tc.tier)
		}
	}
}

// --- TierLabel ---

func TestTierLabel_AllTiersHaveLabels(t *testing.T) {
	t.Parallel()
	for _, tier := range AllTiers() {
		label := TierLabel(tier)
		if label == tier {
			t.Errorf("TierLabel(%q) returned raw tier value — should be human-friendly", tier)
		}
		if label == "" {
			t.Errorf("TierLabel(%q) returned empty string", tier)
		}
	}
}

// --- TierConfidenceRange ---

func TestTierConfidenceRange_HigherTiersHaveHigherMinimum(t *testing.T) {
	t.Parallel()
	structMin, _ := TierConfidenceRange(TierStructural)
	contentMin, _ := TierConfidenceRange(TierContent)
	if structMin <= contentMin {
		t.Errorf("structural min (%.2f) should be > content min (%.2f)", structMin, contentMin)
	}
}

// --- DetectorID registry completeness ---

// allDetectorIDs returns every registered DetectorID constant.
func allDetectorIDs() []string {
	return []string{
		// Structural
		DetectorGoExport, DetectorBracketMessageArray, DetectorBracketFewShot,
		DetectorBracketPromptAssign, DetectorASTMessageArray, DetectorASTSystemPrompt,
		DetectorASTFewShot, DetectorASTPromptBuilder, DetectorASTTemplateCall,
		DetectorFixtureGoTestMain, DetectorFixtureGoHelper, DetectorFixturePyFixture,
		DetectorFixtureJavaLifecycle,
		// Semantic
		DetectorLangChainConstructor, DetectorLlamaIndexConstructor,
		DetectorPromptBuilderFunc, DetectorTemplatePrompt,
		DetectorZodSchema, DetectorPydanticModel, DetectorOpenAITools, DetectorToolDecorator,
		DetectorRAGConstructor, DetectorRAGRetriever, DetectorRAGEmbedding,
		DetectorRAGChunking, DetectorRAGVectorStore, DetectorRAGReranker,
		DetectorRAGQueryBuilder, DetectorRAGDocLoader, DetectorRAGCitation,
		DetectorRAGContextWindow,
		// Pattern
		DetectorExportName, DetectorRouteRegistration, DetectorHandlerName,
		DetectorPathConvention, DetectorImportMatch,
		DetectorFixtureLifecycleHook, DetectorFixtureBuilder,
		DetectorFixtureMockProvider, DetectorFixtureDataLoader,
		// Content
		DetectorContentMarkers, DetectorContentString, DetectorTemplateFile,
		DetectorRAGConfigFile,
	}
}

func TestAllDetectorIDs_HaveValidTierPrefix(t *testing.T) {
	t.Parallel()
	for _, id := range allDetectorIDs() {
		tier := TierFromDetectorID(id)
		if tier == "" {
			t.Errorf("DetectorID %q has no tier prefix", id)
			continue
		}
		if !IsValidTier(tier) {
			t.Errorf("DetectorID %q has invalid tier prefix %q", id, tier)
		}
	}
}

func TestAllDetectorIDs_NoDuplicates(t *testing.T) {
	t.Parallel()
	seen := map[string]bool{}
	for _, id := range allDetectorIDs() {
		if seen[id] {
			t.Errorf("duplicate DetectorID: %s", id)
		}
		seen[id] = true
	}
}

func TestAllDetectorIDs_UseColonSeparator(t *testing.T) {
	t.Parallel()
	for _, id := range allDetectorIDs() {
		if !strings.Contains(id, ":") {
			t.Errorf("DetectorID %q missing colon separator", id)
		}
		parts := strings.SplitN(id, ":", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			t.Errorf("DetectorID %q has malformed tier:name format", id)
		}
	}
}

func TestAllDetectorIDs_NonEmpty(t *testing.T) {
	t.Parallel()
	for _, id := range allDetectorIDs() {
		if id == "" {
			t.Error("detector ID constant should not be empty")
		}
	}
}

// --- ValidateSurfaceTiers ---

func TestValidateSurfaceTiers_DetectsMissingTier(t *testing.T) {
	t.Parallel()
	surfaces := []CodeSurface{
		{SurfaceID: "s:a", DetectionTier: TierStructural, Confidence: 0.95},
		{SurfaceID: "s:b", DetectionTier: "", Confidence: 0.80},
		{SurfaceID: "s:c", DetectionTier: TierPattern, Confidence: 0},
	}

	violations := ValidateSurfaceTiers(surfaces)
	if len(violations) != 2 {
		t.Errorf("expected 2 violations, got %d: %v", len(violations), violations)
	}
}

func TestValidateSurfaceTiers_DetectsInvalidTier(t *testing.T) {
	t.Parallel()
	surfaces := []CodeSurface{
		{SurfaceID: "s:x", DetectionTier: "ast", Confidence: 0.90},
		{SurfaceID: "s:y", DetectionTier: "rag", Confidence: 0.90},
	}

	violations := ValidateSurfaceTiers(surfaces)
	if len(violations) != 2 {
		t.Errorf("expected 2 violations for invalid tiers, got %d: %v", len(violations), violations)
	}
}

func TestValidateSurfaceTiers_PassesForValidSurfaces(t *testing.T) {
	t.Parallel()
	surfaces := []CodeSurface{
		{SurfaceID: "s:1", DetectionTier: TierStructural, Confidence: 0.99},
		{SurfaceID: "s:2", DetectionTier: TierSemantic, Confidence: 0.90},
		{SurfaceID: "s:3", DetectionTier: TierPattern, Confidence: 0.85},
		{SurfaceID: "s:4", DetectionTier: TierContent, Confidence: 0.75},
	}

	violations := ValidateSurfaceTiers(surfaces)
	if len(violations) != 0 {
		t.Errorf("expected 0 violations, got %d: %v", len(violations), violations)
	}
}
