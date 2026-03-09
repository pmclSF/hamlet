package lifecycle

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/pmclSF/hamlet/internal/identity"
	"github.com/pmclSF/hamlet/internal/models"
)

// InferContinuity analyzes two snapshots and classifies test continuity.
// It produces exact matches first, then applies heuristic matching for
// unmatched tests.
func InferContinuity(from, to *models.TestSuiteSnapshot) *ContinuityResult {
	result := &ContinuityResult{}
	if from == nil || to == nil {
		// If no previous snapshot, everything is "added"
		if to != nil {
			for _, tc := range to.TestCases {
				result.Mappings = append(result.Mappings, ContinuityMapping{
					ToTestID:    tc.TestID,
					ToCanonical: tc.CanonicalIdentity,
					ToPath:      tc.FilePath,
					Class:       ContinuityAdded,
					Confidence:  1.0,
					Evidence:    []EvidenceBasis{EvidenceExactID},
					Explanation: "new test, no prior snapshot",
				})
			}
			result.AddedCount = len(to.TestCases)
		}
		return result
	}

	// Index test cases by TestID for O(1) lookup.
	fromByID := indexTestCases(from.TestCases)
	toByID := indexTestCases(to.TestCases)

	// Phase 1: Exact continuity — same TestID in both snapshots.
	matchedFrom := map[string]bool{}
	matchedTo := map[string]bool{}

	// Sort IDs for deterministic iteration order.
	toIDs := make([]string, 0, len(toByID))
	for id := range toByID {
		toIDs = append(toIDs, id)
	}
	sort.Strings(toIDs)

	for _, id := range toIDs {
		toTC := toByID[id]
		if fromTC, ok := fromByID[id]; ok {
			result.Mappings = append(result.Mappings, ContinuityMapping{
				FromTestID:    fromTC.TestID,
				ToTestID:      toTC.TestID,
				FromCanonical: fromTC.CanonicalIdentity,
				ToCanonical:   toTC.CanonicalIdentity,
				FromPath:      fromTC.FilePath,
				ToPath:        toTC.FilePath,
				Class:         ContinuityExact,
				Confidence:    1.0,
				Evidence:      []EvidenceBasis{EvidenceExactID},
				Explanation:   "identical test ID across snapshots",
			})
			matchedFrom[id] = true
			matchedTo[id] = true
			result.ExactCount++
		}
	}

	// Collect unmatched tests.
	var unmatchedFrom []models.TestCase
	for id, tc := range fromByID {
		if !matchedFrom[id] {
			unmatchedFrom = append(unmatchedFrom, tc)
		}
	}
	sort.Slice(unmatchedFrom, func(i, j int) bool {
		return unmatchedFrom[i].TestID < unmatchedFrom[j].TestID
	})

	var unmatchedTo []models.TestCase
	for id, tc := range toByID {
		if !matchedTo[id] {
			unmatchedTo = append(unmatchedTo, tc)
		}
	}
	sort.Slice(unmatchedTo, func(i, j int) bool {
		return unmatchedTo[i].TestID < unmatchedTo[j].TestID
	})

	// Phase 2: Heuristic matching for unmatched tests.
	heuristicMappings := inferHeuristicContinuity(unmatchedFrom, unmatchedTo)

	// Track which unmatched tests got heuristic matches.
	heuristicFrom := map[string]bool{}
	heuristicTo := map[string]bool{}
	for _, m := range heuristicMappings {
		result.Mappings = append(result.Mappings, m)
		if m.FromTestID != "" {
			heuristicFrom[m.FromTestID] = true
		}
		if m.ToTestID != "" {
			heuristicTo[m.ToTestID] = true
		}
		switch m.Class {
		case ContinuityRename:
			result.RenameCount++
		case ContinuityMove:
			result.MoveCount++
		case ContinuitySplit:
			result.SplitCount++
		case ContinuityMerge:
			result.MergeCount++
		case ContinuityAmbiguous:
			result.AmbiguousCount++
		}
	}

	// Phase 3: Remaining unmatched from = removed, unmatched to = added.
	for _, tc := range unmatchedFrom {
		if !heuristicFrom[tc.TestID] {
			result.Mappings = append(result.Mappings, ContinuityMapping{
				FromTestID:    tc.TestID,
				FromCanonical: tc.CanonicalIdentity,
				FromPath:      tc.FilePath,
				Class:         ContinuityRemoved,
				Confidence:    1.0,
				Evidence:      []EvidenceBasis{EvidenceExactID},
				Explanation:   "test no longer present in current snapshot",
			})
			result.RemovedCount++
		}
	}
	for _, tc := range unmatchedTo {
		if !heuristicTo[tc.TestID] {
			result.Mappings = append(result.Mappings, ContinuityMapping{
				ToTestID:    tc.TestID,
				ToCanonical: tc.CanonicalIdentity,
				ToPath:      tc.FilePath,
				Class:       ContinuityAdded,
				Confidence:  1.0,
				Evidence:    []EvidenceBasis{EvidenceExactID},
				Explanation: "new test not present in previous snapshot",
			})
			result.AddedCount++
		}
	}

	// Sort mappings for determinism: by class, then by ToTestID, then FromTestID.
	sort.Slice(result.Mappings, func(i, j int) bool {
		if result.Mappings[i].Class != result.Mappings[j].Class {
			return result.Mappings[i].Class < result.Mappings[j].Class
		}
		if result.Mappings[i].ToTestID != result.Mappings[j].ToTestID {
			return result.Mappings[i].ToTestID < result.Mappings[j].ToTestID
		}
		return result.Mappings[i].FromTestID < result.Mappings[j].FromTestID
	})

	return result
}

