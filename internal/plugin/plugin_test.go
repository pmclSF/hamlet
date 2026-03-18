package plugin

import (
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

// --- Test fixtures ---

type mockFrameworkDetector struct {
	name       string
	frameworks []DetectedFramework
}

func (m *mockFrameworkDetector) Name() string          { return m.name }
func (m *mockFrameworkDetector) Detect(string) []DetectedFramework { return m.frameworks }

type mockScenarioDeriver struct {
	name      string
	scenarios []models.Scenario
}

func (m *mockScenarioDeriver) Name() string { return m.name }
func (m *mockScenarioDeriver) DeriveScenarios(string, []models.CodeSurface, []models.TestFile) []models.Scenario {
	return m.scenarios
}

type mockSignalClassifier struct {
	name           string
	failureMap     map[string]models.SignalType
	regressionMap  map[string]models.SignalType
}

func (m *mockSignalClassifier) Name() string { return m.name }
func (m *mockSignalClassifier) ClassifyFailure(name string) models.SignalType {
	return m.failureMap[name]
}
func (m *mockSignalClassifier) ClassifyRegression(metric string) models.SignalType {
	return m.regressionMap[metric]
}

type mockPolicyRule struct {
	name    string
	signals []models.Signal
}

func (m *mockPolicyRule) Name() string { return m.name }
func (m *mockPolicyRule) Evaluate(*models.TestSuiteSnapshot) []models.Signal {
	return m.signals
}

// --- Tests ---

func TestRegistry_Empty(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	stats := r.Stats()
	for k, v := range stats {
		if v != 0 {
			t.Errorf("%s = %d, want 0", k, v)
		}
	}
}

func TestRegistry_FrameworkDetector(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	r.RegisterFrameworkDetector(&mockFrameworkDetector{
		name: "custom-eval",
		frameworks: []DetectedFramework{
			{Name: "custom-eval", Source: "config", Confidence: 0.9},
		},
	})

	results := r.DetectFrameworks("/tmp")
	if len(results) != 1 {
		t.Fatalf("expected 1 framework, got %d", len(results))
	}
	if results[0].Name != "custom-eval" {
		t.Errorf("framework name = %q", results[0].Name)
	}
}

func TestRegistry_ScenarioDeriver(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	r.RegisterScenarioDeriver(&mockScenarioDeriver{
		name: "custom",
		scenarios: []models.Scenario{
			{ScenarioID: "sc:custom:1", Name: "custom-safety"},
		},
	})

	results := r.DeriveScenarios("/tmp", nil, nil)
	if len(results) != 1 {
		t.Fatalf("expected 1 scenario, got %d", len(results))
	}
}

func TestRegistry_SignalClassifier(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	r.RegisterSignalClassifier(&mockSignalClassifier{
		name: "domain-specific",
		failureMap:    map[string]models.SignalType{"custom-safety": "safetyFailure"},
		regressionMap: map[string]models.SignalType{"custom_accuracy": "accuracyRegression"},
	})

	if got := r.ClassifyFailure("custom-safety"); got != "safetyFailure" {
		t.Errorf("failure = %q, want safetyFailure", got)
	}
	if got := r.ClassifyFailure("unknown"); got != "" {
		t.Errorf("expected empty for unknown, got %q", got)
	}
	if got := r.ClassifyRegression("custom_accuracy"); got != "accuracyRegression" {
		t.Errorf("regression = %q", got)
	}
}

func TestRegistry_SignalClassifier_Priority(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	// First classifier wins.
	r.RegisterSignalClassifier(&mockSignalClassifier{
		name:       "high-priority",
		failureMap: map[string]models.SignalType{"safety": "safetyFailure"},
	})
	r.RegisterSignalClassifier(&mockSignalClassifier{
		name:       "low-priority",
		failureMap: map[string]models.SignalType{"safety": "evalFailure"}, // should NOT win
	})

	if got := r.ClassifyFailure("safety"); got != "safetyFailure" {
		t.Errorf("expected first classifier to win, got %q", got)
	}
}

func TestRegistry_PolicyRule(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	r.RegisterPolicyRule(&mockPolicyRule{
		name: "custom-rule",
		signals: []models.Signal{
			{Type: "policyViolation", Severity: models.SeverityHigh, Explanation: "custom violation"},
		},
	})

	signals := r.EvaluatePolicies(&models.TestSuiteSnapshot{})
	if len(signals) != 1 {
		t.Fatalf("expected 1 signal, got %d", len(signals))
	}
	if signals[0].Explanation != "custom violation" {
		t.Errorf("explanation = %q", signals[0].Explanation)
	}
}

func TestRegistry_MultiplePlugins(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	r.RegisterFrameworkDetector(&mockFrameworkDetector{name: "a"})
	r.RegisterFrameworkDetector(&mockFrameworkDetector{name: "b"})
	r.RegisterScenarioDeriver(&mockScenarioDeriver{name: "c"})
	r.RegisterSignalClassifier(&mockSignalClassifier{name: "d"})
	r.RegisterPolicyRule(&mockPolicyRule{name: "e"})

	stats := r.Stats()
	if stats["frameworkDetectors"] != 2 {
		t.Errorf("frameworkDetectors = %d", stats["frameworkDetectors"])
	}
	if stats["scenarioDeriver"] != 1 {
		t.Errorf("scenarioDeriver = %d", stats["scenarioDeriver"])
	}
	if stats["signalClassifiers"] != 1 {
		t.Errorf("signalClassifiers = %d", stats["signalClassifiers"])
	}
	if stats["policyRules"] != 1 {
		t.Errorf("policyRules = %d", stats["policyRules"])
	}
}

func TestRegistry_Deterministic(t *testing.T) {
	t.Parallel()
	makeRegistry := func() *Registry {
		r := NewRegistry()
		r.RegisterSignalClassifier(&mockSignalClassifier{
			name:       "a",
			failureMap: map[string]models.SignalType{"x": "safetyFailure"},
		})
		r.RegisterSignalClassifier(&mockSignalClassifier{
			name:       "b",
			failureMap: map[string]models.SignalType{"x": "evalFailure"},
		})
		return r
	}

	r1 := makeRegistry()
	r2 := makeRegistry()
	t1 := r1.ClassifyFailure("x")
	t2 := r2.ClassifyFailure("x")
	if t1 != t2 {
		t.Errorf("non-deterministic: %q vs %q", t1, t2)
	}
}
