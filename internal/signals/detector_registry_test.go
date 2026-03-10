package signals

import (
	"testing"

	"github.com/pmclSF/hamlet/internal/models"
)

// stubDetector emits a fixed set of signals for testing.
type stubDetector struct {
	id      string
	signals []models.Signal
}

func (d *stubDetector) Detect(_ *models.TestSuiteSnapshot) []models.Signal {
	return d.signals
}

type dependentStubDetector struct{}

func (d *dependentStubDetector) Detect(snap *models.TestSuiteSnapshot) []models.Signal {
	if len(snap.Signals) == 0 {
		return []models.Signal{{Type: "policyViolation", Category: models.CategoryGovernance, Explanation: "no upstream signals"}}
	}
	return []models.Signal{{Type: "policyViolation", Category: models.CategoryGovernance, Explanation: "upstream signals present"}}
}

func TestNewRegistry_Empty(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	if r.Len() != 0 {
		t.Errorf("new registry Len() = %d, want 0", r.Len())
	}
	if got := r.All(); len(got) != 0 {
		t.Errorf("new registry All() returned %d entries, want 0", len(got))
	}
}

func TestRegistry_Register(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	r.MustRegister(DetectorRegistration{
		Meta:     DetectorMeta{ID: "test.a", Domain: DomainQuality},
		Detector: &stubDetector{id: "a"},
	})
	r.MustRegister(DetectorRegistration{
		Meta:     DetectorMeta{ID: "test.b", Domain: DomainMigration},
		Detector: &stubDetector{id: "b"},
	})

	if r.Len() != 2 {
		t.Fatalf("Len() = %d, want 2", r.Len())
	}

	all := r.All()
	if all[0].Meta.ID != "test.a" || all[1].Meta.ID != "test.b" {
		t.Errorf("registration order not preserved: got %s, %s", all[0].Meta.ID, all[1].Meta.ID)
	}
}

func TestRegistry_ByDomain(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	r.MustRegister(DetectorRegistration{
		Meta:     DetectorMeta{ID: "q1", Domain: DomainQuality},
		Detector: &stubDetector{id: "q1"},
	})
	r.MustRegister(DetectorRegistration{
		Meta:     DetectorMeta{ID: "m1", Domain: DomainMigration},
		Detector: &stubDetector{id: "m1"},
	})
	r.MustRegister(DetectorRegistration{
		Meta:     DetectorMeta{ID: "q2", Domain: DomainQuality},
		Detector: &stubDetector{id: "q2"},
	})

	quality := r.ByDomain(DomainQuality)
	if len(quality) != 2 {
		t.Fatalf("ByDomain(quality) returned %d, want 2", len(quality))
	}
	if quality[0].Meta.ID != "q1" || quality[1].Meta.ID != "q2" {
		t.Errorf("ByDomain order: got %s, %s", quality[0].Meta.ID, quality[1].Meta.ID)
	}

	migration := r.ByDomain(DomainMigration)
	if len(migration) != 1 {
		t.Fatalf("ByDomain(migration) returned %d, want 1", len(migration))
	}

	health := r.ByDomain(DomainHealth)
	if len(health) != 0 {
		t.Errorf("ByDomain(health) returned %d, want 0", len(health))
	}
}

func TestRegistry_Run(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	r.MustRegister(DetectorRegistration{
		Meta: DetectorMeta{ID: "d1", Domain: DomainQuality},
		Detector: &stubDetector{signals: []models.Signal{
			{Type: "weakAssertion", Category: models.CategoryQuality},
		}},
	})
	r.MustRegister(DetectorRegistration{
		Meta: DetectorMeta{ID: "d2", Domain: DomainMigration},
		Detector: &stubDetector{signals: []models.Signal{
			{Type: "deprecatedTestPattern", Category: models.CategoryMigration},
			{Type: "dynamicTestGeneration", Category: models.CategoryMigration},
		}},
	})

	snap := &models.TestSuiteSnapshot{}
	r.Run(snap)

	if len(snap.Signals) != 3 {
		t.Fatalf("Run produced %d signals, want 3", len(snap.Signals))
	}
	if snap.Signals[0].Type != "weakAssertion" {
		t.Errorf("signal[0].Type = %s, want weakAssertion", snap.Signals[0].Type)
	}
	if snap.Signals[1].Type != "deprecatedTestPattern" {
		t.Errorf("signal[1].Type = %s, want deprecatedTestPattern", snap.Signals[1].Type)
	}
	if snap.Signals[2].Type != "dynamicTestGeneration" {
		t.Errorf("signal[2].Type = %s, want dynamicTestGeneration", snap.Signals[2].Type)
	}
}

