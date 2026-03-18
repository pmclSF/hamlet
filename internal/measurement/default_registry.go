package measurement

import "fmt"

// DefaultRegistry returns a registry pre-populated with all standard
// measurement definitions across all posture dimensions.
// Returns an error if any definition has a duplicate ID.
func DefaultRegistry() (*Registry, error) {
	r := NewRegistry()

	groups := [][]Definition{
		HealthMeasurements(),
		CoverageDepthMeasurements(),
		CoverageDiversityMeasurements(),
		StructuralRiskMeasurements(),
		OperationalRiskMeasurements(),
	}

	for _, defs := range groups {
		for _, def := range defs {
			if err := r.Register(def); err != nil {
				return nil, fmt.Errorf("measurement registry: %w", err)
			}
		}
	}

	return r, nil
}
