package clustering

import (
	"testing"

	"github.com/pmclSF/hamlet/internal/models"
	"github.com/pmclSF/hamlet/internal/signals"
)

func TestDetect_NilSnapshot(t *testing.T) {
	result := Detect(nil)
	if result == nil {
		t.Fatal("expected non-nil result for nil snapshot")
	}
	if len(result.Clusters) != 0 {
		t.Errorf("expected 0 clusters, got %d", len(result.Clusters))
	}
	if result.TotalAffectedTests != 0 {
		t.Errorf("expected 0 total affected, got %d", result.TotalAffectedTests)
	}
}

func TestDetect_EmptySnapshot(t *testing.T) {
	snap := &models.TestSuiteSnapshot{}
	result := Detect(snap)
	if len(result.Clusters) != 0 {
		t.Errorf("expected 0 clusters for empty snapshot, got %d", len(result.Clusters))
	}
}

func TestDetect_HealthyCodebase_NoClusters(t *testing.T) {
	// Each test links to a different code unit, no shared dependencies.
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/a_test.go", LinkedCodeUnits: []string{"pkg/a.go:FuncA"}},
			{Path: "test/b_test.go", LinkedCodeUnits: []string{"pkg/b.go:FuncB"}},
			{Path: "test/c_test.go", LinkedCodeUnits: []string{"pkg/c.go:FuncC"}},
		},
	}
	result := Detect(snap)
	if len(result.Clusters) != 0 {
		t.Errorf("expected 0 clusters for healthy codebase, got %d", len(result.Clusters))
	}
}

func TestDetect_SharedAuthHelper_BroadFlaky(t *testing.T) {
	// Multiple tests link to a shared auth helper, all have flaky signals.
	authUnit := "pkg/auth/helper.go:Authenticate"
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{
				Path:            "test/login_test.go",
				LinkedCodeUnits: []string{authUnit},
				Signals:         []models.Signal{{Type: signals.SignalFlakyTest}},
			},
			{
				Path:            "test/checkout_test.go",
				LinkedCodeUnits: []string{authUnit},
				Signals:         []models.Signal{{Type: signals.SignalFlakyTest}},
			},
			{
				Path:            "test/profile_test.go",
				LinkedCodeUnits: []string{authUnit},
				Signals:         []models.Signal{{Type: signals.SignalFlakyTest}},
			},
			{
				Path:            "test/dashboard_test.go",
				LinkedCodeUnits: []string{authUnit},
				Signals:         []models.Signal{{Type: signals.SignalFlakyTest}},
			},
		},
	}

	result := Detect(snap)

	// Should find at least a shared import cluster and a flaky fixture cluster.
	foundSharedImport := false
	foundFlakyFixture := false
	for _, c := range result.Clusters {
		if c.Type == ClusterSharedImport && c.CausePath == authUnit {
			foundSharedImport = true
			if c.AffectedCount != 4 {
				t.Errorf("shared import cluster: expected 4 affected, got %d", c.AffectedCount)
			}
			if c.Confidence <= 0 || c.Confidence > 1.0 {
				t.Errorf("confidence out of range: %f", c.Confidence)
			}
		}
		if c.Type == ClusterDominantFlakyFixture && c.CausePath == authUnit {
			foundFlakyFixture = true
			if c.AffectedCount != 4 {
				t.Errorf("flaky fixture cluster: expected 4 affected, got %d", c.AffectedCount)
			}
		}
	}
	if !foundSharedImport {
		t.Error("expected shared import cluster for auth helper")
	}
	if !foundFlakyFixture {
		t.Error("expected flaky fixture cluster for auth helper")
	}
	if result.TotalAffectedTests != 4 {
		t.Errorf("expected 4 total affected tests, got %d", result.TotalAffectedTests)
	}
}

func TestDetect_SharedSlowPath(t *testing.T) {
	dbUnit := "pkg/db/connection.go:Connect"
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{
				Path:            "test/repo_a_test.go",
				LinkedCodeUnits: []string{dbUnit},
				Signals:         []models.Signal{{Type: signals.SignalSlowTest}},
				RuntimeStats:    &models.RuntimeStats{AvgRuntimeMs: 5000},
			},
			{
				Path:            "test/repo_b_test.go",
				LinkedCodeUnits: []string{dbUnit},
				Signals:         []models.Signal{{Type: signals.SignalSlowTest}},
				RuntimeStats:    &models.RuntimeStats{AvgRuntimeMs: 4000},
			},
			{
				Path:            "test/repo_c_test.go",
				LinkedCodeUnits: []string{dbUnit},
				Signals:         []models.Signal{{Type: signals.SignalSlowTest}},
				RuntimeStats:    &models.RuntimeStats{AvgRuntimeMs: 6000},
			},
		},
	}

	result := Detect(snap)

	foundSlowHelper := false
	for _, c := range result.Clusters {
		if c.Type == ClusterDominantSlowHelper && c.CausePath == dbUnit {
			foundSlowHelper = true
			if c.AffectedCount != 3 {
				t.Errorf("slow helper cluster: expected 3 affected, got %d", c.AffectedCount)
			}
			// Total runtime should be 15000ms.
			if c.ImpactMetric != 15000 {
				t.Errorf("expected impact metric 15000, got %f", c.ImpactMetric)
			}
			if c.ImpactUnit != "total_avg_runtime_ms" {
				t.Errorf("expected impact unit total_avg_runtime_ms, got %s", c.ImpactUnit)
			}
		}
	}
	if !foundSlowHelper {
		t.Error("expected dominant slow helper cluster for db connection")
	}
}

