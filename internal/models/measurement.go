package models

// MeasurementSnapshot captures all measurement results and posture
// computations for a point in time.
//
// This is the serializable model stored in TestSuiteSnapshot.Measurements.
// The measurement package populates these structs during analysis.
type MeasurementSnapshot struct {
	// Posture contains per-dimension posture results.
	Posture []DimensionPostureResult `json:"posture"`

	// Measurements contains all individual measurement results.
	Measurements []MeasurementResult `json:"measurements"`
}

// DimensionPostureResult is the computed posture for a single dimension.
type DimensionPostureResult struct {
	Dimension           string              `json:"dimension"`
	Band                string              `json:"band"`
	Explanation         string              `json:"explanation"`
	DrivingMeasurements []string            `json:"drivingMeasurements,omitempty"`
	Measurements        []MeasurementResult `json:"measurements,omitempty"`
}

// MeasurementResult is the output of a single measurement computation.
type MeasurementResult struct {
	ID          string   `json:"id"`
	Dimension   string   `json:"dimension"`
	Value       float64  `json:"value"`
	Units       string   `json:"units"`
	Band        string   `json:"band,omitempty"`
	Evidence    string   `json:"evidence"`
	Explanation string   `json:"explanation"`
	Inputs      []string `json:"inputs,omitempty"`
	Limitations []string `json:"limitations,omitempty"`
}
