package impact

// PrivacyThreshold is the minimum count required for a bucket to be included
// in a privacy-safe aggregate. Buckets below this threshold are suppressed
// to prevent identification of individual files, tests, or owners.
const PrivacyThreshold = 3

// Aggregate contains privacy-safe aggregate impact statistics.
// No raw file paths, symbol names, or source code — only counts and ratios.
type Aggregate struct {
	// ChangedFileCount is the number of files in the change scope.
	ChangedFileCount int `json:"changedFileCount"`

	// ChangedTestFileCount is the number of test files that were directly changed.
	ChangedTestFileCount int `json:"changedTestFileCount"`

	// ImpactedUnitCount is the number of code units affected.
	ImpactedUnitCount int `json:"impactedUnitCount"`

	// ExportedUnitCount is the number of affected exported/public units.
	ExportedUnitCount int `json:"exportedUnitCount"`

	// ProtectionCounts breaks down impacted units by protection status.
	ProtectionCounts map[string]int `json:"protectionCounts"`

	// GapCount is the total number of protection gaps.
	GapCount int `json:"gapCount"`

	// HighSeverityGapCount is the number of high-severity gaps.
	HighSeverityGapCount int `json:"highSeverityGapCount"`

	// ImpactedTestCount is the number of tests relevant to the change.
	ImpactedTestCount int `json:"impactedTestCount"`

	// SelectedTestCount is the number of tests in the recommended set.
	SelectedTestCount int `json:"selectedTestCount"`

	// OwnerCount is the number of distinct owners affected.
	OwnerCount int `json:"ownerCount"`

	// Posture is the change-risk posture band.
	Posture string `json:"posture"`

	// ConfidenceCounts breaks down impact mappings by confidence level.
	ConfidenceCounts map[string]int `json:"confidenceCounts"`

	// ProtectionRatio is the ratio of protected units to total impacted units.
	// Only included when ImpactedUnitCount >= PrivacyThreshold.
	ProtectionRatio float64 `json:"protectionRatio,omitempty"`

	// ExactConfidenceRatio is the ratio of exact-confidence mappings.
	// Only included when ImpactedUnitCount >= PrivacyThreshold.
	ExactConfidenceRatio float64 `json:"exactConfidenceRatio,omitempty"`

	// SelectionSetKind is the kind of protective test set selected.
	SelectionSetKind string `json:"selectionSetKind,omitempty"`

	// GraphStats contains privacy-safe impact graph statistics.
	GraphStats *GraphStats `json:"graphStats,omitempty"`

	// IsSparse indicates the aggregate is based on limited data
	// and some fields were suppressed for privacy.
	IsSparse bool `json:"isSparse,omitempty"`
}

// BuildAggregate creates a privacy-safe aggregate from an ImpactResult.
func BuildAggregate(result *ImpactResult) *Aggregate {
	agg := &Aggregate{
		ChangedFileCount:  len(result.Scope.ChangedFiles),
		ImpactedUnitCount: len(result.ImpactedUnits),
		GapCount:          len(result.ProtectionGaps),
		ImpactedTestCount: len(result.ImpactedTests),
		SelectedTestCount: len(result.SelectedTests),
		OwnerCount:        len(result.ImpactedOwners),
		Posture:           result.Posture.Band,
		ProtectionCounts:  map[string]int{},
		ConfidenceCounts:  map[string]int{},
	}

	for _, cf := range result.Scope.ChangedFiles {
		if cf.IsTestFile {
			agg.ChangedTestFileCount++
		}
	}

	for _, iu := range result.ImpactedUnits {
		agg.ProtectionCounts[string(iu.ProtectionStatus)]++
		agg.ConfidenceCounts[string(iu.ImpactConfidence)]++
		if iu.Exported {
			agg.ExportedUnitCount++
		}
	}

	for _, gap := range result.ProtectionGaps {
		if gap.Severity == "high" {
			agg.HighSeverityGapCount++
		}
	}

	// Compute ratios only when above privacy threshold.
	if agg.ImpactedUnitCount >= PrivacyThreshold {
		protected := agg.ProtectionCounts["strong"] + agg.ProtectionCounts["partial"]
		agg.ProtectionRatio = float64(protected) / float64(agg.ImpactedUnitCount)

		exact := agg.ConfidenceCounts["exact"]
		agg.ExactConfidenceRatio = float64(exact) / float64(agg.ImpactedUnitCount)
	} else if agg.ImpactedUnitCount > 0 {
		agg.IsSparse = true
	}

	// Include protective set kind.
	if result.ProtectiveSet != nil {
		agg.SelectionSetKind = result.ProtectiveSet.SetKind
	}

	// Include graph stats (already privacy-safe — only counts).
	if result.Graph != nil {
		stats := result.Graph.Stats
		agg.GraphStats = &stats
	}

	// Apply privacy suppression to small buckets.
	applyPrivacySuppression(agg)

	return agg
}

// applyPrivacySuppression zeroes out breakdown fields that could
// identify individual entities when counts are below threshold.
func applyPrivacySuppression(agg *Aggregate) {
	// Suppress protection breakdown if total is below threshold.
	if agg.ImpactedUnitCount < PrivacyThreshold {
		agg.ProtectionCounts = map[string]int{}
		agg.ConfidenceCounts = map[string]int{}
		agg.IsSparse = true
	}

	// Suppress owner count if it could identify specific teams.
	if agg.OwnerCount > 0 && agg.OwnerCount < PrivacyThreshold {
		// Keep the count but mark as sparse.
		agg.IsSparse = true
	}
}
