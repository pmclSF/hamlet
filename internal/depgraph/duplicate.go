package depgraph

import (
	"fmt"
	"sort"
	"strings"
)

// DuplicateResult contains the structural duplicate analysis.
type DuplicateResult struct {
	// All duplicate clusters found.
	Clusters []DuplicateCluster `json:"clusters"`

	// Total number of tests analyzed.
	TestsAnalyzed int `json:"testsAnalyzed"`

	// Number of tests flagged as potential duplicates.
	DuplicateCount int `json:"duplicateCount"`

	// Indicates duplicate analysis was skipped for scale-safety.
	Skipped bool `json:"skipped,omitempty"`

	// Human-readable reason when duplicate analysis is skipped.
	SkipReason string `json:"skipReason,omitempty"`
}

// DuplicateCluster is a group of structurally similar tests.
type DuplicateCluster struct {
	// Unique cluster identifier.
	ID int `json:"id"`

	// Test node IDs in this cluster.
	Tests []string `json:"tests"`

	// Overall similarity score for the cluster (0–1).
	Similarity float64 `json:"similarity"`

	// Breakdown of signal contributions.
	Signals SimilaritySignals `json:"signals"`
}

// SimilaritySignals breaks down the similarity score into components.
type SimilaritySignals struct {
	FixtureOverlap             float64 `json:"fixtureOverlap"`
	HelperOverlap              float64 `json:"helperOverlap"`
	SuitePathSimilarity        float64 `json:"suitePathSimilarity"`
	AssertionPatternSimilarity float64 `json:"assertionPatternSimilarity"`
}

// Similarity weights.
const (
	weightFixture   = 0.25
	weightHelper    = 0.25
	weightSuitePath = 0.30
	weightAssertion = 0.20

	similarityThreshold = 0.60

	// maxBlockSize caps how many members a single blocking key can have.
	// Blocks larger than this are too coarse to meaningfully identify duplicates
	// and would generate O(n²) candidate pairs. Tests sharing specific fixtures
	// or helpers will still be paired via their more specific blocking keys.
	maxBlockSize = 500

	// maxDuplicateTests is a hard safety cap for duplicate clustering.
	// Above this threshold, full pairwise-style clustering is too expensive for
	// interactive CLI usage and benchmark runs.
	maxDuplicateTests = 25000
)

// testFingerprint captures the structural signature of a test.
type testFingerprint struct {
	nodeID           string
	packageID        string
	fixtures         map[string]bool
	helpers          map[string]bool
	suitePath        []string
	assertionPattern string // normalized test name tokens
}

