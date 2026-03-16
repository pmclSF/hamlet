// Package truthcheck validates Terrain's analysis output against a
// ground truth specification. It computes precision, recall, and overlap
// for each analysis category (impact, coverage, redundancy, fanout,
// stability, AI, environment).
//
// The truth spec is a YAML file (typically tests/truth/terrain_truth.yaml)
// that documents expected findings. The checker runs Terrain's pipeline
// and compares actual output to expectations.
package truthcheck

// TruthSpec is the top-level truth specification loaded from YAML.
type TruthSpec struct {
	Impact      *ImpactTruth      `yaml:"impact"`
	Coverage    *CoverageTruth    `yaml:"coverage"`
	Redundancy  *RedundancyTruth  `yaml:"redundancy"`
	Fanout      *FanoutTruth      `yaml:"fanout"`
	Stability   *StabilityTruth   `yaml:"stability"`
	AI          *AITruth          `yaml:"ai"`
	Environment *EnvironmentTruth `yaml:"environment"`
}

// ImpactTruth defines expected impact analysis results.
type ImpactTruth struct {
	Description string        `yaml:"description"`
	Cases       []ImpactCase  `yaml:"cases"`
}

// ImpactCase defines expected impact for a single file change.
type ImpactCase struct {
	Change                   string   `yaml:"change"`
	ExpectedImpactedTests    []string `yaml:"expected_impacted_tests"`
	ExpectedImpactedScenarios []string `yaml:"expected_impacted_scenarios"`
	ExpectedMinImpacted      int      `yaml:"expected_min_impacted"`
	Notes                    string   `yaml:"notes"`
}

// CoverageTruth defines expected coverage gaps.
type CoverageTruth struct {
	Description      string          `yaml:"description"`
	ExpectedUncovered []CoverageItem `yaml:"expected_uncovered"`
	ExpectedWeak     []CoverageItem  `yaml:"expected_weak"`
}

// CoverageItem is an expected uncovered or weakly-covered file.
type CoverageItem struct {
	Path   string `yaml:"path"`
	Reason string `yaml:"reason"`
}

// RedundancyTruth defines expected duplicate test clusters.
type RedundancyTruth struct {
	Description      string             `yaml:"description"`
	ExpectedClusters []RedundancyCluster `yaml:"expected_clusters"`
}

// RedundancyCluster is an expected group of redundant tests.
type RedundancyCluster struct {
	Tests  []string `yaml:"tests"`
	Reason string   `yaml:"reason"`
}

// FanoutTruth defines expected high-fanout nodes.
type FanoutTruth struct {
	Description     string       `yaml:"description"`
	ExpectedFlagged []FanoutNode `yaml:"expected_flagged"`
}

// FanoutNode is an expected high-fanout node.
type FanoutNode struct {
	Node                  string `yaml:"node"`
	Reason                string `yaml:"reason"`
	ExpectedMinDependents int    `yaml:"expected_min_dependents"`
}

// StabilityTruth defines expected stability signals.
type StabilityTruth struct {
	Description     string         `yaml:"description"`
	ExpectedSkipped []SkippedEntry `yaml:"expected_skipped"`
	Notes           string         `yaml:"notes"`
}

// SkippedEntry is an expected skipped test file.
type SkippedEntry struct {
	File   string `yaml:"file"`
	Count  int    `yaml:"count"`
	Reason string `yaml:"reason"`
}

// AITruth defines expected AI/scenario analysis results.
type AITruth struct {
	Description                string             `yaml:"description"`
	ExpectedScenarios          int                `yaml:"expected_scenarios"`
	ExpectedScenarioDuplication []ScenarioDupEntry `yaml:"expected_scenario_duplication"`
	ExpectedPromptSurfaces     []string           `yaml:"expected_prompt_surfaces"`
	ExpectedDatasetSurfaces    []string           `yaml:"expected_dataset_surfaces"`
}

// ScenarioDupEntry is an expected scenario duplication pair.
type ScenarioDupEntry struct {
	Pair    []string `yaml:"pair"`
	Overlap string   `yaml:"overlap"`
	Reason  string   `yaml:"reason"`
}

// EnvironmentTruth defines expected environment analysis.
type EnvironmentTruth struct {
	Description string `yaml:"description"`
	Notes       string `yaml:"notes"`
}

// TruthCheckReport is the complete validation result.
type TruthCheckReport struct {
	RepoRoot   string                `json:"repoRoot"`
	TruthFile  string                `json:"truthFile"`
	Categories []TruthCategoryResult `json:"categories"`
	Summary    ReportSummary         `json:"summary"`
}

// TruthCategoryResult is the validation result for one truth category.
type TruthCategoryResult struct {
	Category    string   `json:"category"`
	Description string   `json:"description,omitempty"`
	Passed      bool     `json:"passed"`
	Score       float64  `json:"score"`       // 0.0-1.0
	Precision   float64  `json:"precision"`   // correct / (correct + unexpected)
	Recall      float64  `json:"recall"`      // correct / (correct + missing)
	Expected    int      `json:"expected"`    // total expected items
	Found       int      `json:"found"`       // total found items
	Matched     int      `json:"matched"`     // correctly matched
	Missing     []string `json:"missing,omitempty"`    // expected but not found
	Unexpected  []string `json:"unexpected,omitempty"` // found but not expected
	Details     []string `json:"details,omitempty"`    // per-check notes
}

// ReportSummary aggregates scores across categories.
type ReportSummary struct {
	TotalCategories int     `json:"totalCategories"`
	PassedCount     int     `json:"passedCount"`
	OverallScore    float64 `json:"overallScore"` // average score
	OverallPrecision float64 `json:"overallPrecision"`
	OverallRecall   float64 `json:"overallRecall"`
}
