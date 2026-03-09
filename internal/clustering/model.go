package clustering

// ClusterType classifies the kind of common-cause cluster.
type ClusterType string

const (
	ClusterDominantSlowHelper   ClusterType = "dominant_slow_helper"
	ClusterDominantFlakyFixture ClusterType = "dominant_flaky_fixture"
	ClusterGlobalSetupPath      ClusterType = "global_setup_path"
	ClusterSharedImport         ClusterType = "shared_import_dependency"
	ClusterRepeatedFailPattern  ClusterType = "repeated_failure_pattern"
)

// Cluster represents a detected common-cause cluster.
type Cluster struct {
	// Type is the cluster classification.
	Type ClusterType `json:"type"`
	// CausePath is the file/module/function believed to be the root cause.
	CausePath string `json:"causePath"`
	// AffectedTests lists test file paths affected by this cause.
	AffectedTests []string `json:"affectedTests"`
	// AffectedCount is len(AffectedTests).
	AffectedCount int `json:"affectedCount"`
	// Confidence is 0.0-1.0 reflecting how confident we are in the causal link.
	Confidence float64 `json:"confidence"`
	// Evidence describes the basis for this cluster.
	Evidence string `json:"evidence"`
	// Explanation is a human-readable summary.
	Explanation string `json:"explanation"`
	// ImpactMetric quantifies impact (e.g., total runtime affected, failure count).
	ImpactMetric float64 `json:"impactMetric"`
	// ImpactUnit describes what ImpactMetric measures.
	ImpactUnit string `json:"impactUnit"`
}

// ClusterResult holds all detected clusters.
type ClusterResult struct {
	Clusters []Cluster `json:"clusters"`
	// TotalAffectedTests is the number of unique test files affected by any cluster.
	TotalAffectedTests int `json:"totalAffectedTests"`
}
