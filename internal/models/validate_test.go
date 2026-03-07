package models

import (
	"testing"
	"time"
)

func TestValidateSnapshot_Nil(t *testing.T) {
	err := ValidateSnapshot(nil)
	if err == nil {
		t.Error("expected error for nil snapshot")
	}
}

func TestValidateSnapshot_Valid(t *testing.T) {
	snap := &TestSuiteSnapshot{
		SnapshotMeta: SnapshotMeta{SchemaVersion: "1.0.0"},
		Repository:   RepositoryMetadata{Name: "test-repo"},
		TestFiles:    []TestFile{{Path: "a.test.js"}},
		Signals: []Signal{{
			Type:        "quality.weak-assertion",
			Category:    CategoryQuality,
			Severity:    SeverityMedium,
			Explanation: "test has no assertions",
		}},
		GeneratedAt: time.Now(),
	}
	err := ValidateSnapshot(snap)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateSnapshot_MissingRepoName(t *testing.T) {
	snap := &TestSuiteSnapshot{
		SnapshotMeta: SnapshotMeta{SchemaVersion: "1.0.0"},
		GeneratedAt:  time.Now(),
	}
	err := ValidateSnapshot(snap)
	if err == nil {
		t.Fatal("expected error for missing repo name")
	}
	ve := err.(*ValidationError)
	if len(ve.Errors) == 0 {
		t.Error("expected validation errors")
	}
}

func TestValidateSnapshot_EmptyTestFilePath(t *testing.T) {
	snap := &TestSuiteSnapshot{
		SnapshotMeta: SnapshotMeta{SchemaVersion: "1.0.0"},
		Repository:   RepositoryMetadata{Name: "test-repo"},
		TestFiles:    []TestFile{{Path: ""}},
		GeneratedAt:  time.Now(),
	}
	err := ValidateSnapshot(snap)
	if err == nil {
		t.Fatal("expected error for empty test file path")
	}
}

func TestValidateSignal_Valid(t *testing.T) {
	s := Signal{
		Type:        "quality.weak-assertion",
		Category:    CategoryQuality,
		Severity:    SeverityMedium,
		Explanation: "no assertions found",
		Confidence:  0.8,
	}
	if err := ValidateSignal(s); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateSignal_EmptyType(t *testing.T) {
	s := Signal{
		Category:    CategoryQuality,
		Severity:    SeverityMedium,
		Explanation: "no assertions found",
	}
	if err := ValidateSignal(s); err == nil {
		t.Error("expected error for empty Type")
	}
}

func TestValidateSignal_InvalidConfidence(t *testing.T) {
	s := Signal{
		Type:        "quality.weak-assertion",
		Category:    CategoryQuality,
		Severity:    SeverityMedium,
		Explanation: "test",
		Confidence:  1.5,
	}
	if err := ValidateSignal(s); err == nil {
		t.Error("expected error for confidence > 1")
	}
}

func TestValidateSignal_InvalidSeverity(t *testing.T) {
	s := Signal{
		Type:        "quality.weak-assertion",
		Category:    CategoryQuality,
		Severity:    "extreme",
		Explanation: "test",
	}
	if err := ValidateSignal(s); err == nil {
		t.Error("expected error for invalid severity")
	}
}

func TestValidateSignal_InvalidCategory(t *testing.T) {
	s := Signal{
		Type:        "quality.weak-assertion",
		Category:    "unknown",
		Severity:    SeverityMedium,
		Explanation: "test",
	}
	if err := ValidateSignal(s); err == nil {
		t.Error("expected error for invalid category")
	}
}

func TestValidateSnapshot_MultipleErrors(t *testing.T) {
	snap := &TestSuiteSnapshot{
		// Missing: repo name, schema version, generatedAt
		Signals: []Signal{
			{Type: "", Category: "", Severity: "", Explanation: ""},
		},
	}
	err := ValidateSnapshot(snap)
	if err == nil {
		t.Fatal("expected validation errors")
	}
	ve := err.(*ValidationError)
	// Should have: repo name, schema version, signal type, signal category, signal severity, signal explanation, generatedAt
	if len(ve.Errors) < 5 {
		t.Errorf("expected at least 5 errors, got %d: %v", len(ve.Errors), ve.Errors)
	}
}
