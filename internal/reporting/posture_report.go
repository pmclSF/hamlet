package reporting

import (
	"fmt"
	"io"
	"strings"

	"github.com/pmclSF/terrain/internal/models"
)

// RenderPostureReport writes a detailed posture breakdown with measurement
// evidence and explanations. This is the "posture explain" output.
func RenderPostureReport(w io.Writer, snap *models.TestSuiteSnapshot) {
	line := func(format string, args ...any) {
		fmt.Fprintf(w, format+"\n", args...)
	}
	blank := func() { fmt.Fprintln(w) }

	line("Terrain Posture Report")
	line(strings.Repeat("=", 60))
	blank()

	if snap.Measurements == nil || len(snap.Measurements.Posture) == 0 {
		line("  No measurement data available.")
		line("  Run `terrain analyze` to generate posture measurements.")
		return
	}

	for _, p := range snap.Measurements.Posture {
		line("%s", strings.ToUpper(p.Dimension))
		line("  Posture: %s", strings.ToUpper(p.Band))
		if p.Explanation != "" {
			line("  %s", p.Explanation)
		}
		blank()

		if len(p.DrivingMeasurements) > 0 {
			line("  Driving measurements: %s", strings.Join(p.DrivingMeasurements, ", "))
		}

		if len(p.Measurements) > 0 {
			line("  Measurements:")
			for _, m := range p.Measurements {
				bandTag := ""
				if m.Band != "" {
					bandTag = fmt.Sprintf(" [%s]", m.Band)
				}
				if m.Units == "ratio" {
					line("    %-40s %.1f%%%s", m.ID, m.Value*100, bandTag)
				} else {
					line("    %-40s %.2f%s", m.ID, m.Value, bandTag)
				}
				line("      Evidence: %s", m.Evidence)
				line("      %s", m.Explanation)
				for _, lim := range m.Limitations {
					line("      * %s", lim)
				}
			}
		}
		blank()
		line(strings.Repeat("-", 60))
		blank()
	}

	line("Next steps:")
	line("  terrain summary       leadership-ready overview")
	line("  terrain metrics       aggregate scorecard")
	line("  terrain posture --json   machine-readable posture data")
	blank()
}
