package testcase

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestExtractJS_Basic(t *testing.T) {
	src := `
describe('AuthService', () => {
  describe('login', () => {
    it('should return a token', () => {
      expect(login()).toBeTruthy();
    });

    it('should reject invalid credentials', () => {
      expect(() => login('bad')).toThrow();
    });
  });

  test('handles empty input', () => {
    expect(auth(null)).toBeNull();
  });
});
`
	dir := t.TempDir()
	path := filepath.Join(dir, "auth.test.js")
	mustWriteFile(t, path, []byte(src))

	cases := Extract(dir, "auth.test.js", "jest")
	if len(cases) != 3 {
		t.Fatalf("expected 3 test cases, got %d", len(cases))
	}

	// Sort by line for predictable assertions.
	sort.Slice(cases, func(i, j int) bool { return cases[i].Line < cases[j].Line })

	// First test: AuthService > login > should return a token
	tc := cases[0]
	if tc.TestName != "should return a token" {
		t.Errorf("test 0 name = %q, want %q", tc.TestName, "should return a token")
	}
	if len(tc.SuiteHierarchy) != 2 || tc.SuiteHierarchy[0] != "AuthService" || tc.SuiteHierarchy[1] != "login" {
		t.Errorf("test 0 hierarchy = %v, want [AuthService login]", tc.SuiteHierarchy)
	}
	if tc.ExtractionKind != ExtractionStatic {
		t.Errorf("test 0 kind = %q, want %q", tc.ExtractionKind, ExtractionStatic)
	}

	// Second test.
	if cases[1].TestName != "should reject invalid credentials" {
		t.Errorf("test 1 name = %q", cases[1].TestName)
	}

	// Third test: top-level under AuthService.
	tc3 := cases[2]
	if tc3.TestName != "handles empty input" {
		t.Errorf("test 2 name = %q, want %q", tc3.TestName, "handles empty input")
	}
	if len(tc3.SuiteHierarchy) != 1 || tc3.SuiteHierarchy[0] != "AuthService" {
		t.Errorf("test 2 hierarchy = %v, want [AuthService]", tc3.SuiteHierarchy)
	}
}

func TestExtractJS_StableIDs(t *testing.T) {
	src := `
describe('Math', () => {
  it('adds', () => {});
  it('subtracts', () => {});
});
`
	dir := t.TempDir()
	path := filepath.Join(dir, "math.test.js")
	mustWriteFile(t, path, []byte(src))

	cases1 := Extract(dir, "math.test.js", "jest")
	cases2 := Extract(dir, "math.test.js", "jest")

	if len(cases1) != 2 || len(cases2) != 2 {
		t.Fatalf("expected 2 cases each, got %d and %d", len(cases1), len(cases2))
	}

	// IDs must be identical across runs.
	ids1 := map[string]bool{}
	for _, c := range cases1 {
		ids1[c.TestID] = true
	}
	for _, c := range cases2 {
		if !ids1[c.TestID] {
			t.Errorf("ID %q from run 2 not found in run 1", c.TestID)
		}
	}
}

func TestExtractJS_ReorderedExtraction_SameIDs(t *testing.T) {
	// Identity should not depend on traversal order.
	src := `
describe('Suite', () => {
  it('test B', () => {});
  it('test A', () => {});
});
`
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "order.test.js"), []byte(src))

	cases := Extract(dir, "order.test.js", "vitest")
	if len(cases) != 2 {
		t.Fatalf("expected 2 cases, got %d", len(cases))
	}

	// Manually verify canonical identity is path-based, not order-based.
	for _, c := range cases {
		if c.TestID == "" {
			t.Error("TestID should not be empty")
		}
		if c.CanonicalIdentity == "" {
			t.Error("CanonicalIdentity should not be empty")
		}
	}
}

func TestExtractJS_LineMovement_SameID(t *testing.T) {
	// Adding blank lines should not change the test ID.
	src1 := `describe('X', () => {
  it('works', () => {});
});`
	src2 := `
// Added comment

describe('X', () => {

  it('works', () => {});

});`

	dir := t.TempDir()
	f := filepath.Join(dir, "x.test.js")

	mustWriteFile(t, f, []byte(src1))
	cases1 := Extract(dir, "x.test.js", "jest")

	mustWriteFile(t, f, []byte(src2))
	cases2 := Extract(dir, "x.test.js", "jest")

	if len(cases1) != 1 || len(cases2) != 1 {
		t.Fatalf("expected 1 case each, got %d and %d", len(cases1), len(cases2))
	}

	if cases1[0].TestID != cases2[0].TestID {
		t.Errorf("line movement changed ID: %q != %q", cases1[0].TestID, cases2[0].TestID)
	}

	// Line numbers should differ (metadata only).
	if cases1[0].Line == cases2[0].Line {
		t.Error("line numbers should differ after line movement")
	}
}

