package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/pmclSF/terrain/internal/analyze"
	"github.com/pmclSF/terrain/internal/depgraph"
	"github.com/pmclSF/terrain/internal/engine"
	"github.com/pmclSF/terrain/internal/impact"
	"github.com/pmclSF/terrain/internal/insights"
	"github.com/pmclSF/terrain/internal/metrics"
)

var updateGolden = flag.Bool("update-golden", false, "update golden snapshot files")

func fixtureRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "tests", "fixtures", "sample-repo")
}

func goldenPath(t *testing.T, name string) string {
	t.Helper()
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), "testdata", name+".golden")
}

// runAnalyzePipeline runs the full pipeline and returns structured output.
func runAnalyzePipeline(t *testing.T, root string) map[string]any {
	t.Helper()
	result, err := engine.RunPipeline(root, engine.PipelineOptions{EngineVersion: "test"})
	if err != nil {
		t.Fatalf("pipeline failed: %v", err)
	}
	snap := result.Snapshot

	return map[string]any{
		"testFileCount":  len(snap.TestFiles),
		"testCaseCount":  len(snap.TestCases),
		"codeUnitCount":  len(snap.CodeUnits),
		"signalCount":    len(snap.Signals),
		"frameworkCount": len(snap.Frameworks),
		"hasImportGraph": len(snap.ImportGraph) > 0,
	}
}

// runInsightsPipeline runs the insights pipeline and returns structured output.
func runInsightsPipeline(t *testing.T, root string) map[string]any {
	t.Helper()
	result, err := engine.RunPipeline(root, engine.PipelineOptions{EngineVersion: "test"})
	if err != nil {
		t.Fatalf("pipeline failed: %v", err)
	}
	snap := result.Snapshot
	ms := metrics.Derive(snap)

	dg := depgraph.Build(snap)
	dgCov := depgraph.AnalyzeCoverage(dg)
	dgDupes := depgraph.DetectDuplicates(dg)
	dgFanout := depgraph.AnalyzeFanout(dg, depgraph.DefaultFanoutThreshold)
	dgInsights := depgraph.ProfileInsights{
		Coverage:   &dgCov,
		Duplicates: &dgDupes,
		Fanout:     &dgFanout,
	}
	dgProfile := depgraph.AnalyzeProfile(dg, dgInsights)
	depgraph.EnrichProfileWithHealthRatios(&dgProfile, ms.Health.SkippedTestRatio, ms.Health.FlakyTestRatio)
	dgEdgeCases := depgraph.DetectEdgeCases(dgProfile, dg, dgInsights)
	dgPolicy := depgraph.ApplyEdgeCasePolicy(dgEdgeCases, dgProfile)

	report := insights.Build(&insights.BuildInput{
		Snapshot:   snap,
		HasPolicy:  result.HasPolicy,
		Coverage:   dgCov,
		Duplicates: dgDupes,
		Fanout:     dgFanout,
		Profile:    dgProfile,
		EdgeCases:  dgEdgeCases,
		Policy:     dgPolicy,
	})

	return map[string]any{
		"healthGrade":       report.HealthGrade,
		"findingCount":      len(report.Findings),
		"recommendationCount": len(report.Recommendations),
		"duplicateClusters": len(dgDupes.Clusters),
		"highFanoutNodes":   dgFanout.FlaggedCount,
		"weakCoverageCount": dgCov.BandCounts[depgraph.CoverageBandLow],
		"repoProfile":       dgProfile,
	}
}

// runImpactPipeline runs impact analysis against the fixture repo's known git
// history (HEAD~1 → HEAD) and returns structured output.
// Mirrors the runImpact() flow in main.go.
// See docs/examples/impact-report.md for the user-facing output this validates.
func runImpactPipeline(t *testing.T, root string) map[string]any {
	t.Helper()
	result, err := engine.RunPipeline(root, engine.PipelineOptions{EngineVersion: "test"})
	if err != nil {
		t.Fatalf("pipeline failed: %v", err)
	}
	snap := result.Snapshot

	absRoot, err := filepath.Abs(root)
	if err != nil {
		t.Fatalf("abs path failed: %v", err)
	}

	cs, err := impact.ChangeSetFromGitDiff(absRoot, "HEAD~1")
	if err != nil {
		t.Fatalf("changeset failed: %v", err)
	}

	impactResult := impact.AnalyzeChangeSet(cs, snap)

	// Apply edge-case policy (mirrors main.go runImpact).
	dg := depgraph.Build(snap)
	dgCov := depgraph.AnalyzeCoverage(dg)
	dgDupes := depgraph.DetectDuplicates(dg)
	dgFanout := depgraph.AnalyzeFanout(dg, depgraph.DefaultFanoutThreshold)
	ms := metrics.Derive(snap)
	pi := depgraph.ProfileInsights{
		Coverage:   &dgCov,
		Duplicates: &dgDupes,
		Fanout:     &dgFanout,
		Snapshot:   analyze.BuildSnapshotProfileData(snap),
	}
	dgProfile := depgraph.AnalyzeProfile(dg, pi)
	depgraph.EnrichProfileWithHealthRatios(&dgProfile, ms.Health.SkippedTestRatio, ms.Health.FlakyTestRatio)
	dgEdgeCases := depgraph.DetectEdgeCases(dgProfile, dg, pi)
	if len(dgEdgeCases) > 0 {
		dgPolicy := depgraph.ApplyEdgeCasePolicy(dgEdgeCases, dgProfile)
		impactResult.ApplyEdgeCasePolicy(dgPolicy.ConfidenceAdjustment, dgPolicy.RiskElevated, dgPolicy.Recommendations)
	}

	// Return stable aggregate fields (no timestamps, SHAs, or paths that vary by machine).
	return map[string]any{
		"changedFileCount":   len(impactResult.Scope.ChangedFiles),
		"impactedUnitCount":  len(impactResult.ImpactedUnits),
		"impactedTestCount":  len(impactResult.ImpactedTests),
		"selectedTestCount":  len(impactResult.SelectedTests),
		"protectionGapCount": len(impactResult.ProtectionGaps),
		"coverageConfidence": impactResult.CoverageConfidence,
		"posture":            impactResult.Posture.Band,
		"policyApplied":      impactResult.PolicyApplied,
		"hasSummary":         impactResult.Summary != "",
	}
}

