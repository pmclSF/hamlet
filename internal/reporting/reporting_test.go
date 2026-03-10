package reporting

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/pmclSF/hamlet/internal/comparison"
	"github.com/pmclSF/hamlet/internal/heatmap"
	"github.com/pmclSF/hamlet/internal/impact"
	"github.com/pmclSF/hamlet/internal/measurement"
	"github.com/pmclSF/hamlet/internal/metrics"
	"github.com/pmclSF/hamlet/internal/migration"
	"github.com/pmclSF/hamlet/internal/models"
	"github.com/pmclSF/hamlet/internal/scoring"
	"github.com/pmclSF/hamlet/internal/testdata"
)

func TestRenderSummaryReport_HealthyBalanced(t *testing.T) {
	t.Parallel()
	snap := testdata.HealthyBalancedSnapshot()
	snap.Risk = scoring.ComputeRisk(snap)
	measReg := measurement.DefaultRegistry()
	snap.Measurements = measReg.ComputeSnapshot(snap).ToModel()

	h := heatmap.Build(snap)
	var buf bytes.Buffer
	RenderSummaryReport(&buf, snap, h)
	output := buf.String()

	checks := []string{"Hamlet Summary", "Key Numbers", "Test files:", "Frameworks:", "Next steps:"}
	for _, c := range checks {
		if !strings.Contains(output, c) {
			t.Errorf("summary report missing %q", c)
		}
	}
}

func TestRenderMetricsReport_Minimal(t *testing.T) {
	t.Parallel()
	snap := testdata.MinimalSnapshot()
	ms := metrics.Derive(snap)

	var buf bytes.Buffer
	RenderMetricsReport(&buf, ms)
	output := buf.String()

	if !strings.Contains(output, "Hamlet Metrics") {
		t.Error("metrics report missing header")
	}
	if !strings.Contains(output, "Test files:") {
		t.Error("metrics report missing test file count")
	}
}

func TestRenderPostureReport_Healthy(t *testing.T) {
	t.Parallel()
	snap := testdata.HealthyBalancedSnapshot()
	measReg := measurement.DefaultRegistry()
	snap.Measurements = measReg.ComputeSnapshot(snap).ToModel()

	var buf bytes.Buffer
	RenderPostureReport(&buf, snap)
	output := buf.String()

	if !strings.Contains(output, "Hamlet Posture") {
		t.Error("posture report missing header")
	}
	if !strings.Contains(output, "Next steps:") {
		t.Error("posture report missing next steps")
	}
}

func TestRenderComparisonReport(t *testing.T) {
	t.Parallel()
	from := testdata.MinimalSnapshot()
	to := testdata.HealthyBalancedSnapshot()
	comp := comparison.Compare(from, to)

	var buf bytes.Buffer
	RenderComparisonReport(&buf, comp)
	output := buf.String()

	if !strings.Contains(output, "Hamlet Snapshot Comparison") {
		t.Error("comparison report missing header")
	}
	if !strings.Contains(output, "Methodology Compatibility") {
		t.Error("comparison report missing methodology section")
	}
	if !strings.Contains(output, "Status: COMPATIBLE") {
		t.Error("comparison report missing compatible status")
	}
}

func TestRenderComparisonReport_MethodologyIncompatibilityVisibleWithoutDeltas(t *testing.T) {
	t.Parallel()
	from := &models.TestSuiteSnapshot{
		GeneratedAt: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		SnapshotMeta: models.SnapshotMeta{
			SchemaVersion:          models.SnapshotSchemaVersion,
			MethodologyFingerprint: "aaa",
		},
	}
	to := &models.TestSuiteSnapshot{
		GeneratedAt: time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC),
		SnapshotMeta: models.SnapshotMeta{
			SchemaVersion:          models.SnapshotSchemaVersion,
			MethodologyFingerprint: "bbb",
		},
	}
	comp := comparison.Compare(from, to)

	var buf bytes.Buffer
	RenderComparisonReport(&buf, comp)
	output := buf.String()

	if !strings.Contains(output, "Methodology Compatibility") {
		t.Error("comparison report missing methodology section")
	}
	if !strings.Contains(output, "Status: INCOMPATIBLE") {
		t.Error("comparison report missing incompatible status")
	}
	if !strings.Contains(output, "Recommended Next Steps") {
		t.Error("comparison report missing recommended next steps for incompatible methodology")
	}
	if !strings.Contains(output, "No meaningful changes detected.") {
		t.Error("comparison report should still show no-delta summary")
	}
}

