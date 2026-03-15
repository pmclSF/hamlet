// Package signals contains the signal registry, signal type constants,
// and detector interfaces for Terrain's intelligence engine.
//
// The canonical Signal struct and its supporting scalar types (SignalType,
// SignalCategory, SignalSeverity, SignalLocation) live in internal/models
// so the snapshot model is self-contained and JSON-serializable without
// circular imports.
//
// This package provides:
//   - signal type constant identifiers
//   - the signal definition registry
//   - future detector interfaces
package signals
