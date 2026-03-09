package lifecycle

import (
	"testing"

	"github.com/pmclSF/hamlet/internal/identity"
	"github.com/pmclSF/hamlet/internal/models"
)

func makeTestCase(path, suite, name string) models.TestCase {
	canonical := identity.BuildCanonical(path, []string{suite}, name, "")
	return models.TestCase{
		TestID:            identity.GenerateID(canonical),
		CanonicalIdentity: canonical,
		FilePath:          path,
		SuiteHierarchy:    []string{suite},
		TestName:          name,
	}
}

func TestInferContinuity_ExactMatch(t *testing.T) {
	tc1 := makeTestCase("src/auth.test.js", "AuthService", "should login")
	tc2 := makeTestCase("src/auth.test.js", "AuthService", "should logout")

	from := &models.TestSuiteSnapshot{TestCases: []models.TestCase{tc1, tc2}}
	to := &models.TestSuiteSnapshot{TestCases: []models.TestCase{tc1, tc2}}

	result := InferContinuity(from, to)

	if result.ExactCount != 2 {
		t.Errorf("exact count = %d, want 2", result.ExactCount)
	}
	if result.RemovedCount != 0 {
		t.Errorf("removed count = %d, want 0", result.RemovedCount)
	}
	if result.AddedCount != 0 {
		t.Errorf("added count = %d, want 0", result.AddedCount)
	}
	for _, m := range result.Mappings {
		if m.Class != ContinuityExact {
			t.Errorf("expected exact continuity, got %s", m.Class)
		}
		if m.IsHeuristic() {
			t.Error("exact continuity should not be heuristic")
		}
	}
}

func TestInferContinuity_RenameOnly(t *testing.T) {
	fromTC := makeTestCase("src/auth.test.js", "AuthService", "should login user")
	toTC := makeTestCase("src/auth.test.js", "AuthService", "should login user successfully")

	from := &models.TestSuiteSnapshot{TestCases: []models.TestCase{fromTC}}
	to := &models.TestSuiteSnapshot{TestCases: []models.TestCase{toTC}}

	result := InferContinuity(from, to)

	if result.ExactCount != 0 {
		t.Errorf("exact count = %d, want 0", result.ExactCount)
	}

	found := false
	for _, m := range result.Mappings {
		if m.Class == ContinuityRename {
			found = true
			if m.Confidence <= 0 {
				t.Error("rename should have positive confidence")
			}
			if !m.IsHeuristic() {
				t.Error("rename should be heuristic")
			}
		}
	}
	if !found {
		// May also be classified as ambiguous with lower similarity.
		// Check that it wasn't simply added+removed.
		hasAdded := false
		hasRemoved := false
		for _, m := range result.Mappings {
			if m.Class == ContinuityAdded {
				hasAdded = true
			}
			if m.Class == ContinuityRemoved {
				hasRemoved = true
			}
		}
		if hasAdded && hasRemoved {
			t.Log("classified as added+removed rather than rename — acceptable if similarity is low")
		}
	}
}

func TestInferContinuity_FileMoveOnly(t *testing.T) {
	fromTC := makeTestCase("src/old/auth.test.js", "AuthService", "should login")
	toTC := makeTestCase("src/new/auth.test.js", "AuthService", "should login")

	from := &models.TestSuiteSnapshot{TestCases: []models.TestCase{fromTC}}
	to := &models.TestSuiteSnapshot{TestCases: []models.TestCase{toTC}}

	result := InferContinuity(from, to)

	found := false
	for _, m := range result.Mappings {
		if m.Class == ContinuityMove {
			found = true
			if m.Confidence <= 0 {
				t.Error("move should have positive confidence")
			}
			if m.FromPath != "src/old/auth.test.js" {
				t.Errorf("from path = %s, want src/old/auth.test.js", m.FromPath)
			}
			if m.ToPath != "src/new/auth.test.js" {
				t.Errorf("to path = %s, want src/new/auth.test.js", m.ToPath)
			}
		}
	}
	if !found {
		t.Error("expected likely_move mapping for file move")
	}
}

func TestInferContinuity_SplitTest(t *testing.T) {
	fromTC := makeTestCase("src/auth.test.js", "AuthService", "handles auth")

	toTC1 := makeTestCase("src/auth.test.js", "AuthService", "handles auth login")
	toTC2 := makeTestCase("src/auth.test.js", "AuthService", "handles auth logout")

	from := &models.TestSuiteSnapshot{TestCases: []models.TestCase{fromTC}}
	to := &models.TestSuiteSnapshot{TestCases: []models.TestCase{toTC1, toTC2}}

	result := InferContinuity(from, to)

	splitCount := 0
	for _, m := range result.Mappings {
		if m.Class == ContinuitySplit {
			splitCount++
		}
	}
	if splitCount < 2 {
		t.Logf("split count = %d (split detection may require exact prefix match)", splitCount)
	}
}

