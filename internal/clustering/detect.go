package clustering

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/pmclSF/terrain/internal/models"
	"github.com/pmclSF/terrain/internal/signals"
)

// minClusterSize is the minimum number of affected tests required to form a cluster.
const minClusterSize = 3

// Detect analyzes a TestSuiteSnapshot and returns common-cause clusters where
// shared helpers, fixtures, or setup paths appear responsible for broad
// instability, slowness, or similar signal patterns.
//
// The approach is candidate-oriented: each cluster represents a hypothesis
// about a shared cause, not a proven root cause.
func Detect(snap *models.TestSuiteSnapshot) *ClusterResult {
	if snap == nil {
		return &ClusterResult{}
	}

	var clusters []Cluster

	clusters = append(clusters, detectSharedImportClusters(snap)...)
	clusters = append(clusters, detectSlowPathClusters(snap)...)
	clusters = append(clusters, detectFlakyFixtureClusters(snap)...)
	clusters = append(clusters, detectSetupPathClusters(snap)...)
	clusters = append(clusters, detectRepeatedFailPatterns(snap)...)

	// Sort clusters by affected count descending for deterministic output.
	sort.Slice(clusters, func(i, j int) bool {
		if clusters[i].AffectedCount != clusters[j].AffectedCount {
			return clusters[i].AffectedCount > clusters[j].AffectedCount
		}
		return clusters[i].CausePath < clusters[j].CausePath
	})

	totalAffected := countUniqueAffectedTests(clusters)

	return &ClusterResult{
		Clusters:           clusters,
		TotalAffectedTests: totalAffected,
	}
}

// detectSharedImportClusters groups test files by their LinkedCodeUnits.
// If many test files share the same code unit, that code unit is a potential
// common cause for any broad pattern (instability, slowness, etc.).
func detectSharedImportClusters(snap *models.TestSuiteSnapshot) []Cluster {
	// Build map: code unit -> test files that link to it.
	unitToTests := make(map[string][]string)
	for _, tf := range snap.TestFiles {
		for _, cu := range tf.LinkedCodeUnits {
			unitToTests[cu] = append(unitToTests[cu], tf.Path)
		}
	}

	var clusters []Cluster
	for unit, tests := range unitToTests {
		if len(tests) < minClusterSize {
			continue
		}
		sorted := sortedCopy(tests)
		confidence := sharedImportConfidence(len(tests), len(snap.TestFiles))
		clusters = append(clusters, Cluster{
			Type:          ClusterSharedImport,
			CausePath:     unit,
			AffectedTests: sorted,
			AffectedCount: len(sorted),
			Confidence:    confidence,
			Evidence:      fmt.Sprintf("%d test files link to code unit %q", len(sorted), unit),
			Explanation: fmt.Sprintf(
				"Code unit %q is a shared dependency of %d test files. "+
					"Changes or instability in this unit may have broad impact across the test suite.",
				unit, len(sorted),
			),
			ImpactMetric: float64(len(sorted)),
			ImpactUnit:   "affected_test_files",
		})
	}
	return clusters
}

// detectSlowPathClusters identifies code units shared by multiple slow tests.
func detectSlowPathClusters(snap *models.TestSuiteSnapshot) []Cluster {
	slowTests := testFilesWithSignalType(snap, signals.SignalSlowTest)
	if len(slowTests) < minClusterSize {
		return nil
	}

	unitToSlowTests := buildUnitToTestMap(slowTests, snap)

	var clusters []Cluster
	for unit, tests := range unitToSlowTests {
		if len(tests) < minClusterSize {
			continue
		}
		sorted := sortedCopy(tests)
		totalRuntime := sumRuntime(sorted, snap)
		confidence := clamp(float64(len(sorted))/float64(len(slowTests)), 0.3, 0.95)
		clusters = append(clusters, Cluster{
			Type:          ClusterDominantSlowHelper,
			CausePath:     unit,
			AffectedTests: sorted,
			AffectedCount: len(sorted),
			Confidence:    confidence,
			Evidence: fmt.Sprintf(
				"%d slow tests share code unit %q; total avg runtime: %.0fms",
				len(sorted), unit, totalRuntime,
			),
			Explanation: fmt.Sprintf(
				"Code unit %q is linked by %d slow tests, suggesting it may be "+
					"a dominant contributor to test suite slowness.",
				unit, len(sorted),
			),
			ImpactMetric: totalRuntime,
			ImpactUnit:   "total_avg_runtime_ms",
		})
	}
	return clusters
}

