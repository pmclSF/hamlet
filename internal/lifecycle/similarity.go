package lifecycle

import (
	"path/filepath"
	"strings"
)

// stringSimilarity computes a normalized similarity score (0.0-1.0)
// between two strings using longest common subsequence ratio.
func stringSimilarity(a, b string) float64 {
	if a == b {
		return 1.0
	}
	if a == "" || b == "" {
		return 0.0
	}

	a = strings.ToLower(a)
	b = strings.ToLower(b)

	lcs := longestCommonSubsequence(a, b)
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}
	return float64(lcs) / float64(maxLen)
}

// longestCommonSubsequence returns the length of the LCS of two strings.
func longestCommonSubsequence(a, b string) int {
	m, n := len(a), len(b)
	// Use two rows to save memory.
	prev := make([]int, n+1)
	curr := make([]int, n+1)

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				curr[j] = prev[j-1] + 1
			} else {
				curr[j] = prev[j]
				if curr[j-1] > curr[j] {
					curr[j] = curr[j-1]
				}
			}
		}
		prev, curr = curr, prev
		// Reset curr for next iteration.
		for j := range curr {
			curr[j] = 0
		}
	}
	return prev[n]
}

// pathSimilarity computes similarity between two file paths,
// considering directory structure and filename.
func pathSimilarity(a, b string) float64 {
	if a == b {
		return 1.0
	}
	if a == "" || b == "" {
		return 0.0
	}

	// Normalize paths.
	a = filepath.ToSlash(a)
	b = filepath.ToSlash(b)

	// Compare directory and filename separately.
	dirA, fileA := filepath.Split(a)
	dirB, fileB := filepath.Split(b)

	dirSim := stringSimilarity(dirA, dirB)
	fileSim := stringSimilarity(fileA, fileB)

	// Weight filename more heavily than directory.
	return 0.4*dirSim + 0.6*fileSim
}

func tokenSimilarity(a, b string) float64 {
	aSet := tokenSet(a)
	bSet := tokenSet(b)
	if len(aSet) == 0 || len(bSet) == 0 {
		return 0.0
	}

	intersection := 0
	for tok := range aSet {
		if bSet[tok] {
			intersection++
		}
	}

	union := len(aSet)
	for tok := range bSet {
		if !aSet[tok] {
			union++
		}
	}
	if union == 0 {
		return 0.0
	}
	return float64(intersection) / float64(union)
}

var lifecycleStopTokens = map[string]bool{
	"test":   true,
	"tests":  true,
	"should": true,
	"it":     true,
	"when":   true,
	"then":   true,
	"and":    true,
	"the":    true,
	"a":      true,
	"an":     true,
}

func tokenSet(v string) map[string]bool {
	if v == "" {
		return nil
	}
	v = strings.ToLower(v)
	replacer := strings.NewReplacer(
		"-", " ",
		"_", " ",
		".", " ",
		":", " ",
		"/", " ",
		"(", " ",
		")", " ",
		"[", " ",
		"]", " ",
		"{", " ",
		"}", " ",
		",", " ",
	)
	v = replacer.Replace(v)

	out := map[string]bool{}
	for _, token := range strings.Fields(v) {
		token = strings.TrimSpace(token)
		if len(token) < 2 || lifecycleStopTokens[token] {
			continue
		}
		out[token] = true
	}
	return out
}
