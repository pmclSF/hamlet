package analysis

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pmclSF/terrain/internal/models"
)

// DeriveBehaviorSurfaces groups related CodeSurfaces into higher-level
// BehaviorSurfaces using conservative, explainable heuristics.
//
// The derivation is optional — callers can safely ignore the result.
// Every BehaviorSurface traces back to concrete CodeSurface IDs.
//
// Five strategies are applied in priority order:
//  1. Route prefix grouping: surfaces sharing an API route prefix
//  2. Class grouping: surfaces sharing a parent class or receiver
//  3. Domain grouping: surfaces sharing a directory-level domain boundary
//  4. Naming convention grouping: surfaces sharing a name prefix
//  5. Module grouping: remaining surfaces grouped by source file
//
// Surfaces may appear in multiple groups (a route handler belongs to both
// the route group and the class group). This is intentional — it reflects
// how the same code participates in multiple behavioral concerns.
func DeriveBehaviorSurfaces(surfaces []models.CodeSurface) []models.BehaviorSurface {
	if len(surfaces) == 0 {
		return nil
	}

	var behaviors []models.BehaviorSurface
	behaviors = append(behaviors, deriveRouteGroups(surfaces)...)
	behaviors = append(behaviors, deriveClassGroups(surfaces)...)
	behaviors = append(behaviors, deriveDomainGroups(surfaces)...)
	behaviors = append(behaviors, deriveNamingGroups(surfaces)...)
	behaviors = append(behaviors, deriveModuleGroups(surfaces)...)

	// Deduplicate: drop groups with only one surface that already appear
	// in a multi-surface group from another strategy.
	behaviors = pruneSingletonDuplicates(behaviors)

	// Sort for deterministic output.
	sort.Slice(behaviors, func(i, j int) bool {
		return behaviors[i].BehaviorID < behaviors[j].BehaviorID
	})

	return behaviors
}

// deriveRouteGroups finds surfaces with routes that share a common prefix
// and groups them into route-based behaviors.
func deriveRouteGroups(surfaces []models.CodeSurface) []models.BehaviorSurface {
	// Collect surfaces that have route information.
	type routeEntry struct {
		surface models.CodeSurface
		parts   []string
	}
	var entries []routeEntry
	for _, s := range surfaces {
		route := s.Route
		if route == "" {
			continue
		}
		parts := splitRoutePath(route)
		if len(parts) == 0 {
			continue
		}
		entries = append(entries, routeEntry{surface: s, parts: parts})
	}

	if len(entries) < 2 {
		return nil
	}

	// Group by the first two significant path segments (e.g., /api/users).
	groups := map[string][]models.CodeSurface{}
	for _, e := range entries {
		prefix := routeGroupKey(e.parts)
		groups[prefix] = append(groups[prefix], e.surface)
	}

	var behaviors []models.BehaviorSurface
	for prefix, group := range groups {
		if len(group) < 2 {
			continue
		}

		ids := make([]string, len(group))
		for i, s := range group {
			ids[i] = s.SurfaceID
		}
		sort.Strings(ids)

		lang := commonLanguage(group)
		behaviors = append(behaviors, models.BehaviorSurface{
			BehaviorID:     fmt.Sprintf("behavior:route:%s", prefix),
			Label:          prefix + "/*",
			Description:    fmt.Sprintf("API routes under %s (inferred from %d endpoints)", prefix, len(group)),
			Kind:           models.BehaviorGroupRoutePrefix,
			CodeSurfaceIDs: ids,
			RoutePrefix:    prefix,
			Language:       lang,
		})
	}

	return behaviors
}

