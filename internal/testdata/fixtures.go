// Package testdata provides standardized test fixtures for Terrain's test suite.
//
// These fixtures represent common real-world scenarios and are reusable across
// unit, integration, golden, and E2E tests.
package testdata

import (
	"fmt"
	"time"

	"github.com/pmclSF/terrain/internal/models"
)

// FixedTime is a deterministic timestamp for reproducible test output.
var FixedTime = time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)

// EmptySnapshot returns a valid but completely empty snapshot.
func EmptySnapshot() *models.TestSuiteSnapshot {
	return &models.TestSuiteSnapshot{
		Repository: models.RepositoryMetadata{
			Name: "empty-repo",
		},
		GeneratedAt: FixedTime,
	}
}

// MinimalSnapshot returns a snapshot with minimal content: one framework,
// one test file, one code unit.
func MinimalSnapshot() *models.TestSuiteSnapshot {
	return &models.TestSuiteSnapshot{
		Repository: models.RepositoryMetadata{
			Name:     "minimal-repo",
			Languages: []string{"javascript"},
		},
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 1, TestCount: 3},
		},
		TestFiles: []models.TestFile{
			{Path: "src/__tests__/utils.test.js", Framework: "jest", TestCount: 3, AssertionCount: 5, LinkedCodeUnits: []string{"formatDate"}},
		},
		CodeUnits: []models.CodeUnit{
			{Name: "formatDate", Path: "src/utils.js", Kind: models.CodeUnitKindFunction, Exported: true},
		},
		GeneratedAt: FixedTime,
	}
}

// HealthyBalancedSnapshot returns a well-tested repo with strong coverage
// across multiple frameworks and code units.
func HealthyBalancedSnapshot() *models.TestSuiteSnapshot {
	return &models.TestSuiteSnapshot{
		Repository: models.RepositoryMetadata{
			Name:     "healthy-balanced",
			Languages: []string{"javascript"},
		},
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 8, TestCount: 45},
			{Name: "playwright", Type: models.FrameworkTypeE2E, FileCount: 3, TestCount: 12},
		},
		TestFiles: []models.TestFile{
			{Path: "src/__tests__/auth.test.js", Framework: "jest", TestCount: 8, AssertionCount: 15, LinkedCodeUnits: []string{"AuthService"}},
			{Path: "src/__tests__/user.test.js", Framework: "jest", TestCount: 6, AssertionCount: 10, LinkedCodeUnits: []string{"UserService"}},
			{Path: "src/__tests__/payment.test.js", Framework: "jest", TestCount: 7, AssertionCount: 12, LinkedCodeUnits: []string{"PaymentProcessor"}},
			{Path: "src/__tests__/config.test.js", Framework: "jest", TestCount: 4, AssertionCount: 6, LinkedCodeUnits: []string{"ConfigLoader"}},
			{Path: "src/__tests__/cache.test.js", Framework: "jest", TestCount: 5, AssertionCount: 8, LinkedCodeUnits: []string{"CacheManager"}},
			{Path: "src/__tests__/logger.test.js", Framework: "jest", TestCount: 3, AssertionCount: 5, LinkedCodeUnits: []string{"Logger"}},
			{Path: "src/__tests__/validator.test.js", Framework: "jest", TestCount: 6, AssertionCount: 10, LinkedCodeUnits: []string{"Validator"}},
			{Path: "src/__tests__/router.test.js", Framework: "jest", TestCount: 6, AssertionCount: 9, LinkedCodeUnits: []string{"Router"}},
			{Path: "e2e/login.spec.js", Framework: "playwright", TestCount: 4, AssertionCount: 6},
			{Path: "e2e/checkout.spec.js", Framework: "playwright", TestCount: 4, AssertionCount: 5},
			{Path: "e2e/dashboard.spec.js", Framework: "playwright", TestCount: 4, AssertionCount: 5},
		},
		CodeUnits: []models.CodeUnit{
			{Name: "AuthService", Path: "src/auth.js", Kind: models.CodeUnitKindClass, Exported: true},
			{Name: "UserService", Path: "src/user.js", Kind: models.CodeUnitKindClass, Exported: true},
			{Name: "PaymentProcessor", Path: "src/payment.js", Kind: models.CodeUnitKindClass, Exported: true},
			{Name: "ConfigLoader", Path: "src/config.js", Kind: models.CodeUnitKindClass, Exported: true},
			{Name: "CacheManager", Path: "src/cache.js", Kind: models.CodeUnitKindClass, Exported: true},
			{Name: "Logger", Path: "src/logger.js", Kind: models.CodeUnitKindClass, Exported: true},
			{Name: "Validator", Path: "src/validator.js", Kind: models.CodeUnitKindClass, Exported: true},
			{Name: "Router", Path: "src/router.js", Kind: models.CodeUnitKindClass, Exported: true},
		},
		Ownership: map[string][]string{
			"src/auth.js":    {"team-platform"},
			"src/user.js":    {"team-platform"},
			"src/payment.js": {"team-payments"},
			"src/config.js":  {"team-platform"},
			"src/cache.js":   {"team-infra"},
			"src/logger.js":  {"team-infra"},
		},
		GeneratedAt: FixedTime,
	}
}