func TestRenderImpactReport_WithGaps(t *testing.T) {
	t.Parallel()
	result := &impact.ImpactResult{
		Scope: impact.ChangeScope{
			ChangedFiles: []impact.ChangedFile{
				{Path: "src/api.js", ChangeKind: impact.ChangeModified},
			},
		},
		ImpactedUnits: []impact.ImpactedCodeUnit{
			{UnitID: "api:ApiClient", Name: "ApiClient", Path: "src/api.js", Exported: true, ChangeKind: impact.ChangeModified, ProtectionStatus: impact.ProtectionNone, Owner: "team-api"},
		},
		ProtectionGaps: []impact.ProtectionGap{
			{GapType: "untested_export", CodeUnitID: "api:ApiClient", Path: "src/api.js", Explanation: "Exported function ApiClient has no coverage.", Severity: "high", SuggestedAction: "Add unit tests."},
		},
		SelectedTests: []impact.ImpactedTest{
			{Path: "test/api.test.js", ImpactConfidence: impact.ConfidenceInferred, Relevance: "in same directory"},
		},
		ImpactedOwners: []string{"team-api"},
		Posture:        impact.ChangeRiskPosture{Band: "high_risk", Explanation: "Significant risk."},
		Summary:        "1 file(s) changed, 1 code unit(s) impacted.",
		Limitations:    []string{"No coverage lineage available."},
	}

	var buf bytes.Buffer
	RenderImpactReport(&buf, result)
	output := buf.String()

	checks := []string{
		"Hamlet Impact Analysis",
		"Change-Risk Posture: HIGH_RISK",
		"Impacted Code Units",
		"Protection Gaps",
		"Recommended Tests",
		"Impacted Owners: team-api",
		"Limitations",
		"Next steps:",
	}
	for _, c := range checks {
		if !strings.Contains(output, c) {
			t.Errorf("impact report missing %q", c)
		}
	}
}

func TestRenderImpactDrilldown_Units(t *testing.T) {
	t.Parallel()
	result := &impact.ImpactResult{
		ImpactedUnits: []impact.ImpactedCodeUnit{
			{Name: "Foo", Path: "src/foo.js", ChangeKind: impact.ChangeModified, Exported: true, ProtectionStatus: impact.ProtectionStrong, ImpactConfidence: impact.ConfidenceExact, CoveringTests: []string{"test/foo.test.js"}},
		},
	}

	var buf bytes.Buffer
	RenderImpactUnits(&buf, result)
	output := buf.String()

	if !strings.Contains(output, "Impacted Code Units (1)") {
		t.Error("units view missing header")
	}
	if !strings.Contains(output, "[exported]") {
		t.Error("units view missing exported tag")
	}
}

func TestRenderImpactDrilldown_Gaps(t *testing.T) {
	t.Parallel()
	result := &impact.ImpactResult{
		ProtectionGaps: []impact.ProtectionGap{
			{GapType: "no_coverage", Severity: "medium", Explanation: "No tests.", Path: "src/bar.js", SuggestedAction: "Add tests."},
			{GapType: "untested_export", Severity: "high", Explanation: "Exported.", Path: "src/foo.js"},
		},
	}

	var buf bytes.Buffer
	RenderImpactGaps(&buf, result)
	output := buf.String()

	if !strings.Contains(output, "Protection Gaps (2)") {
		t.Error("gaps view missing header")
	}
	if !strings.Contains(output, "HIGH severity (1)") {
		t.Error("gaps view missing high severity section")
	}
	if !strings.Contains(output, "MEDIUM severity (1)") {
		t.Error("gaps view missing medium severity section")
	}
}

