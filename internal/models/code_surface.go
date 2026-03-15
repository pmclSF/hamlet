package models

// CodeSurfaceKind describes the kind of behavior anchor a CodeSurface represents.
type CodeSurfaceKind string

const (
	// SurfaceFunction is a standalone exported function.
	SurfaceFunction CodeSurfaceKind = "function"

	// SurfaceMethod is a method on a type/class.
	SurfaceMethod CodeSurfaceKind = "method"

	// SurfaceHandler is an HTTP/RPC handler or middleware.
	SurfaceHandler CodeSurfaceKind = "handler"

	// SurfaceRoute is a registered route/endpoint.
	SurfaceRoute CodeSurfaceKind = "route"

	// SurfaceClass is a class or struct with public surface area.
	SurfaceClass CodeSurfaceKind = "class"
)

// CodeSurface represents an inferred behavior anchor in source code.
//
// Unlike CodeUnit (which tracks individual exported symbols for coverage
// linkage), CodeSurface identifies semantic behavior boundaries: the points
// in code where observable behavior originates. These are the natural
// targets for validation — the things tests should exercise.
//
// CodeSurfaces are inferred automatically from code structure. No manual
// YAML or configuration is required. The inference philosophy: if a function
// is exported, it has surface area. If it registers a route, it has behavior.
// If it's a handler, it transforms input to output. Terrain derives these
// anchors from the code itself.
type CodeSurface struct {
	// SurfaceID is a deterministic stable identifier.
	// Format: "surface:<path>:<name>" or "surface:<path>:<parent>.<name>".
	SurfaceID string `json:"surfaceId"`

	// Name is the local identifier (function name, method name, route path).
	Name string `json:"name"`

	// Path is the repository-relative file path containing this surface.
	Path string `json:"path"`

	// Kind classifies the behavior anchor.
	Kind CodeSurfaceKind `json:"kind"`

	// ParentName is the containing class/struct name for methods.
	ParentName string `json:"parentName,omitempty"`

	// Language is the programming language.
	Language string `json:"language"`

	// Package is the inferred package or module.
	Package string `json:"package,omitempty"`

	// Line is the source line where this surface is defined.
	Line int `json:"line,omitempty"`

	// Receiver is the type receiver for methods (Go-specific: "*Handler").
	Receiver string `json:"receiver,omitempty"`

	// Route is the HTTP route pattern when Kind is SurfaceRoute or SurfaceHandler.
	Route string `json:"route,omitempty"`

	// HTTPMethod is the HTTP method (GET, POST, etc.) when applicable.
	HTTPMethod string `json:"httpMethod,omitempty"`

	// Exported indicates whether this surface is publicly visible.
	Exported bool `json:"exported"`

	// LinkedCodeUnit is the CodeUnit.UnitID that corresponds to this surface,
	// if one exists. This links the behavior anchor to the coverage model.
	LinkedCodeUnit string `json:"linkedCodeUnit,omitempty"`
}

// BuildSurfaceID constructs a deterministic surface ID.
func BuildSurfaceID(path, name, parent string) string {
	if parent != "" {
		return "surface:" + path + ":" + parent + "." + name
	}
	return "surface:" + path + ":" + name
}