// FlakyConcentratedSnapshot returns a repo with flaky tests and concentrated
// test ownership — a common risk pattern.
func FlakyConcentratedSnapshot() *models.TestSuiteSnapshot {
	return &models.TestSuiteSnapshot{
		Repository: models.RepositoryMetadata{
			Name:     "flaky-concentrated",
			Languages: []string{"javascript"},
		},
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 4, TestCount: 20},
			{Name: "cypress", Type: models.FrameworkTypeE2E, FileCount: 6, TestCount: 30},
		},
		TestFiles: []models.TestFile{
			{Path: "src/__tests__/api.test.js", Framework: "jest", TestCount: 8, AssertionCount: 3, LinkedCodeUnits: []string{"ApiClient"}},
			{Path: "src/__tests__/db.test.js", Framework: "jest", TestCount: 6, AssertionCount: 2, MockCount: 5},
			{Path: "src/__tests__/core.test.js", Framework: "jest", TestCount: 4, AssertionCount: 1, MockCount: 8},
			{Path: "src/__tests__/helpers.test.js", Framework: "jest", TestCount: 2, AssertionCount: 1},
			{Path: "cypress/e2e/flow1.cy.js", Framework: "cypress", TestCount: 5, AssertionCount: 3, RuntimeStats: &models.RuntimeStats{AvgRuntimeMs: 12000, PassRate: 0.7, RetryRate: 0.3}},
			{Path: "cypress/e2e/flow2.cy.js", Framework: "cypress", TestCount: 5, AssertionCount: 3, RuntimeStats: &models.RuntimeStats{AvgRuntimeMs: 15000, PassRate: 0.6, RetryRate: 0.4}},
			{Path: "cypress/e2e/flow3.cy.js", Framework: "cypress", TestCount: 5, AssertionCount: 2, RuntimeStats: &models.RuntimeStats{AvgRuntimeMs: 18000, PassRate: 0.65, RetryRate: 0.35}},
			{Path: "cypress/e2e/flow4.cy.js", Framework: "cypress", TestCount: 5, AssertionCount: 2, RuntimeStats: &models.RuntimeStats{AvgRuntimeMs: 20000, PassRate: 0.55, RetryRate: 0.45}},
			{Path: "cypress/e2e/flow5.cy.js", Framework: "cypress", TestCount: 5, AssertionCount: 2, RuntimeStats: &models.RuntimeStats{AvgRuntimeMs: 22000, PassRate: 0.5, RetryRate: 0.5}},
			{Path: "cypress/e2e/flow6.cy.js", Framework: "cypress", TestCount: 5, AssertionCount: 2, RuntimeStats: &models.RuntimeStats{AvgRuntimeMs: 25000, PassRate: 0.5, RetryRate: 0.5}},
		},
		CodeUnits: []models.CodeUnit{
			{Name: "ApiClient", Path: "src/api.js", Kind: models.CodeUnitKindClass, Exported: true},
			{Name: "DbAdapter", Path: "src/db.js", Kind: models.CodeUnitKindClass, Exported: true},
			{Name: "CoreEngine", Path: "src/core.js", Kind: models.CodeUnitKindClass, Exported: true},
			{Name: "UserManager", Path: "src/users.js", Kind: models.CodeUnitKindClass, Exported: true},
			{Name: "AuthHandler", Path: "src/auth.js", Kind: models.CodeUnitKindClass, Exported: true},
		},
		GeneratedAt: FixedTime,
	}
}

