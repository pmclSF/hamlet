package reasoning

import (
	"sort"
	"strings"

	"github.com/pmclSF/terrain/internal/depgraph"
)

// Similarity weights for redundancy scoring.
const (
	WeightFixture   = 0.25
	WeightHelper    = 0.25
	WeightSuitePath = 0.30
	WeightAssertion = 0.20

	// DefaultSimilarityThreshold is the minimum composite similarity for
	// two tests to be considered redundancy candidates.
	DefaultSimilarityThreshold = 0.60

	// DefaultMaxBlockSize caps how many members a single blocking key can have.
	DefaultMaxBlockSize = 500
)

// CandidatePair is a pair of nodes identified as potential duplicates.
type CandidatePair struct {
	A, B       string
	Similarity float64
	Signals    SimilaritySignals
}

// SimilaritySignals breaks down a similarity score into components.
type SimilaritySignals struct {
	FixtureOverlap             float64 `json:"fixtureOverlap"`
	HelperOverlap              float64 `json:"helperOverlap"`
	SuitePathSimilarity        float64 `json:"suitePathSimilarity"`
	AssertionPatternSimilarity float64 `json:"assertionPatternSimilarity"`
}

// CandidateConfig controls redundancy candidate generation.
type CandidateConfig struct {
	// SimilarityThreshold is the minimum composite score. Default: 0.60.
	SimilarityThreshold float64

	// MaxBlockSize caps blocking-key group size. Default: 500.
	MaxBlockSize int
}

// DefaultCandidateConfig returns standard candidate generation parameters.
func DefaultCandidateConfig() CandidateConfig {
	return CandidateConfig{
		SimilarityThreshold: DefaultSimilarityThreshold,
		MaxBlockSize:        DefaultMaxBlockSize,
	}
}

// Fingerprint captures the structural signature of a test node for
// similarity comparison.
type Fingerprint struct {
	NodeID           string
	PackageID        string
	Fixtures         map[string]bool
	Helpers          map[string]bool
	SuitePath        []string
	AssertionPattern string
}

// BuildFingerprints extracts structural fingerprints for all test nodes
// in the graph.
func BuildFingerprints(g *depgraph.Graph) []Fingerprint {
	tests := g.NodesByType(depgraph.NodeTest)
	if len(tests) == 0 {
		return nil
	}

	// Build file→fixtures/helpers index.
	fileFixtures := map[string]map[string]bool{}
	fileHelpers := map[string]map[string]bool{}

	for _, e := range g.Edges() {
		switch e.Type {
		case depgraph.EdgeTestUsesFixture:
			if fileFixtures[e.From] == nil {
				fileFixtures[e.From] = map[string]bool{}
			}
			fileFixtures[e.From][e.To] = true
		case depgraph.EdgeTestUsesHelper:
			if fileHelpers[e.From] == nil {
				fileHelpers[e.From] = map[string]bool{}
			}
			fileHelpers[e.From][e.To] = true
		}
	}

	fps := make([]Fingerprint, len(tests))
	for i, test := range tests {
		fileID := "file:" + test.Path

		fixtures := map[string]bool{}
		for f := range fileFixtures[fileID] {
			fixtures[f] = true
		}
		helpers := map[string]bool{}
		for h := range fileHelpers[fileID] {
			helpers[h] = true
		}

		suitePath := extractSuitePath(g, test.ID)
		pattern := NormalizeTestName(test.Name)

		fps[i] = Fingerprint{
			NodeID:           test.ID,
			PackageID:        test.Package,
			Fixtures:         fixtures,
			Helpers:          helpers,
			SuitePath:        suitePath,
			AssertionPattern: pattern,
		}
	}

	return fps
}

// extractSuitePath walks reverse edges to find the suite hierarchy for a test.
func extractSuitePath(g *depgraph.Graph, testID string) []string {
	var path []string
	for _, e := range g.Incoming(testID) {
		if e.Type == depgraph.EdgeSuiteContainsTest {
			suiteNode := g.Node(e.From)
			if suiteNode != nil {
				path = append(path, suiteNode.Name)
			}
		}
	}
	if len(path) == 0 {
		testNode := g.Node(testID)
		if testNode != nil && testNode.Path != "" {
			path = []string{testNode.Path}
		}
	}
	return path
}

// NormalizeTestName strips noise words, lowercases, sorts tokens.
func NormalizeTestName(name string) string {
	noise := map[string]bool{
		"should": true, "must": true, "can": true, "does": true,
		"will": true, "it": true, "the": true, "a": true, "an": true,
	}
	tokens := strings.FieldsFunc(strings.ToLower(name), func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))
	})
	filtered := make([]string, 0, len(tokens))
	for _, t := range tokens {
		if !noise[t] {
			filtered = append(filtered, t)
		}
	}
	sort.Strings(filtered)
	return strings.Join(filtered, " ")
}

