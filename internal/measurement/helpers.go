package measurement

import "github.com/pmclSF/hamlet/internal/models"

// countSignals counts the number of signals in the snapshot matching any of
// the given signal types.
func countSignals(snap *models.TestSuiteSnapshot, types ...models.SignalType) int {
	typeSet := make(map[models.SignalType]bool, len(types))
	for _, t := range types {
		typeSet[t] = true
	}
	count := 0
	for _, s := range snap.Signals {
		if typeSet[s.Type] {
			count++
		}
	}
	return count
}

// ratioToBand maps a ratio to a qualitative band using three thresholds.
//   - ratio <= low  → "strong"
//   - ratio <= mid  → "moderate"
//   - ratio <= high → "weak"
//   - ratio >  high → "critical"
func ratioToBand(ratio, low, mid, high float64) string {
	switch {
	case ratio <= low:
		return "strong"
	case ratio <= mid:
		return "moderate"
	case ratio <= high:
		return "weak"
	default:
		return "critical"
	}
}

// runtimeEvidence returns the evidence strength based on whether runtime data
// is available in the snapshot.
func runtimeEvidence(snap *models.TestSuiteSnapshot) EvidenceStrength {
	for _, tf := range snap.TestFiles {
		if tf.RuntimeStats != nil && tf.RuntimeStats.AvgRuntimeMs > 0 {
			return EvidenceStrong
		}
	}
	return EvidenceWeak
}

// evidenceLimitations returns standard limitation strings for a given evidence level.
func evidenceLimitations(evidence EvidenceStrength) []string {
	switch evidence {
	case EvidenceWeak:
		return []string{"No runtime data available; result is based on static analysis only."}
	case EvidencePartial:
		return []string{"Partial data available; result may improve with more runtime artifacts."}
	case EvidenceNone:
		return []string{"No data available for this measurement."}
	default:
		return nil
	}
}
