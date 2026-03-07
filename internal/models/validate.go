package models

import (
	"fmt"
	"strings"
)

// ValidationError collects multiple invariant violations.
type ValidationError struct {
	Errors []string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%d validation errors: %s", len(e.Errors), strings.Join(e.Errors, "; "))
}

func (e *ValidationError) add(msg string) {
	e.Errors = append(e.Errors, msg)
}

func (e *ValidationError) addf(format string, args ...any) {
	e.Errors = append(e.Errors, fmt.Sprintf(format, args...))
}

func (e *ValidationError) err() error {
	if len(e.Errors) == 0 {
		return nil
	}
	return e
}

// ValidateSnapshot checks structural invariants on a TestSuiteSnapshot.
// Returns nil if all invariants hold.
func ValidateSnapshot(snap *TestSuiteSnapshot) error {
	if snap == nil {
		return &ValidationError{Errors: []string{"snapshot is nil"}}
	}

	ve := &ValidationError{}

	// Repository name must be non-empty.
	if snap.Repository.Name == "" {
		ve.add("repository name is empty")
	}

	// Schema version must be set.
	if snap.SnapshotMeta.SchemaVersion == "" {
		ve.add("snapshot schema version is empty")
	}

	// All test files must have non-empty paths.
	for i, tf := range snap.TestFiles {
		if tf.Path == "" {
			ve.addf("testFiles[%d] has empty path", i)
		}
	}

	// All signals must have type, category, and severity.
	for i, s := range snap.Signals {
		ValidateSignalInto(s, i, ve)
	}

	// All test cases must have non-empty TestID and FilePath.
	for i, tc := range snap.TestCases {
		if tc.TestID == "" {
			ve.addf("testCases[%d] has empty TestID", i)
		}
		if tc.FilePath == "" {
			ve.addf("testCases[%d] has empty FilePath", i)
		}
	}

	// All code units must have non-empty Name and Path.
	for i, cu := range snap.CodeUnits {
		if cu.Name == "" {
			ve.addf("codeUnits[%d] has empty Name", i)
		}
		if cu.Path == "" {
			ve.addf("codeUnits[%d] has empty Path", i)
		}
	}

	// Risk surfaces must have Type and Scope.
	for i, r := range snap.Risk {
		if r.Type == "" {
			ve.addf("risk[%d] has empty Type", i)
		}
		if r.Scope == "" {
			ve.addf("risk[%d] has empty Scope", i)
		}
	}

	// GeneratedAt must be set.
	if snap.GeneratedAt.IsZero() {
		ve.add("generatedAt is zero")
	}

	return ve.err()
}

// ValidateSignal checks structural invariants on a single Signal.
func ValidateSignal(s Signal) error {
	ve := &ValidationError{}
	ValidateSignalInto(s, 0, ve)
	return ve.err()
}

// ValidateSignalInto checks signal invariants and appends violations to ve.
func ValidateSignalInto(s Signal, idx int, ve *ValidationError) {
	if s.Type == "" {
		ve.addf("signals[%d] has empty Type", idx)
	}
	if s.Category == "" {
		ve.addf("signals[%d] has empty Category", idx)
	}
	if s.Severity == "" {
		ve.addf("signals[%d] has empty Severity", idx)
	}
	if s.Explanation == "" {
		ve.addf("signals[%d] (%s) has empty Explanation", idx, s.Type)
	}

	// Confidence, if set, must be in [0, 1].
	if s.Confidence < 0 || s.Confidence > 1 {
		ve.addf("signals[%d] (%s) has confidence %f outside [0,1]", idx, s.Type, s.Confidence)
	}

	// Valid severity values.
	switch s.Severity {
	case SeverityInfo, SeverityLow, SeverityMedium, SeverityHigh, SeverityCritical, "":
		// ok
	default:
		ve.addf("signals[%d] (%s) has invalid severity %q", idx, s.Type, s.Severity)
	}

	// Valid category values.
	switch s.Category {
	case CategoryStructure, CategoryHealth, CategoryQuality, CategoryMigration, CategoryGovernance, "":
		// ok
	default:
		ve.addf("signals[%d] (%s) has invalid category %q", idx, s.Type, s.Category)
	}
}