// deriveClassGroups finds surfaces that share a parent class or receiver
// and groups them into class-based behaviors.
func deriveClassGroups(surfaces []models.CodeSurface) []models.BehaviorSurface {
	type classKey struct {
		path   string
		parent string
	}

	groups := map[classKey][]models.CodeSurface{}
	for _, s := range surfaces {
		parent := s.ParentName
		if parent == "" && s.Receiver != "" {
			parent = s.Receiver
		}
		if parent == "" {
			continue
		}
		key := classKey{path: s.Path, parent: parent}
		groups[key] = append(groups[key], s)
	}

	var behaviors []models.BehaviorSurface
	for key, group := range groups {
		if len(group) < 2 {
			continue
		}

		ids := make([]string, len(group))
		for i, s := range group {
			ids[i] = s.SurfaceID
		}
		sort.Strings(ids)

		lang := commonLanguage(group)
		pkg := group[0].Package
		behaviors = append(behaviors, models.BehaviorSurface{
			BehaviorID:     fmt.Sprintf("behavior:class:%s:%s", key.path, key.parent),
			Label:          key.parent,
			Description:    fmt.Sprintf("%s in %s (inferred from %d methods)", key.parent, key.path, len(group)),
			Kind:           models.BehaviorGroupClass,
			CodeSurfaceIDs: ids,
			Package:        pkg,
			Language:       lang,
		})
	}

	return behaviors
}

// deriveDomainGroups groups surfaces by directory-level domain boundary.
// Surfaces under the same second-level directory (e.g., src/auth/, src/billing/)
// are grouped into a domain behavior. This captures service/domain boundaries
// that span multiple files.
func deriveDomainGroups(surfaces []models.CodeSurface) []models.BehaviorSurface {
	groups := map[string][]models.CodeSurface{}
	for _, s := range surfaces {
		domain := inferDomain(s.Path)
		if domain == "" {
			continue
		}
		groups[domain] = append(groups[domain], s)
	}

	var behaviors []models.BehaviorSurface
	for domain, group := range groups {
		// Require at least 3 surfaces from at least 2 files to form a domain group.
		// This prevents trivial domains from single-file modules.
		if len(group) < 3 {
			continue
		}
		paths := map[string]bool{}
		for _, s := range group {
			paths[s.Path] = true
		}
		if len(paths) < 2 {
			continue
		}

		ids := make([]string, len(group))
		for i, s := range group {
			ids[i] = s.SurfaceID
		}
		sort.Strings(ids)

		lang := commonLanguage(group)
		domainName := filepath.Base(domain)
		behaviors = append(behaviors, models.BehaviorSurface{
			BehaviorID:     fmt.Sprintf("behavior:domain:%s", domain),
			Label:          domainName,
			Description:    fmt.Sprintf("Domain %q boundary (%d surfaces across %d files)", domainName, len(group), len(paths)),
			Kind:           models.BehaviorGroupDomain,
			CodeSurfaceIDs: ids,
			Package:        domain,
			Language:       lang,
		})
	}

	return behaviors
}

// inferDomain extracts a domain boundary from a file path.
// Returns the first two directory segments for paths with depth >= 2,
// which typically maps to domain directories like src/auth, handlers/user, etc.
// Returns "" for shallow paths that don't imply a domain boundary.
func inferDomain(filePath string) string {
	parts := strings.Split(filepath.ToSlash(filePath), "/")
	if len(parts) < 3 {
		// Need at least dir/subdir/file.ext to infer a domain.
		return ""
	}
	return parts[0] + "/" + parts[1]
}

// deriveNamingGroups groups surfaces by shared name prefix using PascalCase
// or camelCase boundaries. For example, AuthLogin, AuthRegister, AuthLogout
// share the prefix "Auth" and form a naming-based behavior group.
func deriveNamingGroups(surfaces []models.CodeSurface) []models.BehaviorSurface {
	// Only consider surfaces with names that have a clear prefix (at least
	// 3 chars before a case boundary).
	groups := map[string][]models.CodeSurface{}
	for _, s := range surfaces {
		prefix := extractNamePrefix(s.Name)
		if prefix == "" {
			continue
		}
		groups[prefix] = append(groups[prefix], s)
	}

	var behaviors []models.BehaviorSurface
	for prefix, group := range groups {
		if len(group) < 3 {
			continue
		}

		ids := make([]string, len(group))
		for i, s := range group {
			ids[i] = s.SurfaceID
		}
		sort.Strings(ids)

		lang := commonLanguage(group)
		behaviors = append(behaviors, models.BehaviorSurface{
			BehaviorID:     fmt.Sprintf("behavior:naming:%s", prefix),
			Label:          prefix + "*",
			Description:    fmt.Sprintf("Surfaces sharing the %q naming prefix (inferred from %d symbols)", prefix, len(group)),
			Kind:           models.BehaviorGroupNaming,
			CodeSurfaceIDs: ids,
			Language:       lang,
		})
	}

	return behaviors
}

