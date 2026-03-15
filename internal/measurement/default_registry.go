package measurement

// DefaultRegistry returns a registry pre-populated with all standard
// measurement definitions across all posture dimensions.
func DefaultRegistry() *Registry {
	r := NewRegistry()
	for _, def := range HealthMeasurements() {
		r.MustRegister(def)
	}
	for _, def := range CoverageDepthMeasurements() {
		r.MustRegister(def)
	}
	for _, def := range CoverageDiversityMeasurements() {
		r.MustRegister(def)
	}
	for _, def := range StructuralRiskMeasurements() {
		r.MustRegister(def)
	}
	for _, def := range OperationalRiskMeasurements() {
		r.MustRegister(def)
	}
	return r
}