// detectFlakyFixtureClusters identifies code units shared by multiple flaky tests.
func detectFlakyFixtureClusters(snap *models.TestSuiteSnapshot) []Cluster {
	flakyTests := testFilesWithSignalType(snap, signals.SignalFlakyTest)
	if len(flakyTests) < minClusterSize {
		return nil
	}

	unitToFlakyTests := buildUnitToTestMap(flakyTests, snap)

	var clusters []Cluster
	for unit, tests := range unitToFlakyTests {
		if len(tests) < minClusterSize {
			continue
		}
		sorted := sortedCopy(tests)
		confidence := clamp(float64(len(sorted))/float64(len(flakyTests)), 0.4, 0.95)
		clusters = append(clusters, Cluster{
			Type:          ClusterDominantFlakyFixture,
			CausePath:     unit,
			AffectedTests: sorted,
			AffectedCount: len(sorted),
			Confidence:    confidence,
			Evidence: fmt.Sprintf(
				"%d flaky tests share code unit %q",
				len(sorted), unit,
			),
			Explanation: fmt.Sprintf(
				"Code unit %q is linked by %d flaky tests. This shared dependency "+
					"is a candidate root cause for non-deterministic test behavior.",
				unit, len(sorted),
			),
			ImpactMetric: float64(len(sorted)),
			ImpactUnit:   "flaky_test_count",
		})
	}
	return clusters
}

// detectSetupPathClusters looks for directories where many test files share
// the same signal type, suggesting directory-level setup may be the cause.
func detectSetupPathClusters(snap *models.TestSuiteSnapshot) []Cluster {
	// Build map: (dir, signalType) -> test file paths.
	type dirSignalKey struct {
		dir        string
		signalType models.SignalType
	}
	groups := make(map[dirSignalKey][]string)

	for _, tf := range snap.TestFiles {
		dir := filepath.Dir(tf.Path)
		seenTypes := make(map[models.SignalType]bool)
		for _, sig := range tf.Signals {
			if seenTypes[sig.Type] {
				continue
			}
			seenTypes[sig.Type] = true
			key := dirSignalKey{dir: dir, signalType: sig.Type}
			groups[key] = append(groups[key], tf.Path)
		}
	}

	var clusters []Cluster
	for key, tests := range groups {
		if len(tests) < minClusterSize {
			continue
		}
		sorted := sortedCopy(tests)
		confidence := clamp(float64(len(sorted))*0.15, 0.3, 0.85)
		clusters = append(clusters, Cluster{
			Type:          ClusterGlobalSetupPath,
			CausePath:     key.dir,
			AffectedTests: sorted,
			AffectedCount: len(sorted),
			Confidence:    confidence,
			Evidence: fmt.Sprintf(
				"%d tests in directory %q share signal type %q",
				len(sorted), key.dir, key.signalType,
			),
			Explanation: fmt.Sprintf(
				"Directory %q has %d test files all exhibiting %q signals. "+
					"A shared setup path or fixture in this directory may be the common cause.",
				key.dir, len(sorted), key.signalType,
			),
			ImpactMetric: float64(len(sorted)),
			ImpactUnit:   "affected_test_files",
		})
	}
	return clusters
}