func TestRenderImpactDrilldown_Tests(t *testing.T) {
	t.Parallel()
	result := &impact.ImpactResult{
		ImpactedTests: []impact.ImpactedTest{
			{Path: "test/a.test.js", ImpactConfidence: impact.ConfidenceExact, Relevance: "covers unit", CoversUnits: []string{"Foo"}},
			{Path: "test/b.test.js", ImpactConfidence: impact.ConfidenceInferred, Relevance: "in same dir"},
		},
		SelectedTests: []impact.ImpactedTest{
			{Path: "test/a.test.js", ImpactConfidence: impact.ConfidenceExact, Relevance: "covers unit"},
		},
	}

	var buf bytes.Buffer
	RenderImpactTests(&buf, result)
	output := buf.String()

	if !strings.Contains(output, "Impacted Tests (2 total, 1 selected)") {
		t.Error("tests view missing header")
	}
	if !strings.Contains(output, "Recommended (run these first)") {
		t.Error("tests view missing recommended section")
	}
	if !strings.Contains(output, "Additional relevant tests") {
		t.Error("tests view missing additional section")
	}
}

func TestRenderImpactDrilldown_Owners(t *testing.T) {
	t.Parallel()
	result := &impact.ImpactResult{
		ImpactedUnits: []impact.ImpactedCodeUnit{
			{Name: "Foo", Owner: "team-a", ProtectionStatus: impact.ProtectionStrong, ChangeKind: impact.ChangeModified},
			{Name: "Bar", Owner: "team-a", ProtectionStatus: impact.ProtectionNone, ChangeKind: impact.ChangeModified},
			{Name: "Baz", Owner: "team-b", ProtectionStatus: impact.ProtectionWeak, ChangeKind: impact.ChangeAdded},
		},
		ImpactedOwners: []string{"team-a", "team-b"},
	}

	var buf bytes.Buffer
	RenderImpactOwners(&buf, result)
	output := buf.String()

	if !strings.Contains(output, "Impacted Owners (2)") {
		t.Error("owners view missing header")
	}
	if !strings.Contains(output, "team-a (2 unit(s))") {
		t.Error("owners view missing team-a")
	}
	if !strings.Contains(output, "team-b (1 unit(s))") {
		t.Error("owners view missing team-b")
	}
}

func TestRenderImpactDrilldown_EmptyResults(t *testing.T) {
	t.Parallel()
	result := &impact.ImpactResult{}

	var buf bytes.Buffer

	RenderImpactUnits(&buf, result)
	if !strings.Contains(buf.String(), "No impacted code units") {
		t.Error("empty units view should show message")
	}

	buf.Reset()
	RenderImpactGaps(&buf, result)
	if !strings.Contains(buf.String(), "No protection gaps") {
		t.Error("empty gaps view should show message")
	}

	buf.Reset()
	RenderImpactTests(&buf, result)
	if !strings.Contains(buf.String(), "No impacted tests") {
		t.Error("empty tests view should show message")
	}

	buf.Reset()
	RenderImpactOwners(&buf, result)
	if !strings.Contains(buf.String(), "No ownership data") {
		t.Error("empty owners view should show message")
	}
}

func TestRenderMigrationReport(t *testing.T) {
	t.Parallel()
	snap := testdata.MigrationRiskSnapshot()
	readiness := migration.ComputeReadiness(snap)

	var buf bytes.Buffer
	RenderMigrationReport(&buf, readiness)
	output := buf.String()

	if !strings.Contains(output, "Migration") {
		t.Error("migration report missing header")
	}
}