// E2EHeavySnapshot returns a repo that relies heavily on E2E tests with
// shallow unit testing — common during migration from legacy frameworks.
func E2EHeavySnapshot() *models.TestSuiteSnapshot {
	return &models.TestSuiteSnapshot{
		Repository: models.RepositoryMetadata{
			Name:     "e2e-heavy",
			Languages: []string{"javascript"},
		},
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 2, TestCount: 6},
			{Name: "cypress", Type: models.FrameworkTypeE2E, FileCount: 10, TestCount: 50},
		},
		TestFiles: []models.TestFile{
			{Path: "src/__tests__/utils.test.js", Framework: "jest", TestCount: 3, AssertionCount: 5},
			{Path: "src/__tests__/config.test.js", Framework: "jest", TestCount: 3, AssertionCount: 4},
			{Path: "cypress/e2e/login.cy.js", Framework: "cypress", TestCount: 5, AssertionCount: 3},
			{Path: "cypress/e2e/register.cy.js", Framework: "cypress", TestCount: 5, AssertionCount: 3},
			{Path: "cypress/e2e/checkout.cy.js", Framework: "cypress", TestCount: 5, AssertionCount: 3},
			{Path: "cypress/e2e/profile.cy.js", Framework: "cypress", TestCount: 5, AssertionCount: 3},
			{Path: "cypress/e2e/search.cy.js", Framework: "cypress", TestCount: 5, AssertionCount: 3},
			{Path: "cypress/e2e/cart.cy.js", Framework: "cypress", TestCount: 5, AssertionCount: 3},
			{Path: "cypress/e2e/admin.cy.js", Framework: "cypress", TestCount: 5, AssertionCount: 3},
			{Path: "cypress/e2e/settings.cy.js", Framework: "cypress", TestCount: 5, AssertionCount: 3},
			{Path: "cypress/e2e/reports.cy.js", Framework: "cypress", TestCount: 5, AssertionCount: 3},
			{Path: "cypress/e2e/notifications.cy.js", Framework: "cypress", TestCount: 5, AssertionCount: 3},
		},
		CodeUnits: []models.CodeUnit{
			{Name: "AuthService", Path: "src/auth.js", Kind: models.CodeUnitKindClass, Exported: true},
			{Name: "UserService", Path: "src/user.js", Kind: models.CodeUnitKindClass, Exported: true},
			{Name: "CartService", Path: "src/cart.js", Kind: models.CodeUnitKindClass, Exported: true},
			{Name: "SearchEngine", Path: "src/search.js", Kind: models.CodeUnitKindClass, Exported: true},
			{Name: "PaymentGateway", Path: "src/payment.js", Kind: models.CodeUnitKindClass, Exported: true},
			{Name: "NotificationService", Path: "src/notifications.js", Kind: models.CodeUnitKindClass, Exported: true},
		},
		GeneratedAt: FixedTime,
	}
}

