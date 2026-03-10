package analysis

import "testing"

func TestLanguageRegistry_AllRegistered(t *testing.T) {
	t.Parallel()
	expected := []string{"js", "go", "python", "java"}
	for _, lang := range expected {
		a := getLanguageAnalyzer(lang)
		if a == nil {
			t.Errorf("no analyzer registered for %s", lang)
			continue
		}
		if a.Language() != lang {
			t.Errorf("analyzer for %s returns Language()=%s", lang, a.Language())
		}
	}
}

func TestLanguageRegistry_DefaultFallback(t *testing.T) {
	t.Parallel()
	a := getLanguageAnalyzer("unknown")
	if a == nil {
		t.Fatal("expected fallback analyzer")
	}
	if a.Language() != "js" {
		t.Errorf("expected js fallback, got %s", a.Language())
	}
}

func TestJSAnalyzer_CountTests(t *testing.T) {
	t.Parallel()
	src := `
it('should work', () => {});
test('another test', () => {});
describe('suite', () => {});
`
	a := &jsAnalyzer{}
	count := a.CountTests(src)
	if count != 2 {
		t.Errorf("expected 2 tests, got %d", count)
	}
}

func TestGoAnalyzer_CountTests(t *testing.T) {
	t.Parallel()
	src := `
func TestFoo(t *testing.T) {}
func TestBar(t *testing.T) {}
func helperFunc() {}
`
	a := &goAnalyzer{}
	count := a.CountTests(src)
	if count != 2 {
		t.Errorf("expected 2 tests, got %d", count)
	}
}

func TestPythonAnalyzer_CountTests(t *testing.T) {
	t.Parallel()
	src := `
def test_foo():
    pass
def test_bar():
    pass
def helper():
    pass
`
	a := &pythonAnalyzer{}
	count := a.CountTests(src)
	if count != 2 {
		t.Errorf("expected 2 tests, got %d", count)
	}
}

func TestJavaAnalyzer_CountTests(t *testing.T) {
	t.Parallel()
	src := `
@Test
public void testFoo() {}
@Test
public void testBar() {}
public void helper() {}
`
	a := &javaAnalyzer{}
	count := a.CountTests(src)
	if count != 2 {
		t.Errorf("expected 2 tests, got %d", count)
	}
}

func TestFrameworkLanguage_Mapping(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"jest":       "js",
		"vitest":     "js",
		"playwright": "js",
		"go-testing": "go",
		"pytest":     "python",
		"junit5":     "java",
	}
	for fw, want := range cases {
		got := frameworkLanguage(fw)
		if got != want {
			t.Errorf("frameworkLanguage(%q) = %q, want %q", fw, got, want)
		}
	}
}