func TestDetect_DirectoryLevelConcentration(t *testing.T) {
	// Multiple tests in same directory all have the same signal type.
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{
				Path:    "test/integration/api_test.go",
				Signals: []models.Signal{{Type: signals.SignalFlakyTest}},
			},
			{
				Path:    "test/integration/grpc_test.go",
				Signals: []models.Signal{{Type: signals.SignalFlakyTest}},
			},
			{
				Path:    "test/integration/ws_test.go",
				Signals: []models.Signal{{Type: signals.SignalFlakyTest}},
			},
		},
	}

	result := Detect(snap)

	foundSetupPath := false
	for _, c := range result.Clusters {
		if c.Type == ClusterGlobalSetupPath && c.CausePath == "test/integration" {
			foundSetupPath = true
			if c.AffectedCount != 3 {
				t.Errorf("setup path cluster: expected 3 affected, got %d", c.AffectedCount)
			}
		}
	}
	if !foundSetupPath {
		t.Error("expected global setup path cluster for test/integration directory")
	}
}

func TestDetect_RepeatedFailPattern(t *testing.T) {
	// Multiple snapshot-level signals from the same directory.
	snap := &models.TestSuiteSnapshot{
		Signals: []models.Signal{
			{Type: signals.SignalWeakAssertion, Location: models.SignalLocation{File: "pkg/handlers/user.go"}},
			{Type: signals.SignalWeakAssertion, Location: models.SignalLocation{File: "pkg/handlers/order.go"}},
			{Type: signals.SignalWeakAssertion, Location: models.SignalLocation{File: "pkg/handlers/product.go"}},
		},
	}

	result := Detect(snap)

	foundRepeated := false
	for _, c := range result.Clusters {
		if c.Type == ClusterRepeatedFailPattern && c.CausePath == "pkg/handlers" {
			foundRepeated = true
			if c.AffectedCount != 3 {
				t.Errorf("repeated fail pattern: expected 3 affected, got %d", c.AffectedCount)
			}
		}
	}
	if !foundRepeated {
		t.Error("expected repeated failure pattern cluster for pkg/handlers")
	}
}

func TestDetect_MultipleClusters(t *testing.T) {
	authUnit := "pkg/auth/helper.go:Auth"
	dbUnit := "pkg/db/pool.go:GetConn"
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{
				Path:            "test/a_test.go",
				LinkedCodeUnits: []string{authUnit, dbUnit},
				Signals:         []models.Signal{{Type: signals.SignalSlowTest}},
				RuntimeStats:    &models.RuntimeStats{AvgRuntimeMs: 3000},
			},
			{
				Path:            "test/b_test.go",
				LinkedCodeUnits: []string{authUnit, dbUnit},
				Signals:         []models.Signal{{Type: signals.SignalSlowTest}},
				RuntimeStats:    &models.RuntimeStats{AvgRuntimeMs: 2500},
			},
			{
				Path:            "test/c_test.go",
				LinkedCodeUnits: []string{authUnit},
				Signals:         []models.Signal{{Type: signals.SignalSlowTest}},
				RuntimeStats:    &models.RuntimeStats{AvgRuntimeMs: 4000},
			},
			{
				Path:            "test/d_test.go",
				LinkedCodeUnits: []string{dbUnit},
			},
		},
	}

	result := Detect(snap)

	if len(result.Clusters) < 2 {
		t.Errorf("expected at least 2 clusters, got %d", len(result.Clusters))
	}

	// Clusters should be sorted by affected count descending.
	for i := 1; i < len(result.Clusters); i++ {
		if result.Clusters[i].AffectedCount > result.Clusters[i-1].AffectedCount {
			t.Error("clusters should be sorted by affected count descending")
			break
		}
	}

	if result.TotalAffectedTests == 0 {
		t.Error("expected nonzero total affected tests")
	}
}

func TestDetect_BelowMinClusterSize(t *testing.T) {
	// Only 2 tests share a code unit — below the threshold of 3.
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/a_test.go", LinkedCodeUnits: []string{"pkg/shared.go:Helper"}},
			{Path: "test/b_test.go", LinkedCodeUnits: []string{"pkg/shared.go:Helper"}},
		},
	}

	result := Detect(snap)
	if len(result.Clusters) != 0 {
		t.Errorf("expected 0 clusters when below min cluster size, got %d", len(result.Clusters))
	}
}
