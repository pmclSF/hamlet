package migration

import (
	"testing"

	"github.com/pmclSF/hamlet/internal/models"
)

func TestComputeReadiness_NoBlockers(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/a.test.js"},
			{Path: "test/b.test.js"},
		},
		Signals: []models.Signal{
			{Type: "weakAssertion"},
		},
	}

	r := ComputeReadiness(snap)
	if r.ReadinessLevel != "high" {
		t.Errorf("readiness = %q, want high", r.ReadinessLevel)
	}
	if r.TotalBlockers != 0 {
		t.Errorf("totalBlockers = %d, want 0", r.TotalBlockers)
	}
}

func TestComputeReadiness_FewBlockers(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		TestFiles: make([]models.TestFile, 20),
		Signals: []models.Signal{
			{Type: "deprecatedTestPattern", Metadata: map[string]any{"blockerType": "deprecated-pattern"}},
		},
	}

	r := ComputeReadiness(snap)
	if r.ReadinessLevel != "high" {
		t.Errorf("readiness = %q, want high (1/20 = 5%%)", r.ReadinessLevel)
	}
}

func TestComputeReadiness_MediumBlockers(t *testing.T) {
	signals := make([]models.Signal, 0)
	for i := 0; i < 4; i++ {
		signals = append(signals, models.Signal{
			Type:     "deprecatedTestPattern",
			Metadata: map[string]any{"blockerType": "deprecated-pattern"},
		})
	}

	snap := &models.TestSuiteSnapshot{
		TestFiles: make([]models.TestFile, 20),
		Signals:   signals,
	}

	r := ComputeReadiness(snap)
	if r.ReadinessLevel != "medium" {
		t.Errorf("readiness = %q, want medium (4/20 = 20%%)", r.ReadinessLevel)
	}
}

func TestComputeReadiness_ManyBlockers(t *testing.T) {
	signals := make([]models.Signal, 0)
	for i := 0; i < 8; i++ {
		signals = append(signals, models.Signal{
			Type:     "migrationBlocker",
			Metadata: map[string]any{"blockerType": "custom-matcher"},
		})
	}

	snap := &models.TestSuiteSnapshot{
		TestFiles: make([]models.TestFile, 10),
		Signals:   signals,
	}

	r := ComputeReadiness(snap)
	if r.ReadinessLevel != "low" {
		t.Errorf("readiness = %q, want low (8/10 = 80%%)", r.ReadinessLevel)
	}
}

func TestComputeReadiness_NoTestFiles(t *testing.T) {
	snap := &models.TestSuiteSnapshot{}

	r := ComputeReadiness(snap)
	if r.ReadinessLevel != "unknown" {
		t.Errorf("readiness = %q, want unknown for empty repo", r.ReadinessLevel)
	}
}

func TestComputeReadiness_BlockersByType(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		TestFiles: make([]models.TestFile, 10),
		Signals: []models.Signal{
			{Type: "deprecatedTestPattern", Metadata: map[string]any{"blockerType": "deprecated-pattern"}},
			{Type: "customMatcherRisk", Metadata: map[string]any{"blockerType": "custom-matcher"}},
			{Type: "customMatcherRisk", Metadata: map[string]any{"blockerType": "custom-matcher"}},
		},
	}

	r := ComputeReadiness(snap)
	if r.BlockersByType["custom-matcher"] != 2 {
		t.Errorf("custom-matcher count = %d, want 2", r.BlockersByType["custom-matcher"])
	}
	if r.BlockersByType["deprecated-pattern"] != 1 {
		t.Errorf("deprecated-pattern count = %d, want 1", r.BlockersByType["deprecated-pattern"])
	}
}
