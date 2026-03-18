// Package plugin defines the extension point interfaces for Terrain.
//
// Plugins allow new frameworks, detectors, scenario derivers, signal
// classifiers, and policy rules to be added without modifying core code.
//
// Extension points:
//
//   - FrameworkDetector: detect AI/eval frameworks in a repository
//   - ScenarioDeriver: derive scenarios from detected frameworks
//   - SignalClassifier: classify Gauntlet metric names to signal types
//   - PolicyRule: custom CI policy evaluation
//   - SurfaceExtractor: (existing) language-specific surface extraction
//   - Detector: (existing) signal detection from snapshot
//
// All plugin interfaces are designed to be:
//   - stateless
//   - deterministic
//   - safe for concurrent use
//   - fail-open (errors are logged, not fatal)
package plugin

import "github.com/pmclSF/terrain/internal/models"

// FrameworkDetector detects AI/eval frameworks in a repository.
// Implementations examine config files, dependency manifests, or source imports.
type FrameworkDetector interface {
	// Name returns a unique identifier for this detector (e.g., "promptfoo", "custom-eval").
	Name() string

	// Detect examines the repository and returns detected frameworks.
	Detect(root string) []DetectedFramework
}

// DetectedFramework represents a framework found by a FrameworkDetector.
type DetectedFramework struct {
	Name       string  `json:"name"`
	Source     string  `json:"source"`     // "config", "dependency", "import"
	ConfigFile string  `json:"configFile,omitempty"`
	Version    string  `json:"version,omitempty"`
	Confidence float64 `json:"confidence"`
}

// ScenarioDeriver derives eval scenarios from repository content.
// Implementations examine framework configs, test files, or code patterns.
type ScenarioDeriver interface {
	// Name returns a unique identifier for this deriver.
	Name() string

	// DeriveScenarios examines the repository and returns inferred scenarios.
	DeriveScenarios(root string, surfaces []models.CodeSurface, testFiles []models.TestFile) []models.Scenario
}

// SignalClassifier maps metric names or scenario names to specific signal types.
// Implementations provide domain-specific signal classification.
type SignalClassifier interface {
	// Name returns a unique identifier for this classifier.
	Name() string

	// ClassifyFailure maps a failed scenario name to a signal type.
	// Returns empty string if this classifier doesn't handle the scenario.
	ClassifyFailure(scenarioName string) models.SignalType

	// ClassifyRegression maps a regressed metric name to a signal type.
	// Returns empty string if this classifier doesn't handle the metric.
	ClassifyRegression(metricName string) models.SignalType
}

// PolicyRule is a custom CI policy evaluation rule.
type PolicyRule interface {
	// Name returns a unique identifier for this rule.
	Name() string

	// Evaluate checks the snapshot against the rule and returns violations.
	// Returns nil if the rule passes.
	Evaluate(snap *models.TestSuiteSnapshot) []models.Signal
}

// Registry holds registered plugins for all extension points.
type Registry struct {
	frameworkDetectors []FrameworkDetector
	scenarioDrivers    []ScenarioDeriver
	signalClassifiers  []SignalClassifier
	policyRules        []PolicyRule
}

// NewRegistry creates an empty plugin registry.
func NewRegistry() *Registry {
	return &Registry{}
}

// RegisterFrameworkDetector adds a framework detector plugin.
func (r *Registry) RegisterFrameworkDetector(d FrameworkDetector) {
	r.frameworkDetectors = append(r.frameworkDetectors, d)
}

// RegisterScenarioDeriver adds a scenario derivation plugin.
func (r *Registry) RegisterScenarioDeriver(d ScenarioDeriver) {
	r.scenarioDrivers = append(r.scenarioDrivers, d)
}

// RegisterSignalClassifier adds a signal classification plugin.
func (r *Registry) RegisterSignalClassifier(c SignalClassifier) {
	r.signalClassifiers = append(r.signalClassifiers, c)
}

// RegisterPolicyRule adds a custom policy rule plugin.
func (r *Registry) RegisterPolicyRule(p PolicyRule) {
	r.policyRules = append(r.policyRules, p)
}

// FrameworkDetectors returns all registered framework detectors.
func (r *Registry) FrameworkDetectors() []FrameworkDetector {
	return r.frameworkDetectors
}

// ScenarioDrivers returns all registered scenario derivers.
func (r *Registry) ScenarioDrivers() []ScenarioDeriver {
	return r.scenarioDrivers
}

// SignalClassifiers returns all registered signal classifiers.
func (r *Registry) SignalClassifiers() []SignalClassifier {
	return r.signalClassifiers
}

// PolicyRules returns all registered policy rules.
func (r *Registry) PolicyRules() []PolicyRule {
	return r.policyRules
}

// ClassifyFailure runs all registered classifiers and returns the first match.
// Falls back to empty string if no classifier handles the scenario.
func (r *Registry) ClassifyFailure(scenarioName string) models.SignalType {
	for _, c := range r.signalClassifiers {
		if t := c.ClassifyFailure(scenarioName); t != "" {
			return t
		}
	}
	return ""
}

// ClassifyRegression runs all registered classifiers and returns the first match.
func (r *Registry) ClassifyRegression(metricName string) models.SignalType {
	for _, c := range r.signalClassifiers {
		if t := c.ClassifyRegression(metricName); t != "" {
			return t
		}
	}
	return ""
}

// EvaluatePolicies runs all registered policy rules.
func (r *Registry) EvaluatePolicies(snap *models.TestSuiteSnapshot) []models.Signal {
	var signals []models.Signal
	for _, p := range r.policyRules {
		signals = append(signals, p.Evaluate(snap)...)
	}
	return signals
}

// DetectFrameworks runs all registered framework detectors.
func (r *Registry) DetectFrameworks(root string) []DetectedFramework {
	var frameworks []DetectedFramework
	for _, d := range r.frameworkDetectors {
		frameworks = append(frameworks, d.Detect(root)...)
	}
	return frameworks
}

// DeriveScenarios runs all registered scenario derivers.
func (r *Registry) DeriveScenarios(root string, surfaces []models.CodeSurface, testFiles []models.TestFile) []models.Scenario {
	var scenarios []models.Scenario
	for _, d := range r.scenarioDrivers {
		scenarios = append(scenarios, d.DeriveScenarios(root, surfaces, testFiles)...)
	}
	return scenarios
}

// Stats returns counts for all registered plugins.
func (r *Registry) Stats() map[string]int {
	return map[string]int{
		"frameworkDetectors": len(r.frameworkDetectors),
		"scenarioDeriver":    len(r.scenarioDrivers),
		"signalClassifiers":  len(r.signalClassifiers),
		"policyRules":        len(r.policyRules),
	}
}
