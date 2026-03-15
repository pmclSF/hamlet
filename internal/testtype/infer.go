// Package testtype implements test type inference with evidence and confidence.
//
// Each test case is classified as unit, integration, e2e, component, smoke,
// or unknown based on multiple signals: path conventions, framework hints,
// import patterns, and annotations.
package testtype

import (
	"path/filepath"
	"strings"

	"github.com/pmclSF/terrain/internal/models"
)

// TestType constants.
const (
	TypeUnit        = "unit"
	TypeIntegration = "integration"
	TypeE2E         = "e2e"
	TypeComponent   = "component"
	TypeSmoke       = "smoke"
	TypeUnknown     = "unknown"
)

// InferResult contains the inferred type and supporting evidence.
type InferResult struct {
	Type       string   `json:"type"`
	Confidence float64  `json:"confidence"`
	Evidence   []string `json:"evidence"`
}

// InferForTestCase infers the test type for a single test case.
func InferForTestCase(tc *models.TestCase) InferResult {
	var candidates []InferResult

	// Rule 1: Framework-based inference (strongest for e2e frameworks).
	if r := inferFromFramework(tc.Framework); r.Type != TypeUnknown {
		candidates = append(candidates, r)
	}

	// Rule 2: Path-based inference.
	if r := inferFromPath(tc.FilePath); r.Type != TypeUnknown {
		candidates = append(candidates, r)
	}

	// Rule 3: Suite hierarchy naming conventions.
	if r := inferFromSuiteNames(tc.SuiteHierarchy); r.Type != TypeUnknown {
		candidates = append(candidates, r)
	}

	// Rule 4: Test name patterns.
	if r := inferFromTestName(tc.TestName); r.Type != TypeUnknown {
		candidates = append(candidates, r)
	}

	if len(candidates) == 0 {
		return InferResult{
			Type:       TypeUnknown,
			Confidence: 0.0,
			Evidence:   []string{"no signals available for type inference"},
		}
	}

	// Merge: highest-confidence candidate wins, accumulate all evidence.
	best := candidates[0]
	for _, c := range candidates[1:] {
		if c.Confidence > best.Confidence {
			best.Type = c.Type
			best.Confidence = c.Confidence
		}
		best.Evidence = append(best.Evidence, c.Evidence...)
	}

	// If candidates disagree, reduce confidence.
	if hasConflict(candidates) {
		best.Confidence *= 0.8
		best.Evidence = append(best.Evidence, "conflicting signals reduced confidence")
	}

	return best
}

// InferAll applies type inference to all test cases in a snapshot.
func InferAll(cases []models.TestCase) []models.TestCase {
	for i := range cases {
		r := InferForTestCase(&cases[i])
		cases[i].TestType = r.Type
		cases[i].TestTypeConfidence = r.Confidence
		cases[i].TestTypeEvidence = r.Evidence
	}
	return cases
}

func inferFromFramework(framework string) InferResult {
	switch framework {
	case "playwright", "cypress", "puppeteer", "webdriverio", "testcafe":
		return InferResult{
			Type:       TypeE2E,
			Confidence: 0.9,
			Evidence:   []string{"framework " + framework + " is an e2e framework"},
		}
	case "jest", "vitest", "mocha", "jasmine":
		// Unit test frameworks — but could be used for integration tests too.
		return InferResult{
			Type:       TypeUnit,
			Confidence: 0.5,
			Evidence:   []string{"framework " + framework + " is typically used for unit tests"},
		}
	case "go-testing":
		return InferResult{
			Type:       TypeUnit,
			Confidence: 0.5,
			Evidence:   []string{"go-testing is used for both unit and integration tests"},
		}
	case "pytest", "unittest", "nose2":
		return InferResult{
			Type:       TypeUnit,
			Confidence: 0.5,
			Evidence:   []string{"framework " + framework + " is typically used for unit tests"},
		}
	case "junit4", "junit5", "testng":
		return InferResult{
			Type:       TypeUnit,
			Confidence: 0.5,
			Evidence:   []string{"framework " + framework + " is typically used for unit tests"},
		}
	}
	return InferResult{Type: TypeUnknown}
}

