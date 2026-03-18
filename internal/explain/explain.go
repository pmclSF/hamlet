// Package explain produces structured explanations for Terrain decisions.
//
// It operates on the output of impact analysis and the impact graph,
// building reason chains that show *why* a test was selected, *how*
// confidence was derived, and *what* fallback strategies were used.
//
// Core entry points:
//
//   - ExplainTest: explain why a test was selected for a change
//   - ExplainSelection: explain the overall test selection strategy
package explain

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pmclSF/terrain/internal/impact"
	"github.com/pmclSF/terrain/internal/models"
)

// ReasonChain is a sequence of steps from a changed code unit to a test.
type ReasonChain struct {
	// Steps is the ordered sequence of hops in the chain.
	Steps []ChainStep `json:"steps"`

	// Confidence is the composite confidence along this chain.
	Confidence float64 `json:"confidence"`

	// Band is the confidence band (high, medium, low).
	Band string `json:"band"`
}

// ChainStep is one hop in a reason chain.
type ChainStep struct {
	// From is the source node (code unit, file, etc.).
	From string `json:"from"`

	// To is the target node.
	To string `json:"to"`

	// Relationship describes the edge type in human terms.
	Relationship string `json:"relationship"`

	// EdgeKind is the raw edge kind from the impact graph.
	EdgeKind string `json:"edgeKind"`

	// EdgeConfidence is the confidence of this specific edge.
	EdgeConfidence float64 `json:"edgeConfidence"`
}

// TestExplanation is the structured explanation for why a test is relevant
// to a change.
type TestExplanation struct {
	// Target is the test being explained.
	Target TestTarget `json:"target"`

	// Verdict is a one-line summary of the explanation.
	Verdict string `json:"verdict"`

	// StrongestPath is the highest-confidence reason chain.
	StrongestPath *ReasonChain `json:"strongestPath"`

	// AlternativePaths are additional supporting reason chains.
	AlternativePaths []ReasonChain `json:"alternativePaths,omitempty"`

	// Confidence is the overall confidence for this test's relevance.
	Confidence float64 `json:"confidence"`

	// ConfidenceBand is the discrete band (high, medium, low).
	ConfidenceBand string `json:"confidenceBand"`

	// ReasonCategory is why this test was selected (directDependency, etc.).
	ReasonCategory string `json:"reasonCategory"`

	// CoversUnits lists code unit IDs this test covers.
	CoversUnits []string `json:"coversUnits,omitempty"`

	// FallbackUsed describes fallback strategy, if any.
	FallbackUsed *FallbackDetail `json:"fallbackUsed,omitempty"`

	// Limitations notes caveats affecting this explanation.
	Limitations []string `json:"limitations,omitempty"`
}

// TestTarget identifies the test being explained.
type TestTarget struct {
	// Path is the test file path.
	Path string `json:"path"`

	// Framework is the test framework.
	Framework string `json:"framework,omitempty"`

	// TestID is the test identifier if known.
	TestID string `json:"testId,omitempty"`

	// Owner is the test owner if known.
	Owner string `json:"owner,omitempty"`
}

// FallbackDetail describes how fallback affected this test's selection.
type FallbackDetail struct {
	// Level is the fallback strategy (package, directory, all).
	Level string `json:"level"`

	// Reason explains why fallback was triggered.
	Reason string `json:"reason"`
}