// DetectDuplicates finds structurally similar tests using fingerprinting,
// blocking-key candidate generation, weighted similarity scoring, and
// union-find clustering.
//
// Algorithm:
//  1. Build a fingerprint for each test (fixtures, helpers, suite path, name tokens)
//  2. Generate candidate pairs using blocking keys (shared package, fixture, helper)
//  3. Score each pair using weighted Jaccard + LCS similarity
//  4. Cluster pairs exceeding the threshold using union-find
func DetectDuplicates(g *Graph) DuplicateResult {
	tests := g.NodesByType(NodeTest)
	if len(tests) == 0 {
		return DuplicateResult{Clusters: []DuplicateCluster{}}
	}
	if len(tests) > maxDuplicateTests {
		return DuplicateResult{
			Clusters:      []DuplicateCluster{},
			TestsAnalyzed: len(tests),
			Skipped:       true,
			SkipReason:    fmt.Sprintf("skipped duplicate clustering for %d tests (limit %d)", len(tests), maxDuplicateTests),
		}
	}

	// Step 1: Build fingerprints.
	fps := buildFingerprints(g, tests)

	// Step 2: Generate candidate pairs.
	pairs := generateCandidates(fps)

	// Step 3: Score pairs and cluster.
	uf := newUnionFind(len(fps))
	pairScores := map[[2]int]float64{}
	pairSignals := map[[2]int]SimilaritySignals{}

	for _, pair := range pairs {
		i, j := pair[0], pair[1]

		// Skip pairs where neither test has fixtures or helpers — with
		// empty-set Jaccard = 0.0, max composite score is 0.50 < threshold.
		if len(fps[i].fixtures) == 0 && len(fps[j].fixtures) == 0 &&
			len(fps[i].helpers) == 0 && len(fps[j].helpers) == 0 {
			continue
		}

		// Skip pairs with zero overlap on both structural dimensions.
		if !hasSetOverlap(fps[i].fixtures, fps[j].fixtures) &&
			!hasSetOverlap(fps[i].helpers, fps[j].helpers) {
			continue
		}

		score, signals := scoreSimilarity(fps[i], fps[j])
		if score >= similarityThreshold {
			uf.union(i, j)
			pairScores[pair] = score
			pairSignals[pair] = signals
		}
	}

	// Step 4: Build clusters from union-find.
	clusterMap := map[int][]int{} // root → member indices
	for i := range fps {
		root := uf.find(i)
		clusterMap[root] = append(clusterMap[root], i)
	}

	var clusters []DuplicateCluster
	clusterID := 0
	dupCount := 0

	for _, members := range clusterMap {
		if len(members) < 2 {
			continue
		}

		// Average the pairwise scores and signals.
		var totalScore float64
		var totalSignals SimilaritySignals
		pairCount := 0

		for a := 0; a < len(members); a++ {
			for b := a + 1; b < len(members); b++ {
				key := [2]int{members[a], members[b]}
				if members[a] > members[b] {
					key = [2]int{members[b], members[a]}
				}
				if s, ok := pairScores[key]; ok {
					totalScore += s
					sig := pairSignals[key]
					totalSignals.FixtureOverlap += sig.FixtureOverlap
					totalSignals.HelperOverlap += sig.HelperOverlap
					totalSignals.SuitePathSimilarity += sig.SuitePathSimilarity
					totalSignals.AssertionPatternSimilarity += sig.AssertionPatternSimilarity
					pairCount++
				}
			}
		}

		if pairCount == 0 {
			continue
		}

		avgScore := totalScore / float64(pairCount)
		avgSignals := SimilaritySignals{
			FixtureOverlap:             totalSignals.FixtureOverlap / float64(pairCount),
			HelperOverlap:              totalSignals.HelperOverlap / float64(pairCount),
			SuitePathSimilarity:        totalSignals.SuitePathSimilarity / float64(pairCount),
			AssertionPatternSimilarity: totalSignals.AssertionPatternSimilarity / float64(pairCount),
		}

		testIDs := make([]string, len(members))
		for k, m := range members {
			testIDs[k] = fps[m].nodeID
		}
		sort.Strings(testIDs)

		clusters = append(clusters, DuplicateCluster{
			ID:         clusterID,
			Tests:      testIDs,
			Similarity: avgScore,
			Signals:    avgSignals,
		})
		clusterID++
		dupCount += len(members)
	}

	// Sort clusters by similarity descending.
	sort.Slice(clusters, func(i, j int) bool {
		return clusters[i].Similarity > clusters[j].Similarity
	})

	return DuplicateResult{
		Clusters:       clusters,
		TestsAnalyzed:  len(tests),
		DuplicateCount: dupCount,
	}
}

// buildFingerprints extracts structural signatures for each test.
func buildFingerprints(g *Graph, tests []*Node) []testFingerprint {
	fps := make([]testFingerprint, len(tests))

	// Build file→fixtures/helpers index.
	fileFixtures := map[string]map[string]bool{}
	fileHelpers := map[string]map[string]bool{}

	for _, e := range g.Edges() {
		switch e.Type {
		case EdgeTestUsesFixture:
			if fileFixtures[e.From] == nil {
				fileFixtures[e.From] = map[string]bool{}
			}
			fileFixtures[e.From][e.To] = true
		case EdgeTestUsesHelper:
			if fileHelpers[e.From] == nil {
				fileHelpers[e.From] = map[string]bool{}
			}
			fileHelpers[e.From][e.To] = true
		}
	}

	for i, test := range tests {
		fileID := "file:" + test.Path

		// Collect fixtures and helpers for this test's file.
		fixtures := map[string]bool{}
		for f := range fileFixtures[fileID] {
			fixtures[f] = true
		}
		helpers := map[string]bool{}
		for h := range fileHelpers[fileID] {
			helpers[h] = true
		}

		// Extract suite path from the graph.
		suitePath := extractSuitePath(g, test.ID)

		// Normalize test name into assertion pattern tokens.
		pattern := normalizeTestName(test.Name)

		fps[i] = testFingerprint{
			nodeID:           test.ID,
			packageID:        test.Package,
			fixtures:         fixtures,
			helpers:          helpers,
			suitePath:        suitePath,
			assertionPattern: pattern,
		}
	}

	return fps
}

// extractSuitePath walks reverse edges to find the suite hierarchy for a test.
func extractSuitePath(g *Graph, testID string) []string {
	// Test nodes are connected via TestDefinedInFile → file.
	// Suite nodes sit between file and test via SuiteContainsTest.
	// Walk incoming SuiteContainsTest edges to collect the suite chain.
	var path []string
	for _, e := range g.Incoming(testID) {
		if e.Type == EdgeSuiteContainsTest {
			suiteNode := g.Node(e.From)
			if suiteNode != nil {
				path = append(path, suiteNode.Name)
			}
		}
	}
	// Also check the test node itself for suite info.
	testNode := g.Node(testID)
	if testNode != nil && len(path) == 0 {
		// Fall back to inferring from the test ID format.
		// test:path:line:name → use file path as implicit suite.
		if testNode.Path != "" {
			path = []string{testNode.Path}
		}
	}
	return path
}

