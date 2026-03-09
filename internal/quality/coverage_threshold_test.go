package quality

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/pmclSF/hamlet/internal/models"
)

func TestCoverageThresholdDetector_BelowThreshold(t *testing.T) {
	// Create temp dir with coverage summary
	dir := t.TempDir()
	covDir := filepath.Join(dir, "coverage")
	if err := os.MkdirAll(covDir, 0o755); err != nil {
		t.Fatalf("mkdir coverage: %v", err)
	}

	summary := map[string]any{
		"total": map[string]any{
			"lines":      map[string]any{"pct": 65.0},
			"branches":   map[string]any{"pct": 45.0},
			"functions":  map[string]any{"pct": 72.0},
			"statements": map[string]any{"pct": 68.0},
		},
	}
	data, _ := json.Marshal(summary)
	if err := os.WriteFile(filepath.Join(covDir, "coverage-summary.json"), data, 0o644); err != nil {
		t.Fatalf("write coverage summary: %v", err)
	}

	snap := &models.TestSuiteSnapshot{
		Repository: models.RepositoryMetadata{RootPath: dir},
	}

	d := &CoverageThresholdDetector{Threshold: 80}
	signals := d.Detect(snap)

	if len(signals) != 4 {
		t.Fatalf("expected 4 signals (all metrics below 80%%), got %d", len(signals))
	}

	for _, s := range signals {
		if s.Type != "coverageThresholdBreak" {
			t.Errorf("type = %q, want coverageThresholdBreak", s.Type)
		}
	}
}

func TestCoverageThresholdDetector_AboveThreshold(t *testing.T) {
	dir := t.TempDir()
	covDir := filepath.Join(dir, "coverage")
	if err := os.MkdirAll(covDir, 0o755); err != nil {
		t.Fatalf("mkdir coverage: %v", err)
	}

	summary := map[string]any{
		"total": map[string]any{
			"lines":      map[string]any{"pct": 92.0},
			"branches":   map[string]any{"pct": 85.0},
			"functions":  map[string]any{"pct": 90.0},
			"statements": map[string]any{"pct": 91.0},
		},
	}
	data, _ := json.Marshal(summary)
	if err := os.WriteFile(filepath.Join(covDir, "coverage-summary.json"), data, 0o644); err != nil {
		t.Fatalf("write coverage summary: %v", err)
	}

	snap := &models.TestSuiteSnapshot{
		Repository: models.RepositoryMetadata{RootPath: dir},
	}

	d := &CoverageThresholdDetector{Threshold: 80}
	signals := d.Detect(snap)

	if len(signals) != 0 {
		t.Errorf("expected 0 signals when above threshold, got %d", len(signals))
	}
}

func TestCoverageThresholdDetector_NoCoverageData(t *testing.T) {
	dir := t.TempDir()

	snap := &models.TestSuiteSnapshot{
		Repository: models.RepositoryMetadata{RootPath: dir},
	}

	d := &CoverageThresholdDetector{Threshold: 80}
	signals := d.Detect(snap)

	if len(signals) != 0 {
		t.Errorf("expected 0 signals with no coverage data, got %d", len(signals))
	}
}
