package runtime

import (
	"sort"
	"strings"

	"github.com/pmclSF/terrain/internal/models"
)

// ResolveTestIDs attempts to join runtime TestResults to the extracted
// TestCase inventory by matching file paths and test names.
//
// Resolution strategy (in priority order):
//  1. Exact file + name match: runtime file matches TestCase.FilePath and
//     runtime name matches TestCase.TestName
//  2. Suffix file + name match: runtime file path ends with TestCase.FilePath
//     (handles absolute vs relative path differences)
//  3. Name-only match within same file scope: when file matches but names
//     differ slightly (e.g., parameterized test instances)
//
// When resolution succeeds, TestResult.TestID is set. When it fails,
// TestID remains empty — downstream consumers must handle this gracefully.
//
// Returns the count of successfully resolved results.
func ResolveTestIDs(results []TestResult, testCases []models.TestCase) int {
	if len(results) == 0 || len(testCases) == 0 {
		return 0
	}

	// Build lookup indexes.
	type tcKey struct {
		file string
		name string
	}

	// Exact file+name → testID.
	exact := make(map[tcKey]string, len(testCases))
	// File → test cases in that file (for fuzzy name matching).
	byFile := make(map[string][]models.TestCase)

	for _, tc := range testCases {
		exact[tcKey{file: tc.FilePath, name: tc.TestName}] = tc.TestID
		byFile[tc.FilePath] = append(byFile[tc.FilePath], tc)
	}

	resolved := 0
	for i := range results {
		r := &results[i]

		// Strategy 1: exact file + name.
		if id, ok := exact[tcKey{file: r.File, name: r.Name}]; ok {
			r.TestID = id
			resolved++
			continue
		}

		// Strategy 2: suffix file match + exact name.
		// Collect all matches and sort for deterministic selection.
		var suffixMatches []string
		for filePath := range byFile {
			if pathSuffixMatch(r.File, filePath) {
				suffixMatches = append(suffixMatches, filePath)
			}
		}
		sort.Strings(suffixMatches)
		matchedFile := ""
		if len(suffixMatches) > 0 {
			matchedFile = suffixMatches[0]
		}
		if matchedFile != "" {
			if id, ok := exact[tcKey{file: matchedFile, name: r.Name}]; ok {
				r.TestID = id
				resolved++
				continue
			}
		}

		// Strategy 3: fuzzy name match within matched file.
		searchFile := r.File
		if matchedFile != "" {
			searchFile = matchedFile
		}
		if cases, ok := byFile[searchFile]; ok {
			if id := fuzzyNameMatch(r.Name, cases); id != "" {
				r.TestID = id
				resolved++
			}
		}
	}

	return resolved
}

// pathSuffixMatch returns true if runtimePath ends with extractedPath,
// accounting for path separator differences.
func pathSuffixMatch(runtimePath, extractedPath string) bool {
	if runtimePath == extractedPath {
		return true
	}
	// Normalize separators.
	rp := strings.ReplaceAll(runtimePath, "\\", "/")
	ep := strings.ReplaceAll(extractedPath, "\\", "/")
	return strings.HasSuffix(rp, "/"+ep)
}

// fuzzyNameMatch attempts to find a test case whose name is contained in
// or contains the runtime test name. Used for parameterized tests where
// the runtime name may include parameter values.
func fuzzyNameMatch(runtimeName string, cases []models.TestCase) string {
	// Prefer exact containment.
	for _, tc := range cases {
		if strings.Contains(runtimeName, tc.TestName) {
			return tc.TestID
		}
	}
	return ""
}