// detectRepeatedFailPatterns groups snapshot-level signals by type and
// directory to find concentrated failure patterns.
func detectRepeatedFailPatterns(snap *models.TestSuiteSnapshot) []Cluster {
	type dirTypeKey struct {
		dir        string
		signalType models.SignalType
	}
	groups := make(map[dirTypeKey][]string)

	for _, sig := range snap.Signals {
		if sig.Location.File == "" {
			continue
		}
		dir := filepath.Dir(sig.Location.File)
		key := dirTypeKey{dir: dir, signalType: sig.Type}
		// Deduplicate: only add each file once per group.
		found := false
		for _, f := range groups[key] {
			if f == sig.Location.File {
				found = true
				break
			}
		}
		if !found {
			groups[key] = append(groups[key], sig.Location.File)
		}
	}

	var clusters []Cluster
	for key, files := range groups {
		if len(files) < minClusterSize {
			continue
		}
		sorted := sortedCopy(files)
		confidence := clamp(float64(len(sorted))*0.12, 0.25, 0.80)
		clusters = append(clusters, Cluster{
			Type:          ClusterRepeatedFailPattern,
			CausePath:     key.dir,
			AffectedTests: sorted,
			AffectedCount: len(sorted),
			Confidence:    confidence,
			Evidence: fmt.Sprintf(
				"%d files in %q have %q signals",
				len(sorted), key.dir, key.signalType,
			),
			Explanation: fmt.Sprintf(
				"Directory %q shows a concentrated pattern of %q signals across %d files. "+
					"This concentration suggests a systemic issue rather than isolated occurrences.",
				key.dir, key.signalType, len(sorted),
			),
			ImpactMetric: float64(len(sorted)),
			ImpactUnit:   "affected_files",
		})
	}
	return clusters
}

// --- helpers ---

// testFilesWithSignalType returns test file paths that have at least one signal
// of the given type.
func testFilesWithSignalType(snap *models.TestSuiteSnapshot, st models.SignalType) []string {
	var result []string
	for _, tf := range snap.TestFiles {
		for _, sig := range tf.Signals {
			if sig.Type == st {
				result = append(result, tf.Path)
				break
			}
		}
	}
	return result
}

// buildUnitToTestMap builds a map from code unit ID to test file paths,
// restricted to test files in the given set.
func buildUnitToTestMap(testPaths []string, snap *models.TestSuiteSnapshot) map[string][]string {
	pathSet := make(map[string]bool, len(testPaths))
	for _, p := range testPaths {
		pathSet[p] = true
	}

	result := make(map[string][]string)
	for _, tf := range snap.TestFiles {
		if !pathSet[tf.Path] {
			continue
		}
		for _, cu := range tf.LinkedCodeUnits {
			result[cu] = append(result[cu], tf.Path)
		}
	}
	return result
}

// sumRuntime returns the sum of AvgRuntimeMs for the given test file paths.
func sumRuntime(testPaths []string, snap *models.TestSuiteSnapshot) float64 {
	pathSet := make(map[string]bool, len(testPaths))
	for _, p := range testPaths {
		pathSet[p] = true
	}

	var total float64
	for _, tf := range snap.TestFiles {
		if pathSet[tf.Path] && tf.RuntimeStats != nil {
			total += tf.RuntimeStats.AvgRuntimeMs
		}
	}
	return total
}

// sharedImportConfidence calculates confidence for shared-import clusters
// based on what fraction of the test suite depends on the code unit.
func sharedImportConfidence(sharedCount, totalTests int) float64 {
	if totalTests == 0 {
		return 0.0
	}
	ratio := float64(sharedCount) / float64(totalTests)
	// Higher ratio means higher confidence that changes will have broad impact.
	return clamp(ratio*1.2, 0.3, 0.95)
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func sortedCopy(s []string) []string {
	cp := make([]string, len(s))
	copy(cp, s)
	sort.Strings(cp)
	return cp
}

func countUniqueAffectedTests(clusters []Cluster) int {
	seen := make(map[string]bool)
	for _, c := range clusters {
		for _, t := range c.AffectedTests {
			seen[t] = true
		}
	}
	return len(seen)
}
