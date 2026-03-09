// Package ownership implements Hamlet's normalized ownership subsystem.
//
// Ownership is a routing layer, not a blame layer. It exists to make
// findings actionable by connecting risk, health, quality, and migration
// data to the people and teams who can act on it.
//
// The ownership model supports:
//   - single owner, multiple owners, and unknown/unowned
//   - confidence and evidence tracking
//   - inherited vs directly assigned ownership
//   - provenance (which source produced the assignment)
//   - consistent attachment to files, modules, test cases, code units, and insights
package ownership

import "strings"

// SourceType identifies where an ownership assignment came from.
type SourceType string

const (
	// SourceCodeowners is ownership derived from a CODEOWNERS file.
	SourceCodeowners SourceType = "codeowners"

	// SourceExplicitConfig is ownership from .hamlet/ownership.yaml.
	SourceExplicitConfig SourceType = "explicit_config"

	// SourcePackageMetadata is ownership from package.json, pom.xml, etc.
	SourcePackageMetadata SourceType = "package_metadata"

	// SourcePathMapping is ownership from path-prefix mapping rules.
	SourcePathMapping SourceType = "path_mapping"

	// SourceGitHistory is ownership inferred from recent git commit author history.
	// This source is opt-in and only used when configured.
	SourceGitHistory SourceType = "git_history"

	// SourceDirectoryFallback is ownership inferred from the top-level directory name.
	SourceDirectoryFallback SourceType = "directory_fallback"

	// SourceUnknown means no ownership source matched.
	SourceUnknown SourceType = "unknown"
)

// Confidence describes how trustworthy an ownership assignment is.
type Confidence string

const (
	// ConfidenceHigh means the assignment comes from an explicit, maintained source.
	ConfidenceHigh Confidence = "high"

	// ConfidenceMedium means the assignment is reasonable but inferred.
	ConfidenceMedium Confidence = "medium"

	// ConfidenceLow means the assignment is a weak heuristic or fallback.
	ConfidenceLow Confidence = "low"

	// ConfidenceNone means no ownership could be determined.
	ConfidenceNone Confidence = "none"
)

// InheritanceKind describes whether ownership was directly assigned or inherited.
type InheritanceKind string

const (
	// InheritanceDirect means the entity was directly matched by an ownership rule.
	InheritanceDirect InheritanceKind = "direct"

	// InheritanceInherited means the entity inherited ownership from its parent
	// (e.g., a code unit inherits from its file, a test case from its test file).
	InheritanceInherited InheritanceKind = "inherited"
)

// EntityType identifies what kind of entity an ownership assignment applies to.
type EntityType string

const (
	EntityFile     EntityType = "file"
	EntityModule   EntityType = "module"
	EntityTestCase EntityType = "test_case"
	EntityCodeUnit EntityType = "code_unit"
	EntityHotspot  EntityType = "hotspot"
	EntityInsight  EntityType = "insight"
)

// Owner represents a resolved owner identity.
type Owner struct {
	// ID is the canonical identifier (e.g., "@team-auth", "team-platform").
	// Stripped of leading @ for consistency.
	ID string `json:"id"`

	// DisplayName is a human-friendly label. Defaults to ID if not set.
	DisplayName string `json:"displayName,omitempty"`
}

// OwnershipAssignment connects an entity to one or more owners with
// provenance, confidence, and inheritance metadata.
type OwnershipAssignment struct {
	// Owners are the resolved owners for this entity.
	// Multiple owners are supported (shared ownership).
	// Empty means unowned.
	Owners []Owner `json:"owners"`

	// Source identifies where this assignment came from.
	Source SourceType `json:"source"`

	// Confidence describes how trustworthy this assignment is.
	Confidence Confidence `json:"confidence"`

	// Inheritance indicates whether this was directly assigned or inherited.
	Inheritance InheritanceKind `json:"inheritance"`

	// MatchedRule is the rule or pattern that produced this assignment.
	// For CODEOWNERS: the pattern line. For explicit config: the path rule.
	// Empty when source is unknown or directory fallback.
	MatchedRule string `json:"matchedRule,omitempty"`

	// SourceFile is the file that contained the ownership rule.
	// E.g., ".github/CODEOWNERS" or ".hamlet/ownership.yaml".
	SourceFile string `json:"sourceFile,omitempty"`
}

// OwnedEntityRef identifies an entity that has ownership attached.
type OwnedEntityRef struct {
	// Type is the kind of entity.
	Type EntityType `json:"type"`

	// ID is the entity's unique identifier (file path, unit ID, test ID, etc.).
	ID string `json:"id"`

	// Assignment is the ownership assignment for this entity.
	Assignment OwnershipAssignment `json:"assignment"`
}

