package measurement

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pmclSF/terrain/internal/models"
)

// Registry holds measurement definitions and executes them.
type Registry struct {
	definitions []Definition
	ids         map[string]bool
}

// NewRegistry creates an empty measurement registry.
func NewRegistry() *Registry {
	return &Registry{
		ids: map[string]bool{},
	}
}

// Register adds a measurement definition. Returns an error on duplicate ID.
func (r *Registry) Register(def Definition) error {
	if r.ids[def.ID] {
		return fmt.Errorf("measurement: duplicate ID %q", def.ID)
	}
	r.ids[def.ID] = true
	r.definitions = append(r.definitions, def)
	return nil
}

// MustRegister adds a measurement definition, panicking on duplicate ID.
//
// Deprecated: Use Register and propagate errors instead.
// Retained only for backward compatibility with tests.
func (r *Registry) MustRegister(def Definition) {
	if err := r.Register(def); err != nil {
		panic(err)
	}
}

// All returns all registered definitions in registration order.
func (r *Registry) All() []Definition {
	out := make([]Definition, len(r.definitions))
	copy(out, r.definitions)
	return out
}

// ByDimension returns definitions matching the given dimension.
func (r *Registry) ByDimension(dim Dimension) []Definition {
	var out []Definition
	for _, d := range r.definitions {
		if d.Dimension == dim {
			out = append(out, d)
		}
	}
	return out
}

// Len returns the number of registered measurements.
func (r *Registry) Len() int {
	return len(r.definitions)
}

// Run executes all registered measurements against the snapshot.
func (r *Registry) Run(snap *models.TestSuiteSnapshot) []Result {
	results := make([]Result, len(r.definitions))
	for i, def := range r.definitions {
		results[i] = def.Compute(snap)
	}
	return results
}

// RunDimension executes measurements for a specific dimension.
func (r *Registry) RunDimension(snap *models.TestSuiteSnapshot, dim Dimension) []Result {
	var results []Result
	for _, def := range r.definitions {
		if def.Dimension == dim {
			results = append(results, def.Compute(snap))
		}
	}
	return results
}

// ComputeSnapshot runs all measurements and computes posture for each dimension.
func (r *Registry) ComputeSnapshot(snap *models.TestSuiteSnapshot) *Snapshot {
	results := r.Run(snap)

	// Group results by dimension.
	byDim := map[Dimension][]Result{}
	for _, res := range results {
		byDim[res.Dimension] = append(byDim[res.Dimension], res)
	}

	// Compute posture for each dimension.
	dims := []Dimension{
		DimensionHealth,
		DimensionCoverageDepth,
		DimensionCoverageDiversity,
		DimensionStructuralRisk,
		DimensionOperationalRisk,
	}

	var posture []DimensionPosture
	for _, dim := range dims {
		dimResults := byDim[dim]
		if len(dimResults) == 0 {
			posture = append(posture, DimensionPosture{
				Dimension:   dim,
				Band:        PostureUnknown,
				Explanation: "No measurements available for this dimension.",
			})
			continue
		}
		posture = append(posture, computeDimensionPosture(dim, dimResults))
	}

	return &Snapshot{
		Posture:      posture,
		Measurements: results,
	}
}

// computeDimensionPosture derives a posture band from a set of measurement results.
//
// The algorithm:
//  1. Count measurements by evidence strength
//  2. Identify the worst band among ratio/band measurements
//  3. Weight by concentration: many weak measurements = worse posture
//  4. Downgrade when evidence is partial/weak
func computeDimensionPosture(dim Dimension, results []Result) DimensionPosture {
	if len(results) == 0 {
		return DimensionPosture{
			Dimension:   dim,
			Band:        PostureUnknown,
			Explanation: "No measurements available.",
		}
	}

	// Collect band-like assessments from measurements.
	var bandValues []string
	var drivers []string
	var limitations []string
	hasStrongEvidence := false

	for _, r := range results {
		if r.Evidence == EvidenceStrong || r.Evidence == EvidencePartial {
			hasStrongEvidence = true
		}
		if r.Band != "" {
			bandValues = append(bandValues, r.Band)
			if r.Band == "weak" || r.Band == "elevated" || r.Band == "critical" {
				drivers = append(drivers, r.ID)
			}
		}
		limitations = append(limitations, r.Limitations...)
	}

	// Determine overall band from constituent measurements.
	band := resolvePostureBand(bandValues)

	// If no strong evidence, cap at moderate.
	if !hasStrongEvidence && band == PostureStrong {
		band = PostureModerate
	}

	explanation := buildPostureExplanation(dim, band, drivers, len(results), limitations)

	return DimensionPosture{
		Dimension:           dim,
		Band:                band,
		Explanation:         explanation,
		DrivingMeasurements: drivers,
		Measurements:        results,
	}
}