func compareSnapshot(t *testing.T, name string, data map[string]any) {
	t.Helper()
	golden := goldenPath(t, name)

	actual, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	if *updateGolden {
		if err := os.WriteFile(golden, actual, 0o644); err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}
		t.Logf("updated golden file: %s", golden)
		return
	}

	expected, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("golden file not found: %s\nRun with -update-golden to create it.", golden)
	}

	// Normalize line endings.
	actualStr := strings.TrimSpace(string(actual))
	expectedStr := strings.TrimSpace(string(expected))

	if actualStr != expectedStr {
		t.Errorf("snapshot mismatch for %s\n\nExpected:\n%s\n\nActual:\n%s\n\nRun with -update-golden to update.",
			name, expectedStr, actualStr)
	}
}

// TestSnapshot_Analyze validates that the analyze pipeline produces stable output.
// See docs/examples/analyze-report.md for the user-facing output this validates.
func TestSnapshot_Analyze(t *testing.T) {
	t.Parallel()
	root := fixtureRoot(t)
	if _, err := os.Stat(root); os.IsNotExist(err) {
		t.Skipf("fixture not found: %s", root)
	}

	data := runAnalyzePipeline(t, root)

	// Structural assertions — these should be stable.
	if data["testFileCount"].(int) < 5 {
		t.Errorf("expected at least 5 test files, got %d", data["testFileCount"])
	}
	if data["signalCount"].(int) < 1 {
		t.Errorf("expected at least 1 signal, got %d", data["signalCount"])
	}

	compareSnapshot(t, "analyze", data)
}

// TestSnapshot_Insights validates that the insights pipeline produces stable output.
// See docs/examples/insights-report.md for the user-facing output this validates.
func TestSnapshot_Insights(t *testing.T) {
	t.Parallel()
	root := fixtureRoot(t)
	if _, err := os.Stat(root); os.IsNotExist(err) {
		t.Skipf("fixture not found: %s", root)
	}

	data := runInsightsPipeline(t, root)

	if data["healthGrade"].(string) == "" {
		t.Error("expected health grade to be populated")
	}

	profile := data["repoProfile"].(depgraph.RepoProfile)
	if profile.TestVolume == "" {
		t.Error("expected test volume classification")
	}

	compareSnapshot(t, "insights", data)
}

// TestSnapshot_Impact validates that impact analysis against the fixture repo's
// known git history produces stable, expected output.
// See docs/examples/impact-report.md for the user-facing output this validates.
func TestSnapshot_Impact(t *testing.T) {
	t.Parallel()
	root := fixtureRoot(t)
	if _, err := os.Stat(root); os.IsNotExist(err) {
		t.Skipf("fixture not found: %s", root)
	}

	data := runImpactPipeline(t, root)

	// Structural assertions — the fixture diff (HEAD~1 → HEAD) adds 2 files.
	if data["changedFileCount"].(int) < 1 {
		t.Errorf("expected at least 1 changed file, got %d", data["changedFileCount"])
	}
	if data["coverageConfidence"].(string) == "" {
		t.Error("expected coverage confidence to be populated")
	}
	if data["hasSummary"].(bool) != true {
		t.Error("expected summary to be populated")
	}

	compareSnapshot(t, "impact", data)
}

// TestSnapshot_Explain validates that explain produces stable per-test metadata.
// See docs/examples/explain-report.md for the user-facing output this validates.
func TestSnapshot_Explain(t *testing.T) {
	t.Parallel()
	root := fixtureRoot(t)
	if _, err := os.Stat(root); os.IsNotExist(err) {
		t.Skipf("fixture not found: %s", root)
	}

	result, err := engine.RunPipeline(root, engine.PipelineOptions{EngineVersion: "test"})
	if err != nil {
		t.Fatalf("pipeline failed: %v", err)
	}
	snap := result.Snapshot

	// Explain should find a test file.
	found := false
	for _, tf := range snap.TestFiles {
		if strings.Contains(tf.Path, "login.test") {
			found = true

			// Verify explain-worthy fields.
			if tf.Framework == "" {
				t.Error("expected framework to be detected")
			}
			if tf.TestCount < 1 {
				t.Error("expected at least 1 test")
			}

			var buf bytes.Buffer
			data := map[string]any{
				"path":      tf.Path,
				"framework": tf.Framework,
				"testCount": tf.TestCount,
			}

			actual, _ := json.MarshalIndent(data, "", "  ")
			buf.Write(actual)

			compareSnapshot(t, "explain", data)
			break
		}
	}

	if !found {
		t.Error("expected to find login.test file in fixture")
	}
}
