package runtime

import (
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

func TestResolveTestIDs_ExactMatch(t *testing.T) {
	t.Parallel()
	results := []TestResult{
		{Name: "validates token", File: "src/auth/auth.test.js"},
		{Name: "creates user", File: "src/api/api.test.js"},
	}
	cases := []models.TestCase{
		{TestID: "t1", FilePath: "src/auth/auth.test.js", TestName: "validates token"},
		{TestID: "t2", FilePath: "src/api/api.test.js", TestName: "creates user"},
	}

	resolved := ResolveTestIDs(results, cases)
	if resolved != 2 {
		t.Errorf("resolved: got %d, want 2", resolved)
	}
	if results[0].TestID != "t1" {
		t.Errorf("results[0].TestID: got %q, want t1", results[0].TestID)
	}
	if results[1].TestID != "t2" {
		t.Errorf("results[1].TestID: got %q, want t2", results[1].TestID)
	}
}

func TestResolveTestIDs_SuffixMatch(t *testing.T) {
	t.Parallel()
	results := []TestResult{
		{Name: "validates token", File: "/home/ci/project/src/auth/auth.test.js"},
	}
	cases := []models.TestCase{
		{TestID: "t1", FilePath: "src/auth/auth.test.js", TestName: "validates token"},
	}

	resolved := ResolveTestIDs(results, cases)
	if resolved != 1 {
		t.Errorf("resolved: got %d, want 1", resolved)
	}
	if results[0].TestID != "t1" {
		t.Errorf("results[0].TestID: got %q, want t1", results[0].TestID)
	}
}

func TestResolveTestIDs_FuzzyNameMatch(t *testing.T) {
	t.Parallel()
	results := []TestResult{
		{Name: "validates token with param=42", File: "src/auth/auth.test.js"},
	}
	cases := []models.TestCase{
		{TestID: "t1", FilePath: "src/auth/auth.test.js", TestName: "validates token"},
	}

	resolved := ResolveTestIDs(results, cases)
	if resolved != 1 {
		t.Errorf("resolved: got %d, want 1", resolved)
	}
}

func TestResolveTestIDs_NoMatch(t *testing.T) {
	t.Parallel()
	results := []TestResult{
		{Name: "unknown test", File: "unknown/file.js"},
	}
	cases := []models.TestCase{
		{TestID: "t1", FilePath: "src/auth/auth.test.js", TestName: "validates token"},
	}

	resolved := ResolveTestIDs(results, cases)
	if resolved != 0 {
		t.Errorf("resolved: got %d, want 0", resolved)
	}
	if results[0].TestID != "" {
		t.Errorf("results[0].TestID should be empty, got %q", results[0].TestID)
	}
}

func TestResolveTestIDs_EmptyInputs(t *testing.T) {
	t.Parallel()
	if ResolveTestIDs(nil, nil) != 0 {
		t.Error("nil inputs should return 0")
	}
	if ResolveTestIDs([]TestResult{}, nil) != 0 {
		t.Error("empty results should return 0")
	}
	if ResolveTestIDs(nil, []models.TestCase{}) != 0 {
		t.Error("empty cases should return 0")
	}
}