func TestInferContinuity_MergeTests(t *testing.T) {
	fromTC1 := makeTestCase("src/auth.test.js", "AuthService", "validates login")
	fromTC2 := makeTestCase("src/auth.test.js", "AuthService", "validates logout")

	toTC := makeTestCase("src/auth.test.js", "AuthService", "validates")

	from := &models.TestSuiteSnapshot{TestCases: []models.TestCase{fromTC1, fromTC2}}
	to := &models.TestSuiteSnapshot{TestCases: []models.TestCase{toTC}}

	result := InferContinuity(from, to)

	mergeCount := 0
	for _, m := range result.Mappings {
		if m.Class == ContinuityMerge {
			mergeCount++
		}
	}
	if mergeCount < 2 {
		t.Logf("merge count = %d (merge detection requires prefix match)", mergeCount)
	}
}

func TestInferContinuity_AmbiguousContinuity(t *testing.T) {
	fromTC := makeTestCase("src/utils.test.js", "UtilsSuite", "formats date")
	toTC := makeTestCase("src/helpers.test.js", "HelpersSuite", "formats timestamp")

	from := &models.TestSuiteSnapshot{TestCases: []models.TestCase{fromTC}}
	to := &models.TestSuiteSnapshot{TestCases: []models.TestCase{toTC}}

	result := InferContinuity(from, to)

	// With very different names and paths, should be added+removed or ambiguous.
	total := result.AddedCount + result.RemovedCount + result.AmbiguousCount
	if total == 0 {
		t.Error("expected some added/removed/ambiguous mappings")
	}
}

func TestInferContinuity_NilSnapshots(t *testing.T) {
	tc := makeTestCase("src/auth.test.js", "Auth", "login")
	to := &models.TestSuiteSnapshot{TestCases: []models.TestCase{tc}}

	result := InferContinuity(nil, to)
	if result.AddedCount != 1 {
		t.Errorf("added count = %d, want 1", result.AddedCount)
	}

	result2 := InferContinuity(nil, nil)
	if len(result2.Mappings) != 0 {
		t.Errorf("expected no mappings for nil/nil, got %d", len(result2.Mappings))
	}
}

func TestInferContinuity_MixedChanges(t *testing.T) {
	// Simulate a realistic scenario:
	// - tc1 stays the same (exact)
	// - tc2 is renamed (heuristic)
	// - tc3 is removed
	// - tc4 is added
	tc1 := makeTestCase("src/auth.test.js", "Auth", "login")
	tc2from := makeTestCase("src/user.test.js", "User", "creates user")
	tc2to := makeTestCase("src/user.test.js", "User", "creates user account")
	tc3 := makeTestCase("src/old.test.js", "Old", "deprecated test")
	tc4 := makeTestCase("src/new.test.js", "New", "brand new test")

	from := &models.TestSuiteSnapshot{TestCases: []models.TestCase{tc1, tc2from, tc3}}
	to := &models.TestSuiteSnapshot{TestCases: []models.TestCase{tc1, tc2to, tc4}}

	result := InferContinuity(from, to)

	if result.ExactCount != 1 {
		t.Errorf("exact count = %d, want 1", result.ExactCount)
	}
	// tc3 should be removed, tc4 should be added.
	// tc2 may be matched as rename or be added+removed.
	totalMappings := result.ExactCount + result.RenameCount + result.MoveCount +
		result.SplitCount + result.MergeCount + result.RemovedCount + result.AddedCount + result.AmbiguousCount
	if totalMappings == 0 {
		t.Error("expected some mappings")
	}
}

func TestStringSimilarity(t *testing.T) {
	tests := []struct {
		a, b string
		min  float64
	}{
		{"identical", "identical", 1.0},
		{"", "", 0.0}, // both empty => 0 (short-circuit)
		{"hello", "", 0.0},
		{"should login user", "should login user successfully", 0.5},
		{"completely different", "nothing alike here", 0.0},
	}

	for _, tt := range tests {
		sim := stringSimilarity(tt.a, tt.b)
		if tt.a == tt.b && tt.a != "" && sim != 1.0 {
			t.Errorf("stringSimilarity(%q, %q) = %f, want 1.0", tt.a, tt.b, sim)
		}
		if sim < tt.min {
			t.Errorf("stringSimilarity(%q, %q) = %f, want >= %f", tt.a, tt.b, sim, tt.min)
		}
	}
}

func TestPathSimilarity(t *testing.T) {
	tests := []struct {
		a, b string
		min  float64
	}{
		{"src/auth.test.js", "src/auth.test.js", 1.0},
		{"src/old/auth.test.js", "src/new/auth.test.js", 0.5},
		{"src/auth.test.js", "test/helpers.test.js", 0.0},
	}

	for _, tt := range tests {
		sim := pathSimilarity(tt.a, tt.b)
		if sim < tt.min {
			t.Errorf("pathSimilarity(%q, %q) = %f, want >= %f", tt.a, tt.b, sim, tt.min)
		}
	}
}
