package engine

import (
	"testing"

	"github.com/pmclSF/terrain/internal/policy"
	"github.com/pmclSF/terrain/internal/signals"
)

func TestDefaultRegistry_WithoutPolicy(t *testing.T) {
	t.Parallel()
	r := DefaultRegistry(Config{RepoRoot: "."})

	// Should have 11 detectors (4 quality + 2 coverage + 5 migration, no governance).
	if r.Len() != 11 {
		t.Errorf("DefaultRegistry without policy: Len() = %d, want 11", r.Len())
	}

	quality := r.ByDomain(signals.DomainQuality)
	if len(quality) != 4 {
		t.Errorf("quality detectors = %d, want 4", len(quality))
	}

	coverage := r.ByDomain(signals.DomainCoverage)
	if len(coverage) != 2 {
		t.Errorf("coverage detectors = %d, want 2", len(coverage))
	}

	migration := r.ByDomain(signals.DomainMigration)
	if len(migration) != 5 {
		t.Errorf("migration detectors = %d, want 5", len(migration))
	}

	governance := r.ByDomain(signals.DomainGovernance)
	if len(governance) != 0 {
		t.Errorf("governance detectors = %d, want 0 (no policy)", len(governance))
	}
}

func TestDefaultRegistry_WithPolicy(t *testing.T) {
	t.Parallel()
	boolTrue := true
	cfg := Config{
		RepoRoot: ".",
		PolicyConfig: &policy.Config{
			Rules: policy.Rules{
				DisallowSkippedTests: &boolTrue,
			},
		},
	}
	r := DefaultRegistry(cfg)

	// Should have 12 detectors (4 quality + 2 coverage + 5 migration + 1 governance).
	if r.Len() != 12 {
		t.Errorf("DefaultRegistry with policy: Len() = %d, want 12", r.Len())
	}

	governance := r.ByDomain(signals.DomainGovernance)
	if len(governance) != 1 {
		t.Errorf("governance detectors = %d, want 1", len(governance))
	}
}

func TestDefaultRegistry_DetectorIDs(t *testing.T) {
	t.Parallel()
	r := DefaultRegistry(Config{RepoRoot: "."})

	expectedIDs := []string{
		"quality.weak-assertion",
		"quality.mock-heavy",
		"quality.snapshot-heavy",
		"quality.untested-export",
		"quality.coverage-threshold",
		"coverage.blind-spot",
		"migration.deprecated-pattern",
		"migration.dynamic-test-generation",
		"migration.custom-matcher",
		"migration.unsupported-setup",
		"migration.framework-migration",
	}

	all := r.All()
	if len(all) != len(expectedIDs) {
		t.Fatalf("got %d detectors, want %d", len(all), len(expectedIDs))
	}

	for i, expected := range expectedIDs {
		if all[i].Meta.ID != expected {
			t.Errorf("detector[%d].ID = %q, want %q", i, all[i].Meta.ID, expected)
		}
	}
}

func TestDefaultRegistry_GovernanceIsLast(t *testing.T) {
	t.Parallel()
	boolTrue := true
	cfg := Config{
		RepoRoot: ".",
		PolicyConfig: &policy.Config{
			Rules: policy.Rules{
				DisallowSkippedTests: &boolTrue,
			},
		},
	}
	r := DefaultRegistry(cfg)

	all := r.All()
	last := all[len(all)-1]
	if last.Meta.Domain != signals.DomainGovernance {
		t.Errorf("last detector domain = %s, want governance", last.Meta.Domain)
	}
	if !last.Meta.DependsOnSignals {
		t.Error("governance detector should have DependsOnSignals=true")
	}
}

func TestGovernanceDetector_ImplementsInterface(t *testing.T) {
	t.Parallel()
	var _ signals.Detector = &GovernanceDetector{}
}

func TestDefaultRegistry_DetectorMetaFields(t *testing.T) {
	t.Parallel()
	r := DefaultRegistry(Config{RepoRoot: "."})

	for _, reg := range r.All() {
		if reg.Meta.ID == "" {
			t.Error("detector has empty ID")
		}
		if reg.Meta.Domain == "" {
			t.Errorf("detector %s has empty Domain", reg.Meta.ID)
		}
		if reg.Meta.EvidenceType == "" {
			t.Errorf("detector %s has empty EvidenceType", reg.Meta.ID)
		}
		if reg.Meta.Description == "" {
			t.Errorf("detector %s has empty Description", reg.Meta.ID)
		}
		if len(reg.Meta.SignalTypes) == 0 {
			t.Errorf("detector %s has no SignalTypes", reg.Meta.ID)
		}
	}
}

func TestDefaultRegistry_EmptyPolicy(t *testing.T) {
	t.Parallel()
	// An empty policy config should NOT register governance detector.
	cfg := Config{
		RepoRoot:     ".",
		PolicyConfig: &policy.Config{},
	}
	r := DefaultRegistry(cfg)

	governance := r.ByDomain(signals.DomainGovernance)
	if len(governance) != 0 {
		t.Errorf("empty policy should not register governance, got %d", len(governance))
	}
}
