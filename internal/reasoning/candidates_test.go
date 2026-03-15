package reasoning

import (
	"math"
	"testing"
)

func TestNormalizeTestName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"should handle empty input", "empty handle input"},
		{"TestCalculateSum", "testcalculatesum"},
		{"it renders the component", "component renders"},
		{"", ""},
	}
	for _, tt := range tests {
		got := NormalizeTestName(tt.name)
		if got != tt.want {
			t.Errorf("NormalizeTestName(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestJaccardSets(t *testing.T) {
	tests := []struct {
		name string
		a, b map[string]bool
		want float64
	}{
		{"both empty", nil, nil, 0.0},
		{"identical", map[string]bool{"x": true, "y": true}, map[string]bool{"x": true, "y": true}, 1.0},
		{"disjoint", map[string]bool{"x": true}, map[string]bool{"y": true}, 0.0},
		{"partial overlap", map[string]bool{"x": true, "y": true}, map[string]bool{"y": true, "z": true}, 1.0 / 3.0},
	}
	for _, tt := range tests {
		got := JaccardSets(tt.a, tt.b)
		if math.Abs(got-tt.want) > 1e-9 {
			t.Errorf("JaccardSets(%s) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestLCSRatio(t *testing.T) {
	tests := []struct {
		name string
		a, b []string
		want float64
	}{
		{"both empty", nil, nil, 1.0},
		{"identical", []string{"a", "b", "c"}, []string{"a", "b", "c"}, 1.0},
		{"one empty", []string{"a"}, nil, 0.0},
		{"partial", []string{"a", "b", "c"}, []string{"a", "c"}, 2.0 / 3.0},
	}
	for _, tt := range tests {
		got := LCSRatio(tt.a, tt.b)
		if math.Abs(got-tt.want) > 1e-9 {
			t.Errorf("LCSRatio(%s) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestTokenJaccard(t *testing.T) {
	got := TokenJaccard("add calculate sum", "calculate sum total")
	// Tokens: {add, calculate, sum} vs {calculate, sum, total}
	// Intersection: {calculate, sum} = 2, Union: {add, calculate, sum, total} = 4
	want := 2.0 / 4.0
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("TokenJaccard = %v, want %v", got, want)
	}
}

func TestHasSetOverlap(t *testing.T) {
	if HasSetOverlap(nil, nil) {
		t.Error("nil sets should not overlap")
	}
	if HasSetOverlap(map[string]bool{"a": true}, map[string]bool{"b": true}) {
		t.Error("disjoint sets should not overlap")
	}
	if !HasSetOverlap(map[string]bool{"a": true}, map[string]bool{"a": true, "b": true}) {
		t.Error("overlapping sets should overlap")
	}
}

func TestScoreSimilarity(t *testing.T) {
	a := Fingerprint{
		Fixtures:         map[string]bool{"fix:db": true},
		Helpers:          map[string]bool{"hlp:auth": true},
		SuitePath:        []string{"user", "login"},
		AssertionPattern: "login success validate",
	}
	b := Fingerprint{
		Fixtures:         map[string]bool{"fix:db": true},
		Helpers:          map[string]bool{"hlp:auth": true},
		SuitePath:        []string{"user", "login"},
		AssertionPattern: "login success validate",
	}

	score, signals := ScoreSimilarity(a, b)
	if score < 0.99 {
		t.Errorf("identical fingerprints should score ~1.0, got %v", score)
	}
	if signals.FixtureOverlap < 0.99 {
		t.Errorf("fixture overlap should be ~1.0, got %v", signals.FixtureOverlap)
	}
}

func TestGenerateCandidates_Empty(t *testing.T) {
	pairs := GenerateCandidates(nil, DefaultCandidateConfig())
	if pairs != nil {
		t.Errorf("expected nil for empty fingerprints, got %v", pairs)
	}
}

func TestGenerateCandidates_SimilarPair(t *testing.T) {
	fps := []Fingerprint{
		{
			NodeID:           "test:1",
			PackageID:        "pkg:main",
			Fixtures:         map[string]bool{"fix:db": true},
			Helpers:          map[string]bool{"hlp:auth": true},
			SuitePath:        []string{"user"},
			AssertionPattern: "login success",
		},
		{
			NodeID:           "test:2",
			PackageID:        "pkg:main",
			Fixtures:         map[string]bool{"fix:db": true},
			Helpers:          map[string]bool{"hlp:auth": true},
			SuitePath:        []string{"user"},
			AssertionPattern: "login success",
		},
	}

	pairs := GenerateCandidates(fps, DefaultCandidateConfig())
	if len(pairs) != 1 {
		t.Fatalf("expected 1 candidate pair, got %d", len(pairs))
	}
	if pairs[0].Similarity < DefaultSimilarityThreshold {
		t.Errorf("pair similarity %v below threshold", pairs[0].Similarity)
	}
}

func TestGenerateCandidates_DissimilarPair(t *testing.T) {
	fps := []Fingerprint{
		{
			NodeID:           "test:1",
			PackageID:        "pkg:a",
			Fixtures:         map[string]bool{"fix:db": true},
			Helpers:          map[string]bool{},
			SuitePath:        []string{"auth"},
			AssertionPattern: "login",
		},
		{
			NodeID:           "test:2",
			PackageID:        "pkg:b",
			Fixtures:         map[string]bool{"fix:redis": true},
			Helpers:          map[string]bool{},
			SuitePath:        []string{"payments"},
			AssertionPattern: "checkout",
		},
	}

	pairs := GenerateCandidates(fps, DefaultCandidateConfig())
	if len(pairs) != 0 {
		t.Errorf("expected 0 candidate pairs for dissimilar tests, got %d", len(pairs))
	}
}