// SelectionExplanation explains the overall test selection strategy
// for a change.
type SelectionExplanation struct {
	// Summary is a one-line overview.
	Summary string `json:"summary"`

	// Strategy is the selection strategy used (exact, near_minimal, fallback_broad).
	Strategy string `json:"strategy"`

	// TotalSelected is the number of tests selected.
	TotalSelected int `json:"totalSelected"`

	// CoverageConfidence is the overall coverage confidence.
	CoverageConfidence string `json:"coverageConfidence"`

	// ReasonBreakdown counts tests by reason category.
	ReasonBreakdown map[string]int `json:"reasonBreakdown"`

	// HighConfidenceTests lists tests with high-confidence paths.
	HighConfidenceTests []TestExplanation `json:"highConfidenceTests,omitempty"`

	// MediumConfidenceTests lists tests with medium-confidence paths.
	MediumConfidenceTests []TestExplanation `json:"mediumConfidenceTests,omitempty"`

	// LowConfidenceTests lists tests with low-confidence paths.
	LowConfidenceTests []TestExplanation `json:"lowConfidenceTests,omitempty"`

	// FallbackUsed describes fallback strategy if any.
	FallbackUsed *FallbackDetail `json:"fallbackUsed,omitempty"`

	// GapCount is the number of protection gaps.
	GapCount int `json:"gapCount"`

	// Limitations notes caveats affecting the selection.
	Limitations []string `json:"limitations,omitempty"`
}

// ExplainTest produces a structured explanation for why a specific test
// is relevant to the change described by the ImpactResult.
//
// The target can be a test file path, test ID, or partial match.
// Returns an error if the test is not found in the impact result.
func ExplainTest(target string, result *impact.ImpactResult) (*TestExplanation, error) {
	if result == nil {
		return nil, fmt.Errorf("no impact result available")
	}

	// Find the test in impacted tests.
	test, found := findTest(target, result)
	if !found {
		return nil, fmt.Errorf("test not found in impact analysis: %s", target)
	}

	explanation := &TestExplanation{
		Target: TestTarget{
			Path:      test.Path,
			Framework: test.Framework,
			TestID:    test.TestID,
		},
		Confidence:     confidenceScore(test.ImpactConfidence),
		ConfidenceBand: string(test.ImpactConfidence),
		CoversUnits:    test.CoversUnits,
	}

	// Build reason chains from the impact graph.
	chains := buildReasonChains(test, result)

	if len(chains) > 0 {
		// Sort by confidence descending.
		sort.Slice(chains, func(i, j int) bool {
			return chains[i].Confidence > chains[j].Confidence
		})
		explanation.StrongestPath = &chains[0]
		if len(chains) > 1 {
			explanation.AlternativePaths = chains[1:]
		}
		explanation.Confidence = chains[0].Confidence
		explanation.ConfidenceBand = chains[0].Band
	}

	// Classify reason category.
	explanation.ReasonCategory = classifyReason(test)

	// Check for fallback usage.
	if result.Fallback.Level != "none" && result.Fallback.Level != "" {
		explanation.FallbackUsed = &FallbackDetail{
			Level:  result.Fallback.Level,
			Reason: result.Fallback.Reason,
		}
	}

	// Build verdict.
	explanation.Verdict = buildVerdict(explanation)

	// Note limitations.
	explanation.Limitations = buildTestLimitations(test, result)

	return explanation, nil
}

// ExplainSelection produces a structured explanation for the overall
// test selection strategy used in the ImpactResult.
func ExplainSelection(result *impact.ImpactResult) (*SelectionExplanation, error) {
	if result == nil {
		return nil, fmt.Errorf("no impact result available")
	}

	sel := &SelectionExplanation{
		TotalSelected:      len(result.SelectedTests),
		CoverageConfidence: result.CoverageConfidence,
		GapCount:           len(result.ProtectionGaps),
		Limitations:        result.Limitations,
	}

	// Strategy from protective set.
	if result.ProtectiveSet != nil {
		sel.Strategy = result.ProtectiveSet.SetKind
	} else if len(result.SelectedTests) > 0 {
		sel.Strategy = "exact"
	} else {
		sel.Strategy = "none"
	}

	// Reason breakdown.
	sel.ReasonBreakdown = map[string]int{
		"directDependency":  result.ReasonCategories.DirectDependency,
		"fixtureDependency": result.ReasonCategories.FixtureDependency,
		"directlyChanged":   result.ReasonCategories.DirectlyChanged,
		"directoryProximity": result.ReasonCategories.DirectoryProximity,
	}

	// Build per-test explanations and bucket by confidence.
	for _, test := range result.SelectedTests {
		te, err := ExplainTest(test.Path, result)
		if err != nil {
			continue
		}
		switch te.ConfidenceBand {
		case "exact", "high":
			sel.HighConfidenceTests = append(sel.HighConfidenceTests, *te)
		case "inferred", "medium":
			sel.MediumConfidenceTests = append(sel.MediumConfidenceTests, *te)
		default:
			sel.LowConfidenceTests = append(sel.LowConfidenceTests, *te)
		}
	}

	// Fallback info.
	if result.Fallback.Level != "none" && result.Fallback.Level != "" {
		sel.FallbackUsed = &FallbackDetail{
			Level:  result.Fallback.Level,
			Reason: result.Fallback.Reason,
		}
	}

	// Summary.
	sel.Summary = buildSelectionSummary(sel, result)

	return sel, nil
}

