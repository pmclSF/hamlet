package models

// BehaviorSurface represents a derived, higher-level behavior grouping
// inferred from related CodeSurfaces.
//
// While CodeSurface identifies individual behavior anchors (a single route,
// a single handler function), BehaviorSurface groups related anchors into
// a cohesive behavioral unit. For example, the "user authentication" behavior
// might group a POST /api/login route, a loginHandler function, and a
// validateToken utility.
//
// BehaviorSurfaces are derived automatically — users never define them.
// They are optional: all analysis pipelines (coverage, impact, risk) work
// with or without them. When present, they provide higher-level explanations
// that map to how teams think about their system.
//
// Every BehaviorSurface traces back to concrete CodeSurfaces. Explanations
// always reference the underlying code anchors, keeping results grounded
// and auditable.
type BehaviorSurface struct {
	// BehaviorID is a deterministic stable identifier.
	// Format: "behavior:<groupKey>:<label>" where groupKey encodes the
	// derivation source (route prefix, package, class name, etc.).
	BehaviorID string `json:"behaviorId"`

	// Label is a human-readable name for this behavior group.
	// Examples: "POST /api/auth/*", "package:auth", "UserController".
	Label string `json:"label"`

	// Description summarizes what this behavior group represents.
	Description string `json:"description,omitempty"`

	// Kind indicates the derivation strategy that produced this group.
	Kind BehaviorGroupKind `json:"kind"`

	// CodeSurfaceIDs lists the SurfaceIDs of the CodeSurfaces that
	// constitute this behavior group. Always non-empty.
	CodeSurfaceIDs []string `json:"codeSurfaceIds"`

	// Package is the common package when surfaces share one.
	Package string `json:"package,omitempty"`

	// RoutePrefix is the shared route prefix for route-grouped behaviors.
	RoutePrefix string `json:"routePrefix,omitempty"`

	// Language is the common language when surfaces share one.
	Language string `json:"language,omitempty"`
}

// BehaviorGroupKind describes how a BehaviorSurface was derived.
type BehaviorGroupKind string

const (
	// BehaviorGroupRoutePrefix groups surfaces by shared API route prefix.
	// Example: all routes under /api/users become one behavior.
	BehaviorGroupRoutePrefix BehaviorGroupKind = "route_prefix"

	// BehaviorGroupClass groups surfaces by containing class or receiver.
	// Example: all methods on UserController become one behavior.
	BehaviorGroupClass BehaviorGroupKind = "class"

	// BehaviorGroupModule groups surfaces by source module (file).
	// Example: all exports from auth.ts become one behavior.
	BehaviorGroupModule BehaviorGroupKind = "module"

	// BehaviorGroupDomain groups surfaces by directory-level domain boundary.
	// Example: all surfaces under src/auth/ become one behavior.
	BehaviorGroupDomain BehaviorGroupKind = "domain"

	// BehaviorGroupNaming groups surfaces by shared naming prefix.
	// Example: AuthLogin, AuthRegister, AuthLogout → behavior "Auth".
	BehaviorGroupNaming BehaviorGroupKind = "naming"
)
