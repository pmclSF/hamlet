package analysis

import (
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/pmclSF/terrain/internal/models"
)

// InferCapabilities populates the Capability field on scenarios by
// analyzing scenario names, file paths, folder structure, and covered
// surface context.
//
// Inference strategy (in priority order):
//  1. Explicit: scenario already has Capability set (e.g., from YAML)
//  2. Folder path: scenarios in "evals/refund/" → "refund"
//  3. Scenario name: "billing-lookup-accuracy" → "billing-lookup"
//  4. Covered surface context: if all covered surfaces are in "src/billing/" → "billing"
//
// The inferred capability is a normalized, hyphen-separated label.
func InferCapabilities(scenarios []models.Scenario, surfaces []models.CodeSurface) {
	surfaceByID := map[string]models.CodeSurface{}
	for _, s := range surfaces {
		surfaceByID[s.SurfaceID] = s
	}

	for i := range scenarios {
		if scenarios[i].Capability != "" {
			continue // already set (e.g., from YAML)
		}

		// Strategy 1: infer from folder path.
		if cap := capabilityFromPath(scenarios[i].Path); cap != "" {
			scenarios[i].Capability = cap
			continue
		}

		// Strategy 2: infer from scenario name.
		if cap := capabilityFromName(scenarios[i].Name); cap != "" {
			scenarios[i].Capability = cap
			continue
		}

		// Strategy 3: infer from covered surface paths.
		if cap := capabilityFromSurfaces(scenarios[i].CoveredSurfaceIDs, surfaceByID); cap != "" {
			scenarios[i].Capability = cap
		}
	}
}

// capabilityFromPath extracts a capability label from a scenario file path.
// Looks for domain directories under eval/evals/tests paths.
// e.g., "evals/refund/accuracy.test.ts" → "refund"
//       "tests/eval/billing/lookup.test.ts" → "billing"
func capabilityFromPath(path string) string {
	if path == "" {
		return ""
	}
	parts := strings.Split(filepath.ToSlash(path), "/")

	// Look for a domain folder after eval/evals/tests/eval.
	for i, p := range parts {
		lower := strings.ToLower(p)
		if lower == "eval" || lower == "evals" || lower == "evaluations" {
			if i+1 < len(parts) {
				next := parts[i+1]
				// Skip generic names.
				if isGenericFolderName(next) {
					continue
				}
				return normalizeCapability(next)
			}
		}
	}

	return ""
}

// capabilityFromName extracts a capability from a scenario name by stripping
// common suffixes like "-accuracy", "-safety", "-regression", "-eval".
func capabilityFromName(name string) string {
	if name == "" {
		return ""
	}
	// Strip known eval-category suffixes.
	cleaned := name
	for _, suffix := range []string{
		"-accuracy", "-safety", "-regression", "-eval", "-test",
		"-quality", "-completeness", "-correctness", "-latency",
		"_accuracy", "_safety", "_regression", "_eval", "_test",
		"_quality", "_completeness", "_correctness", "_latency",
	} {
		cleaned = strings.TrimSuffix(cleaned, suffix)
	}

	// If stripping produced a meaningful shorter name, use it.
	if cleaned != name && len(cleaned) >= 3 {
		return normalizeCapability(cleaned)
	}

	// If the name itself looks like a capability (not a generic eval term).
	if !isGenericEvalName(name) && len(name) >= 3 {
		return normalizeCapability(name)
	}

	return ""
}

// capabilityFromSurfaces infers capability from the common path prefix of
// covered surfaces. If all surfaces share a domain folder, use that.
func capabilityFromSurfaces(surfaceIDs []string, surfaces map[string]models.CodeSurface) string {
	if len(surfaceIDs) == 0 {
		return ""
	}

	// Collect unique parent directories of covered surfaces.
	dirs := map[string]int{}
	for _, sid := range surfaceIDs {
		s, ok := surfaces[sid]
		if !ok || s.Path == "" {
			continue
		}
		parts := strings.Split(filepath.ToSlash(s.Path), "/")
		// Use the last meaningful directory before the filename.
		for i := len(parts) - 2; i >= 0; i-- {
			if !isGenericFolderName(parts[i]) {
				dirs[parts[i]]++
				break
			}
		}
	}

	if len(dirs) == 0 {
		return ""
	}

	// If all surfaces share the same domain directory, use it.
	if len(dirs) == 1 {
		for d := range dirs {
			return normalizeCapability(d)
		}
	}

	// If one directory dominates (>50% of surfaces), use it.
	total := 0
	best := ""
	bestCount := 0
	for d, c := range dirs {
		total += c
		if c > bestCount {
			bestCount = c
			best = d
		}
	}
	if bestCount*2 > total {
		return normalizeCapability(best)
	}

	return ""
}