// findTest locates a test in the impact result by path, ID, or partial match.
func findTest(target string, result *impact.ImpactResult) (*impact.ImpactedTest, bool) {
	target = strings.TrimSpace(target)

	// Check selected tests first (they have richer data).
	for i := range result.SelectedTests {
		t := &result.SelectedTests[i]
		if t.Path == target || t.TestID == target {
			return t, true
		}
	}

	// Check all impacted tests.
	for i := range result.ImpactedTests {
		t := &result.ImpactedTests[i]
		if t.Path == target || t.TestID == target {
			return t, true
		}
	}

	// Partial match: suffix match on path.
	for i := range result.ImpactedTests {
		t := &result.ImpactedTests[i]
		if strings.HasSuffix(t.Path, target) {
			return t, true
		}
	}

	return nil, false
}

// buildReasonChains constructs reason chains from the impact graph.
func buildReasonChains(test *impact.ImpactedTest, result *impact.ImpactResult) []ReasonChain {
	if result.Graph == nil {
		return buildSyntheticChains(test, result)
	}

	// Get edges connecting this test to code units.
	unitIDs := result.Graph.UnitsForTest(test.Path)
	if len(unitIDs) == 0 {
		return buildSyntheticChains(test, result)
	}

	var chains []ReasonChain
	seen := map[string]bool{}

	for _, unitID := range unitIDs {
		// Check if this unit is actually impacted.
		if !isImpactedUnit(unitID, result) {
			continue
		}

		edge := result.Graph.EdgeBetween(unitID, test.Path)
		if edge == nil {
			continue
		}

		// Deduplicate by unit ID.
		if seen[unitID] {
			continue
		}
		seen[unitID] = true

		conf := confidenceScore(edge.Confidence)
		chain := ReasonChain{
			Steps: []ChainStep{
				{
					From:           unitID,
					To:             test.Path,
					Relationship:   edgeKindLabel(edge.Kind),
					EdgeKind:       string(edge.Kind),
					EdgeConfidence: conf,
				},
			},
			Confidence: conf,
			Band:       classifyConfidenceBand(conf),
		}
		chains = append(chains, chain)
	}

	if len(chains) == 0 {
		return buildSyntheticChains(test, result)
	}

	return chains
}

// buildSyntheticChains creates chains from test metadata when no graph is available.
func buildSyntheticChains(test *impact.ImpactedTest, result *impact.ImpactResult) []ReasonChain {
	var chains []ReasonChain

	if test.IsDirectlyChanged {
		chains = append(chains, ReasonChain{
			Steps: []ChainStep{
				{
					From:           test.Path,
					To:             test.Path,
					Relationship:   "test file directly changed",
					EdgeKind:       "directly_changed",
					EdgeConfidence: 1.0,
				},
			},
			Confidence: 1.0,
			Band:       "high",
		})
		return chains
	}

	if len(test.CoversUnits) > 0 {
		for _, unitID := range test.CoversUnits {
			conf := confidenceScore(test.ImpactConfidence)
			chains = append(chains, ReasonChain{
				Steps: []ChainStep{
					{
						From:           unitID,
						To:             test.Path,
						Relationship:   "covers code unit",
						EdgeKind:       "structural_link",
						EdgeConfidence: conf,
					},
				},
				Confidence: conf,
				Band:       classifyConfidenceBand(conf),
			})
		}
		return chains
	}

	// Minimal synthetic chain from relevance text.
	conf := confidenceScore(test.ImpactConfidence)
	chains = append(chains, ReasonChain{
		Steps: []ChainStep{
			{
				From:           "(changed code)",
				To:             test.Path,
				Relationship:   test.Relevance,
				EdgeKind:       "inferred",
				EdgeConfidence: conf,
			},
		},
		Confidence: conf,
		Band:       classifyConfidenceBand(conf),
	})

	return chains
}