// inferHeuristicContinuity attempts to match unmatched tests using
// name similarity, path similarity, and suite hierarchy.
func inferHeuristicContinuity(unmatchedFrom, unmatchedTo []models.TestCase) []ContinuityMapping {
	if len(unmatchedFrom) == 0 || len(unmatchedTo) == 0 {
		return nil
	}

	var mappings []ContinuityMapping

	// Score all candidate pairs.
	type candidate struct {
		from     models.TestCase
		to       models.TestCase
		score    float64
		evidence []EvidenceBasis
		class    ContinuityClass
	}

	var candidates []candidate

	for _, f := range unmatchedFrom {
		for _, t := range unmatchedTo {
			score, evidence, class := scorePair(f, t)
			if score >= 0.4 { // Minimum threshold for consideration
				candidates = append(candidates, candidate{
					from:     f,
					to:       t,
					score:    score,
					evidence: evidence,
					class:    class,
				})
			}
		}
	}

	// Sort candidates by score descending for greedy matching.
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	// Greedy 1:1 matching — each test can only be matched once.
	usedFrom := map[string]bool{}
	usedTo := map[string]bool{}

	for _, c := range candidates {
		if usedFrom[c.from.TestID] || usedTo[c.to.TestID] {
			continue
		}

		confidence := c.score
		if confidence > 1.0 {
			confidence = 1.0
		}

		explanation := buildExplanation(c.class, c.from, c.to)

		mappings = append(mappings, ContinuityMapping{
			FromTestID:    c.from.TestID,
			ToTestID:      c.to.TestID,
			FromCanonical: c.from.CanonicalIdentity,
			ToCanonical:   c.to.CanonicalIdentity,
			FromPath:      c.from.FilePath,
			ToPath:        c.to.FilePath,
			Class:         c.class,
			Confidence:    confidence,
			Evidence:      c.evidence,
			Explanation:   explanation,
		})

		usedFrom[c.from.TestID] = true
		usedTo[c.to.TestID] = true
	}

	// Check for split patterns: one old test -> multiple new tests with same name prefix.
	// Check for merge patterns: multiple old tests -> one new test.
	splitMerge := detectSplitMerge(unmatchedFrom, unmatchedTo, usedFrom, usedTo)
	mappings = append(mappings, splitMerge...)

	return mappings
}

// scorePair computes a similarity score between two test cases and
// determines the most likely continuity class.
func scorePair(from, to models.TestCase) (float64, []EvidenceBasis, ContinuityClass) {
	var score float64
	var evidence []EvidenceBasis

	// Name similarity (strongest heuristic signal).
	nameSim := stringSimilarity(from.TestName, to.TestName)
	if nameSim >= 0.8 {
		score += 0.4
		evidence = append(evidence, EvidenceNameSimilar)
	} else if nameSim >= 0.5 {
		score += 0.2
		evidence = append(evidence, EvidenceNameSimilar)
	}

	// Suite hierarchy similarity.
	suiteSim := stringSimilarity(
		strings.Join(from.SuiteHierarchy, " > "),
		strings.Join(to.SuiteHierarchy, " > "),
	)
	if suiteSim >= 0.8 {
		score += 0.2
		evidence = append(evidence, EvidenceSuiteHierarchy)
	} else if suiteSim >= 0.5 {
		score += 0.1
		evidence = append(evidence, EvidenceSuiteHierarchy)
	}

	// Path similarity.
	pathSim := pathSimilarity(from.FilePath, to.FilePath)
	if pathSim >= 0.8 {
		score += 0.2
		evidence = append(evidence, EvidencePathSimilar)
	} else if pathSim >= 0.5 {
		score += 0.1
		evidence = append(evidence, EvidencePathSimilar)
	}

	// Canonical identity similarity (combines all components).
	canonSim := stringSimilarity(from.CanonicalIdentity, to.CanonicalIdentity)
	if canonSim >= 0.7 {
		score += 0.2
		evidence = append(evidence, EvidenceCanonicalSimilar)
	}

	// Determine class based on what changed.
	class := classifyChange(from, to, nameSim, pathSim)

	return score, evidence, class
}

