package policy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTerrainConfig_NotFound(t *testing.T) {
	t.Parallel()
	cfg, err := LoadTerrainConfig(t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg != nil {
		t.Error("expected nil config when file does not exist")
	}
}

func TestLoadTerrainConfig_Malformed(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "terrain.yaml"), []byte("{{invalid"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadTerrainConfig(dir)
	if err == nil {
		t.Error("expected error for malformed YAML")
	}
}

func TestLoadTerrainConfig_Valid(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	yaml := `manual_coverage:
  - name: billing regression
    area: billing-core
    source: testrail
    owner: qa-billing
    criticality: high
    frequency: per-release
  - name: onboarding flow
    area: onboarding
    source: jira
    owner: qa-platform
    criticality: medium
ci_duration_seconds: 120
`
	if err := os.WriteFile(filepath.Join(dir, "terrain.yaml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadTerrainConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if len(cfg.ManualCoverage) != 2 {
		t.Fatalf("expected 2 manual coverage entries, got %d", len(cfg.ManualCoverage))
	}
	if cfg.ManualCoverage[0].Name != "billing regression" {
		t.Errorf("expected 'billing regression', got %s", cfg.ManualCoverage[0].Name)
	}
	if cfg.ManualCoverage[0].Source != "testrail" {
		t.Errorf("expected 'testrail', got %s", cfg.ManualCoverage[0].Source)
	}
	if cfg.CIDurationSeconds == nil || *cfg.CIDurationSeconds != 120 {
		t.Error("expected ci_duration_seconds = 120")
	}
}

func TestToManualCoverageArtifacts(t *testing.T) {
	t.Parallel()
	cfg := &TerrainConfig{
		ManualCoverage: []ManualCoverageEntry{
			{Name: "billing regression", Area: "billing-core", Source: "testrail", Criticality: "high", Owner: "qa-billing"},
			{Name: "onboarding flow", Area: "onboarding", Source: "", Criticality: "", Owner: "qa-platform"},
			{Name: "", Area: "empty-name", Source: "manual"}, // skipped — no name
			{Name: "no-area", Area: "", Source: "manual"},    // skipped — no area
		},
	}

	artifacts := cfg.ToManualCoverageArtifacts()
	if len(artifacts) != 2 {
		t.Fatalf("expected 2 artifacts (skipping empty name/area), got %d", len(artifacts))
	}

	// First artifact.
	a := artifacts[0]
	if a.Name != "billing regression" {
		t.Errorf("expected 'billing regression', got %s", a.Name)
	}
	if a.Source != "testrail" {
		t.Errorf("expected 'testrail', got %s", a.Source)
	}
	if a.Criticality != "high" {
		t.Errorf("expected 'high', got %s", a.Criticality)
	}
	if a.ArtifactID == "" {
		t.Error("expected non-empty artifact ID")
	}
	if a.ArtifactID[:len("manual:testrail:")] != "manual:testrail:" {
		t.Errorf("expected ID prefix 'manual:testrail:', got %s", a.ArtifactID)
	}

	// Second artifact — defaults applied.
	b := artifacts[1]
	if b.Source != "manual" {
		t.Errorf("expected default source 'manual', got %s", b.Source)
	}
	if b.Criticality != "medium" {
		t.Errorf("expected default criticality 'medium', got %s", b.Criticality)
	}
}

func TestToManualCoverageArtifacts_Nil(t *testing.T) {
	t.Parallel()
	var cfg *TerrainConfig
	artifacts := cfg.ToManualCoverageArtifacts()
	if len(artifacts) != 0 {
		t.Errorf("expected empty for nil config, got %d", len(artifacts))
	}
}

func TestTerrainConfig_IsEmpty(t *testing.T) {
	t.Parallel()
	var nilCfg *TerrainConfig
	if !nilCfg.IsEmpty() {
		t.Error("nil config should be empty")
	}

	emptyCfg := &TerrainConfig{}
	if !emptyCfg.IsEmpty() {
		t.Error("zero-value config should be empty")
	}

	dur := 120
	nonEmpty := &TerrainConfig{CIDurationSeconds: &dur}
	if nonEmpty.IsEmpty() {
		t.Error("config with CI duration should not be empty")
	}
}

func TestManualArtifactID_Deterministic(t *testing.T) {
	t.Parallel()
	id1 := manualArtifactID("testrail", "billing regression")
	id2 := manualArtifactID("testrail", "billing regression")
	if id1 != id2 {
		t.Errorf("expected deterministic IDs, got %s and %s", id1, id2)
	}

	// Case-insensitive.
	id3 := manualArtifactID("testrail", "Billing Regression")
	if id1 != id3 {
		t.Errorf("expected case-insensitive match, got %s and %s", id1, id3)
	}

	// Different source = different ID.
	id4 := manualArtifactID("jira", "billing regression")
	if id1 == id4 {
		t.Error("different sources should produce different IDs")
	}
}
