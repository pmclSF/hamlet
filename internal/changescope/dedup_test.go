package changescope

import (
	"bytes"
	"strings"
	"testing"
)

func TestDeduplicateFindings_RemovesDuplicates(t *testing.T) {
	t.Parallel()
	findings := []ChangeScopedFinding{
		{Type: "protection_gap", Path: "src/auth.ts", Severity: "high", Explanation: "Function login has no test coverage."},
		{Type: "protection_gap", Path: "src/auth.ts", Severity: "high", Explanation: "Function login has no test coverage."},
		{Type: "protection_gap", Path: "src/auth.ts", Severity: "medium", Explanation: "Different finding."},
		{Type: "existing_signal", Path: "src/auth.ts", Severity: "medium", Explanation: "Weak assertion."},
	}

	deduped := DeduplicateFindings(findings)
	if len(deduped) != 3 {
		t.Errorf("expected 3 unique findings, got %d", len(deduped))
	}
}

func TestDeduplicateFindings_NormalizesExplanation(t *testing.T) {
	t.Parallel()
	findings := []ChangeScopedFinding{
		{Type: "protection_gap", Path: "src/a.ts", Severity: "high", Explanation: "No test coverage."},
		{Type: "protection_gap", Path: "src/a.ts", Severity: "high", Explanation: "No test coverage"},  // missing period
		{Type: "protection_gap", Path: "src/a.ts", Severity: "high", Explanation: "  No test coverage. "}, // extra whitespace
	}

	deduped := DeduplicateFindings(findings)
	if len(deduped) != 1 {
		t.Errorf("expected 1 finding after normalization, got %d", len(deduped))
	}
}

func TestDeduplicateFindings_Empty(t *testing.T) {
	t.Parallel()
	deduped := DeduplicateFindings(nil)
	if len(deduped) != 0 {
		t.Errorf("expected 0, got %d", len(deduped))
	}
}

func TestClassifyFindings_SeparatesNewAndExisting(t *testing.T) {
	t.Parallel()
	findings := []ChangeScopedFinding{
		{Type: "protection_gap", Severity: "high"},
		{Type: "existing_signal", Severity: "medium"},
		{Type: "protection_gap", Severity: "medium"},
		{Type: "existing_signal", Severity: "low"},
	}

	newRisk, existing := ClassifyFindings(findings)
	if len(newRisk) != 2 {
		t.Errorf("expected 2 new risks, got %d", len(newRisk))
	}
	if len(existing) != 2 {
		t.Errorf("expected 2 existing, got %d", len(existing))
	}
}

func TestGroupTestsByPackage(t *testing.T) {
	t.Parallel()
	paths := []string{
		"tests/unit/auth/login.test.ts",
		"tests/unit/auth/session.test.ts",
		"tests/unit/billing/invoice.test.ts",
		"tests/e2e/checkout.test.ts",
	}

	groups := GroupTestsByPackage(paths)
	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(groups))
	}
	// Largest group first.
	if groups[0].Count != 2 {
		t.Errorf("expected largest group to have 2, got %d", groups[0].Count)
	}
}

func TestMergeRecommendation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		band     string
		findings []ChangeScopedFinding
		wantRec  string
	}{
		{"well_protected", nil, "Safe to merge"},
		{"evidence_limited", nil, "Informational only"},
		{"partially_protected", []ChangeScopedFinding{{Severity: "high"}}, "Merge with caution"},
		{"high_risk", []ChangeScopedFinding{{Severity: "medium"}}, "Merge blocked"},
		{"partially_protected", []ChangeScopedFinding{{Severity: "medium"}}, "Merge with caution"},
	}

	for _, tt := range tests {
		rec, _ := MergeRecommendation(tt.band, tt.findings)
		if rec != tt.wantRec {
			t.Errorf("MergeRecommendation(%q, %d findings) = %q, want %q",
				tt.band, len(tt.findings), rec, tt.wantRec)
		}
	}
}