// classifyChange determines the continuity class based on what changed
// between from and to test cases.
func classifyChange(from, to models.TestCase, nameSim, pathSim float64) ContinuityClass {
	samePath := from.FilePath == to.FilePath
	sameName := from.TestName == to.TestName
	sameDir := filepath.Dir(from.FilePath) == filepath.Dir(to.FilePath)

	switch {
	case samePath && nameSim >= 0.7 && !sameName:
		// Same file, similar but not identical name -> rename
		return ContinuityRename
	case !samePath && sameName:
		// Different file, same name -> move
		return ContinuityMove
	case !samePath && nameSim >= 0.7:
		// Different file, similar name -> rename+move, classify as move
		if sameDir {
			return ContinuityRename
		}
		return ContinuityMove
	case pathSim >= 0.5 && nameSim >= 0.5:
		// Moderate similarity on both dimensions
		if nameSim >= 0.7 {
			return ContinuityRename
		}
		return ContinuityAmbiguous
	default:
		return ContinuityAmbiguous
	}
}

// detectSplitMerge looks for 1:N (split) and N:1 (merge) patterns among
// remaining unmatched tests.
func detectSplitMerge(unmatchedFrom, unmatchedTo []models.TestCase, usedFrom, usedTo map[string]bool) []ContinuityMapping {
	var mappings []ContinuityMapping

	// Check for splits: one old test name appears in multiple new test names.
	for _, f := range unmatchedFrom {
		if usedFrom[f.TestID] {
			continue
		}
		var splitTargets []models.TestCase
		for _, t := range unmatchedTo {
			if usedTo[t.TestID] {
				continue
			}
			// Split detection: new test name contains old test name as prefix,
			// and they share the same file or directory.
			if strings.HasPrefix(identity.NormalizeName(t.TestName), identity.NormalizeName(f.TestName)) &&
				(t.FilePath == f.FilePath || filepath.Dir(t.FilePath) == filepath.Dir(f.FilePath)) {
				splitTargets = append(splitTargets, t)
			}
		}
		if len(splitTargets) >= 2 {
			for _, t := range splitTargets {
				mappings = append(mappings, ContinuityMapping{
					FromTestID:    f.TestID,
					ToTestID:      t.TestID,
					FromCanonical: f.CanonicalIdentity,
					ToCanonical:   t.CanonicalIdentity,
					FromPath:      f.FilePath,
					ToPath:        t.FilePath,
					Class:         ContinuitySplit,
					Confidence:    0.6,
					Evidence:      []EvidenceBasis{EvidenceNameSimilar, EvidencePathSimilar},
					Explanation:   f.TestName + " appears to have been split into multiple tests",
				})
				usedTo[t.TestID] = true
			}
			usedFrom[f.TestID] = true
		}
	}

	// Check for merges: multiple old test names share a prefix that matches
	// a new test name.
	for _, t := range unmatchedTo {
		if usedTo[t.TestID] {
			continue
		}
		var mergeSourcesForThis []models.TestCase
		for _, f := range unmatchedFrom {
			if usedFrom[f.TestID] {
				continue
			}
			if strings.HasPrefix(identity.NormalizeName(f.TestName), identity.NormalizeName(t.TestName)) &&
				(f.FilePath == t.FilePath || filepath.Dir(f.FilePath) == filepath.Dir(t.FilePath)) {
				mergeSourcesForThis = append(mergeSourcesForThis, f)
			}
		}
		if len(mergeSourcesForThis) >= 2 {
			for _, f := range mergeSourcesForThis {
				mappings = append(mappings, ContinuityMapping{
					FromTestID:    f.TestID,
					ToTestID:      t.TestID,
					FromCanonical: f.CanonicalIdentity,
					ToCanonical:   t.CanonicalIdentity,
					FromPath:      f.FilePath,
					ToPath:        t.FilePath,
					Class:         ContinuityMerge,
					Confidence:    0.6,
					Evidence:      []EvidenceBasis{EvidenceNameSimilar, EvidencePathSimilar},
					Explanation:   "multiple tests appear to have been merged into " + t.TestName,
				})
				usedFrom[f.TestID] = true
			}
			usedTo[t.TestID] = true
		}
	}

	return mappings
}

// indexTestCases builds a map from TestID to TestCase.
func indexTestCases(cases []models.TestCase) map[string]models.TestCase {
	m := make(map[string]models.TestCase, len(cases))
	for _, tc := range cases {
		if tc.TestID != "" {
			m[tc.TestID] = tc
		}
	}
	return m
}

// buildExplanation creates a human-readable explanation for a mapping.
func buildExplanation(class ContinuityClass, from, to models.TestCase) string {
	switch class {
	case ContinuityRename:
		return from.TestName + " appears renamed to " + to.TestName
	case ContinuityMove:
		return from.TestName + " appears moved from " + from.FilePath + " to " + to.FilePath
	case ContinuityAmbiguous:
		return "ambiguous relationship between " + from.TestName + " and " + to.TestName
	default:
		return string(class) + " relationship inferred"
	}
}
