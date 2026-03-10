package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSnapshotJSONRoundTrip(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 3, 6, 12, 0, 0, 0, time.UTC)

	snapshot := TestSuiteSnapshot{
		Repository: RepositoryMetadata{
			Name:              "example-repo",
			RootPath:          "/workspace/example-repo",
			Languages:         []string{"javascript", "typescript"},
			PackageManagers:   []string{"npm"},
			CISystems:         []string{"github-actions"},
			SnapshotTimestamp: now,
			CommitSHA:         "abc123",
			Branch:            "main",
		},
		Frameworks: []Framework{
			{
				Name:      "jest",
				Version:   "29.7.0",
				Type:      FrameworkTypeUnit,
				FileCount: 42,
				TestCount: 185,
			},
		},
		TestFiles: []TestFile{
			{
				Path:           "src/__tests__/auth.test.js",
				Framework:      "jest",
				TestCount:      8,
				AssertionCount: 12,
				MockCount:      2,
			},
		},
		CodeUnits: []CodeUnit{
			{
				Name:     "authenticate",
				Path:     "src/auth.js",
				Kind:     CodeUnitKindFunction,
				Exported: true,
			},
		},
		Signals: []Signal{
			{
				Type:        "weakAssertion",
				Category:    CategoryQuality,
				Severity:    SeverityMedium,
				Confidence:  0.8,
				Explanation: "Low assertion density in auth.test.js",
				Location: SignalLocation{
					File: "src/__tests__/auth.test.js",
				},
			},
		},
		Risk: []RiskSurface{
			{
				Type:      "change",
				Scope:     "file",
				ScopeName: "src/auth.js",
				Band:      RiskBandMedium,
				Score:     0.6,
			},
		},
		GeneratedAt: now,
	}

	data, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded TestSuiteSnapshot
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if decoded.Repository.Name != "example-repo" {
		t.Errorf("Repository.Name = %q, want %q", decoded.Repository.Name, "example-repo")
	}
	if len(decoded.Frameworks) != 1 || decoded.Frameworks[0].Name != "jest" {
		t.Errorf("Frameworks mismatch: got %+v", decoded.Frameworks)
	}
	if len(decoded.TestFiles) != 1 || decoded.TestFiles[0].Path != "src/__tests__/auth.test.js" {
		t.Errorf("TestFiles mismatch: got %+v", decoded.TestFiles)
	}
	if len(decoded.CodeUnits) != 1 || decoded.CodeUnits[0].Name != "authenticate" {
		t.Errorf("CodeUnits mismatch: got %+v", decoded.CodeUnits)
	}
	if len(decoded.Signals) != 1 || decoded.Signals[0].Type != "weakAssertion" {
		t.Errorf("Signals mismatch: got %+v", decoded.Signals)
	}
	if len(decoded.Risk) != 1 || decoded.Risk[0].Band != RiskBandMedium {
		t.Errorf("Risk mismatch: got %+v", decoded.Risk)
	}
	if !decoded.GeneratedAt.Equal(now) {
		t.Errorf("GeneratedAt = %v, want %v", decoded.GeneratedAt, now)
	}
}

func TestSnapshotJSONFieldNames(t *testing.T) {
	t.Parallel()
	snapshot := TestSuiteSnapshot{
		Repository: RepositoryMetadata{
			Name:              "test",
			SnapshotTimestamp: time.Now(),
		},
		GeneratedAt: time.Now(),
	}

	data, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("json.Unmarshal to map failed: %v", err)
	}

	requiredKeys := []string{"repository", "generatedAt"}
	for _, key := range requiredKeys {
		if _, ok := raw[key]; !ok {
			t.Errorf("JSON output missing required key %q", key)
		}
	}

	optionalKeys := []string{"frameworks", "testFiles", "codeUnits", "signals", "risk", "ownership", "policies", "metadata"}
	for _, key := range optionalKeys {
		if _, ok := raw[key]; ok {
			t.Errorf("JSON output should omit empty optional key %q", key)
		}
	}
}