func TestRenderPRSummaryMarkdown_Deterministic(t *testing.T) {
	t.Parallel()
	pr := &PRAnalysis{
		PostureBand:        "partially_protected",
		ChangedFileCount:   5,
		ChangedSourceCount: 3,
		ChangedTestCount:   2,
		ImpactedUnitCount:  4,
		ProtectionGapCount: 2,
		TotalTestCount:     100,
		NewFindings: []ChangeScopedFinding{
			{Type: "protection_gap", Scope: "direct", Path: "src/auth.ts", Severity: "high", Explanation: "No coverage"},
			{Type: "protection_gap", Scope: "indirect", Path: "src/utils.ts", Severity: "medium", Explanation: "Indirectly impacted"},
			{Type: "existing_signal", Scope: "direct", Path: "src/db.ts", Severity: "medium", Explanation: "[weakAssertion] Weak"},
		},
		TestSelections: []TestSelection{
			{Path: "tests/auth.test.ts", Confidence: "exact", CoversUnits: []string{"src/auth.ts:login"}},
		},
		RecommendedTests: []string{"tests/auth.test.ts"},
	}

	var buf1, buf2 bytes.Buffer
	RenderPRSummaryMarkdown(&buf1, pr)
	RenderPRSummaryMarkdown(&buf2, pr)

	if buf1.String() != buf2.String() {
		t.Error("markdown render is not deterministic")
	}

	output := buf1.String()
	if !strings.Contains(output, "Merge with caution") {
		t.Error("expected merge recommendation in output")
	}
	if !strings.Contains(output, "New Risks (directly changed)") {
		t.Error("expected direct risks section")
	}
	if !strings.Contains(output, "Indirectly impacted") {
		t.Error("expected indirect risks section")
	}
	if !strings.Contains(output, "Pre-existing issues") {
		t.Error("expected pre-existing issues section")
	}
	if !strings.Contains(output, "1 of 100") {
		t.Error("expected suite size context in tests row")
	}
}

func TestRenderPRSummaryMarkdown_LargeTestSet_GroupsByPackage(t *testing.T) {
	t.Parallel()

	var selections []TestSelection
	var recommended []string
	for i := 0; i < 25; i++ {
		pkg := "tests/unit/auth"
		if i >= 10 {
			pkg = "tests/unit/billing"
		}
		if i >= 20 {
			pkg = "tests/e2e"
		}
		path := pkg + "/" + strings.Repeat("x", i) + ".test.ts"
		selections = append(selections, TestSelection{Path: path, Confidence: "inferred"})
		recommended = append(recommended, path)
	}

	pr := &PRAnalysis{
		PostureBand:      "partially_protected",
		TestSelections:   selections,
		RecommendedTests: recommended,
	}

	var buf bytes.Buffer
	RenderPRSummaryMarkdown(&buf, pr)
	output := buf.String()

	// Should group by package, not list 25 individual files.
	if strings.Contains(output, "| Confidence |") {
		t.Error("expected grouped table, not individual test table for 25+ tests")
	}
	if !strings.Contains(output, "Package") {
		t.Error("expected Package grouping header")
	}
}

func TestRenderPRSummaryMarkdown_FindingTruncation(t *testing.T) {
	t.Parallel()

	var findings []ChangeScopedFinding
	for i := 0; i < 20; i++ {
		findings = append(findings, ChangeScopedFinding{
			Type:        "protection_gap",
			Path:        "src/file_" + strings.Repeat("a", i) + ".ts",
			Severity:    "medium",
			Explanation: "No coverage",
		})
	}

	pr := &PRAnalysis{
		PostureBand: "weakly_protected",
		NewFindings: findings,
	}

	var buf bytes.Buffer
	RenderPRSummaryMarkdown(&buf, pr)
	output := buf.String()

	if !strings.Contains(output, "... and 10 more") {
		t.Error("expected truncation message for >10 findings")
	}
}

func TestClassifyFindingsDetailed_ThreeWay(t *testing.T) {
	t.Parallel()
	findings := []ChangeScopedFinding{
		{Type: "protection_gap", Scope: "direct", Severity: "high", Path: "src/auth.ts"},
		{Type: "protection_gap", Scope: "indirect", Severity: "medium", Path: "src/utils.ts"},
		{Type: "protection_gap", Scope: "direct", Severity: "medium", Path: "src/db.ts"},
		{Type: "existing_signal", Scope: "direct", Severity: "low", Path: "src/auth.ts"},
	}

	direct, indirect, existing := ClassifyFindingsDetailed(findings)
	if len(direct) != 2 {
		t.Errorf("expected 2 direct, got %d", len(direct))
	}
	if len(indirect) != 1 {
		t.Errorf("expected 1 indirect, got %d", len(indirect))
	}
	if len(existing) != 1 {
		t.Errorf("expected 1 existing, got %d", len(existing))
	}
}

