package models

// FrameworkType describes the broad category of testing framework.
//
// This is intentionally high-level. Hamlet should be able to reason about
// a repository using categories like "unit" or "e2e" even when the specific
// framework differs.
type FrameworkType string

const (
	FrameworkTypeUnit          FrameworkType = "unit"
	FrameworkTypeIntegration   FrameworkType = "integration"
	FrameworkTypeE2E           FrameworkType = "e2e"
	FrameworkTypePerformance   FrameworkType = "performance"
	FrameworkTypeVisual        FrameworkType = "visual"
	FrameworkTypeContract      FrameworkType = "contract"
	FrameworkTypePropertyBased FrameworkType = "property-based"
	FrameworkTypeUnknown       FrameworkType = "unknown"
)

// Framework represents a testing framework detected in the repository.
type Framework struct {
	// Name is the canonical framework name.
	// Examples: jest, vitest, playwright, cypress, pytest, junit.
	Name string `json:"name"`

	// Version is optional and may be unavailable during early analysis.
	Version string `json:"version,omitempty"`

	// Type is the broad category of this framework.
	Type FrameworkType `json:"type"`

	// FileCount is the number of test files associated with this framework.
	FileCount int `json:"fileCount"`

	// TestCount is the estimated number of test cases associated with this framework.
	TestCount int `json:"testCount"`
}
