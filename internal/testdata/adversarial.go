package testdata

import (
	"fmt"

	"github.com/pmclSF/terrain/internal/models"
)

// MixedFrameworkSnapshot returns a repo with 5+ frameworks — a common
// adversarial case for migration readiness and framework detection.
func MixedFrameworkSnapshot() *models.TestSuiteSnapshot {
	return &models.TestSuiteSnapshot{
		SnapshotMeta: models.SnapshotMeta{SchemaVersion: models.SnapshotSchemaVersion},
		Repository: models.RepositoryMetadata{
			Name:      "mixed-frameworks",
			Languages: []string{"go", "java", "javascript", "python"},
		},
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 10, TestCount: 50},
			{Name: "vitest", Type: models.FrameworkTypeUnit, FileCount: 5, TestCount: 25},
			{Name: "pytest", Type: models.FrameworkTypeUnit, FileCount: 8, TestCount: 40},
			{Name: "go-testing", Type: models.FrameworkTypeUnit, FileCount: 12, TestCount: 60},
			{Name: "junit5", Type: models.FrameworkTypeUnit, FileCount: 6, TestCount: 30},
			{Name: "playwright", Type: models.FrameworkTypeE2E, FileCount: 3, TestCount: 15},
		},
		TestFiles: []models.TestFile{
			{Path: "js/auth.test.js", Framework: "jest", TestCount: 5, AssertionCount: 8},
			{Path: "js/user.test.ts", Framework: "vitest", TestCount: 5, AssertionCount: 6},
			{Path: "py/test_auth.py", Framework: "pytest", TestCount: 5, AssertionCount: 5},
			{Path: "go/auth_test.go", Framework: "go-testing", TestCount: 5, AssertionCount: 5},
			{Path: "java/AuthTest.java", Framework: "junit5", TestCount: 5, AssertionCount: 5},
			{Path: "e2e/login.spec.ts", Framework: "playwright", TestCount: 5, AssertionCount: 3},
		},
		CodeUnits: []models.CodeUnit{
			{UnitID: "js/auth.js:AuthService", Name: "AuthService", Path: "js/auth.js", Kind: models.CodeUnitKindClass, Exported: true, Language: "js"},
			{UnitID: "py/auth.py:authenticate", Name: "authenticate", Path: "py/auth.py", Kind: models.CodeUnitKindFunction, Exported: true, Language: "python"},
			{UnitID: "go/auth.go:Authenticate", Name: "Authenticate", Path: "go/auth.go", Kind: models.CodeUnitKindFunction, Exported: true, Language: "go"},
		},
		GeneratedAt: FixedTime,
	}
}

// ZeroSignalSnapshot returns a snapshot that should produce zero signals —
// every test file has strong assertions, no mocks, good coverage.
func ZeroSignalSnapshot() *models.TestSuiteSnapshot {
	return &models.TestSuiteSnapshot{
		SnapshotMeta: models.SnapshotMeta{SchemaVersion: models.SnapshotSchemaVersion},
		Repository: models.RepositoryMetadata{
			Name:      "zero-signals",
			Languages: []string{"javascript"},
		},
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 3, TestCount: 15},
		},
		TestFiles: []models.TestFile{
			{Path: "src/__tests__/auth.test.js", Framework: "jest", TestCount: 5, AssertionCount: 10, LinkedCodeUnits: []string{"AuthService"}},
			{Path: "src/__tests__/user.test.js", Framework: "jest", TestCount: 5, AssertionCount: 10, LinkedCodeUnits: []string{"UserService"}},
			{Path: "src/__tests__/api.test.js", Framework: "jest", TestCount: 5, AssertionCount: 10, LinkedCodeUnits: []string{"ApiClient"}},
		},
		CodeUnits: []models.CodeUnit{
			{UnitID: "src/auth.js:AuthService", Name: "AuthService", Path: "src/auth.js", Kind: models.CodeUnitKindClass, Exported: true},
			{UnitID: "src/user.js:UserService", Name: "UserService", Path: "src/user.js", Kind: models.CodeUnitKindClass, Exported: true},
			{UnitID: "src/api.js:ApiClient", Name: "ApiClient", Path: "src/api.js", Kind: models.CodeUnitKindClass, Exported: true},
		},
		GeneratedAt: FixedTime,
	}
}