func resolvePostureBand(bands []string) PostureBand {
	if len(bands) == 0 {
		return PostureUnknown
	}

	order := map[string]int{
		"critical": 5,
		"elevated": 4,
		"weak":     3,
		"moderate": 2,
		"strong":   1,
	}

	// Filter out "unknown" bands — they represent missing data, not assessments.
	// Only resolved bands participate in posture computation.
	worst := 0
	weakCount := 0
	resolvedCount := 0
	for _, b := range bands {
		o := order[b]
		if o == 0 {
			// Unknown or unrecognized band — skip.
			continue
		}
		resolvedCount++
		if o > worst {
			worst = o
		}
		if o >= 3 {
			weakCount++
		}
	}

	// If no bands could be resolved, the dimension is unknown.
	if resolvedCount == 0 {
		return PostureUnknown
	}

	// If majority of resolved measurements are weak+, escalate.
	if resolvedCount > 1 && weakCount > resolvedCount/2 && worst < 4 {
		worst = 3 // at least weak
	}

	// Map back to PostureBand.
	bandMap := map[int]PostureBand{
		1: PostureStrong,
		2: PostureModerate,
		3: PostureWeak,
		4: PostureElevated,
		5: PostureCritical,
	}
	if b, ok := bandMap[worst]; ok {
		return b
	}
	return PostureUnknown
}

func buildPostureExplanation(dim Dimension, band PostureBand, drivers []string, total int, limitations []string) string {
	dimName := string(dim)
	switch band {
	case PostureStrong:
		return fmt.Sprintf("%s posture is strong across %d measurement(s).", dimName, total)
	case PostureModerate:
		return fmt.Sprintf("%s posture is moderate. Some measurements indicate room for improvement.", dimName)
	case PostureWeak:
		if len(drivers) > 0 {
			sort.Strings(drivers)
			return fmt.Sprintf("%s posture is weak. Driven by: %s.", dimName, joinMax(drivers, 3))
		}
		return fmt.Sprintf("%s posture is weak across %d measurement(s).", dimName, total)
	case PostureElevated:
		return fmt.Sprintf("%s posture is elevated. Significant issues detected in %s.", dimName, joinMax(drivers, 3))
	case PostureCritical:
		return fmt.Sprintf("%s posture is critical. Immediate attention needed.", dimName)
	default:
		return fmt.Sprintf("%s posture could not be determined.", dimName)
	}
}

func joinMax(items []string, max int) string {
	if len(items) <= max {
		return join(items)
	}
	return join(items[:max]) + fmt.Sprintf(" (+%d more)", len(items)-max)
}

// ToModel converts a measurement Snapshot to the serializable models type
// suitable for embedding in TestSuiteSnapshot.
func (s *Snapshot) ToModel() *models.MeasurementSnapshot {
	ms := &models.MeasurementSnapshot{}

	for _, p := range s.Posture {
		dp := models.DimensionPostureResult{
			Dimension:           string(p.Dimension),
			Band:                string(p.Band),
			Explanation:         p.Explanation,
			DrivingMeasurements: p.DrivingMeasurements,
		}
		for _, m := range p.Measurements {
			dp.Measurements = append(dp.Measurements, resultToModel(m))
		}
		ms.Posture = append(ms.Posture, dp)
	}

	for _, m := range s.Measurements {
		ms.Measurements = append(ms.Measurements, resultToModel(m))
	}

	return ms
}

func resultToModel(r Result) models.MeasurementResult {
	return models.MeasurementResult{
		ID:          r.ID,
		Dimension:   string(r.Dimension),
		Value:       r.Value,
		Units:       string(r.Units),
		Band:        r.Band,
		Evidence:    string(r.Evidence),
		Explanation: r.Explanation,
		Inputs:      r.Inputs,
		Limitations: r.Limitations,
	}
}

func join(items []string) string {
	return strings.Join(items, ", ")
}
