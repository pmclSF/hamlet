package policy

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/pmclSF/terrain/internal/models"
)

// TerrainConfigFileName is the expected config file path relative to repo root.
const TerrainConfigFileName = "terrain.yaml"

// TerrainConfig represents the top-level terrain.yaml configuration.
type TerrainConfig struct {
	// ManualCoverage declares manual validation activities that supplement
	// automated CI coverage. These are overlays — they influence coverage
	// and risk reporting but never participate as executable CI validation.
	ManualCoverage []ManualCoverageEntry `yaml:"manual_coverage"`

	// CIDurationSeconds is the known CI duration in seconds.
	// Used by edge-case detection (FAST_CI_ALREADY).
	CIDurationSeconds *int `yaml:"ci_duration_seconds"`
}

// ManualCoverageEntry is a single manual coverage declaration in terrain.yaml.
type ManualCoverageEntry struct {
	// Name is a human-readable label (required).
	Name string `yaml:"name"`

	// Area is the code area or package this coverage applies to (required).
	// Examples: "billing-core", "auth/login", "checkout/*".
	Area string `yaml:"area"`

	// Source identifies the origin system.
	// Values: "testrail", "jira", "qase", "checklist", "exploratory", "manual".
	Source string `yaml:"source"`

	// Owner is the team or individual responsible.
	Owner string `yaml:"owner"`

	// Criticality indicates importance to release confidence.
	// Values: "high", "medium", "low". Defaults to "medium".
	Criticality string `yaml:"criticality"`

	// Frequency is the expected execution cadence.
	// Values: "per-release", "weekly", "monthly", "ad-hoc".
	Frequency string `yaml:"frequency"`

	// LastExecuted is when this was last executed (ISO 8601 date).
	LastExecuted string `yaml:"last_executed"`

	// Surfaces lists specific CodeSurface or BehaviorSurface IDs this covers.
	// Optional — when empty, area-based matching is used.
	Surfaces []string `yaml:"surfaces"`
}

// LoadTerrainConfig reads terrain.yaml from the given repository root.
//
// Behavior:
//   - If the file does not exist, returns nil config and no error.
//   - If the file exists but is malformed, returns an actionable error.
//   - If the file is valid, returns the parsed TerrainConfig.
func LoadTerrainConfig(repoRoot string) (*TerrainConfig, error) {
	path := filepath.Join(repoRoot, TerrainConfigFileName)

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read terrain config %s: %w", path, err)
	}

	var cfg TerrainConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("malformed terrain config %s: %w", path, err)
	}

	return &cfg, nil
}

// ToManualCoverageArtifacts converts config entries into model artifacts
// suitable for inclusion in the snapshot.
func (c *TerrainConfig) ToManualCoverageArtifacts() []models.ManualCoverageArtifact {
	if c == nil || len(c.ManualCoverage) == 0 {
		return nil
	}

	artifacts := make([]models.ManualCoverageArtifact, 0, len(c.ManualCoverage))
	for _, entry := range c.ManualCoverage {
		if entry.Name == "" || entry.Area == "" {
			continue
		}

		source := entry.Source
		if source == "" {
			source = "manual"
		}

		criticality := entry.Criticality
		if criticality == "" {
			criticality = "medium"
		}

		artifactID := manualArtifactID(source, entry.Name)

		artifacts = append(artifacts, models.ManualCoverageArtifact{
			ArtifactID:        artifactID,
			Name:              entry.Name,
			Area:              entry.Area,
			Source:            source,
			Owner:             entry.Owner,
			Criticality:       criticality,
			LastExecuted:      entry.LastExecuted,
			Frequency:         entry.Frequency,
			CoveredSurfaceIDs: entry.Surfaces,
		})
	}

	return artifacts
}

// manualArtifactID generates a stable artifact ID from source and name.
func manualArtifactID(source, name string) string {
	h := sha256.Sum256([]byte(strings.ToLower(name)))
	return fmt.Sprintf("manual:%s:%s", source, fmt.Sprintf("%x", h[:8]))
}

// IsEmpty returns true if no configuration is present.
func (c *TerrainConfig) IsEmpty() bool {
	return c == nil || (len(c.ManualCoverage) == 0 && c.CIDurationSeconds == nil)
}