// extractNamePrefix extracts a PascalCase or camelCase prefix from a name.
// Returns the prefix if it is at least 3 characters and followed by an
// uppercase letter (indicating a word boundary). Returns "" otherwise.
//
// Examples:
//
//	"AuthLogin"     → "Auth"
//	"authLogin"     → "auth"
//	"getUserById"   → "get"
//	"validateEmail" → "validate"
//	"Go"            → "" (too short)
//	"GET /api/foo"  → "" (not a code identifier)
func extractNamePrefix(name string) string {
	// Skip names that look like routes or have spaces.
	if strings.ContainsAny(name, " /") {
		return ""
	}
	// Find the first uppercase boundary after position 2.
	for i := 3; i < len(name); i++ {
		if name[i] >= 'A' && name[i] <= 'Z' {
			return name[:i]
		}
		// Also split on underscore for snake_case.
		if name[i] == '_' {
			return name[:i]
		}
	}
	return ""
}

// deriveModuleGroups groups surfaces by their source file. This is the
// lowest-confidence grouping — used as a catch-all for surfaces not
// captured by route or class grouping.
func deriveModuleGroups(surfaces []models.CodeSurface) []models.BehaviorSurface {
	groups := map[string][]models.CodeSurface{}
	for _, s := range surfaces {
		groups[s.Path] = append(groups[s.Path], s)
	}

	var behaviors []models.BehaviorSurface
	for path, group := range groups {
		if len(group) < 2 {
			continue
		}

		ids := make([]string, len(group))
		for i, s := range group {
			ids[i] = s.SurfaceID
		}
		sort.Strings(ids)

		lang := commonLanguage(group)
		pkg := group[0].Package
		baseName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		behaviors = append(behaviors, models.BehaviorSurface{
			BehaviorID:     fmt.Sprintf("behavior:module:%s", path),
			Label:          baseName,
			Description:    fmt.Sprintf("Exported surfaces from %s (inferred from %d symbols)", path, len(group)),
			Kind:           models.BehaviorGroupModule,
			CodeSurfaceIDs: ids,
			Package:        pkg,
			Language:       lang,
		})
	}

	return behaviors
}

// pruneSingletonDuplicates removes single-surface behavior groups when
// that surface already belongs to a multi-surface group. This prevents
// noise from trivial groupings.
func pruneSingletonDuplicates(behaviors []models.BehaviorSurface) []models.BehaviorSurface {
	// Build set of surface IDs that appear in multi-surface groups.
	inMulti := map[string]bool{}
	for _, b := range behaviors {
		if len(b.CodeSurfaceIDs) >= 2 {
			for _, id := range b.CodeSurfaceIDs {
				inMulti[id] = true
			}
		}
	}

	var result []models.BehaviorSurface
	for _, b := range behaviors {
		if len(b.CodeSurfaceIDs) == 1 && inMulti[b.CodeSurfaceIDs[0]] {
			continue
		}
		result = append(result, b)
	}
	return result
}

// splitRoutePath splits a route like "/api/users/:id" into ["api", "users"].
// Leading slashes and parameter segments are stripped.
func splitRoutePath(route string) []string {
	var parts []string
	for _, seg := range strings.Split(route, "/") {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		// Skip parameter segments (:id, {id}, etc.)
		if strings.HasPrefix(seg, ":") || strings.HasPrefix(seg, "{") {
			continue
		}
		parts = append(parts, seg)
	}
	return parts
}

// routeGroupKey returns a grouping key from route parts.
// Uses up to the first two significant segments: "/api/users".
func routeGroupKey(parts []string) string {
	n := len(parts)
	if n > 2 {
		n = 2
	}
	return "/" + strings.Join(parts[:n], "/")
}

// commonLanguage returns the shared language if all surfaces have the same one.
func commonLanguage(surfaces []models.CodeSurface) string {
	if len(surfaces) == 0 {
		return ""
	}
	lang := surfaces[0].Language
	for _, s := range surfaces[1:] {
		if s.Language != lang {
			return ""
		}
	}
	return lang
}
