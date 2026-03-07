package signals

import "github.com/pmclSF/hamlet/internal/models"

// Detector is the interface for all Hamlet signal detectors.
//
// A detector examines the snapshot and emits zero or more Signal values.
// Detectors should be:
//   - stateless
//   - deterministic given the same snapshot
//   - honest about confidence
type Detector interface {
	// Detect examines the snapshot and returns signals found.
	Detect(snap *models.TestSuiteSnapshot) []models.Signal
}
