package identity

import "testing"

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"src/test/foo.test.js", "src/test/foo.test.js"},
		{"./src/test/foo.test.js", "src/test/foo.test.js"},
		{"/src/test/foo.test.js", "src/test/foo.test.js"},
		{"src\\test\\foo.test.js", "src/test/foo.test.js"},
		{".\\src\\test\\foo.test.js", "src/test/foo.test.js"},
		{"", ""},
	}
	for _, tt := range tests {
		got := NormalizePath(tt.input)
		if got != tt.want {
			t.Errorf("NormalizePath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"should do something", "should do something"},
		{"  should  do   something  ", "should do something"},
		{"should\tdo\nsomething", "should do something"},
		{"", ""},
		{"  ", ""},
		{"CaseSensitive", "CaseSensitive"},
	}
	for _, tt := range tests {
		got := NormalizeName(tt.input)
		if got != tt.want {
			t.Errorf("NormalizeName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestNormalizeSuiteHierarchy(t *testing.T) {
	tests := []struct {
		parts []string
		want  string
	}{
		{[]string{"AuthService", "login"}, "AuthService > login"},
		{[]string{"  AuthService ", " login "}, "AuthService > login"},
		{[]string{"", "login"}, "login"},
		{nil, ""},
		{[]string{"Single"}, "Single"},
	}
	for _, tt := range tests {
		got := NormalizeSuiteHierarchy(tt.parts)
		if got != tt.want {
			t.Errorf("NormalizeSuiteHierarchy(%v) = %q, want %q", tt.parts, got, tt.want)
		}
	}
}

func TestBuildCanonical(t *testing.T) {
	tests := []struct {
		path      string
		hierarchy []string
		name      string
		param     string
		want      string
	}{
		{
			"src/__tests__/auth.test.js",
			[]string{"AuthService"},
			"should login",
			"",
			"src/__tests__/auth.test.js::AuthService::should login",
		},
		{
			"./src/__tests__/auth.test.js",
			[]string{"AuthService", "login"},
			"returns token",
			"",
			"src/__tests__/auth.test.js::AuthService > login::returns token",
		},
		{
			"test/utils.test.js",
			nil,
			"handles empty input",
			"",
			"test/utils.test.js::::handles empty input",
		},
		{
			"test/math.test.js",
			[]string{"add"},
			"adds numbers",
			"1 + 2 = 3",
			"test/math.test.js::add::adds numbers::1 + 2 = 3",
		},
	}
	for _, tt := range tests {
		got := BuildCanonical(tt.path, tt.hierarchy, tt.name, tt.param)
		if got != tt.want {
			t.Errorf("BuildCanonical(%q, %v, %q, %q) = %q, want %q",
				tt.path, tt.hierarchy, tt.name, tt.param, got, tt.want)
		}
	}
}

func TestParseCanonical(t *testing.T) {
	path, suite, name, param := ParseCanonical("src/test.js::Suite > Sub::test name::param")
	if path != "src/test.js" {
		t.Errorf("path = %q, want %q", path, "src/test.js")
	}
	if suite != "Suite > Sub" {
		t.Errorf("suite = %q, want %q", suite, "Suite > Sub")
	}
	if name != "test name" {
		t.Errorf("name = %q, want %q", name, "test name")
	}
	if param != "param" {
		t.Errorf("param = %q, want %q", param, "param")
	}
}

func TestGenerateID_Deterministic(t *testing.T) {
	canonical := "src/__tests__/auth.test.js::AuthService::should login"
	id1 := GenerateID(canonical)
	id2 := GenerateID(canonical)
	if id1 != id2 {
		t.Errorf("GenerateID not deterministic: %q != %q", id1, id2)
	}
	if len(id1) != HashLength {
		t.Errorf("GenerateID length = %d, want %d", len(id1), HashLength)
	}
}

func TestGenerateID_Different(t *testing.T) {
	id1 := GenerateID("path::suite::test1")
	id2 := GenerateID("path::suite::test2")
	if id1 == id2 {
		t.Error("different canonical identities should produce different IDs")
	}
}

func TestSameSource_SameID(t *testing.T) {
	// Same content with different whitespace normalization should produce same ID.
	c1 := BuildCanonical("src/test.js", []string{"  Suite "}, "  should  work  ", "")
	c2 := BuildCanonical("./src/test.js", []string{"Suite"}, "should work", "")
	if c1 != c2 {
		t.Errorf("canonical should match after normalization: %q != %q", c1, c2)
	}
	if GenerateID(c1) != GenerateID(c2) {
		t.Error("same source should produce same ID")
	}
}

func TestPathNormalization_Stability(t *testing.T) {
	// Different path representations should normalize the same.
	paths := []string{
		"src/test/auth.test.js",
		"./src/test/auth.test.js",
		"src\\test\\auth.test.js",
	}
	normalized := NormalizePath(paths[0])
	for _, p := range paths[1:] {
		got := NormalizePath(p)
		if got != normalized {
			t.Errorf("NormalizePath(%q) = %q, want %q", p, got, normalized)
		}
	}
}
