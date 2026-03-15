package signals

import "github.com/pmclSF/terrain/internal/models"

// Detector is the interface for all Terrain signal detectors.
//
// A detector examines the snapshot and emits zero or more Signal values.
// Detectors should be:
//   - stateless
//   - deterministic given the same snapshot
//   - honest about confidence
//   - read-only with respect to the snapshot (MUST NOT mutate snap)
//
// Non-dependent detectors run concurrently. Their results are collected
// and appended to snap.Signals only after all concurrent detectors complete.
type Detector interface {
	// Detect examines the snapshot and returns signals found.
	// Implementations MUST treat snap as read-only.
	Detect(snap *models.TestSuiteSnapshot) []models.Signal
}