// isImpactedUnit checks if a unit ID is in the impacted units list.
func isImpactedUnit(unitID string, result *impact.ImpactResult) bool {
	for _, u := range result.ImpactedUnits {
		if u.UnitID == unitID {
			return true
		}
	}
	return false
}

// classifyReason maps an impacted test to a reason category string.
func classifyReason(test *impact.ImpactedTest) string {
	switch {
	case test.IsDirectlyChanged:
		return "directlyChanged"
	case test.ImpactConfidence == impact.ConfidenceExact:
		return "directDependency"
	case test.Relevance == "in same directory tree as changed code":
		return "directoryProximity"
	default:
		return "fixtureDependency"
	}
}

// confidenceScore converts an impact.Confidence or impact.EdgeKind to a numeric score.
func confidenceScore(conf impact.Confidence) float64 {
	switch conf {
	case impact.ConfidenceExact:
		return 0.95
	case impact.ConfidenceInferred:
		return 0.65
	case impact.ConfidenceWeak:
		return 0.30
	default:
		return 0.50
	}
}

// classifyConfidenceBand maps a numeric confidence to a band label.
func classifyConfidenceBand(conf float64) string {
	if conf >= 0.7 {
		return "high"
	}
	if conf >= 0.4 {
		return "medium"
	}
	return "low"
}

// edgeKindLabel converts an EdgeKind to a human-readable label.
func edgeKindLabel(kind impact.EdgeKind) string {
	switch kind {
	case impact.EdgeExactCoverage:
		return "exact per-test coverage"
	case impact.EdgeBucketCoverage:
		return "file-level coverage link"
	case impact.EdgeStructuralLink:
		return "import/export dependency"
	case impact.EdgeDirectoryProximity:
		return "directory proximity"
	case impact.EdgeNameConvention:
		return "naming convention match"
	default:
		return string(kind)
	}
}

// buildVerdict creates a one-line summary for a test explanation.
func buildVerdict(te *TestExplanation) string {
	if te.StrongestPath == nil {
		return fmt.Sprintf("Test %s is relevant to this change (confidence: %s).",
			te.Target.Path, te.ConfidenceBand)
	}

	strongest := te.StrongestPath
	if len(strongest.Steps) == 0 {
		return fmt.Sprintf("Test %s is relevant to this change (confidence: %s).",
			te.Target.Path, te.ConfidenceBand)
	}

	step := strongest.Steps[0]
	altCount := len(te.AlternativePaths)

	base := fmt.Sprintf("Selected via %s from %s (confidence: %s)",
		step.Relationship, step.From, te.ConfidenceBand)

	if altCount > 0 {
		return fmt.Sprintf("%s, plus %d alternative path(s).", base, altCount)
	}
	return base + "."
}

// buildTestLimitations generates limitation notes for a test explanation.
func buildTestLimitations(test *impact.ImpactedTest, result *impact.ImpactResult) []string {
	var lims []string

	if test.ImpactConfidence == impact.ConfidenceWeak {
		lims = append(lims, "Low confidence mapping — based on structural heuristics, not coverage data.")
	}

	if result.Graph != nil && result.Graph.Stats.ExactEdges == 0 {
		lims = append(lims, "No per-test coverage lineage available; all paths are inferred.")
	}

	return lims
}

