package testtype

import (
	"testing"

	"github.com/pmclSF/hamlet/internal/models"
)

func TestInfer_E2EFramework(t *testing.T) {
	tc := &models.TestCase{
		Framework: "playwright",
		FilePath:  "e2e/login.spec.ts",
	}
	r := InferForTestCase(tc)
	if r.Type != TypeE2E {
		t.Errorf("type = %q, want e2e", r.Type)
	}
	if r.Confidence < 0.8 {
		t.Errorf("confidence = %f, want >= 0.8", r.Confidence)
	}
}

func TestInfer_UnitByPath(t *testing.T) {
	tc := &models.TestCase{
		Framework: "jest",
		FilePath:  "src/__tests__/utils.test.js",
	}
	r := InferForTestCase(tc)
	if r.Type != TypeUnit {
		t.Errorf("type = %q, want unit", r.Type)
	}
}

func TestInfer_IntegrationByPath(t *testing.T) {
	tc := &models.TestCase{
		Framework: "jest",
		FilePath:  "test/integration/db.test.js",
	}
	r := InferForTestCase(tc)
	if r.Type != TypeIntegration {
		t.Errorf("type = %q, want integration", r.Type)
	}
	// Confidence is reduced by conflict (jest=unit vs path=integration).
	if r.Confidence < 0.5 {
		t.Errorf("confidence = %f, want >= 0.5", r.Confidence)
	}
}

func TestInfer_E2EByPath(t *testing.T) {
	tc := &models.TestCase{
		Framework: "jest",
		FilePath:  "e2e/checkout.e2e.test.js",
	}
	r := InferForTestCase(tc)
	if r.Type != TypeE2E {
		t.Errorf("type = %q, want e2e", r.Type)
	}
}

func TestInfer_SmokeByPath(t *testing.T) {
	tc := &models.TestCase{
		Framework: "jest",
		FilePath:  "test/smoke/health.test.js",
	}
	r := InferForTestCase(tc)
	if r.Type != TypeSmoke {
		t.Errorf("type = %q, want smoke", r.Type)
	}
}

func TestInfer_BySuiteName(t *testing.T) {
	tc := &models.TestCase{
		Framework:      "mocha",
		FilePath:       "test/api.test.js",
		SuiteHierarchy: []string{"Integration Tests", "API"},
	}
	r := InferForTestCase(tc)
	if r.Type != TypeIntegration {
		t.Errorf("type = %q, want integration", r.Type)
	}
}

func TestInfer_CypressFileExtension(t *testing.T) {
	tc := &models.TestCase{
		Framework: "cypress",
		FilePath:  "cypress/e2e/login.cy.js",
	}
	r := InferForTestCase(tc)
	if r.Type != TypeE2E {
		t.Errorf("type = %q, want e2e", r.Type)
	}
}

func TestInfer_UnknownWhenAmbiguous(t *testing.T) {
	tc := &models.TestCase{
		Framework: "unknown",
		FilePath:  "lib/something.js",
	}
	r := InferForTestCase(tc)
	// With no signals, should be unknown.
	if r.Type != TypeUnknown && r.Confidence > 0.5 {
		t.Errorf("ambiguous case should be unknown or low confidence, got %q at %f", r.Type, r.Confidence)
	}
}

func TestInfer_ConflictReducesConfidence(t *testing.T) {
	// Path says integration, framework says e2e.
	tc := &models.TestCase{
		Framework: "playwright",
		FilePath:  "test/integration/visual.spec.ts",
	}
	r := InferForTestCase(tc)
	// Should have evidence of conflict.
	hasConflictEvidence := false
	for _, e := range r.Evidence {
		if e == "conflicting signals reduced confidence" {
			hasConflictEvidence = true
		}
	}
	if !hasConflictEvidence {
		t.Error("expected conflict evidence when framework and path disagree")
	}
}

func TestInferAll(t *testing.T) {
	cases := []models.TestCase{
		{Framework: "jest", FilePath: "src/__tests__/a.test.js", TestName: "works"},
		{Framework: "playwright", FilePath: "e2e/b.spec.ts", TestName: "loads"},
	}
	result := InferAll(cases)
	if result[0].TestType != TypeUnit {
		t.Errorf("case 0 type = %q, want unit", result[0].TestType)
	}
	if result[1].TestType != TypeE2E {
		t.Errorf("case 1 type = %q, want e2e", result[1].TestType)
	}
	if len(result[0].TestTypeEvidence) == 0 {
		t.Error("case 0 should have evidence")
	}
}
