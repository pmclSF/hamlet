package main

import (
	"fmt"
	"os"

	"github.com/pmclSF/terrain/internal/engine"
)

// isInteractive returns true if stderr is a terminal (TTY).
// Progress output goes to stderr so it doesn't interfere with
// stdout (which carries JSON or report output).
func isInteractive() bool {
	info, err := os.Stderr.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// newProgressFunc returns a ProgressFunc appropriate for the current output
// mode. Returns nil if progress should be suppressed (JSON mode, non-TTY).
//
// In interactive mode (TTY), progress is rendered as step-based lines:
//
//	[1/5] Scanning repository
//	[2/5] Building graph
//	...
//
// Each step overwrites the previous line using \r for a clean UX.
// In non-interactive mode (pipe/redirect), no progress is emitted
// to keep output parseable.
func newProgressFunc(jsonOutput bool) engine.ProgressFunc {
	if jsonOutput || !isInteractive() {
		return nil
	}
	return func(step, total int, label string) {
		if step < total {
			// Overwrite current line with \r for TTY.
			fmt.Fprintf(os.Stderr, "\r[%d/%d] %s", step, total, label)
		} else {
			// Final step: clear the progress line.
			fmt.Fprintf(os.Stderr, "\r[%d/%d] %s\n", step, total, label)
		}
	}
}
