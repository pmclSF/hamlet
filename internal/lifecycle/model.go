package lifecycle

// ContinuityClass classifies the relationship between tests across snapshots.
type ContinuityClass string

const (
	ContinuityExact     ContinuityClass = "exact_continuity"
	ContinuityRename    ContinuityClass = "likely_rename"
	ContinuityMove      ContinuityClass = "likely_move"
	ContinuitySplit     ContinuityClass = "likely_split"
	ContinuityMerge     ContinuityClass = "likely_merge"
	ContinuityRemoved   ContinuityClass = "removed"
	ContinuityAdded     ContinuityClass = "added"
	ContinuityAmbiguous ContinuityClass = "ambiguous"
)

// EvidenceBasis describes what evidence supports a continuity inference.
type EvidenceBasis string

const (
	EvidenceExactID            EvidenceBasis = "exact_id_match"
	EvidenceCanonicalSimilar   EvidenceBasis = "canonical_similarity"
	EvidenceSuiteHierarchy     EvidenceBasis = "suite_hierarchy_match"
	EvidencePathSimilar        EvidenceBasis = "path_similarity"
	EvidenceCoverageContinuity EvidenceBasis = "coverage_continuity"
	EvidenceNameSimilar        EvidenceBasis = "name_similarity"
)

// ContinuityMapping represents the inferred relationship between a test in
// the "from" snapshot and a test in the "to" snapshot.
type ContinuityMapping struct {
	// FromTestID is the test ID in the previous snapshot ("" for added tests).
	FromTestID string
	// ToTestID is the test ID in the current snapshot ("" for removed tests).
	ToTestID string
	// Class is the continuity classification.
	Class ContinuityClass
	// Confidence is 0.0-1.0 reflecting inference strength.
	Confidence float64
	// Evidence lists the basis for this classification.
	Evidence []EvidenceBasis
	// Explanation is a human-readable description of the mapping.
	Explanation string

	// FromCanonical is the canonical identity from the previous snapshot.
	FromCanonical string
	// ToCanonical is the canonical identity in the current snapshot.
	ToCanonical string
	// FromPath is the file path in the previous snapshot.
	FromPath string
	// ToPath is the file path in the current snapshot.
	ToPath string
}

// ContinuityResult holds the complete lifecycle analysis between two snapshots.
type ContinuityResult struct {
	Mappings       []ContinuityMapping
	ExactCount     int
	RenameCount    int
	MoveCount      int
	SplitCount     int
	MergeCount     int
	RemovedCount   int
	AddedCount     int
	AmbiguousCount int
}

// IsHeuristic returns true if this mapping is based on heuristic inference
// rather than exact identity match.
func (m ContinuityMapping) IsHeuristic() bool {
	return m.Class != ContinuityExact
}
