package insights

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/pmclSF/terrain/internal/depgraph"
	"github.com/pmclSF/terrain/internal/models"
)

var updateGolden = flag.Bool("update-golden", false, "update golden snapshot files")

func goldenPath(t *testing.T, name string) string {
	t.Helper()
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), "testdata", name+".golden")
}

// goldenReport extracts a stable subset of the report for golden comparison.
type goldenReport struct {
	HealthGrade     string                       `json:"healthGrade"`
	FindingCount    int                          `json:"findingCount"`
	RecCount        int                          `json:"recCount"`
	Categories      map[Category]int             `json:"categories"`
	TopFinding      string                       `json:"topFinding,omitempty"`
	TopRec          string                       `json:"topRec,omitempty"`
	LimitationCount int                          `json:"limitationCount"`
	DataSources     int                          `json:"dataSources"`
}

func extractGolden(r *Report) goldenReport {
	gr := goldenReport{
		HealthGrade:     r.HealthGrade,
		FindingCount:    len(r.Findings),
		RecCount:        len(r.Recommendations),
		Categories:      map[Category]int{},
		LimitationCount: len(r.Limitations),
		DataSources:     len(r.DataCompleteness),
	}
	for _, f := range r.Findings {
		gr.Categories[f.Category]++
	}
	if len(r.Findings) > 0 {
		gr.TopFinding = r.Findings[0].Title
	}
	if len(r.Recommendations) > 0 {
		gr.TopRec = r.Recommendations[0].Action
	}
	return gr
}

func compareGolden(t *testing.T, name string, gr goldenReport) {
	t.Helper()
	golden := goldenPath(t, name)

	actual, err := json.MarshalIndent(gr, "", "  ")
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

	actualStr := strings.TrimSpace(string(actual))
	expectedStr := strings.TrimSpace(string(expected))

	if actualStr != expectedStr {
		t.Errorf("snapshot mismatch for %s\n\nExpected:\n%s\n\nActual:\n%s\n\nRun with -update-golden to update.",
			name, expectedStr, actualStr)
	}
}

func TestGolden_EmptyRepo(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{}
	input := &BuildInput{
		Snapshot: snap,
		Coverage: depgraph.CoverageResult{BandCounts: map[depgraph.CoverageBand]int{}},
	}

	r := Build(input)
	compareGolden(t, "insights-empty-repo", extractGolden(r))
}

func TestGolden_HealthySmallRepo(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/auth.test.js", Framework: "jest", TestCount: 5},
			{Path: "test/api.test.js", Framework: "jest", TestCount: 8},
		},
		CodeUnits: []models.CodeUnit{
			{Path: "src/auth.js", Name: "login"},
			{Path: "src/api.js", Name: "handle"},
		},
	}
	input := &BuildInput{
		Snapshot: snap,
		Coverage: depgraph.CoverageResult{
			SourceCount: 2,
			BandCounts: map[depgraph.CoverageBand]int{
				depgraph.CoverageBandHigh: 2,
			},
		},
	}

	r := Build(input)
	compareGolden(t, "insights-healthy-small", extractGolden(r))
}

func TestGolden_ProblematicRepo(t *testing.T) {
	t.Parallel()

	signals := []models.Signal{
		{Type: "skippedTest", Severity: models.SeverityMedium, Category: models.CategoryHealth},
		{Type: "skippedTest", Severity: models.SeverityMedium, Category: models.CategoryHealth},
		{Type: "skippedTest", Severity: models.SeverityMedium, Category: models.CategoryHealth},
		{Type: "flakyTest", Severity: models.SeverityHigh, Category: models.CategoryHealth},
		{Type: "flakyTest", Severity: models.SeverityHigh, Category: models.CategoryHealth},
		{Type: "weakAssertion", Severity: models.SeverityCritical, Category: models.CategoryQuality},
		{Type: "untestedExport", Severity: models.SeverityHigh, Category: models.CategoryQuality},
	}

	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/auth.test.js", Framework: "jest", TestCount: 50},
			{Path: "test/api.test.js", Framework: "jest", TestCount: 30},
		},
		Signals: signals,
	}

	input := &BuildInput{
		Snapshot: snap,
		Coverage: depgraph.CoverageResult{
			SourceCount: 20,
			BandCounts: map[depgraph.CoverageBand]int{
				depgraph.CoverageBandLow:    12,
				depgraph.CoverageBandMedium: 5,
				depgraph.CoverageBandHigh:   3,
			},
			Sources: []depgraph.SourceCoverage{
				{Path: "src/billing.js", TestCount: 0, Band: depgraph.CoverageBandLow},
				{Path: "src/checkout.js", TestCount: 0, Band: depgraph.CoverageBandLow},
			},
		},
		Duplicates: depgraph.DuplicateResult{
			DuplicateCount: 25,
			TestsAnalyzed:  80,
			Clusters: []depgraph.DuplicateCluster{
				{ID: 1, Tests: []string{"t1", "t2", "t3"}, Similarity: 0.82},
			},
		},
		Fanout: depgraph.FanoutResult{
			FlaggedCount: 2,
			Threshold:    10,
			NodeCount:    50,
			Entries: []depgraph.FanoutEntry{
				{NodeID: "n1", Path: "fixtures/auth.js", NodeType: "helper", TransitiveFanout: 150, Flagged: true},
				{NodeID: "n2", Path: "fixtures/db.js", NodeType: "helper", TransitiveFanout: 40, Flagged: true},
			},
		},
	}

	r := Build(input)
	compareGolden(t, "insights-problematic", extractGolden(r))
}
