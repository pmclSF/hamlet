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

	"github.com/pmclSF/hamlet/internal/benchmark"
	"github.com/pmclSF/hamlet/internal/depgraph"
	"github.com/pmclSF/hamlet/internal/engine"
	"github.com/pmclSF/hamlet/internal/graph"
	"github.com/pmclSF/hamlet/internal/heatmap"
	"github.com/pmclSF/hamlet/internal/metrics"
	"github.com/pmclSF/hamlet/internal/summary"
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
		"testFileCount":   len(snap.TestFiles),
		"testCaseCount":   len(snap.TestCases),
		"codeUnitCount":   len(snap.CodeUnits),
		"signalCount":     len(snap.Signals),
		"frameworkCount":  len(snap.Frameworks),
		"hasImportGraph":  snap.ImportGraph != nil && len(snap.ImportGraph) > 0,
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

	g := graph.Build(snap)
	h := heatmap.BuildWithGraph(snap, g)
	ms := metrics.Derive(snap)
	seg := &benchmark.BuildExport(snap, ms, result.HasPolicy).Segment

	es := summary.Build(&summary.BuildInput{
		Snapshot:  snap,
		Heatmap:   h,
		Metrics:   ms,
		Segment:   seg,
		HasPolicy: result.HasPolicy,
	})

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

	return map[string]any{
		"recommendationCount": len(es.Recommendations),
		"hasPosture":          es.Posture.OverallBand != "",
		"duplicateClusters":   len(dgDupes.Clusters),
		"highFanoutNodes":     dgFanout.FlaggedCount,
		"weakCoverageCount":   dgCov.BandCounts[depgraph.CoverageBandLow],
		"repoProfile":         dgProfile,
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

func TestSnapshot_Analyze(t *testing.T) {
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

func TestSnapshot_Insights(t *testing.T) {
	root := fixtureRoot(t)
	if _, err := os.Stat(root); os.IsNotExist(err) {
		t.Skipf("fixture not found: %s", root)
	}

	data := runInsightsPipeline(t, root)

	if !data["hasPosture"].(bool) {
		t.Error("expected posture to be populated")
	}

	profile := data["repoProfile"].(depgraph.RepoProfile)
	if profile.TestVolume == "" {
		t.Error("expected test volume classification")
	}

	compareSnapshot(t, "insights", data)
}

func TestSnapshot_Explain(t *testing.T) {
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
