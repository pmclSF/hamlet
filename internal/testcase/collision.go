package testcase

import (
	"fmt"
	"sort"

	"github.com/pmclSF/hamlet/internal/identity"
	"github.com/pmclSF/hamlet/internal/models"
)

// CollisionDiagnostic reports a detected identity collision.
type CollisionDiagnostic struct {
	// CanonicalIdentity is the colliding identity string.
	CanonicalIdentity string `json:"canonicalIdentity"`

	// TestID is the colliding hash.
	TestID string `json:"testId"`

	// Occurrences lists the colliding test cases by file and line.
	Occurrences []CollisionOccurrence `json:"occurrences"`

	// Resolution describes how the collision was resolved.
	Resolution string `json:"resolution"`
}

// CollisionOccurrence identifies one participant in a collision.
type CollisionOccurrence struct {
	FilePath string `json:"filePath"`
	Line     int    `json:"line"`
	TestName string `json:"testName"`
}

// DetectAndResolveCollisions finds canonical identity collisions in a set of
// test cases and applies deterministic disambiguation.
//
// Returns the deduplicated test cases and any collision diagnostics.
func DetectAndResolveCollisions(cases []models.TestCase) ([]models.TestCase, []CollisionDiagnostic) {
	// Group by canonical identity.
	groups := map[string][]int{}
	for i, tc := range cases {
		groups[tc.CanonicalIdentity] = append(groups[tc.CanonicalIdentity], i)
	}

	var diagnostics []CollisionDiagnostic
	result := make([]models.TestCase, 0, len(cases))
	seen := map[int]bool{}

	// Sort keys for deterministic output.
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, canonical := range keys {
		indices := groups[canonical]
		if len(indices) == 1 {
			result = append(result, cases[indices[0]])
			seen[indices[0]] = true
			continue
		}

		// Collision detected. Disambiguate by appending line-based suffix.
		// Sort by line number for deterministic disambiguation order.
		sort.Slice(indices, func(a, b int) bool {
			return cases[indices[a]].Line < cases[indices[b]].Line
		})

		var occurrences []CollisionOccurrence
		for disambigIdx, origIdx := range indices {
			tc := cases[origIdx]
			occurrences = append(occurrences, CollisionOccurrence{
				FilePath: tc.FilePath,
				Line:     tc.Line,
				TestName: tc.TestName,
			})

			// Disambiguate by appending #N suffix to the canonical identity.
			if disambigIdx > 0 {
				tc.CanonicalIdentity = fmt.Sprintf("%s#%d", canonical, disambigIdx+1)
				tc.TestID = identity.GenerateID(tc.CanonicalIdentity)
			}
			result = append(result, tc)
			seen[origIdx] = true
		}

		diagnostics = append(diagnostics, CollisionDiagnostic{
			CanonicalIdentity: canonical,
			TestID:            cases[indices[0]].TestID,
			Occurrences:       occurrences,
			Resolution:        fmt.Sprintf("disambiguated %d collisions with line-based suffix", len(indices)),
		})
	}

	return result, diagnostics
}
