package models

import "testing"

func TestTestCase_ImplementsValidationTarget(t *testing.T) {
	t.Parallel()
	var vt ValidationTarget = TestCase{
		TestID:   "test-123",
		TestName: "should login",
		FilePath: "test/auth.test.ts",
	}

	if vt.ValidationID() != "test-123" {
		t.Errorf("expected ID test-123, got %q", vt.ValidationID())
	}
	if vt.ValidationName() != "should login" {
		t.Errorf("expected name 'should login', got %q", vt.ValidationName())
	}
	if vt.ValidationKindOf() != ValidationKindTest {
		t.Errorf("expected kind test, got %q", vt.ValidationKindOf())
	}
	if vt.ValidationPath() != "test/auth.test.ts" {
		t.Errorf("expected path, got %q", vt.ValidationPath())
	}
	if !vt.IsExecutable() {
		t.Error("tests should be executable")
	}
}

func TestScenario_ImplementsValidationTarget(t *testing.T) {
	t.Parallel()
	var vt ValidationTarget = Scenario{
		ScenarioID: "scenario:auth:login-flow",
		Name:       "Login flow",
		Path:       "evals/auth.yaml",
		Owner:      "ml-team",
		Executable: true,
	}

	if vt.ValidationID() != "scenario:auth:login-flow" {
		t.Errorf("expected ID, got %q", vt.ValidationID())
	}
	if vt.ValidationKindOf() != ValidationKindScenario {
		t.Errorf("expected kind scenario, got %q", vt.ValidationKindOf())
	}
	if vt.ValidationOwner() != "ml-team" {
		t.Errorf("expected owner ml-team, got %q", vt.ValidationOwner())
	}
	if !vt.IsExecutable() {
		t.Error("executable scenario should be executable")
	}
}

func TestScenario_NonExecutable(t *testing.T) {
	t.Parallel()
	var vt ValidationTarget = Scenario{
		ScenarioID: "scenario:derived:checkout",
		Executable: false,
	}
	if vt.IsExecutable() {
		t.Error("non-executable scenario should not be executable")
	}
}

func TestManualCoverageArtifact_ImplementsValidationTarget(t *testing.T) {
	t.Parallel()
	var vt ValidationTarget = ManualCoverageArtifact{
		ArtifactID: "manual:testrail:login-suite",
		Name:       "Login regression suite",
		Source:     "testrail",
		Owner:      "qa-team",
	}

	if vt.ValidationID() != "manual:testrail:login-suite" {
		t.Errorf("expected ID, got %q", vt.ValidationID())
	}
	if vt.ValidationKindOf() != ValidationKindManual {
		t.Errorf("expected kind manual, got %q", vt.ValidationKindOf())
	}
	if vt.ValidationPath() != "" {
		t.Errorf("manual coverage should have no path, got %q", vt.ValidationPath())
	}
	if vt.IsExecutable() {
		t.Error("manual coverage should never be executable")
	}
}

func TestCollectValidationTargets(t *testing.T) {
	t.Parallel()
	snap := &TestSuiteSnapshot{
		TestCases: []TestCase{
			{TestID: "t1", TestName: "test one"},
			{TestID: "t2", TestName: "test two"},
		},
		Scenarios: []Scenario{
			{ScenarioID: "s1", Name: "scenario one", Executable: true},
		},
		ManualCoverage: []ManualCoverageArtifact{
			{ArtifactID: "m1", Name: "manual one"},
		},
	}

	targets := CollectValidationTargets(snap)

	if len(targets) != 4 {
		t.Fatalf("expected 4 targets, got %d", len(targets))
	}

	// Verify order: tests, then scenarios, then manual.
	if targets[0].ValidationKindOf() != ValidationKindTest {
		t.Error("first targets should be tests")
	}
	if targets[2].ValidationKindOf() != ValidationKindScenario {
		t.Error("third target should be scenario")
	}
	if targets[3].ValidationKindOf() != ValidationKindManual {
		t.Error("fourth target should be manual")
	}
}

func TestCollectValidationTargets_NilSnapshot(t *testing.T) {
	t.Parallel()
	targets := CollectValidationTargets(nil)
	if len(targets) != 0 {
		t.Errorf("expected 0 targets from nil snapshot, got %d", len(targets))
	}
}

func TestCollectValidationTargets_EmptySnapshot(t *testing.T) {
	t.Parallel()
	targets := CollectValidationTargets(&TestSuiteSnapshot{})
	if len(targets) != 0 {
		t.Errorf("expected 0 targets from empty snapshot, got %d", len(targets))
	}
}