// buildSelectionSummary creates a one-line summary for the selection explanation.
func buildSelectionSummary(sel *SelectionExplanation, result *impact.ImpactResult) string {
	parts := []string{
		fmt.Sprintf("%d test(s) selected", sel.TotalSelected),
	}

	if sel.Strategy != "" && sel.Strategy != "none" {
		parts = append(parts, fmt.Sprintf("strategy: %s", sel.Strategy))
	}

	parts = append(parts, fmt.Sprintf("coverage confidence: %s", sel.CoverageConfidence))

	if sel.GapCount > 0 {
		parts = append(parts, fmt.Sprintf("%d protection gap(s)", sel.GapCount))
	}

	return strings.Join(parts, ", ") + "."
}

// ScenarioExplanation is the structured explanation for why a scenario
// is impacted by a change.
type ScenarioExplanation struct {
	// ScenarioID is the scenario identifier.
	ScenarioID string `json:"scenarioId"`

	// Name is the human-readable scenario name.
	Name string `json:"name"`

	// Category is the scenario classification (safety, accuracy, etc.).
	Category string `json:"category,omitempty"`

	// Framework is the eval framework (promptfoo, deepeval, etc.).
	Framework string `json:"framework,omitempty"`

	// Verdict is a one-line summary of why this scenario is impacted.
	Verdict string `json:"verdict"`

	// ChangedSurfaces lists the specific code surfaces that triggered the impact.
	ChangedSurfaces []string `json:"changedSurfaces"`

	// Confidence is the impact confidence.
	Confidence string `json:"confidence"`

	// Relevance explains the impact relationship.
	Relevance string `json:"relevance"`

	// Capability is the inferred business capability this scenario validates.
	Capability string `json:"capability,omitempty"`

	// RelatedSurfaces groups all covered surfaces by kind for rich detail.
	RelatedSurfaces *RelatedSurfaces `json:"relatedSurfaces,omitempty"`

	// Signals lists AI signals related to this scenario.
	Signals []ScenarioSignal `json:"signals,omitempty"`

	// PolicyDecision describes the CI policy outcome for this scenario.
	PolicyDecision string `json:"policyDecision,omitempty"`
}

// RelatedSurfaces groups scenario-covered surfaces by kind.
type RelatedSurfaces struct {
	Prompts    []SurfaceRef `json:"prompts,omitempty"`
	Contexts   []SurfaceRef `json:"contexts,omitempty"`
	Datasets   []SurfaceRef `json:"datasets,omitempty"`
	ToolDefs   []SurfaceRef `json:"toolDefinitions,omitempty"`
	Retrievals []SurfaceRef `json:"retrievalSurfaces,omitempty"`
	Agents     []SurfaceRef `json:"agentSurfaces,omitempty"`
	EvalDefs   []SurfaceRef `json:"evalDefinitions,omitempty"`
	Other      []SurfaceRef `json:"other,omitempty"`
}

// SurfaceRef identifies a surface with its name, path, and kind.
type SurfaceRef struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Path          string  `json:"path"`
	Kind          string  `json:"kind"`
	Changed       bool    `json:"changed,omitempty"`
	DetectionTier string  `json:"detectionTier,omitempty"`
	Confidence    float64 `json:"confidence,omitempty"`
	Reason        string  `json:"reason,omitempty"`
}

// ScenarioSignal is an AI signal related to a scenario.
type ScenarioSignal struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Explanation string `json:"explanation"`
}

// ExplainScenario produces a structured explanation for why a specific
// scenario is impacted by the change described by the ImpactResult.
//
// The target can be a scenario ID or name.
func ExplainScenario(target string, result *impact.ImpactResult) (*ScenarioExplanation, error) {
	if result == nil {
		return nil, fmt.Errorf("no impact result available")
	}

	target = strings.TrimSpace(target)

	for _, sc := range result.ImpactedScenarios {
		if sc.ScenarioID == target || sc.Name == target {
			verdict := fmt.Sprintf(
				"Scenario %q is impacted because %d of its covered code surfaces changed.",
				sc.Name, len(sc.CoversSurfaces))
			if len(sc.CoversSurfaces) == 1 {
				verdict = fmt.Sprintf(
					"Scenario %q is impacted because its covered surface %s changed.",
					sc.Name, sc.CoversSurfaces[0])
			}

			return &ScenarioExplanation{
				ScenarioID:      sc.ScenarioID,
				Name:            sc.Name,
				Category:        sc.Category,
				Framework:       sc.Framework,
				Verdict:         verdict,
				ChangedSurfaces: sc.CoversSurfaces,
				Confidence:      string(sc.ImpactConfidence),
				Relevance:       sc.Relevance,
				Capability:      sc.Capability,
			}, nil
		}
	}

	return nil, fmt.Errorf("scenario not found in impact analysis: %s", target)
}