// InferAICapabilities analyzes all code surfaces and scenarios to determine
// which canonical AI capabilities are present in the codebase. This operates
// above individual framework detection — a retrieval_augmented_generation
// capability is inferred whether the code uses LangChain, LlamaIndex, DSPy,
// or raw vector store calls.
//
// Returns capabilities with their supporting evidence (surfaces + scenarios).
func InferAICapabilities(surfaces []models.CodeSurface, scenarios []models.Scenario) []models.InferredCapability {
	// Map surface kinds to capabilities.
	capSurfaces := map[models.AICapability][]string{}
	capConfidence := map[models.AICapability]float64{}

	for _, s := range surfaces {
		caps := surfaceKindToCapabilities(s.Kind, s.Name, s.Reason)
		for _, cap := range caps {
			capSurfaces[cap] = append(capSurfaces[cap], s.SurfaceID)
			// Track highest confidence surface per capability.
			if s.Confidence > capConfidence[cap] {
				capConfidence[cap] = s.Confidence
			}
		}
	}

	// Map scenarios to capabilities.
	capScenarios := map[models.AICapability][]string{}
	for _, sc := range scenarios {
		if sc.Capability == "" {
			continue
		}
		// Try to match scenario's free-text capability to a canonical one.
		if canonical := matchCanonicalCapability(sc.Capability); canonical != "" {
			capScenarios[models.AICapability(canonical)] = append(
				capScenarios[models.AICapability(canonical)], sc.ScenarioID)
		}
	}

	// Build result list.
	var result []models.InferredCapability
	for _, cap := range models.AllAICapabilities() {
		sids := capSurfaces[cap]
		if len(sids) == 0 {
			continue
		}
		// Deduplicate surface IDs.
		sids = uniqueSorted(sids)

		ic := models.InferredCapability{
			Capability:  cap,
			Label:       models.CapabilityLabel(cap),
			SurfaceIDs:  sids,
			ScenarioIDs: capScenarios[cap],
			Confidence:  capConfidence[cap],
			Covered:     len(capScenarios[cap]) > 0,
		}
		result = append(result, ic)
	}

	return result
}

// surfaceKindToCapabilities maps a CodeSurface to zero or more canonical
// AI capabilities based on its Kind, name patterns, and detection evidence.
func surfaceKindToCapabilities(kind models.CodeSurfaceKind, name, reason string) []models.AICapability {
	var caps []models.AICapability
	lower := strings.ToLower(name)
	lowerReason := strings.ToLower(reason)

	switch kind {
	case models.SurfaceRetrieval:
		caps = append(caps, models.CapabilityRAG)
		// Citation surfaces are a sub-capability of RAG.
		if strings.Contains(lower, "citation") || strings.Contains(lower, "source") ||
			strings.Contains(lowerReason, "citation") {
			caps = append(caps, models.CapabilityCitation)
		}

	case models.SurfaceToolDef:
		caps = append(caps, models.CapabilityToolUse)
		// Structured output is also inferred from schema/tool patterns.
		if strings.Contains(lower, "schema") || strings.Contains(lower, "structured") ||
			strings.Contains(lower, "output") || strings.Contains(lowerReason, "structured output") {
			caps = append(caps, models.CapabilityStructuredOutput)
		}

	case models.SurfacePrompt:
		caps = append(caps, models.CapabilityPromptGeneration)

	case models.SurfaceContext:
		caps = append(caps, models.CapabilityPromptGeneration)
		// Safety overlays are a sub-capability of context.
		if strings.Contains(lower, "safety") || strings.Contains(lower, "guardrail") ||
			strings.Contains(lower, "filter") || strings.Contains(lower, "moderat") {
			caps = append(caps, models.CapabilitySafety)
		}
		// Memory patterns.
		if strings.Contains(lower, "memory") || strings.Contains(lower, "history") ||
			strings.Contains(lower, "session") || strings.Contains(lower, "conversation") {
			caps = append(caps, models.CapabilityMemory)
		}

	case models.SurfaceAgent:
		caps = append(caps, models.CapabilityAgentOrchestration)
		// Agents often also imply tool use.
		if strings.Contains(lower, "tool") || strings.Contains(lower, "router") {
			caps = append(caps, models.CapabilityToolUse)
		}

	case models.SurfaceEvalDef:
		caps = append(caps, models.CapabilityEvaluation)

	case models.SurfaceDataset:
		// Datasets support evaluation or RAG depending on context.
		if strings.Contains(lower, "eval") || strings.Contains(lower, "benchmark") ||
			strings.Contains(lower, "grading") {
			caps = append(caps, models.CapabilityEvaluation)
		}
	}

	return caps
}

