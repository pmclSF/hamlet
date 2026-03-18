package models

// FixtureKind classifies the type of test fixture.
type FixtureKind string

const (
	// FixtureSetupHook is a lifecycle hook (beforeEach, setUp, TestMain).
	FixtureSetupHook FixtureKind = "setup_hook"

	// FixtureTeardownHook is a lifecycle hook (afterEach, tearDown).
	FixtureTeardownHook FixtureKind = "teardown_hook"

	// FixtureBuilder is a helper that constructs test objects (factories, builders).
	FixtureBuilder FixtureKind = "builder"

	// FixtureMockProvider creates mock/stub/fake services for tests.
	FixtureMockProvider FixtureKind = "mock_provider"

	// FixtureDataLoader loads test datasets or seed data.
	FixtureDataLoader FixtureKind = "data_loader"

	// FixtureHelper is a general test helper function.
	FixtureHelper FixtureKind = "helper"
)

// FixtureSurface represents a shared test fixture detected in test code.
// Fixtures are the shared infrastructure that tests depend on — setup hooks,
// builders, mock providers, and data loaders. High-fanout fixtures (used by
// many tests) are fragility hotspots: a single change can break many tests.
type FixtureSurface struct {
	// FixtureID is a deterministic stable identifier.
	// Format: "fixture:<path>:<name>" or "fixture:<path>:<scope>.<name>".
	FixtureID string `json:"fixtureId"`

	// Name is the fixture identifier (function name, hook type, variable name).
	Name string `json:"name"`

	// Path is the repository-relative file path containing this fixture.
	Path string `json:"path"`

	// Kind classifies the fixture type.
	Kind FixtureKind `json:"kind"`

	// Scope describes the fixture's lifecycle scope.
	// Values: "test" (per-test), "suite" (per-suite/describe), "module" (per-file),
	// "session" (per-run), "unknown".
	Scope string `json:"scope"`

	// Language is the programming language.
	Language string `json:"language"`

	// Framework is the test framework this fixture belongs to.
	Framework string `json:"framework,omitempty"`

	// Line is the source line where this fixture is defined.
	Line int `json:"line,omitempty"`

	// Shared indicates whether this fixture is used across multiple files.
	// Detected when the fixture is exported or defined in a conftest/setup file.
	Shared bool `json:"shared"`

	// DetectionTier records the inference method.
	DetectionTier string `json:"detectionTier,omitempty"`

	// Confidence is the detection confidence (0.0–1.0).
	Confidence float64 `json:"confidence,omitempty"`

	// Reason explains why this fixture was detected and classified.
	// Format: "[detectorID] description" for traceability.
	Reason string `json:"reason,omitempty"`
}

// Evidence returns a unified DetectionEvidence view.
func (fs *FixtureSurface) Evidence() DetectionEvidence {
	return DetectionEvidence{
		Tier:       fs.DetectionTier,
		Confidence: fs.Confidence,
		FilePath:   fs.Path,
		Symbol:     fs.Name,
		Line:       fs.Line,
		Reason:     fs.Reason,
	}
}

// BuildFixtureID constructs a deterministic fixture ID.
func BuildFixtureID(path, name, scope string) string {
	if scope != "" {
		return "fixture:" + path + ":" + scope + "." + name
	}
	return "fixture:" + path + ":" + name
}