func inferFromPath(filePath string) InferResult {
	lower := strings.ToLower(filepath.ToSlash(filePath))
	parts := strings.Split(lower, "/")

	for _, p := range parts {
		switch {
		case p == "e2e" || p == "end-to-end" || p == "end2end":
			return InferResult{
				Type:       TypeE2E,
				Confidence: 0.85,
				Evidence:   []string{"path contains e2e directory: " + p},
			}
		case p == "integration" || p == "integ" || p == "int-tests":
			return InferResult{
				Type:       TypeIntegration,
				Confidence: 0.85,
				Evidence:   []string{"path contains integration directory: " + p},
			}
		case p == "component" || p == "components":
			// Only if it looks like a test directory, not a source directory.
			if containsTestIndicator(lower) {
				return InferResult{
					Type:       TypeComponent,
					Confidence: 0.7,
					Evidence:   []string{"path contains component test directory"},
				}
			}
		case p == "smoke":
			return InferResult{
				Type:       TypeSmoke,
				Confidence: 0.85,
				Evidence:   []string{"path contains smoke directory"},
			}
		case p == "unit" || p == "__tests__":
			return InferResult{
				Type:       TypeUnit,
				Confidence: 0.75,
				Evidence:   []string{"path contains unit test directory: " + p},
			}
		}
	}

	// File name patterns.
	base := filepath.Base(lower)
	switch {
	case strings.Contains(base, ".e2e.") || strings.Contains(base, ".e2e-"):
		return InferResult{
			Type:       TypeE2E,
			Confidence: 0.8,
			Evidence:   []string{"filename contains .e2e. pattern"},
		}
	case strings.Contains(base, ".integration.") || strings.Contains(base, ".integ."):
		return InferResult{
			Type:       TypeIntegration,
			Confidence: 0.8,
			Evidence:   []string{"filename contains .integration. pattern"},
		}
	case strings.Contains(base, ".component."):
		return InferResult{
			Type:       TypeComponent,
			Confidence: 0.7,
			Evidence:   []string{"filename contains .component. pattern"},
		}
	case strings.HasSuffix(base, ".cy.js") || strings.HasSuffix(base, ".cy.ts"):
		return InferResult{
			Type:       TypeE2E,
			Confidence: 0.9,
			Evidence:   []string{"cypress test file extension (.cy.js/.cy.ts)"},
		}
	case strings.HasSuffix(base, ".spec.js") || strings.HasSuffix(base, ".spec.ts"):
		// .spec files are ambiguous — could be unit or e2e.
		return InferResult{
			Type:       TypeUnit,
			Confidence: 0.4,
			Evidence:   []string{".spec extension is ambiguous (could be unit or e2e)"},
		}
	}

	return InferResult{Type: TypeUnknown}
}

func inferFromSuiteNames(hierarchy []string) InferResult {
	for _, name := range hierarchy {
		lower := strings.ToLower(name)
		switch {
		case strings.Contains(lower, "integration"):
			return InferResult{
				Type:       TypeIntegration,
				Confidence: 0.7,
				Evidence:   []string{"suite name contains 'integration': " + name},
			}
		case strings.Contains(lower, "e2e") || strings.Contains(lower, "end to end"):
			return InferResult{
				Type:       TypeE2E,
				Confidence: 0.7,
				Evidence:   []string{"suite name contains 'e2e': " + name},
			}
		case strings.Contains(lower, "component"):
			return InferResult{
				Type:       TypeComponent,
				Confidence: 0.6,
				Evidence:   []string{"suite name contains 'component': " + name},
			}
		case strings.Contains(lower, "smoke"):
			return InferResult{
				Type:       TypeSmoke,
				Confidence: 0.7,
				Evidence:   []string{"suite name contains 'smoke': " + name},
			}
		}
	}
	return InferResult{Type: TypeUnknown}
}

func inferFromTestName(name string) InferResult {
	lower := strings.ToLower(name)
	switch {
	case strings.HasPrefix(lower, "e2e") || strings.HasPrefix(lower, "end-to-end"):
		return InferResult{
			Type:       TypeE2E,
			Confidence: 0.6,
			Evidence:   []string{"test name starts with 'e2e'"},
		}
	case strings.HasPrefix(lower, "integration") || strings.HasPrefix(lower, "integ"):
		return InferResult{
			Type:       TypeIntegration,
			Confidence: 0.6,
			Evidence:   []string{"test name starts with 'integration'"},
		}
	case strings.HasPrefix(lower, "smoke"):
		return InferResult{
			Type:       TypeSmoke,
			Confidence: 0.6,
			Evidence:   []string{"test name starts with 'smoke'"},
		}
	}
	return InferResult{Type: TypeUnknown}
}

func containsTestIndicator(path string) bool {
	return strings.Contains(path, "test") || strings.Contains(path, "spec")
}

func hasConflict(candidates []InferResult) bool {
	if len(candidates) < 2 {
		return false
	}
	first := candidates[0].Type
	for _, c := range candidates[1:] {
		if c.Type != first && c.Type != TypeUnknown {
			return true
		}
	}
	return false
}
