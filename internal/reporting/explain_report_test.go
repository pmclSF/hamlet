package reporting

import (
	"bytes"
	"strings"
	"testing"

	"github.com/pmclSF/terrain/internal/explain"
)

func TestRenderTestExplanation_VerboseShowsEvidence(t *testing.T) {
	t.Parallel()
	te := &explain.TestExplanation{
		Target: explain.TestTarget{
			Path:      "test/auth.test.ts",
			Framework: "jest",
		},
		Verdict:        "Selected via import/export dependency from src/auth.ts:login (confidence: high).",
		Confidence:     0.95,
		ConfidenceBand: "high",
		ReasonCategory: "directDependency",
		StrongestPath: &explain.ReasonChain{
			Steps: []explain.ChainStep{
				{
					From:           "src/auth.ts:login",
					To:             "test/auth.test.ts",
					Relationship:   "import/export dependency",
					EdgeKind:       "structural_link",
					EdgeConfidence: 0.95,
				},
			},
			Confidence: 0.95,
			Band:       "high",
		},
		CoversUnits: []string{"src/auth.ts:login", "src/auth.ts:register"},
	}

	// Non-verbose: should NOT contain "Evidence detail".
	var buf bytes.Buffer
	RenderTestExplanation(&buf, te)
	output := buf.String()
	if strings.Contains(output, "Evidence detail") {
		t.Error("non-verbose output should not contain 'Evidence detail'")
	}
	if !strings.Contains(output, "Verdict:") {
		t.Error("expected Verdict in output")
	}
	if !strings.Contains(output, "Confidence: high") {
		t.Error("expected confidence band in output")
	}

	// Verbose: should contain evidence detail section.
	buf.Reset()
	RenderTestExplanation(&buf, te, true)
	verboseOutput := buf.String()
	if !strings.Contains(verboseOutput, "Evidence detail") {
		t.Error("verbose output should contain 'Evidence detail'")
	}
	if !strings.Contains(verboseOutput, "Edge: src/auth.ts:login") {
		t.Error("verbose output should contain edge source")
	}
	if !strings.Contains(verboseOutput, "Edge kind:    structural_link") {
		t.Error("verbose output should contain edge kind")
	}
	if !strings.Contains(verboseOutput, "Confidence:   95%") {
		t.Error("verbose output should contain edge confidence percentage")
	}
}

func TestRenderTestExplanation_VerboseAlternativePaths(t *testing.T) {
	t.Parallel()
	te := &explain.TestExplanation{
		Target:         explain.TestTarget{Path: "test/utils.test.ts"},
		Verdict:        "Selected.",
		Confidence:     0.95,
		ConfidenceBand: "high",
		ReasonCategory: "directDependency",
		StrongestPath: &explain.ReasonChain{
			Steps: []explain.ChainStep{
				{From: "src/utils.ts:parse", To: "test/utils.test.ts", Relationship: "import", EdgeKind: "structural_link", EdgeConfidence: 0.95},
			},
			Confidence: 0.95,
			Band:       "high",
		},
		AlternativePaths: []explain.ReasonChain{
			{
				Steps: []explain.ChainStep{
					{From: "src/utils.ts:format", To: "test/utils.test.ts", Relationship: "import", EdgeKind: "structural_link", EdgeConfidence: 0.80},
				},
				Confidence: 0.80,
				Band:       "high",
			},
		},
	}

	var buf bytes.Buffer
	RenderTestExplanation(&buf, te, true)
	output := buf.String()
	if !strings.Contains(output, "Alternative 1:") {
		t.Error("verbose output should show alternative paths")
	}
	if !strings.Contains(output, "src/utils.ts:format") {
		t.Error("verbose output should contain alternative path source")
	}
}

