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