// AllSignalTypesSnapshot returns a snapshot pre-loaded with one signal of
// every category and severity — useful for testing renderers and filters.
func AllSignalTypesSnapshot() *models.TestSuiteSnapshot {
	snap := &models.TestSuiteSnapshot{
		SnapshotMeta: models.SnapshotMeta{SchemaVersion: models.SnapshotSchemaVersion},
		Repository: models.RepositoryMetadata{
			Name:      "all-signal-types",
			Languages: []string{"javascript"},
		},
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 5, TestCount: 25},
		},
		TestFiles: []models.TestFile{
			{Path: "src/__tests__/app.test.js", Framework: "jest", TestCount: 5, AssertionCount: 2},
		},
		GeneratedAt: FixedTime,
	}

	categories := []models.SignalCategory{
		models.CategoryQuality, models.CategoryMigration, models.CategoryHealth,
		models.CategoryGovernance, models.CategoryStructure,
	}
	severities := []models.SignalSeverity{
		models.SeverityInfo, models.SeverityLow, models.SeverityMedium,
		models.SeverityHigh, models.SeverityCritical,
	}

	for _, cat := range categories {
		for _, sev := range severities {
			snap.Signals = append(snap.Signals, models.Signal{
				Type:        models.SignalType(fmt.Sprintf("%s.test-%s", cat, sev)),
				Category:    cat,
				Severity:    sev,
				Location:    models.SignalLocation{File: "src/__tests__/app.test.js"},
				Explanation: fmt.Sprintf("Test signal: %s/%s", cat, sev),
			})
		}
	}

	return snap
}

// DeepNestingSnapshot returns a repo with deeply nested directory structures.
func DeepNestingSnapshot() *models.TestSuiteSnapshot {
	snap := &models.TestSuiteSnapshot{
		SnapshotMeta: models.SnapshotMeta{SchemaVersion: models.SnapshotSchemaVersion},
		Repository: models.RepositoryMetadata{
			Name:      "deep-nesting",
			Languages: []string{"javascript"},
		},
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 10, TestCount: 50},
		},
		GeneratedAt: FixedTime,
	}

	for i := 0; i < 10; i++ {
		path := ""
		for d := 0; d <= i; d++ {
			path += fmt.Sprintf("level%d/", d)
		}
		path += fmt.Sprintf("module%d.test.js", i)
		snap.TestFiles = append(snap.TestFiles, models.TestFile{
			Path:           path,
			Framework:      "jest",
			TestCount:      5,
			AssertionCount: 3,
		})
	}

	return snap
}

// OwnershipFragmentedSnapshot returns a repo with many distinct owners
// and gaps in ownership coverage — tests ownership resolution, propagation,
// and coordination risk.
func OwnershipFragmentedSnapshot() *models.TestSuiteSnapshot {
	snap := &models.TestSuiteSnapshot{
		SnapshotMeta: models.SnapshotMeta{SchemaVersion: models.SnapshotSchemaVersion},
		Repository: models.RepositoryMetadata{
			Name:      "ownership-fragmented",
			Languages: []string{"javascript"},
		},
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 10, TestCount: 50},
		},
		GeneratedAt: FixedTime,
	}

	teams := []string{"team-alpha", "team-beta", "team-gamma", "team-delta", "team-epsilon", "team-zeta", "team-eta", "team-theta"}
	snap.Ownership = map[string][]string{}
	for i := 0; i < 10; i++ {
		path := fmt.Sprintf("src/pkg%d/service.js", i)
		testPath := fmt.Sprintf("src/pkg%d/__tests__/service.test.js", i)
		snap.TestFiles = append(snap.TestFiles, models.TestFile{
			Path: testPath, Framework: "jest", TestCount: 5, AssertionCount: 4,
			LinkedCodeUnits: []string{fmt.Sprintf("Service%d", i)},
		})
		snap.CodeUnits = append(snap.CodeUnits, models.CodeUnit{
			UnitID: fmt.Sprintf("%s:Service%d", path, i), Name: fmt.Sprintf("Service%d", i),
			Path: path, Kind: models.CodeUnitKindClass, Exported: true,
		})
		if i < 8 {
			snap.Ownership[path] = []string{teams[i%len(teams)]}
		}
		// Last 2 files have no ownership — gap.
	}
	return snap
}

