package reporting

import (
	"fmt"
	"io"
	"strings"

	"github.com/pmclSF/hamlet/internal/governance"
)

// RenderPolicyReport writes a human-readable policy check report to w.
func RenderPolicyReport(w io.Writer, policyPath string, result *governance.Result) {
	line := func(format string, args ...any) {
		fmt.Fprintf(w, format+"\n", args...)
	}
	blank := func() { fmt.Fprintln(w) }

	line("Hamlet Policy Check")
	line(strings.Repeat("=", 40))
	blank()

	// Policy file
	line("Policy file")
	if policyPath != "" {
		line("  %s", policyPath)
	} else {
		line("  (none)")
	}
	blank()

	// Violations
	line("Violations")
	line(strings.Repeat("-", 40))
	if len(result.Violations) == 0 {
		line("  (none)")
	} else {
		for _, v := range result.Violations {
			loc := v.Location.File
			if loc == "" {
				loc = v.Location.Repository
			}
			line("  - %s: %s", v.Type, v.Explanation)
			if loc != "" {
				line("    location: %s", loc)
			}
		}
	}
	blank()

	// Status
	line("Status")
	if result.Pass {
		line("  PASS")
	} else {
		line("  FAIL")
	}
	blank()
}
