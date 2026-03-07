package models

// CodeUnitKind describes the kind of code element represented by a CodeUnit.
type CodeUnitKind string

const (
	CodeUnitKindFunction CodeUnitKind = "function"
	CodeUnitKindMethod   CodeUnitKind = "method"
	CodeUnitKindClass    CodeUnitKind = "class"
	CodeUnitKindModule   CodeUnitKind = "module"
	CodeUnitKindUnknown  CodeUnitKind = "unknown"
)

// CodeUnit represents a source code element that may be covered by tests.
//
// Hamlet uses code units to connect source structure to test effectiveness.
// This enables signals like:
//   - untested exports
//   - coverage blind spots
//   - change risk by module
type CodeUnit struct {
	// UnitID is the deterministic stable identifier for this code unit.
	// Format: normalized_path:symbol_name (or path:parent.symbol for methods).
	UnitID string `json:"unitId,omitempty"`

	// Name is the local identifier for the code unit.
	Name string `json:"name"`

	// Path is the repository-relative path containing the code unit.
	Path string `json:"path"`

	// Kind indicates the shape of the unit.
	Kind CodeUnitKind `json:"kind"`

	// Exported indicates whether the code unit is externally visible.
	Exported bool `json:"exported"`

	// ParentName is the containing class/struct name for methods.
	ParentName string `json:"parentName,omitempty"`

	// Language is the programming language.
	Language string `json:"language,omitempty"`

	// StartLine is the approximate start line of the unit definition.
	StartLine int `json:"startLine,omitempty"`

	// EndLine is the approximate end line of the unit definition.
	// May be zero if not determinable.
	EndLine int `json:"endLine,omitempty"`

	// Complexity is an optional lightweight complexity estimate.
	Complexity float64 `json:"complexity,omitempty"`

	// Coverage is an optional normalized coverage ratio for this unit.
	// 0.0 means no observed coverage, 1.0 means fully covered by the current model.
	Coverage float64 `json:"coverage,omitempty"`

	// LinkedTestFiles contains test file paths associated with this code unit.
	LinkedTestFiles []string `json:"linkedTestFiles,omitempty"`

	// Owner is the resolved owner for the code unit if known.
	Owner string `json:"owner,omitempty"`
}