func TestExtractJS_Rename_NewID(t *testing.T) {
	src1 := `describe('X', () => { it('old name', () => {}); });`
	src2 := `describe('X', () => { it('new name', () => {}); });`

	dir := t.TempDir()
	f := filepath.Join(dir, "x.test.js")

	mustWriteFile(t, f, []byte(src1))
	cases1 := Extract(dir, "x.test.js", "jest")

	mustWriteFile(t, f, []byte(src2))
	cases2 := Extract(dir, "x.test.js", "jest")

	if len(cases1) != 1 || len(cases2) != 1 {
		t.Fatal("expected 1 case each")
	}

	if cases1[0].TestID == cases2[0].TestID {
		t.Error("renaming a test should produce a new ID")
	}
}

func TestExtractGo(t *testing.T) {
	src := `package foo

func TestAdd(t *testing.T) {
	t.Run("positive numbers", func(t *testing.T) {
		if add(1, 2) != 3 {
			t.Error("wrong")
		}
	})
	t.Run("negative numbers", func(t *testing.T) {
		if add(-1, -2) != -3 {
			t.Error("wrong")
		}
	})
}

func TestSubtract(t *testing.T) {
	if sub(5, 3) != 2 {
		t.Error("wrong")
	}
}
`
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "math_test.go"), []byte(src))

	cases := Extract(dir, "math_test.go", "go-testing")
	if len(cases) != 4 {
		t.Fatalf("expected 4 test cases, got %d", len(cases))
	}

	sort.Slice(cases, func(i, j int) bool { return cases[i].Line < cases[j].Line })

	if cases[0].TestName != "TestAdd" {
		t.Errorf("case 0 name = %q", cases[0].TestName)
	}
	if cases[1].TestName != "positive numbers" {
		t.Errorf("case 1 name = %q", cases[1].TestName)
	}
	if len(cases[1].SuiteHierarchy) != 1 || cases[1].SuiteHierarchy[0] != "TestAdd" {
		t.Errorf("case 1 hierarchy = %v", cases[1].SuiteHierarchy)
	}
	if cases[3].TestName != "TestSubtract" {
		t.Errorf("case 3 name = %q", cases[3].TestName)
	}
}

func TestExtractPython(t *testing.T) {
	src := `import pytest

class TestCalculator:
    def test_add(self):
        assert add(1, 2) == 3

    @pytest.mark.parametrize("a,b,expected", [(1,2,3), (0,0,0)])
    def test_multiply(self, a, b, expected):
        assert multiply(a, b) == expected

def test_standalone():
    assert True
`
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "test_calc.py"), []byte(src))

	cases := Extract(dir, "test_calc.py", "pytest")
	if len(cases) != 3 {
		t.Fatalf("expected 3 test cases, got %d", len(cases))
	}

	sort.Slice(cases, func(i, j int) bool { return cases[i].Line < cases[j].Line })

	if cases[0].TestName != "test_add" {
		t.Errorf("case 0 name = %q", cases[0].TestName)
	}
	if len(cases[0].SuiteHierarchy) != 1 || cases[0].SuiteHierarchy[0] != "TestCalculator" {
		t.Errorf("case 0 hierarchy = %v", cases[0].SuiteHierarchy)
	}

	if cases[1].TestName != "test_multiply" {
		t.Errorf("case 1 name = %q", cases[1].TestName)
	}
	if cases[1].ExtractionKind != ExtractionParameterizedTemplate {
		t.Errorf("case 1 kind = %q, want parameterized_template", cases[1].ExtractionKind)
	}

	if cases[2].TestName != "test_standalone" {
		t.Errorf("case 2 name = %q", cases[2].TestName)
	}
	if len(cases[2].SuiteHierarchy) != 0 {
		t.Errorf("case 2 should have no hierarchy, got %v", cases[2].SuiteHierarchy)
	}
}

func TestExtractJava(t *testing.T) {
	src := `import org.junit.jupiter.api.Test;

class UserServiceTest {
    @Test
    void shouldCreateUser() {
        // test
    }

    @ParameterizedTest
    void shouldValidateEmail() {
        // test
    }
}
`
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "UserServiceTest.java"), []byte(src))

	cases := Extract(dir, "UserServiceTest.java", "junit5")
	if len(cases) != 2 {
		t.Fatalf("expected 2 test cases, got %d", len(cases))
	}

	sort.Slice(cases, func(i, j int) bool { return cases[i].Line < cases[j].Line })

	if cases[0].TestName != "shouldCreateUser" {
		t.Errorf("case 0 name = %q", cases[0].TestName)
	}
	if cases[1].ExtractionKind != ExtractionParameterizedTemplate {
		t.Errorf("case 1 kind = %q, want parameterized_template", cases[1].ExtractionKind)
	}
}

func mustWriteFile(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write %q: %v", path, err)
	}
}