func TestRenderPRSummaryMarkdown_SuiteSizeContext(t *testing.T) {
	t.Parallel()
	pr := &PRAnalysis{
		PostureBand:      "well_protected",
		TotalTestCount:   200,
		RecommendedTests: []string{"a.test.ts", "b.test.ts"},
		TestSelections: []TestSelection{
			{Path: "a.test.ts", Confidence: "exact"},
			{Path: "b.test.ts", Confidence: "inferred"},
		},
	}
	var buf bytes.Buffer
	RenderPRSummaryMarkdown(&buf, pr)
	output := buf.String()
	if !strings.Contains(output, "2 of 200") {
		t.Errorf("expected suite size '2 of 200', got:\n%s", output)
	}
	if !strings.Contains(output, "1%% of suite") && !strings.Contains(output, "1% of suite") {
		t.Errorf("expected percentage in output, got:\n%s", output)
	}
}

func TestRenderPRSummaryMarkdown_DirectVsIndirectSections(t *testing.T) {
	t.Parallel()
	pr := &PRAnalysis{
		PostureBand: "weakly_protected",
		NewFindings: []ChangeScopedFinding{
			{Type: "protection_gap", Scope: "direct", Path: "src/a.ts", Severity: "high", Explanation: "Direct gap"},
			{Type: "protection_gap", Scope: "indirect", Path: "src/b.ts", Severity: "medium", Explanation: "Indirect gap"},
		},
	}
	var buf bytes.Buffer
	RenderPRSummaryMarkdown(&buf, pr)
	output := buf.String()

	// Direct risks in main section
	if !strings.Contains(output, "New Risks (directly changed)") {
		t.Error("expected 'New Risks (directly changed)' heading")
	}
	if !strings.Contains(output, "`src/a.ts`: Direct gap") {
		t.Error("expected direct finding in main section")
	}

	// Indirect risks in collapsed section
	if !strings.Contains(output, "Indirectly impacted protection gaps (1)") {
		t.Error("expected indirect section with count")
	}
	if !strings.Contains(output, "`src/b.ts`: Indirect gap") {
		t.Error("expected indirect finding in collapsed section")
	}
}

func TestRenderChangeScopedReport_ShowsTestReduction(t *testing.T) {
	t.Parallel()
	pr := &PRAnalysis{
		PostureBand:      "well_protected",
		TotalTestCount:   50,
		RecommendedTests: []string{"a.test.ts"},
		TestSelections:   []TestSelection{{Path: "a.test.ts", Confidence: "exact"}},
	}
	var buf bytes.Buffer
	RenderChangeScopedReport(&buf, pr)
	output := buf.String()
	if !strings.Contains(output, "1 of 50") {
		t.Errorf("expected '1 of 50' in text report, got:\n%s", output)
	}
}

func TestSummarizeFindingsBySeverity(t *testing.T) {
	t.Parallel()
	findings := []ChangeScopedFinding{
		{Severity: "high"}, {Severity: "high"}, {Severity: "medium"}, {Severity: "low"},
	}
	counts := SummarizeFindingsBySeverity(findings)
	if counts["high"] != 2 || counts["medium"] != 1 || counts["low"] != 1 {
		t.Errorf("unexpected counts: %v", counts)
	}
}

// --- AI PR Section Tests ---