// ExplainScenarioRich produces an enriched explanation with surface-kind
// breakdowns, signals, and policy decisions.
func ExplainScenarioRich(target string, result *impact.ImpactResult, snap *models.TestSuiteSnapshot) (*ScenarioExplanation, error) {
	base, err := ExplainScenario(target, result)
	if err != nil {
		return nil, err
	}

	// Find the source scenario for covered surface IDs.
	var sourceScenario *models.Scenario
	for i, sc := range snap.Scenarios {
		if sc.ScenarioID == base.ScenarioID || sc.Name == base.Name {
			sourceScenario = &snap.Scenarios[i]
			break
		}
	}
	if sourceScenario == nil {
		return base, nil
	}

	// Build surface index.
	surfaceByID := map[string]models.CodeSurface{}
	for _, cs := range snap.CodeSurfaces {
		surfaceByID[cs.SurfaceID] = cs
	}

	// Build changed set.
	changedSet := map[string]bool{}
	for _, sid := range base.ChangedSurfaces {
		changedSet[sid] = true
	}

	// Group covered surfaces by kind.
	related := &RelatedSurfaces{}
	for _, sid := range sourceScenario.CoveredSurfaceIDs {
		cs, ok := surfaceByID[sid]
		if !ok {
			related.Other = append(related.Other, SurfaceRef{
				ID: sid, Name: sid, Kind: "unknown", Changed: changedSet[sid],
			})
			continue
		}
		ref := SurfaceRef{
			ID: cs.SurfaceID, Name: cs.Name, Path: cs.Path,
			Kind: string(cs.Kind), Changed: changedSet[cs.SurfaceID],
			DetectionTier: cs.DetectionTier, Confidence: cs.Confidence,
			Reason: cs.Reason,
		}
		switch cs.Kind {
		case models.SurfacePrompt:
			related.Prompts = append(related.Prompts, ref)
		case models.SurfaceContext:
			related.Contexts = append(related.Contexts, ref)
		case models.SurfaceDataset:
			related.Datasets = append(related.Datasets, ref)
		case models.SurfaceToolDef:
			related.ToolDefs = append(related.ToolDefs, ref)
		case models.SurfaceRetrieval:
			related.Retrievals = append(related.Retrievals, ref)
		case models.SurfaceAgent:
			related.Agents = append(related.Agents, ref)
		case models.SurfaceEvalDef:
			related.EvalDefs = append(related.EvalDefs, ref)
		default:
			related.Other = append(related.Other, ref)
		}
	}
	base.RelatedSurfaces = related

	// Collect signals for this scenario.
	for _, sig := range snap.Signals {
		if sig.Category == models.CategoryAI && sig.Location.ScenarioID == base.ScenarioID {
			base.Signals = append(base.Signals, ScenarioSignal{
				Type: string(sig.Type), Severity: string(sig.Severity),
				Explanation: sig.Explanation,
			})
		}
	}

	// Policy decision from governance signals.
	for _, sig := range snap.Signals {
		if sig.Category == models.CategoryGovernance {
			if md, ok := sig.Metadata["rule"]; ok {
				rule, _ := md.(string)
				if strings.HasPrefix(rule, "block_on_") || rule == "blocking_signal_types" {
					base.PolicyDecision = "blocked: " + sig.Explanation
					break
				}
			}
		}
	}
	if base.PolicyDecision == "" && len(base.Signals) > 0 {
		base.PolicyDecision = "warn"
	} else if base.PolicyDecision == "" {
		base.PolicyDecision = "pass"
	}

	return base, nil
}