// OwnerAggregate summarizes ownership statistics for a single owner
// across all entity types.
type OwnerAggregate struct {
	// Owner is the aggregated owner.
	Owner Owner `json:"owner"`

	// FileCount is the number of files owned.
	FileCount int `json:"fileCount"`

	// TestCaseCount is the number of test cases owned (directly or inherited).
	TestCaseCount int `json:"testCaseCount"`

	// CodeUnitCount is the number of code units owned (directly or inherited).
	CodeUnitCount int `json:"codeUnitCount"`

	// ExportedCodeUnitCount is the number of exported code units owned.
	ExportedCodeUnitCount int `json:"exportedCodeUnitCount"`

	// SignalCount is the number of signals attributed to this owner.
	SignalCount int `json:"signalCount"`

	// CriticalSignalCount is the number of critical-severity signals.
	CriticalSignalCount int `json:"criticalSignalCount"`

	// HealthSignalCount is the number of health-category signals.
	HealthSignalCount int `json:"healthSignalCount"`

	// MigrationBlockerCount is the number of migration signals.
	MigrationBlockerCount int `json:"migrationBlockerCount"`

	// UncoveredExportedCount is the number of uncovered exported code units.
	UncoveredExportedCount int `json:"uncoveredExportedCount"`
}

// OwnershipSummary is a snapshot-level overview of ownership coverage.
type OwnershipSummary struct {
	// TotalFiles is the total number of files in scope.
	TotalFiles int `json:"totalFiles"`

	// OwnedFiles is the number of files with at least one owner.
	OwnedFiles int `json:"ownedFiles"`

	// UnownedFiles is the number of files with no owner.
	UnownedFiles int `json:"unownedFiles"`

	// TotalCodeUnits is the total number of code units.
	TotalCodeUnits int `json:"totalCodeUnits"`

	// OwnedCodeUnits is the number of code units with ownership.
	OwnedCodeUnits int `json:"ownedCodeUnits"`

	// TotalTestCases is the total number of test cases.
	TotalTestCases int `json:"totalTestCases"`

	// OwnedTestCases is the number of test cases with ownership.
	OwnedTestCases int `json:"ownedTestCases"`

	// OwnerCount is the number of distinct owners.
	OwnerCount int `json:"ownerCount"`

	// CoveragePosture is the qualitative ownership coverage posture.
	// Values: "strong", "partial", "weak", "none".
	CoveragePosture string `json:"coveragePosture"`

	// Sources lists which ownership sources were used.
	Sources []SourceType `json:"sources,omitempty"`

	// Owners is the per-owner aggregate breakdown.
	Owners []OwnerAggregate `json:"owners,omitempty"`

	// Diagnostics holds any warnings or issues from ownership resolution.
	Diagnostics []Diagnostic `json:"diagnostics,omitempty"`
}

// Diagnostic represents a warning or issue encountered during ownership resolution.
type Diagnostic struct {
	// Level is "warning" or "info".
	Level string `json:"level"`

	// Message describes the issue.
	Message string `json:"message"`

	// Source identifies where the issue was found.
	Source string `json:"source,omitempty"`

	// Line is the line number in the source file, if applicable.
	Line int `json:"line,omitempty"`
}

// IsUnowned returns true if the assignment has no owners.
func (a *OwnershipAssignment) IsUnowned() bool {
	return len(a.Owners) == 0
}

// PrimaryOwner returns the first owner, or an empty Owner if unowned.
func (a *OwnershipAssignment) PrimaryOwner() Owner {
	if len(a.Owners) == 0 {
		return Owner{ID: unknownOwner}
	}
	return a.Owners[0]
}

// PrimaryOwnerID returns the primary owner's ID string.
// Returns "unknown" if unowned.
func (a *OwnershipAssignment) PrimaryOwnerID() string {
	return a.PrimaryOwner().ID
}

// HasOwner returns true if the given owner ID is among the assigned owners.
func (a *OwnershipAssignment) HasOwner(ownerID string) bool {
	for _, o := range a.Owners {
		if o.ID == ownerID {
			return true
		}
	}
	return false
}

// NormalizeOwnerID strips leading @ and whitespace from an owner identifier.
func NormalizeOwnerID(raw string) string {
	id := strings.TrimSpace(raw)
	id = strings.TrimPrefix(id, "@")
	return strings.TrimSpace(id)
}

// SourceConfidence returns the default confidence for a given source type.
func SourceConfidence(source SourceType) Confidence {
	switch source {
	case SourceExplicitConfig:
		return ConfidenceHigh
	case SourceCodeowners:
		return ConfidenceHigh
	case SourcePackageMetadata:
		return ConfidenceMedium
	case SourcePathMapping:
		return ConfidenceMedium
	case SourceGitHistory:
		return ConfidenceLow
	case SourceDirectoryFallback:
		return ConfidenceLow
	default:
		return ConfidenceNone
	}
}