func TestRenderPRSummaryMarkdown_AISection(t *testing.T) {
	t.Parallel()
	pr := &PRAnalysis{
		PostureBand: "partially_protected",
		AI: &AIValidationSummary{
			ImpactedCapabilities: []string{"refund-explanation", "enterprise-search"},
			SelectedScenarios:    3,
			TotalScenarios:       8,
			Scenarios: []AIScenarioSummary{
				{Name: "refund-accuracy", Capability: "refund-explanation", Reason: "context template changed (policyBlock)"},
				{Name: "search-citations", Capability: "enterprise-search", Reason: "retrieval config changed (chunkConfig)"},
				{Name: "safety-guardrail", Capability: "refund-explanation", Reason: "prompt changed (safetyOverlay)"},
			},
			BlockingSignals: []AISignalSummary{
				{Type: "safetyFailure", Severity: "critical", Explanation: "Safety eval failed"},
			},
			WarningSignals: []AISignalSummary{
				{Type: "latencyRegression", Severity: "medium", Explanation: "p95 latency regressed"},
			},
			UncoveredContexts: []string{"customerContext (src/context.ts)"},
		},
	}

	var buf bytes.Buffer
	RenderPRSummaryMarkdown(&buf, pr)
	output := buf.String()

	// AI section header.
	if !strings.Contains(output, "### AI Validation") {
		t.Error("expected AI Validation section")
	}
	// Capabilities.
	if !strings.Contains(output, "refund-explanation") {
		t.Error("expected refund-explanation capability")
	}
	if !strings.Contains(output, "enterprise-search") {
		t.Error("expected enterprise-search capability")
	}
	// Scenario counts.
	if !strings.Contains(output, "3 of 8 selected") {
		t.Error("expected scenario count '3 of 8 selected'")
	}
	// Blocking signals.
	if !strings.Contains(output, "Blocking signals") {
		t.Error("expected blocking signals section")
	}
	if !strings.Contains(output, "safetyFailure") {
		t.Error("expected safetyFailure in output")
	}
	// Warning signals (collapsed).
	if !strings.Contains(output, "Warning signals") {
		t.Error("expected warning signals section")
	}
	// Uncovered contexts.
	if !strings.Contains(output, "customerContext") {
		t.Error("expected uncovered context")
	}
}

func TestRenderPRSummaryMarkdown_NoAISection(t *testing.T) {
	t.Parallel()
	pr := &PRAnalysis{
		PostureBand: "well_protected",
		// No AI field — traditional-only PR.
	}

	var buf bytes.Buffer
	RenderPRSummaryMarkdown(&buf, pr)
	output := buf.String()

	if strings.Contains(output, "AI Validation") {
		t.Error("expected no AI section for traditional PR")
	}
}

func TestRenderPRSummaryMarkdown_MixedTraditionalAndAI(t *testing.T) {
	t.Parallel()
	pr := &PRAnalysis{
		PostureBand:        "partially_protected",
		ChangedFileCount:   5,
		ChangedSourceCount: 3,
		ChangedTestCount:   2,
		ImpactedUnitCount:  4,
		ProtectionGapCount: 1,
		TotalTestCount:     50,
		NewFindings: []ChangeScopedFinding{
			{Type: "protection_gap", Scope: "direct", Path: "src/auth.ts", Severity: "high", Explanation: "No coverage"},
		},
		RecommendedTests: []string{"tests/auth.test.ts"},
		TestSelections:   []TestSelection{{Path: "tests/auth.test.ts", Confidence: "exact"}},
		AI: &AIValidationSummary{
			ImpactedCapabilities: []string{"search"},
			SelectedScenarios:    1,
			TotalScenarios:       3,
			Scenarios: []AIScenarioSummary{
				{Name: "search-quality", Capability: "search", Reason: "retrieval config changed"},
			},
		},
	}

	var buf bytes.Buffer
	RenderPRSummaryMarkdown(&buf, pr)
	output := buf.String()

	// Both traditional and AI sections present.
	if !strings.Contains(output, "New Risks") {
		t.Error("expected traditional New Risks section")
	}
	if !strings.Contains(output, "Recommended Tests") {
		t.Error("expected traditional Recommended Tests section")
	}
	if !strings.Contains(output, "### AI Validation") {
		t.Error("expected AI Validation section")
	}
	if !strings.Contains(output, "search-quality") {
		t.Error("expected AI scenario in output")
	}
}

func TestRenderPRSummaryMarkdown_AISection_Deterministic(t *testing.T) {
	t.Parallel()
	pr := &PRAnalysis{
		PostureBand: "well_protected",
		AI: &AIValidationSummary{
			ImpactedCapabilities: []string{"billing", "auth"},
			SelectedScenarios:    2,
			TotalScenarios:       5,
			Scenarios: []AIScenarioSummary{
				{Name: "billing-accuracy", Capability: "billing", Reason: "prompt changed"},
				{Name: "auth-safety", Capability: "auth", Reason: "context changed"},
			},
		},
	}

	var buf1, buf2 bytes.Buffer
	RenderPRSummaryMarkdown(&buf1, pr)
	RenderPRSummaryMarkdown(&buf2, pr)

	if buf1.String() != buf2.String() {
		t.Error("AI section rendering is not deterministic")
	}
}
