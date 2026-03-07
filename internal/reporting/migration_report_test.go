package reporting

import (
	"bytes"
	"strings"
	"testing"

	"github.com/pmclSF/hamlet/internal/migration"
	"github.com/pmclSF/hamlet/internal/models"
)

func TestRenderMigrationPreview_WithBlockers(t *testing.T) {
	preview := &migration.PreviewResult{
		File:             "test/auth/login.test.js",
		SourceFramework:  "jest",
		SuggestedTarget:  "vitest",
		Difficulty:       "medium",
		PreviewAvailable: true,
		Explanation:      "Source framework: jest. Suggested target: vitest. 2 blocker(s).",
		Blockers: []migration.PreviewBlocker{
			{
				Type:        "deprecated-pattern",
				Pattern:     "done-callback",
				Explanation: "done() callbacks are deprecated.",
				Remediation: "Convert to async/await.",
			},
			{
				Type:        "custom-matcher",
				Pattern:     "custom-matcher",
				Explanation: "Custom matchers need rewriting.",
				Remediation: "Create equivalent for target framework.",
			},
		},
		SafePatterns: []string{
			"standard test structure (describe/it/test)",
			"expect() assertions",
		},
		Limitations: []string{
			"Preview is based on structural pattern analysis.",
		},
	}

	var buf bytes.Buffer
	RenderMigrationPreview(&buf, preview)
	output := buf.String()

	checks := []string{
		"Migration Preview",
		"test/auth/login.test.js",
		"jest",
		"vitest",
		"MEDIUM",
		"deprecated-pattern",
		"custom-matcher",
		"Safe Patterns",
		"Limitations",
	}
	for _, c := range checks {
		if !strings.Contains(output, c) {
			t.Errorf("output missing %q", c)
		}
	}
}

func TestRenderMigrationPreview_NotAvailable(t *testing.T) {
	preview := &migration.PreviewResult{
		File:             "test_main.go",
		SourceFramework:  "go-testing",
		Difficulty:       "unknown",
		PreviewAvailable: false,
		Explanation:      "Migration preview is currently supported for JavaScript/TypeScript frameworks.",
		Limitations: []string{
			"No preview support for go frameworks yet.",
		},
	}

	var buf bytes.Buffer
	RenderMigrationPreview(&buf, preview)
	output := buf.String()

	if !strings.Contains(output, "Preview Not Available") {
		t.Error("expected 'Preview Not Available' in output")
	}
	if !strings.Contains(output, "JavaScript/TypeScript") {
		t.Error("expected JS/TS limitation message")
	}
}

func TestRenderMigrationPreviewScope(t *testing.T) {
	previews := []*migration.PreviewResult{
		{File: "test/hard.test.js", SourceFramework: "jest", Difficulty: "high", Blockers: make([]migration.PreviewBlocker, 3)},
		{File: "test/med.test.js", SourceFramework: "jest", Difficulty: "medium", Blockers: make([]migration.PreviewBlocker, 1)},
		{File: "test/easy.test.js", SourceFramework: "jest", Difficulty: "low"},
	}

	var buf bytes.Buffer
	RenderMigrationPreviewScope(&buf, previews)
	output := buf.String()

	checks := []string{
		"Files analyzed: 3",
		"High-Difficulty",
		"Medium-Difficulty",
		"Low-Difficulty",
		"test/hard.test.js",
		"test/easy.test.js",
	}
	for _, c := range checks {
		if !strings.Contains(output, c) {
			t.Errorf("output missing %q", c)
		}
	}
}

func TestRenderMigrationReport_WithAllSections(t *testing.T) {
	readiness := &migration.ReadinessSummary{
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 42},
		},
		TotalBlockers:  3,
		BlockersByType: map[string]int{"deprecated-pattern": 2, "custom-matcher": 1},
		RepresentativeBlockers: []migration.BlockerExample{
			{Type: "deprecatedTestPattern", File: "test/a.js", Explanation: "Done callback used"},
		},
		ReadinessLevel: "medium",
		Explanation:    "Some blockers found.",
		QualityFactors: []migration.QualityFactor{
			{SignalType: "weakAssertion", AffectedFiles: 2, Explanation: "2 files have weak assertions."},
		},
		AreaAssessments: []migration.AreaAssessment{
			{Directory: "test/auth", Classification: "risky", Explanation: "Blockers + quality issues."},
		},
		CoverageGuidance: []migration.CoverageGuidanceItem{
			{Directory: "test/auth", Reason: "migration blockers", Priority: "high"},
		},
	}

	var buf bytes.Buffer
	RenderMigrationReport(&buf, readiness)
	output := buf.String()

	checks := []string{
		"Migration Readiness",
		"jest",
		"MEDIUM",
		"deprecated-pattern",
		"Quality Factors",
		"Area Assessments",
		"RISKY",
		"Coverage Reduces Migration Risk",
	}
	for _, c := range checks {
		if !strings.Contains(output, c) {
			t.Errorf("output missing %q", c)
		}
	}
}

func TestRenderMigrationBlockers_ZeroBlockers(t *testing.T) {
	readiness := &migration.ReadinessSummary{
		TotalBlockers:  0,
		BlockersByType: map[string]int{},
		ReadinessLevel: "high",
	}

	var buf bytes.Buffer
	RenderMigrationBlockers(&buf, readiness)
	output := buf.String()

	if !strings.Contains(output, "No migration blockers detected") {
		t.Error("expected zero-blockers message")
	}
}
