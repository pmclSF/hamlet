// Package logging provides structured logging for the Terrain CLI using
// log/slog. All diagnostic output is written to stderr so that stdout remains
// clean for reports and JSON output.
//
// Three verbosity levels are supported:
//
//   - quiet:   only warn and error messages
//   - default: info, warn, and error messages
//   - debug:   all messages including debug-level tracing
//
// The package exposes a single configurable global logger. Library code
// obtains the logger via logging.L() and emits structured messages:
//
//	logging.L().Info("coverage ingested", "path", path, "files", count)
//	logging.L().Warn("runtime ingestion failed", "error", err)
//	logging.L().Debug("detector registered", "id", id)
package logging

import (
	"io"
	"log/slog"
	"os"
	"sync/atomic"
)

// Level represents the CLI verbosity level.
type Level int

const (
	// LevelQuiet suppresses info; only warn and error are emitted.
	LevelQuiet Level = iota
	// LevelDefault emits info, warn, and error.
	LevelDefault
	// LevelDebug emits all messages including debug.
	LevelDebug
)

// globalLogger is the package-level logger. Atomically swapped on Init.
var globalLogger atomic.Pointer[slog.Logger]

func init() {
	// Default: info-level, text handler, stderr.
	globalLogger.Store(newLogger(os.Stderr, LevelDefault))
}

// Init configures the global logger with the given verbosity level.
// It should be called once at CLI startup before any log output.
func Init(level Level) {
	globalLogger.Store(newLogger(os.Stderr, level))
}

// InitWithWriter configures the global logger to write to w.
// Primarily useful for testing.
func InitWithWriter(w io.Writer, level Level) {
	globalLogger.Store(newLogger(w, level))
}

// L returns the global structured logger.
func L() *slog.Logger {
	return globalLogger.Load()
}

// ParseLevel converts a CLI verbosity string to a Level.
// Accepted values: "quiet", "debug", "" (empty = default).
func ParseLevel(s string) Level {
	switch s {
	case "quiet", "q":
		return LevelQuiet
	case "debug", "d":
		return LevelDebug
	default:
		return LevelDefault
	}
}

func newLogger(w io.Writer, level Level) *slog.Logger {
	var slogLevel slog.Level
	switch level {
	case LevelQuiet:
		slogLevel = slog.LevelWarn
	case LevelDebug:
		slogLevel = slog.LevelDebug
	default:
		slogLevel = slog.LevelInfo
	}
	handler := slog.NewTextHandler(w, &slog.HandlerOptions{
		Level: slogLevel,
	})
	return slog.New(handler)
}