// normalizeTestName strips noise words, lowercases, and tokenizes a test name.
func normalizeTestName(name string) string {
	noise := map[string]bool{
		"should": true, "must": true, "can": true, "does": true,
		"will": true, "it": true, "the": true, "a": true, "an": true,
	}
	// Tokenize on non-alphanumeric characters.
	tokens := strings.FieldsFunc(strings.ToLower(name), func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))
	})
	// Filter noise words.
	filtered := make([]string, 0, len(tokens))
	for _, t := range tokens {
		if !noise[t] {
			filtered = append(filtered, t)
		}
	}
	sort.Strings(filtered)
	return strings.Join(filtered, " ")
}

// generateCandidates produces candidate pairs using blocking keys to avoid
// O(n²) comparison. Tests sharing a package, fixture, or helper are candidates.
func generateCandidates(fps []testFingerprint) [][2]int {
	blocks := map[string][]int{}

	for i, fp := range fps {
		// Block by package.
		if fp.packageID != "" {
			key := "pkg:" + fp.packageID
			blocks[key] = append(blocks[key], i)
		}
		// Block by shared fixture.
		for f := range fp.fixtures {
			key := "fix:" + f
			blocks[key] = append(blocks[key], i)
		}
		// Block by shared helper.
		for h := range fp.helpers {
			key := "hlp:" + h
			blocks[key] = append(blocks[key], i)
		}
	}

	// Emit unique pairs from each block.
	seen := map[[2]int]bool{}
	var pairs [][2]int

	for _, members := range blocks {
		if len(members) < 2 || len(members) > maxBlockSize {
			continue
		}
		for a := 0; a < len(members); a++ {
			for b := a + 1; b < len(members); b++ {
				i, j := members[a], members[b]
				if i > j {
					i, j = j, i
				}
				pair := [2]int{i, j}
				if !seen[pair] {
					seen[pair] = true
					pairs = append(pairs, pair)
				}
			}
		}
	}

	return pairs
}

// scoreSimilarity computes a weighted similarity score between two fingerprints.
func scoreSimilarity(a, b testFingerprint) (float64, SimilaritySignals) {
	fixtureScore := jaccardSets(a.fixtures, b.fixtures)
	helperScore := jaccardSets(a.helpers, b.helpers)
	suiteScore := lcsRatio(a.suitePath, b.suitePath)
	assertionScore := tokenJaccard(a.assertionPattern, b.assertionPattern)

	composite := weightFixture*fixtureScore +
		weightHelper*helperScore +
		weightSuitePath*suiteScore +
		weightAssertion*assertionScore

	return composite, SimilaritySignals{
		FixtureOverlap:             fixtureScore,
		HelperOverlap:              helperScore,
		SuitePathSimilarity:        suiteScore,
		AssertionPatternSimilarity: assertionScore,
	}
}

// hasSetOverlap returns true if sets a and b share at least one element.
// This is an O(min(|a|,|b|)) check used as a fast pre-filter.
func hasSetOverlap(a, b map[string]bool) bool {
	// Iterate over the smaller set for efficiency.
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

// jaccardSets computes Jaccard similarity between two sets.
func jaccardSets(a, b map[string]bool) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 0.0 // Both empty → no evidence of similarity.
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

// lcsRatio computes the LCS (longest common subsequence) ratio between
// two string slices.
func lcsRatio(a, b []string) float64 {
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

	// Standard LCS dynamic programming.
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

// tokenJaccard computes Jaccard similarity on space-separated tokens.
func tokenJaccard(a, b string) float64 {
	tokA := map[string]bool{}
	for _, t := range strings.Fields(a) {
		tokA[t] = true
	}
	tokB := map[string]bool{}
	for _, t := range strings.Fields(b) {
		tokB[t] = true
	}
	return jaccardSets(tokA, tokB)
}

// unionFind implements a disjoint-set data structure with rank-based
// path compression for efficient clustering.
type unionFind struct {
	parent []int
	rank   []int
}

func newUnionFind(n int) *unionFind {
	parent := make([]int, n)
	rank := make([]int, n)
	for i := range parent {
		parent[i] = i
	}
	return &unionFind{parent: parent, rank: rank}
}

func (uf *unionFind) find(x int) int {
	if uf.parent[x] != x {
		uf.parent[x] = uf.find(uf.parent[x]) // path compression
	}
	return uf.parent[x]
}

func (uf *unionFind) union(x, y int) {
	rx, ry := uf.find(x), uf.find(y)
	if rx == ry {
		return
	}
	if uf.rank[rx] < uf.rank[ry] {
		rx, ry = ry, rx
	}
	uf.parent[ry] = rx
	if uf.rank[rx] == uf.rank[ry] {
		uf.rank[rx]++
	}
}