// SuppressionHeavySnapshot returns a repo that relies heavily on test
// suppression (skipped, disabled, or quarantined tests). Tests suppression
// detection and governance enforcement.
func SuppressionHeavySnapshot() *models.TestSuiteSnapshot {
	snap := &models.TestSuiteSnapshot{
		SnapshotMeta: models.SnapshotMeta{SchemaVersion: models.SnapshotSchemaVersion},
		Repository: models.RepositoryMetadata{
			Name:      "suppression-heavy",
			Languages: []string{"javascript"},
		},
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 6, TestCount: 30},
		},
		GeneratedAt: FixedTime,
	}

	for i := 0; i < 6; i++ {
		snap.TestFiles = append(snap.TestFiles, models.TestFile{
			Path:           fmt.Sprintf("src/__tests__/mod%d.test.js", i),
			Framework:      "jest",
			TestCount:      5,
			AssertionCount: 3,
		})
	}

	// Add suppression signals.
	for i := 0; i < 3; i++ {
		snap.Signals = append(snap.Signals, models.Signal{
			Type:     "suppressedTest",
			Category: models.CategoryHealth,
			Severity: models.SeverityMedium,
			Location: models.SignalLocation{File: fmt.Sprintf("src/__tests__/mod%d.test.js", i)},
		})
	}
	return snap
}

// ChangeScopedPRSnapshot returns a snapshot paired with a typical PR change
// scope — useful for impact analysis and test selection E2E tests.
func ChangeScopedPRSnapshot() (*models.TestSuiteSnapshot, []string) {
	snap := HealthyBalancedSnapshot()
	changedFiles := []string{
		"src/auth.js",
		"src/payment.js",
		"src/__tests__/auth.test.js",
	}
	return snap, changedFiles
}

// VeryLargeSnapshot returns a snapshot with 2000+ test files for
// performance and memory pressure testing.
func VeryLargeSnapshot() *models.TestSuiteSnapshot {
	snap := &models.TestSuiteSnapshot{
		SnapshotMeta: models.SnapshotMeta{SchemaVersion: models.SnapshotSchemaVersion},
		Repository: models.RepositoryMetadata{
			Name:      "very-large",
			Languages: []string{"javascript", "go", "python"},
		},
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 1500, TestCount: 7500},
			{Name: "go-testing", Type: models.FrameworkTypeUnit, FileCount: 300, TestCount: 1500},
			{Name: "pytest", Type: models.FrameworkTypeUnit, FileCount: 200, TestCount: 1000},
		},
		GeneratedAt: FixedTime,
	}

	for i := 0; i < 1500; i++ {
		snap.TestFiles = append(snap.TestFiles, models.TestFile{
			Path:           fmt.Sprintf("src/pkg%d/__tests__/mod%d.test.js", i/10, i),
			Framework:      "jest",
			TestCount:      5,
			AssertionCount: i % 5, // some with zero assertions
			MockCount:      i % 7, // some mock-heavy
		})
	}
	for i := 0; i < 300; i++ {
		snap.TestFiles = append(snap.TestFiles, models.TestFile{
			Path:           fmt.Sprintf("internal/pkg%d/handler%d_test.go", i/10, i),
			Framework:      "go-testing",
			TestCount:      5,
			AssertionCount: 5,
		})
	}
	for i := 0; i < 200; i++ {
		snap.TestFiles = append(snap.TestFiles, models.TestFile{
			Path:           fmt.Sprintf("tests/test_module%d.py", i),
			Framework:      "pytest",
			TestCount:      5,
			AssertionCount: 4,
		})
	}

	for i := 0; i < 500; i++ {
		snap.CodeUnits = append(snap.CodeUnits, models.CodeUnit{
			UnitID:   fmt.Sprintf("src/pkg%d/mod%d.js:Export%d", i/10, i, i),
			Name:     fmt.Sprintf("Export%d", i),
			Path:     fmt.Sprintf("src/pkg%d/mod%d.js", i/10, i),
			Kind:     models.CodeUnitKindFunction,
			Exported: true,
		})
	}

	return snap
}
