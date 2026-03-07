// Package identity implements deterministic test identity construction
// and normalization for Hamlet's longitudinal test tracking.
//
// Test identity is critical for:
//   - snapshot comparison across runs
//   - trend analysis over time
//   - flake tracking by stable test ID
//   - coverage attribution to specific tests
//
// Identity is explicitly NOT based on:
//   - random UUIDs
//   - traversal/discovery order
//   - runtime execution order
//   - line numbers (stored as metadata only)
package identity

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// NormalizePath normalizes a file path for identity construction.
//
// Rules:
//   - convert to forward slashes (OS-independent)
//   - ensure repository-relative (strip leading ./ or /)
//   - preserve case (paths are case-sensitive on Linux)
//   - NFC unicode normalization (via ValidString check)
func NormalizePath(p string) string {
	// Forward slashes only (explicit replace, since filepath.ToSlash is a no-op on Unix).
	p = strings.ReplaceAll(p, "\\", "/")

	// Strip leading ./
	p = strings.TrimPrefix(p, "./")

	// Strip leading /
	p = strings.TrimPrefix(p, "/")

	// Ensure valid UTF-8 (strip invalid sequences).
	if !utf8.ValidString(p) {
		p = strings.ToValidUTF8(p, "")
	}

	return p
}

// NormalizeName normalizes a test or suite name for identity construction.
//
// Rules:
//   - collapse internal whitespace to single spaces
//   - trim leading/trailing whitespace
//   - ensure valid UTF-8
//   - preserve case (test names are case-sensitive)
func NormalizeName(name string) string {
	// Ensure valid UTF-8.
	if !utf8.ValidString(name) {
		name = strings.ToValidUTF8(name, "")
	}

	// Trim leading/trailing whitespace.
	name = strings.TrimSpace(name)

	// Collapse internal whitespace to single spaces.
	name = collapseWhitespace(name)

	return name
}

// NormalizeSuiteHierarchy normalizes and joins a suite hierarchy.
// Each element is individually normalized, then joined with " > ".
func NormalizeSuiteHierarchy(parts []string) string {
	var normalized []string
	for _, p := range parts {
		n := NormalizeName(p)
		if n != "" {
			normalized = append(normalized, n)
		}
	}
	return strings.Join(normalized, " > ")
}

// collapseWhitespace replaces runs of whitespace with a single space.
func collapseWhitespace(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	inSpace := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			if !inSpace {
				b.WriteRune(' ')
				inSpace = true
			}
		} else {
			b.WriteRune(r)
			inSpace = false
		}
	}
	return b.String()
}