// matchCanonicalCapability attempts to match a free-text scenario capability
// to one of the canonical AICapability values.
func matchCanonicalCapability(capability string) string {
	lower := strings.ToLower(capability)

	// Check specific capabilities before general ones to avoid
	// "prompt-safety" matching "prompt" instead of "safety".
	switch {
	case strings.Contains(lower, "safety") || strings.Contains(lower, "guardrail") ||
		strings.Contains(lower, "moderat"):
		return string(models.CapabilitySafety)
	case strings.Contains(lower, "retriev") || strings.Contains(lower, "rag") ||
		strings.Contains(lower, "search") || strings.Contains(lower, "vector"):
		return string(models.CapabilityRAG)
	case strings.Contains(lower, "citation") || strings.Contains(lower, "source-attrib"):
		return string(models.CapabilityCitation)
	case strings.Contains(lower, "memory") || strings.Contains(lower, "conversation"):
		return string(models.CapabilityMemory)
	case strings.Contains(lower, "agent") || strings.Contains(lower, "orchestrat") ||
		strings.Contains(lower, "workflow"):
		return string(models.CapabilityAgentOrchestration)
	case strings.Contains(lower, "tool") || strings.Contains(lower, "function-call"):
		return string(models.CapabilityToolUse)
	case strings.Contains(lower, "structured") || strings.Contains(lower, "schema") ||
		strings.Contains(lower, "extract"):
		return string(models.CapabilityStructuredOutput)
	case strings.Contains(lower, "eval") || strings.Contains(lower, "accuracy") ||
		strings.Contains(lower, "quality"):
		return string(models.CapabilityEvaluation)
	case strings.Contains(lower, "prompt") || strings.Contains(lower, "template"):
		return string(models.CapabilityPromptGeneration)
	default:
		return ""
	}
}

func uniqueSorted(ss []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range ss {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}

// CollectImpactedCapabilities returns unique capability names from impacted scenarios.
func CollectImpactedCapabilities(scenarios []models.Scenario, impactedIDs []string) []string {
	idSet := map[string]bool{}
	for _, id := range impactedIDs {
		idSet[id] = true
	}

	capSet := map[string]bool{}
	for _, sc := range scenarios {
		if idSet[sc.ScenarioID] && sc.Capability != "" {
			capSet[sc.Capability] = true
		}
	}

	caps := make([]string, 0, len(capSet))
	for c := range capSet {
		caps = append(caps, c)
	}
	sort.Strings(caps)
	return caps
}

// --- Helpers ---

var camelSplitPattern = regexp.MustCompile(`([a-z])([A-Z])`)

func normalizeCapability(s string) string {
	// Convert camelCase to kebab-case.
	s = camelSplitPattern.ReplaceAllString(s, "${1}-${2}")
	s = strings.ToLower(s)
	// Replace underscores and spaces with hyphens.
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.ReplaceAll(s, " ", "-")
	// Collapse multiple hyphens.
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	return strings.Trim(s, "-")
}

func isGenericFolderName(name string) bool {
	lower := strings.ToLower(name)
	switch lower {
	case "src", "lib", "test", "tests", "spec", "specs", "eval", "evals",
		"unit", "integration", "e2e", "fixtures", "utils", "helpers",
		"common", "shared", "internal", "core", "ai", "ml":
		return true
	}
	return false
}

func isGenericEvalName(name string) bool {
	lower := strings.ToLower(name)
	switch lower {
	case "accuracy", "safety", "regression", "eval", "test", "quality",
		"latency", "correctness", "completeness":
		return true
	}
	return false
}
