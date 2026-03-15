// Package policy implements Terrain's local, repo-native policy system.
//
// A policy is loaded from .terrain/policy.yaml in the analyzed repository.
// It declares constraints that the repository's test suite should satisfy.
// Policy evaluation produces governance signals when violations are found.
//
// The policy system is intentionally simple and explicit:
//   - no magic defaults
//   - no central management
//   - no complex DSL
//
// Future stages may layer centralized or organization-wide policy on top
// of this local model, but the local file remains the ground truth for
// any single repository.
package policy

// Config represents a local Terrain policy loaded from .terrain/policy.yaml.
//
// All fields are pointers or zero-value-safe so that a partial policy
// file works correctly — only explicitly set rules are enforced.
type Config struct {
	Rules Rules `yaml:"rules"`
}

// Rules contains the individual policy rules.
//
// Each rule is a pointer type so we can distinguish "not set" from
// "set to zero/false". A nil pointer means the rule is not active.
type Rules struct {
	// DisallowSkippedTests, when true, flags any skipped/pending tests
	// as a policy violation.
	DisallowSkippedTests *bool `yaml:"disallow_skipped_tests"`

	// DisallowFrameworks lists framework names that are not permitted.
	// If the repository contains test files using any of these frameworks,
	// a legacyFrameworkUsage governance signal is emitted.
	DisallowFrameworks []string `yaml:"disallow_frameworks"`

	// MaxTestRuntimeMs sets the maximum allowed average test runtime in
	// milliseconds. If any test file's average runtime exceeds this,
	// a runtimeBudgetExceeded governance signal is emitted.
	MaxTestRuntimeMs *float64 `yaml:"max_test_runtime_ms"`

	// MinimumCoveragePercent sets the minimum required coverage percentage.
	// This is evaluated against existing coverage threshold signals.
	// Value should be 0-100 (e.g., 80 means 80%).
	MinimumCoveragePercent *float64 `yaml:"minimum_coverage_percent"`

	// MaxWeakAssertions sets the maximum allowed number of weakAssertion
	// signals before a policy violation is raised.
	MaxWeakAssertions *int `yaml:"max_weak_assertions"`

	// MaxMockHeavyTests sets the maximum allowed number of mockHeavyTest
	// signals before a policy violation is raised.
	MaxMockHeavyTests *int `yaml:"max_mock_heavy_tests"`
}

// IsEmpty returns true if no rules are configured.
func (c *Config) IsEmpty() bool {
	r := c.Rules
	return r.DisallowSkippedTests == nil &&
		len(r.DisallowFrameworks) == 0 &&
		r.MaxTestRuntimeMs == nil &&
		r.MinimumCoveragePercent == nil &&
		r.MaxWeakAssertions == nil &&
		r.MaxMockHeavyTests == nil
}
