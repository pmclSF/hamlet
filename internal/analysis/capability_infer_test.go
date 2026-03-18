package analysis

import (
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

func TestCapabilityFromPath_EvalSubfolder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		path string
		want string
	}{
		{"evals/refund/accuracy.test.ts", "refund"},
		{"evals/billing/lookup.test.ts", "billing"},
		{"tests/eval/safety/prompt-injection.test.ts", "safety"},
		{"evaluations/enterprise-search/retrieval.py", "enterprise-search"},
		{"evals/unit/test_basic.py", ""},  // "unit" is generic
		{"src/services/auth.ts", ""},      // no eval path
		{"", ""},
	}
	for _, tt := range tests {
		got := capabilityFromPath(tt.path)
		if got != tt.want {
			t.Errorf("capabilityFromPath(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestCapabilityFromName_StripsSuffixes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		want string
	}{
		{"refund-explanation-accuracy", "refund-explanation"},
		{"billing-lookup-safety", "billing-lookup"},
		{"document-qa-regression", "document-qa"},
		{"enterprise-search-eval", "enterprise-search"},
		{"support-triage-quality", "support-triage"},
		{"meeting-summarization-test", "meeting-summarization"},
		{"accuracy", ""},          // generic, stripped to empty
		{"safety", ""},            // generic
		{"ab", ""},                // too short
		{"", ""},
	}
	for _, tt := range tests {
		got := capabilityFromName(tt.name)
		if got != tt.want {
			t.Errorf("capabilityFromName(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestCapabilityFromSurfaces_SharedDomain(t *testing.T) {
	t.Parallel()
	surfaces := map[string]models.CodeSurface{
		"s1": {Path: "src/billing/charge.ts"},
		"s2": {Path: "src/billing/refund.ts"},
		"s3": {Path: "src/billing/invoice.ts"},
	}
	got := capabilityFromSurfaces([]string{"s1", "s2", "s3"}, surfaces)
	if got != "billing" {
		t.Errorf("expected 'billing', got %q", got)
	}
}

func TestCapabilityFromSurfaces_MixedDomains(t *testing.T) {
	t.Parallel()
	surfaces := map[string]models.CodeSurface{
		"s1": {Path: "src/billing/charge.ts"},
		"s2": {Path: "src/auth/login.ts"},
	}
	got := capabilityFromSurfaces([]string{"s1", "s2"}, surfaces)
	// Neither dominates — no capability inferred.
	if got != "" {
		t.Errorf("expected empty for mixed domains, got %q", got)
	}
}

func TestCapabilityFromSurfaces_DominantDomain(t *testing.T) {
	t.Parallel()
	surfaces := map[string]models.CodeSurface{
		"s1": {Path: "src/search/retriever.ts"},
		"s2": {Path: "src/search/reranker.ts"},
		"s3": {Path: "src/utils/logger.ts"},
	}
	got := capabilityFromSurfaces([]string{"s1", "s2", "s3"}, surfaces)
	if got != "search" {
		t.Errorf("expected 'search' (dominant), got %q", got)
	}
}

func TestInferCapabilities_Integration(t *testing.T) {
	t.Parallel()
	scenarios := []models.Scenario{
		{
			ScenarioID: "s1",
			Name:       "refund-explanation-accuracy",
			Path:       "evals/refund/accuracy.test.ts",
		},
		{
			ScenarioID: "s2",
			Name:       "enterprise-search-eval",
			Path:       "evals/search/retrieval.test.ts",
		},
		{
			ScenarioID:        "s3",
			Name:              "safety-check",
			CoveredSurfaceIDs: []string{"s1", "s2"},
		},
		{
			ScenarioID: "s4",
			Name:       "pre-set-capability",
			Capability: "billing", // already set — should not be overridden
		},
	}
	surfaces := []models.CodeSurface{
		{SurfaceID: "s1", Path: "src/ai/prompts.ts"},
		{SurfaceID: "s2", Path: "src/ai/safety.ts"},
	}

	InferCapabilities(scenarios, surfaces)

	// s1: path-based → "refund"
	if scenarios[0].Capability != "refund" {
		t.Errorf("s1 capability = %q, want 'refund'", scenarios[0].Capability)
	}
	// s2: path-based → "search"
	if scenarios[1].Capability != "search" {
		t.Errorf("s2 capability = %q, want 'search'", scenarios[1].Capability)
	}
	// s3: name-based → "safety-check" (no suffix stripped, but it's a valid name)
	if scenarios[2].Capability != "safety-check" {
		t.Errorf("s3 capability = %q, want 'safety-check'", scenarios[2].Capability)
	}
	// s4: pre-set — should not be overridden
	if scenarios[3].Capability != "billing" {
		t.Errorf("s4 capability = %q, want 'billing' (pre-set)", scenarios[3].Capability)
	}
}

func TestCollectImpactedCapabilities(t *testing.T) {
	t.Parallel()
	scenarios := []models.Scenario{
		{ScenarioID: "s1", Capability: "refund-explanation"},
		{ScenarioID: "s2", Capability: "enterprise-search"},
		{ScenarioID: "s3", Capability: "refund-explanation"}, // duplicate
		{ScenarioID: "s4", Capability: ""},                    // no capability
	}
	caps := CollectImpactedCapabilities(scenarios, []string{"s1", "s2", "s3", "s4"})
	if len(caps) != 2 {
		t.Fatalf("expected 2 unique capabilities, got %d: %v", len(caps), caps)
	}
	if caps[0] != "enterprise-search" || caps[1] != "refund-explanation" {
		t.Errorf("expected [enterprise-search, refund-explanation], got %v", caps)
	}
}

func TestNormalizeCapability(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  string
	}{
		{"refundExplanation", "refund-explanation"},
		{"BillingLookup", "billing-lookup"},
		{"enterprise_search", "enterprise-search"},
		{"REFUND", "refund"},
		{"meeting-summarization", "meeting-summarization"},
	}
	for _, tt := range tests {
		got := normalizeCapability(tt.input)
		if got != tt.want {
			t.Errorf("normalizeCapability(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
