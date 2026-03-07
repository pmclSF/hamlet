package signals

import "github.com/pmclSF/hamlet/internal/models"

// RunDetectors runs all provided detectors against the snapshot and
// appends the resulting signals to snap.Signals.
func RunDetectors(snap *models.TestSuiteSnapshot, detectors ...Detector) {
	for _, d := range detectors {
		found := d.Detect(snap)
		snap.Signals = append(snap.Signals, found...)
	}
}
