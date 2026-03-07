package engine

import (
	"fmt"
	"io"
	"time"
)

// PipelineDiagnostics captures timing and counts for each pipeline step.
type PipelineDiagnostics struct {
	Steps []StepDiagnostic `json:"steps"`
	Total time.Duration    `json:"totalDuration"`
}

// StepDiagnostic records metrics for a single pipeline step.
type StepDiagnostic struct {
	Name     string        `json:"name"`
	Duration time.Duration `json:"duration"`
	Count    int           `json:"count,omitempty"` // items produced/processed
}

func (d *PipelineDiagnostics) add(name string, dur time.Duration, count int) {
	d.Steps = append(d.Steps, StepDiagnostic{
		Name:     name,
		Duration: dur,
		Count:    count,
	})
}

// Render writes a human-readable diagnostics report to w.
func (d *PipelineDiagnostics) Render(w io.Writer) {
	fmt.Fprintf(w, "Pipeline Diagnostics\n")
	fmt.Fprintf(w, "%-30s %10s %8s\n", "Step", "Duration", "Count")
	fmt.Fprintf(w, "%-30s %10s %8s\n", "----", "--------", "-----")
	for _, s := range d.Steps {
		fmt.Fprintf(w, "%-30s %10s %8d\n", s.Name, s.Duration.Round(time.Microsecond), s.Count)
	}
	fmt.Fprintf(w, "%-30s %10s\n", "TOTAL", d.Total.Round(time.Microsecond))
}
