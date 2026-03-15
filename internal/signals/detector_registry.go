package signals

import (
	"fmt"
	"sort"
	"sync"

	"github.com/pmclSF/terrain/internal/models"
)

// Domain classifies a detector's area of concern.
type Domain string

const (
	DomainQuality    Domain = "quality"
	DomainMigration  Domain = "migration"
	DomainGovernance Domain = "governance"
	DomainHealth     Domain = "health"
	DomainCoverage   Domain = "coverage"
)

// EvidenceType describes how a detector obtains its evidence.
type EvidenceType string

const (
	EvidenceStructuralPattern EvidenceType = "structural-pattern"
	EvidencePathName          EvidenceType = "path-name"
	EvidenceRuntime           EvidenceType = "runtime"
	EvidenceCoverage          EvidenceType = "coverage"
	EvidencePolicy            EvidenceType = "policy"
	EvidenceCodeowners        EvidenceType = "codeowners"
)

// DetectorMeta describes a detector's identity and capabilities.
type DetectorMeta struct {
	// ID is a unique, stable identifier for the detector (e.g., "quality.weak-assertion").
	ID string

	// Domain is the detector's area of concern.
	Domain Domain

	// EvidenceType describes how the detector obtains evidence.
	EvidenceType EvidenceType

	// Description is a short human-readable summary.
	Description string

	// SignalTypes lists the signal types this detector may emit.
	SignalTypes []models.SignalType

	// RequiresFileIO indicates the detector reads files from disk beyond the snapshot.
	RequiresFileIO bool

	// DependsOnSignals indicates this detector reads signals from prior detectors.
	DependsOnSignals bool
}

// DetectorRegistration pairs a Detector with its metadata.
type DetectorRegistration struct {
	Meta     DetectorMeta
	Detector Detector
}

// DetectorRegistry holds an ordered set of detector registrations.
//
// Detectors are executed in registration order. Detectors that depend on
// prior signals (DependsOnSignals=true) should be registered after the
// detectors whose signals they read.
type DetectorRegistry struct {
	registrations []DetectorRegistration
}

// NewRegistry creates an empty DetectorRegistry.
func NewRegistry() *DetectorRegistry {
	return &DetectorRegistry{}
}

// Register adds a detector to the registry.
// Returns an error if a non-signal-dependent detector is registered after a
// signal-dependent one, since dependent detectors must run last.
func (r *DetectorRegistry) Register(reg DetectorRegistration) error {
	if !reg.Meta.DependsOnSignals {
		for _, existing := range r.registrations {
			if existing.Meta.DependsOnSignals {
				return fmt.Errorf("signals: cannot register detector %q (DependsOnSignals=false) after dependent detector %q", reg.Meta.ID, existing.Meta.ID)
			}
		}
	}
	r.registrations = append(r.registrations, reg)
	return nil
}

// MustRegister adds a detector to the registry, panicking on error.
// Use only for compile-time-known registrations (e.g., in DefaultRegistry).
func (r *DetectorRegistry) MustRegister(reg DetectorRegistration) {
	if err := r.Register(reg); err != nil {
		panic(err)
	}
}

// All returns all registrations in registration order.
func (r *DetectorRegistry) All() []DetectorRegistration {
	out := make([]DetectorRegistration, len(r.registrations))
	copy(out, r.registrations)
	return out
}

// ByDomain returns registrations matching the given domain.
func (r *DetectorRegistry) ByDomain(domain Domain) []DetectorRegistration {
	var out []DetectorRegistration
	for _, reg := range r.registrations {
		if reg.Meta.Domain == domain {
			out = append(out, reg)
		}
	}
	return out
}

// Detectors returns the Detector instances in registration order.
func (r *DetectorRegistry) Detectors() []Detector {
	out := make([]Detector, len(r.registrations))
	for i, reg := range r.registrations {
		out[i] = reg.Detector
	}
	return out
}

// Run executes all registered detectors against the snapshot in
// registration order, appending signals to snap.Signals.
func (r *DetectorRegistry) Run(snap *models.TestSuiteSnapshot) {
	if snap == nil || len(r.registrations) == 0 {
		return
	}

	type result struct {
		idx     int
		signals []models.Signal
	}

	var (
		wg         sync.WaitGroup
		mu         sync.Mutex
		results    []result
		dependents []DetectorRegistration
	)

	for idx, reg := range r.registrations {
		if reg.Meta.DependsOnSignals {
			dependents = append(dependents, reg)
			continue
		}

		wg.Add(1)
		go func(idx int, reg DetectorRegistration) {
			defer wg.Done()
			found := reg.Detector.Detect(snap)
			mu.Lock()
			results = append(results, result{idx: idx, signals: found})
			mu.Unlock()
		}(idx, reg)
	}
	wg.Wait()

	sort.Slice(results, func(i, j int) bool {
		return results[i].idx < results[j].idx
	})
	for _, res := range results {
		snap.Signals = append(snap.Signals, res.signals...)
	}

	// Dependent detectors run after independent outputs are available.
	for _, reg := range dependents {
		found := reg.Detector.Detect(snap)
		snap.Signals = append(snap.Signals, found...)
	}
}

// RunDomain executes only detectors matching the given domain.
func (r *DetectorRegistry) RunDomain(snap *models.TestSuiteSnapshot, domain Domain) {
	for _, reg := range r.registrations {
		if reg.Meta.Domain == domain {
			found := reg.Detector.Detect(snap)
			snap.Signals = append(snap.Signals, found...)
		}
	}
}

// Len returns the number of registered detectors.
func (r *DetectorRegistry) Len() int {
	return len(r.registrations)
}
