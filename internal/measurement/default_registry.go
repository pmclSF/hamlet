package measurement

// DefaultRegistry returns a registry pre-populated with all standard
// measurement definitions across all posture dimensions.
func DefaultRegistry() *Registry {
	r := NewRegistry()
	for _, def := range HealthMeasurements() {
		r.Register(def)
	}
	for _, def := range CoverageDepthMeasurements() {
		r.Register(def)
	}
	for _, def := range CoverageDiversityMeasurements() {
		r.Register(def)
	}
	for _, def := range StructuralRiskMeasurements() {
		r.Register(def)
	}
	for _, def := range OperationalRiskMeasurements() {
		r.Register(def)
	}
	return r
}
