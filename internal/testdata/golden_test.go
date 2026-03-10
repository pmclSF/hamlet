package testdata

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/pmclSF/hamlet/internal/benchmark"
	"github.com/pmclSF/hamlet/internal/comparison"
	"github.com/pmclSF/hamlet/internal/heatmap"
	"github.com/pmclSF/hamlet/internal/impact"
	"github.com/pmclSF/hamlet/internal/measurement"
	"github.com/pmclSF/hamlet/internal/metrics"
	"github.com/pmclSF/hamlet/internal/portfolio"
	"github.com/pmclSF/hamlet/internal/reporting"
	"github.com/pmclSF/hamlet/internal/scoring"
)

var update = flag.Bool("update", false, "update golden files")
var goldenWriteMu sync.Mutex

func goldenPath(name string) string {
	return filepath.Join("golden", name)
}

func assertGolden(t *testing.T, name string, got []byte) {
	t.Helper()
	path := goldenPath(name)

	if *update {
		tmpPath := filepath.Join(t.TempDir(), filepath.Base(path))
		if err := os.WriteFile(tmpPath, got, 0o644); err != nil {
			t.Fatalf("write temp golden: %v", err)
		}
		goldenWriteMu.Lock()
		defer goldenWriteMu.Unlock()
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.Rename(tmpPath, path); err != nil {
			// Fallback for cross-device rename edge cases.
			data, readErr := os.ReadFile(tmpPath)
			if readErr != nil {
				t.Fatalf("read temp golden: %v", readErr)
			}
			if err := os.WriteFile(path, data, 0o644); err != nil {
				t.Fatalf("write golden: %v", err)
			}
		}
		if err := os.Chmod(path, 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		return
	}

	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v (run with -update to create)", path, err)
	}

	if !bytes.Equal(got, want) {
		t.Errorf("golden mismatch for %s\n--- want (first 500 chars) ---\n%s\n--- got (first 500 chars) ---\n%s",
			name, truncate(want, 500), truncate(got, 500))
	}
}

func truncate(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "..."
}

func TestGolden_MetricsJSON(t *testing.T) {
	t.Parallel()
	snap := MinimalSnapshot()
	ms := metrics.Derive(snap)
	// Zero out time-dependent field for determinism.
	ms.GeneratedAt = FixedTime

	data, err := json.MarshalIndent(ms, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	assertGolden(t, "metrics-minimal.json", data)
}

func TestGolden_ExportJSON(t *testing.T) {
	t.Parallel()
	snap := MinimalSnapshot()
	ms := metrics.Derive(snap)
	// Zero out time-dependent fields for determinism.
	ms.GeneratedAt = FixedTime

	// Compute measurements and portfolio so all export fields are populated.
	measReg := measurement.DefaultRegistry()
	snap.Measurements = measReg.ComputeSnapshot(snap).ToModel()
	snap.Portfolio = portfolio.Analyze(snap).ToModel()

	export := benchmark.BuildExport(snap, ms, false)
	export.ExportedAt = FixedTime

	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	assertGolden(t, "export-minimal.json", data)
}

func TestGolden_AnalyzeText(t *testing.T) {
	t.Parallel()
	snap := MinimalSnapshot()
	snap.Risk = scoring.ComputeRisk(snap)
	measReg := measurement.DefaultRegistry()
	snap.Measurements = measReg.ComputeSnapshot(snap).ToModel()
	snap.Portfolio = portfolio.Analyze(snap).ToModel()

	var buf bytes.Buffer
	reporting.RenderAnalyzeReport(&buf, snap)
	assertGolden(t, "analyze-minimal.txt", buf.Bytes())
}

func TestGolden_PortfolioText(t *testing.T) {
	t.Parallel()
	snap := FlakyConcentratedSnapshot()
	snap.Risk = scoring.ComputeRisk(snap)
	measReg := measurement.DefaultRegistry()
	snap.Measurements = measReg.ComputeSnapshot(snap).ToModel()
	snap.Portfolio = portfolio.Analyze(snap).ToModel()

	var buf bytes.Buffer
	reporting.RenderPortfolioReport(&buf, snap)
	assertGolden(t, "portfolio-flaky.txt", buf.Bytes())
}

func TestGolden_SummaryText(t *testing.T) {
	t.Parallel()
	snap := MinimalSnapshot()
	snap.Risk = scoring.ComputeRisk(snap)
	measReg := measurement.DefaultRegistry()
	snap.Measurements = measReg.ComputeSnapshot(snap).ToModel()

	h := heatmap.Build(snap)

	var buf bytes.Buffer
	reporting.RenderSummaryReport(&buf, snap, h)
	assertGolden(t, "summary-minimal.txt", buf.Bytes())
}

func TestGolden_PostureText(t *testing.T) {
	t.Parallel()
	snap := MinimalSnapshot()
	measReg := measurement.DefaultRegistry()
	snap.Measurements = measReg.ComputeSnapshot(snap).ToModel()

	var buf bytes.Buffer
	reporting.RenderPostureReport(&buf, snap)
	assertGolden(t, "posture-minimal.txt", buf.Bytes())
}

func TestGolden_ImpactText(t *testing.T) {
	t.Parallel()
	snap := HealthyBalancedSnapshot()
	measReg := measurement.DefaultRegistry()
	snap.Measurements = measReg.ComputeSnapshot(snap).ToModel()

	scope := impact.ChangeScopeFromPaths(
		[]string{"src/auth.js", "src/payment.js"},
		impact.ChangeModified,
	)
	result := impact.Analyze(scope, snap)

	var buf bytes.Buffer
	reporting.RenderImpactReport(&buf, result)
	assertGolden(t, "impact-balanced.txt", buf.Bytes())
}

func TestGolden_CompareText(t *testing.T) {
	t.Parallel()
	from := FlakyConcentratedSnapshot()
	to := HealthyBalancedSnapshot()

	comp := comparison.Compare(from, to)

	var buf bytes.Buffer
	reporting.RenderComparisonReport(&buf, comp)
	assertGolden(t, "compare-trend.txt", buf.Bytes())
}

func TestGolden_ImpactAggregateJSON(t *testing.T) {
	t.Parallel()
	snap := HealthyBalancedSnapshot()
	measReg := measurement.DefaultRegistry()
	snap.Measurements = measReg.ComputeSnapshot(snap).ToModel()

	scope := impact.ChangeScopeFromPaths(
		[]string{"src/auth.js", "src/payment.js", "src/__tests__/auth.test.js"},
		impact.ChangeModified,
	)
	result := impact.Analyze(scope, snap)
	agg := impact.BuildAggregate(result)

	data, err := json.MarshalIndent(agg, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	assertGolden(t, "impact-aggregate.json", data)
}
