package models

// No external imports — Signal is now in the models package.

// TestFile represents a discovered test file and the key facts Terrain knows
// about it.
//
// Over time, this model will become one of the central objects in the product.
// It connects:
//   - structure (where the file lives, what framework it uses)
//   - quality (assertions, mocks, snapshots)
//   - runtime behavior (duration, retry rate)
//   - intelligence (signals attached to the file)
type TestFile struct {
	// Path is the repository-relative path to the test file.
	Path string `json:"path"`

	// Framework is the primary framework Terrain believes this file uses.
	Framework string `json:"framework,omitempty"`

	// FrameworkConfidence is the detection confidence (0.0–1.0).
	FrameworkConfidence float64 `json:"frameworkConfidence,omitempty"`

	// FrameworkSource describes how the framework was detected.
	// Values: "import", "config-file", "project-fallback", "convention".
	FrameworkSource string `json:"frameworkSource,omitempty"`

	// Owner is the resolved owner for this file if known.
	// This may come from CODEOWNERS, config, or future ownership inference.
	Owner string `json:"owner,omitempty"`

	// TestCount is the estimated number of tests in the file.
	TestCount int `json:"testCount"`

	// AssertionCount is the estimated number of assertions in the file.
	AssertionCount int `json:"assertionCount"`

	// MockCount is the estimated number of mocks or mock-like constructs in the file.
	MockCount int `json:"mockCount"`

	// SnapshotCount is the estimated number of snapshot assertions or snapshot artifacts used.
	SnapshotCount int `json:"snapshotCount"`

	// RuntimeStats contains runtime evidence for this file when available.
	RuntimeStats *RuntimeStats `json:"runtimeStats,omitempty"`

	// LinkedCodeUnits identifies code units this test file is believed to exercise.
	LinkedCodeUnits []string `json:"linkedCodeUnits,omitempty"`

	// EnvironmentIDs lists the environments this test file is known to execute in.
	// Inferred from CI config matrices, Docker configs, or manual annotation.
	// Format: "env:<canonical-name>" matching Environment.EnvironmentID.
	EnvironmentIDs []string `json:"environmentIds,omitempty"`

	// DeviceIDs lists the devices or browsers this test file targets.
	// Inferred from Playwright configs, BrowserStack, Xcode schemes, etc.
	// Format: "device:<canonical-name>" matching DeviceConfig.DeviceID.
	DeviceIDs []string `json:"deviceIds,omitempty"`

	// Signals contains signal identifiers or full signal objects associated with this file.
	Signals []Signal `json:"signals,omitempty"`
}
