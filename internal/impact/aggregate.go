package impact

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

	return agg
}