// GenerateCandidates produces candidate pairs using blocking keys to avoid
// O(n²) comparison. Tests sharing a package, fixture, or helper are candidates.
func GenerateCandidates(fps []Fingerprint, cfg CandidateConfig) []CandidatePair {
	if len(fps) < 2 {
		return nil
	}
	if cfg.SimilarityThreshold <= 0 {
		cfg.SimilarityThreshold = DefaultSimilarityThreshold
	}
	if cfg.MaxBlockSize <= 0 {
		cfg.MaxBlockSize = DefaultMaxBlockSize
	}

	// Build blocking-key index.
	blocks := map[string][]int{}
	for i, fp := range fps {
		if fp.PackageID != "" {
			key := "pkg:" + fp.PackageID
			blocks[key] = append(blocks[key], i)
		}
		for f := range fp.Fixtures {
			key := "fix:" + f
			blocks[key] = append(blocks[key], i)
		}
		for h := range fp.Helpers {
			key := "hlp:" + h
			blocks[key] = append(blocks[key], i)
		}
	}

	// Emit unique pairs from each block.
	seen := map[[2]int]bool{}
	var pairs []CandidatePair

	for _, members := range blocks {
		if len(members) < 2 || len(members) > cfg.MaxBlockSize {
			continue
		}
		for a := 0; a < len(members); a++ {
			for b := a + 1; b < len(members); b++ {
				i, j := members[a], members[b]
				if i > j {
					i, j = j, i
				}
				key := [2]int{i, j}
				if seen[key] {
					continue
				}
				seen[key] = true

				// Skip pairs with no structural overlap.
				if !HasSetOverlap(fps[i].Fixtures, fps[j].Fixtures) &&
					!HasSetOverlap(fps[i].Helpers, fps[j].Helpers) {
					// With empty Jaccard = 0, max composite is 0.50 < threshold.
					if len(fps[i].Fixtures) == 0 && len(fps[j].Fixtures) == 0 &&
						len(fps[i].Helpers) == 0 && len(fps[j].Helpers) == 0 {
						continue
					}
					continue
				}

				score, signals := ScoreSimilarity(fps[i], fps[j])
				if score >= cfg.SimilarityThreshold {
					pairs = append(pairs, CandidatePair{
						A:          fps[i].NodeID,
						B:          fps[j].NodeID,
						Similarity: score,
						Signals:    signals,
					})
				}
			}
		}
	}

	// Sort by similarity descending.
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Similarity > pairs[j].Similarity
	})

	return pairs
}

// ScoreSimilarity computes a weighted similarity score between two fingerprints.
func ScoreSimilarity(a, b Fingerprint) (float64, SimilaritySignals) {
	fixtureScore := JaccardSets(a.Fixtures, b.Fixtures)
	helperScore := JaccardSets(a.Helpers, b.Helpers)
	suiteScore := LCSRatio(a.SuitePath, b.SuitePath)
	assertionScore := TokenJaccard(a.AssertionPattern, b.AssertionPattern)

	composite := WeightFixture*fixtureScore +
		WeightHelper*helperScore +
		WeightSuitePath*suiteScore +
		WeightAssertion*assertionScore

	return composite, SimilaritySignals{
		FixtureOverlap:             fixtureScore,
		HelperOverlap:              helperScore,
		SuitePathSimilarity:        suiteScore,
		AssertionPatternSimilarity: assertionScore,
	}
}

// HasSetOverlap returns true if sets a and b share at least one element.
func HasSetOverlap(a, b map[string]bool) bool {
	if len(a) > len(b) {
		a, b = b, a
	}
	for k := range a {
		if b[k] {
			return true
		}
	}
	return false
}

// JaccardSets computes Jaccard similarity between two sets.
func JaccardSets(a, b map[string]bool) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 0.0
	}
	inter := 0
	for k := range a {
		if b[k] {
			inter++
		}
	}
	union := len(a) + len(b) - inter
	return float64(inter) / float64(union)
}

// LCSRatio computes the LCS ratio between two string slices.
func LCSRatio(a, b []string) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 1.0
	}
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}
	if maxLen == 0 {
		return 1.0
	}

	dp := make([][]int, len(a)+1)
	for i := range dp {
		dp[i] = make([]int, len(b)+1)
	}
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] > dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}

	return float64(dp[len(a)][len(b)]) / float64(maxLen)
}

// TokenJaccard computes Jaccard similarity on space-separated tokens.
func TokenJaccard(a, b string) float64 {
	tokA := map[string]bool{}
	for _, t := range strings.Fields(a) {
		tokA[t] = true
	}
	tokB := map[string]bool{}
	for _, t := range strings.Fields(b) {
		tokB[t] = true
	}
	return JaccardSets(tokA, tokB)
}
