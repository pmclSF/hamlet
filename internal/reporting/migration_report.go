package reporting

import (
	"fmt"
	"io"
	"strings"

	"github.com/pmclSF/hamlet/internal/migration"
)

// RenderMigrationReport writes a migration readiness report to w.
func RenderMigrationReport(w io.Writer, readiness *migration.ReadinessSummary) {
	line := func(format string, args ...any) {
		fmt.Fprintf(w, format+"\n", args...)
	}
	blank := func() { fmt.Fprintln(w) }

	line("Hamlet Migration Readiness")
	line(strings.Repeat("=", 40))
	blank()

	// Frameworks
	line("Frameworks")
	line(strings.Repeat("-", 40))
	if len(readiness.Frameworks) == 0 {
		line("  (no frameworks detected)")
	} else {
		for _, fw := range readiness.Frameworks {
			typeBadge := ""
			if fw.Type != "" {
				typeBadge = fmt.Sprintf(" [%s]", fw.Type)
			}
			line("  %-20s %4d files%s", fw.Name, fw.FileCount, typeBadge)
		}
	}
	blank()

	// Readiness
	line("Readiness Level: %s", strings.ToUpper(readiness.ReadinessLevel))
	line("  %s", readiness.Explanation)
	blank()

	// Blockers
	line("Migration Blockers: %d", readiness.TotalBlockers)
	line(strings.Repeat("-", 40))
	if readiness.TotalBlockers == 0 {
		line("  (none detected)")
	} else {
		for bt, count := range readiness.BlockersByType {
			line("  %-26s %d", bt, count)
		}
	}
	blank()

	// Representative examples
	if len(readiness.RepresentativeBlockers) > 0 {
		line("Representative Blockers")
		line(strings.Repeat("-", 40))
		for _, ex := range readiness.RepresentativeBlockers {
			line("  [%s] %s", ex.Type, ex.Explanation)
			if ex.File != "" {
				line("    %s", ex.File)
			}
		}
		blank()
	}
}