// MigrationRiskSnapshot returns a repo mid-migration from one framework
// to another, with migration blockers and quality issues.
func MigrationRiskSnapshot() *models.TestSuiteSnapshot {
	return &models.TestSuiteSnapshot{
		Repository: models.RepositoryMetadata{
			Name:     "migration-risk",
			Languages: []string{"javascript"},
		},
		Frameworks: []models.Framework{
			{Name: "jasmine", Type: models.FrameworkTypeUnit, FileCount: 6, TestCount: 30},
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 4, TestCount: 20},
			{Name: "protractor", Type: models.FrameworkTypeE2E, FileCount: 3, TestCount: 15},
		},
		TestFiles: []models.TestFile{
			{Path: "spec/auth.spec.js", Framework: "jasmine", TestCount: 6, AssertionCount: 4},
			{Path: "spec/user.spec.js", Framework: "jasmine", TestCount: 5, AssertionCount: 3, MockCount: 4},
			{Path: "spec/api.spec.js", Framework: "jasmine", TestCount: 5, AssertionCount: 2, MockCount: 6},
			{Path: "spec/db.spec.js", Framework: "jasmine", TestCount: 5, AssertionCount: 2},
			{Path: "spec/cache.spec.js", Framework: "jasmine", TestCount: 5, AssertionCount: 3},
			{Path: "spec/logger.spec.js", Framework: "jasmine", TestCount: 4, AssertionCount: 2},
			{Path: "src/__tests__/auth.test.js", Framework: "jest", TestCount: 6, AssertionCount: 8, LinkedCodeUnits: []string{"AuthService"}},
			{Path: "src/__tests__/user.test.js", Framework: "jest", TestCount: 5, AssertionCount: 7, LinkedCodeUnits: []string{"UserService"}},
			{Path: "src/__tests__/api.test.js", Framework: "jest", TestCount: 5, AssertionCount: 6, LinkedCodeUnits: []string{"ApiClient"}},
			{Path: "src/__tests__/cache.test.js", Framework: "jest", TestCount: 4, AssertionCount: 5, LinkedCodeUnits: []string{"CacheManager"}},
			{Path: "e2e/login.spec.js", Framework: "protractor", TestCount: 5, AssertionCount: 3},
			{Path: "e2e/checkout.spec.js", Framework: "protractor", TestCount: 5, AssertionCount: 3},
			{Path: "e2e/admin.spec.js", Framework: "protractor", TestCount: 5, AssertionCount: 3},
		},
		CodeUnits: []models.CodeUnit{
			{Name: "AuthService", Path: "src/auth.js", Kind: models.CodeUnitKindClass, Exported: true, UnitID: "src/auth.js:AuthService"},
			{Name: "UserService", Path: "src/user.js", Kind: models.CodeUnitKindClass, Exported: true, UnitID: "src/user.js:UserService"},
			{Name: "ApiClient", Path: "src/api.js", Kind: models.CodeUnitKindClass, Exported: true, UnitID: "src/api.js:ApiClient"},
			{Name: "CacheManager", Path: "src/cache.js", Kind: models.CodeUnitKindClass, Exported: true, UnitID: "src/cache.js:CacheManager"},
			{Name: "DbAdapter", Path: "src/db.js", Kind: models.CodeUnitKindClass, Exported: true, UnitID: "src/db.js:DbAdapter"},
			{Name: "Logger", Path: "src/logger.js", Kind: models.CodeUnitKindClass, Exported: true, UnitID: "src/logger.js:Logger"},
		},
		CoverageSummary: &models.CoverageSummary{
			TotalCodeUnits:     6,
			CoveredByUnitTests: 2,
			CoveredByE2E:       4,
			CoveredOnlyByE2E:   3,
			UncoveredExported:  1,
			LineCoveragePct:    42.5,
		},
		CoverageInsights: []models.CoverageInsight{
			{Type: "e2e_only_coverage", Severity: "medium", Path: "src/db.js", UnitID: "src/db.js:DbAdapter", Description: "DbAdapter covered only by e2e"},
			{Type: "e2e_only_coverage", Severity: "medium", Path: "src/cache.js", UnitID: "src/cache.js:CacheManager", Description: "CacheManager covered only by e2e"},
			{Type: "e2e_only_coverage", Severity: "medium", Path: "src/logger.js", UnitID: "src/logger.js:Logger", Description: "Logger covered only by e2e"},
		},
		GeneratedAt: FixedTime,
	}
}

// LargeScaleSnapshot returns a snapshot with 500+ test files for scale testing.
func LargeScaleSnapshot() *models.TestSuiteSnapshot {
	snap := &models.TestSuiteSnapshot{
		Repository: models.RepositoryMetadata{
			Name:     "large-scale",
			Languages: []string{"javascript"},
		},
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 400, TestCount: 2000},
			{Name: "playwright", Type: models.FrameworkTypeE2E, FileCount: 100, TestCount: 500},
			{Name: "vitest", Type: models.FrameworkTypeUnit, FileCount: 50, TestCount: 250},
		},
		GeneratedAt: FixedTime,
	}

	// Generate test files.
	dirs := []string{"src/auth", "src/user", "src/payment", "src/admin", "src/api", "src/core", "src/utils", "src/db", "src/cache", "src/config"}
	for i := 0; i < 400; i++ {
		dir := dirs[i%len(dirs)]
		snap.TestFiles = append(snap.TestFiles, models.TestFile{
			Path:           fmt.Sprintf("%s/__tests__/module%d.test.js", dir, i),
			Framework:      "jest",
			TestCount:      5,
			AssertionCount: 8,
		})
	}
	for i := 0; i < 100; i++ {
		snap.TestFiles = append(snap.TestFiles, models.TestFile{
			Path:           fmt.Sprintf("e2e/flow%d.spec.js", i),
			Framework:      "playwright",
			TestCount:      5,
			AssertionCount: 3,
		})
	}
	for i := 0; i < 50; i++ {
		snap.TestFiles = append(snap.TestFiles, models.TestFile{
			Path:           fmt.Sprintf("src/new/__tests__/module%d.test.js", i),
			Framework:      "vitest",
			TestCount:      5,
			AssertionCount: 6,
		})
	}

	// Generate code units.
	for i := 0; i < 200; i++ {
		dir := dirs[i%len(dirs)]
		snap.CodeUnits = append(snap.CodeUnits, models.CodeUnit{
			Name:     fmt.Sprintf("Module%d", i),
			Path:     fmt.Sprintf("%s/module%d.js", dir, i),
			Kind:     models.CodeUnitKindClass,
			Exported: i%3 != 0, // ~67% exported
		})
	}

	return snap
}
