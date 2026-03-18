package engine

import (
	"strings"
	"testing"

	"github.com/pmclSF/terrain/internal/models"
	"github.com/pmclSF/terrain/internal/signals"
)

func TestDefaultRegistry_ReturnsNoError(t *testing.T) {
	t.Parallel()
	_, err := DefaultRegistry(Config{RepoRoot: "."})
	if err != nil {
		t.Fatalf("DefaultRegistry should succeed with valid config: %v", err)
	}
}

func TestDefaultRegistry_DetectorCountStable(t *testing.T) {
	t.Parallel()
	reg, err := DefaultRegistry(Config{RepoRoot: "."})
	if err != nil {
		t.Fatal(err)
	}
	// Without policy, should have 12 detectors.
	count := len(reg.All())
	if count < 12 {
		t.Errorf("expected at least 12 detectors, got %d", count)
	}
}

func TestRegistry_OrderingViolationReturnsError(t *testing.T) {
	t.Parallel()
	r := signals.NewRegistry()

	// Register a dependent detector first.
	depReg := signals.DetectorRegistration{
		Meta: signals.DetectorMeta{
			ID:               "test.dependent",
			Domain:           signals.DomainGovernance,
			SignalTypes:       []models.SignalType{"test-signal"},
			DependsOnSignals: true,
		},
		Detector: &stubDetector{},
	}
	if err := r.Register(depReg); err != nil {
		t.Fatalf("dependent Register should succeed: %v", err)
	}

	// Now try to register a non-dependent detector after a dependent one.
	// This should fail because non-dependent detectors must come before dependent ones.
	nonDepReg := signals.DetectorRegistration{
		Meta: signals.DetectorMeta{
			ID:               "test.non-dependent",
			Domain:           signals.DomainQuality,
			SignalTypes:       []models.SignalType{"test-signal-2"},
			DependsOnSignals: false,
		},
		Detector: &stubDetector{},
	}
	err := r.Register(nonDepReg)
	if err == nil {
		t.Fatal("ordering violation should return error, not nil")
	}
	if !strings.Contains(err.Error(), "cannot register") {
		t.Errorf("error should explain the ordering violation, got: %v", err)
	}
}

func TestRegistry_OrderingViolationDoesNotPanic(t *testing.T) {
	t.Parallel()
	r := signals.NewRegistry()

	// Register dependent first, then try non-dependent — should error, not panic.
	_ = r.Register(signals.DetectorRegistration{
		Meta: signals.DetectorMeta{
			ID: "dep", Domain: signals.DomainGovernance,
			SignalTypes: []models.SignalType{"s"}, DependsOnSignals: true,
		},
		Detector: &stubDetector{},
	})

	err := r.Register(signals.DetectorRegistration{
		Meta: signals.DetectorMeta{
			ID: "nondep", Domain: signals.DomainQuality,
			SignalTypes: []models.SignalType{"s2"}, DependsOnSignals: false,
		},
		Detector: &stubDetector{},
	})
	if err == nil {
		t.Fatal("expected error for ordering violation, got nil")
	}
	// If we got here, no panic occurred.
}

// stubDetector is a minimal detector for testing.
type stubDetector struct{}

func (d *stubDetector) Detect(_ *models.TestSuiteSnapshot) []models.Signal {
	return nil
}