func TestRegistry_RunDomain(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	r.MustRegister(DetectorRegistration{
		Meta: DetectorMeta{ID: "q1", Domain: DomainQuality},
		Detector: &stubDetector{signals: []models.Signal{
			{Type: "weakAssertion"},
		}},
	})
	r.MustRegister(DetectorRegistration{
		Meta: DetectorMeta{ID: "m1", Domain: DomainMigration},
		Detector: &stubDetector{signals: []models.Signal{
			{Type: "deprecatedTestPattern"},
		}},
	})

	snap := &models.TestSuiteSnapshot{}
	r.RunDomain(snap, DomainQuality)

	if len(snap.Signals) != 1 {
		t.Fatalf("RunDomain(quality) produced %d signals, want 1", len(snap.Signals))
	}
	if snap.Signals[0].Type != "weakAssertion" {
		t.Errorf("signal[0].Type = %s, want weakAssertion", snap.Signals[0].Type)
	}
}

func TestRegistry_Detectors(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	d1 := &stubDetector{id: "a"}
	d2 := &stubDetector{id: "b"}
	r.MustRegister(DetectorRegistration{
		Meta:     DetectorMeta{ID: "a"},
		Detector: d1,
	})
	r.MustRegister(DetectorRegistration{
		Meta:     DetectorMeta{ID: "b"},
		Detector: d2,
	})

	detectors := r.Detectors()
	if len(detectors) != 2 {
		t.Fatalf("Detectors() returned %d, want 2", len(detectors))
	}
}

func TestRegistry_DeterministicOutput_DifferentRegistrationOrder(t *testing.T) {
	t.Parallel()
	makeSignals := func(order []string) []models.Signal {
		r := NewRegistry()
		for _, id := range order {
			var sigs []models.Signal
			switch id {
			case "a":
				sigs = []models.Signal{{Type: "weakAssertion"}}
			case "b":
				sigs = []models.Signal{{Type: "mockHeavyTest"}}
			case "c":
				sigs = []models.Signal{{Type: "deprecatedTestPattern"}}
			}
			r.MustRegister(DetectorRegistration{
				Meta:     DetectorMeta{ID: id, Domain: DomainQuality},
				Detector: &stubDetector{signals: sigs},
			})
		}
		snap := &models.TestSuiteSnapshot{}
		r.Run(snap)
		return snap.Signals
	}

	// Same registration order -> same signal order.
	s1 := makeSignals([]string{"a", "b", "c"})
	s2 := makeSignals([]string{"a", "b", "c"})
	for i := range s1 {
		if s1[i].Type != s2[i].Type {
			t.Errorf("signal[%d] mismatch: %s vs %s", i, s1[i].Type, s2[i].Type)
		}
	}

	// Different registration order -> different signal order (deterministic per registration).
	s3 := makeSignals([]string{"c", "a", "b"})
	if s3[0].Type != "deprecatedTestPattern" {
		t.Errorf("reordered signal[0] = %s, want deprecatedTestPattern", s3[0].Type)
	}
}

func TestRegistry_All_ReturnsCopy(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	r.MustRegister(DetectorRegistration{
		Meta:     DetectorMeta{ID: "a"},
		Detector: &stubDetector{},
	})

	all := r.All()
	all[0].Meta.ID = "mutated"

	// Original should be unchanged.
	if r.All()[0].Meta.ID != "a" {
		t.Error("All() did not return a copy — mutation leaked")
	}
}

func TestRegistry_Register_DependencyOrderEnforced(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	if err := r.Register(DetectorRegistration{
		Meta:     DetectorMeta{ID: "dependent", DependsOnSignals: true},
		Detector: &stubDetector{},
	}); err != nil {
		t.Fatalf("registering dependent detector failed: %v", err)
	}

	err := r.Register(DetectorRegistration{
		Meta:     DetectorMeta{ID: "non-dependent", DependsOnSignals: false},
		Detector: &stubDetector{},
	})
	if err == nil {
		t.Fatal("expected error when registering non-dependent detector after dependent detector")
	}
}

func TestRegistry_Run_DependentDetectorsAfterIndependentPhase(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	r.MustRegister(DetectorRegistration{
		Meta: DetectorMeta{ID: "q", Domain: DomainQuality},
		Detector: &stubDetector{signals: []models.Signal{
			{Type: "weakAssertion", Category: models.CategoryQuality, Explanation: "weak"},
		}},
	})
	r.MustRegister(DetectorRegistration{
		Meta:     DetectorMeta{ID: "g", Domain: DomainGovernance, DependsOnSignals: true},
		Detector: &dependentStubDetector{},
	})

	snap := &models.TestSuiteSnapshot{}
	r.Run(snap)
	if len(snap.Signals) != 2 {
		t.Fatalf("Run produced %d signals, want 2", len(snap.Signals))
	}
	last := snap.Signals[len(snap.Signals)-1]
	if last.Type != "policyViolation" {
		t.Fatalf("last signal type = %q, want policyViolation", last.Type)
	}
	if last.Explanation != "upstream signals present" {
		t.Fatalf("dependent detector did not observe upstream signals, explanation=%q", last.Explanation)
	}
}