func TestRenderSelectionExplanation_VerboseShowsPerTestEvidence(t *testing.T) {
	t.Parallel()
	sel := &explain.SelectionExplanation{
		Summary:            "2 test(s) selected, strategy: exact.",
		Strategy:           "exact",
		TotalSelected:      2,
		CoverageConfidence: "high",
		ReasonBreakdown: map[string]int{
			"directDependency": 2,
		},
		HighConfidenceTests: []explain.TestExplanation{
			{
				Target:         explain.TestTarget{Path: "test/a.test.ts"},
				Confidence:     0.95,
				ConfidenceBand: "high",
				StrongestPath: &explain.ReasonChain{
					Steps: []explain.ChainStep{
						{From: "src/a.ts:foo", To: "test/a.test.ts", Relationship: "import", EdgeKind: "structural_link", EdgeConfidence: 0.95},
					},
					Confidence: 0.95,
					Band:       "high",
				},
			},
		},
	}

	// Non-verbose.
	var buf bytes.Buffer
	RenderSelectionExplanation(&buf, sel)
	output := buf.String()
	if strings.Contains(output, "edge:") {
		t.Error("non-verbose selection should not contain edge details")
	}

	// Verbose.
	buf.Reset()
	RenderSelectionExplanation(&buf, sel, true)
	verboseOutput := buf.String()
	if !strings.Contains(verboseOutput, "edge:") {
		t.Error("verbose selection should contain per-test edge details")
	}
	if !strings.Contains(verboseOutput, "structural_link") {
		t.Error("verbose selection should contain edge kind")
	}
}

func TestRenderSurfaceEvidence(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	RenderSurfaceEvidence(&buf, "system_prompt", "src/prompts.ts", 5, "structural", 0.95, "[ast:system-prompt] system prompt assignment")

	output := buf.String()
	if !strings.Contains(output, "system_prompt") {
		t.Error("expected surface name in output")
	}
	if !strings.Contains(output, "src/prompts.ts:5") {
		t.Error("expected file:line in output")
	}
	if !strings.Contains(output, "tier: structural") {
		t.Error("expected tier in output")
	}
	if !strings.Contains(output, "confidence: 95%") {
		t.Error("expected confidence percentage in output")
	}
	if !strings.Contains(output, "evidence: [ast:system-prompt]") {
		t.Error("expected evidence reason in output")
	}
}

func TestRenderTestExplanation_DeterministicOutput(t *testing.T) {
	t.Parallel()
	te := &explain.TestExplanation{
		Target:         explain.TestTarget{Path: "test/auth.test.ts", Framework: "jest"},
		Verdict:        "Selected.",
		Confidence:     0.90,
		ConfidenceBand: "high",
		ReasonCategory: "directDependency",
		StrongestPath: &explain.ReasonChain{
			Steps: []explain.ChainStep{
				{From: "src/auth.ts:login", To: "test/auth.test.ts", Relationship: "import", EdgeKind: "structural_link", EdgeConfidence: 0.90},
			},
			Confidence: 0.90,
			Band:       "high",
		},
		CoversUnits: []string{"src/auth.ts:login"},
	}

	var buf1, buf2 bytes.Buffer
	RenderTestExplanation(&buf1, te, true)
	RenderTestExplanation(&buf2, te, true)

	if buf1.String() != buf2.String() {
		t.Error("rendering should be deterministic — two calls with same input produced different output")
	}
}

func TestRenderSelectionExplanation_DeterministicOutput(t *testing.T) {
	t.Parallel()
	sel := &explain.SelectionExplanation{
		Summary:            "1 test(s) selected.",
		Strategy:           "exact",
		TotalSelected:      1,
		CoverageConfidence: "high",
		ReasonBreakdown:    map[string]int{"directDependency": 1},
		HighConfidenceTests: []explain.TestExplanation{
			{
				Target:         explain.TestTarget{Path: "test/a.test.ts"},
				Confidence:     0.95,
				ConfidenceBand: "high",
			},
		},
	}

	var buf1, buf2 bytes.Buffer
	RenderSelectionExplanation(&buf1, sel, true)
	RenderSelectionExplanation(&buf2, sel, true)

	if buf1.String() != buf2.String() {
		t.Error("selection rendering should be deterministic")
	}
}
