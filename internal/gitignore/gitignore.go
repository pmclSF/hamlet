// Package gitignore implements a pragmatic subset of .gitignore semantics
// as a shared skip filter for repository walkers.
//
// Supported:
//   - Blank lines and `#` comments are skipped.
//   - Trailing whitespace is trimmed (no escape support).
//   - `!pattern` negates a previous match.
//   - Patterns ending in `/` apply to directories only.
//   - Patterns starting with `/` are anchored to the repo root.
//   - Other patterns are floating and match at any depth.
//   - Glob characters are matched via filepath.Match (`*`, `?`, `[abc]`).
//
// Not yet supported:
//   - Nested `.gitignore` files in subdirectories (root only is read).
//   - `**` globstar across path segments.
//   - Negation precedence inside an anchored parent directory.
//
// The matcher is a best-effort skip filter, not a full git-spec
// implementation. It exists so that scanning a repo doesn't surface or waste
// time walking vendored trees, generated artifacts, and local-only
// directories the user has already declared off-limits. Files that slip
// through it are still subject to each walker's hardcoded skip set.
package gitignore

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// Matcher holds the parsed rules from a repo-root .gitignore.
type Matcher struct {
	rules []rule
}

type rule struct {
	pattern  string // pattern with leading `/` and trailing `/` stripped
	negated  bool
	dirOnly  bool
	anchored bool
}

// Load reads `<root>/.gitignore` and returns a matcher. Returns a non-nil
// matcher even when the file is missing; that matcher simply matches nothing.
func Load(root string) *Matcher {
	m := &Matcher{}
	f, err := os.Open(filepath.Join(root, ".gitignore"))
	if err != nil {
		return m
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		raw := scanner.Text()
		// Strip carriage returns, trailing whitespace.
		raw = strings.TrimRight(raw, " \t\r")
		if raw == "" || strings.HasPrefix(raw, "#") {
			continue
		}

		r := rule{}
		if strings.HasPrefix(raw, "!") {
			r.negated = true
			raw = raw[1:]
		}
		if strings.HasSuffix(raw, "/") {
			r.dirOnly = true
			raw = strings.TrimSuffix(raw, "/")
		}
		if strings.HasPrefix(raw, "/") {
			r.anchored = true
			raw = raw[1:]
		}
		if raw == "" {
			continue
		}
		r.pattern = raw
		m.rules = append(m.rules, r)
	}

	return m
}

// Match reports whether relPath (cleaned, forward-slash) is excluded by the
// loaded `.gitignore`. isDir indicates whether the path refers to a directory
// (dir-only patterns only apply to directories).
//
// The decision walks all rules in order, allowing a later negation to
// override an earlier match — matching git's documented semantics, modulo the
// scope not yet supported. For files, a dirOnly rule applies if any ancestor
// directory of the file matches; a file inside an ignored directory is itself
// ignored unless explicitly negated.
func (m *Matcher) Match(relPath string, isDir bool) bool {
	if m == nil || len(m.rules) == 0 || relPath == "" || relPath == "." {
		return false
	}
	relPath = filepath.ToSlash(filepath.Clean(relPath))

	// For files, evaluate against the file path AND each ancestor directory.
	// Each rule contributes its decision once; the last winning rule decides.
	var probes []string
	probes = append(probes, relPath)
	if !isDir {
		dir := filepath.ToSlash(filepath.Dir(relPath))
		for dir != "." && dir != "/" && dir != "" {
			probes = append(probes, dir)
			parent := filepath.ToSlash(filepath.Dir(dir))
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	excluded := false
	for _, r := range m.rules {
		for i, probe := range probes {
			probeIsDir := isDir || i > 0
			if r.dirOnly && !probeIsDir {
				continue
			}
			if matchesRule(r, probe) {
				excluded = !r.negated
				break
			}
		}
	}
	return excluded
}

func matchesRule(r rule, relPath string) bool {
	pattern := r.pattern
	if r.anchored {
		ok, _ := filepath.Match(pattern, relPath)
		if ok {
			return true
		}
		// Anchored directory patterns also match descendants.
		return strings.HasPrefix(relPath, pattern+"/")
	}

	// Floating: match against the full path or any path segment.
	parts := strings.Split(relPath, "/")
	for i := range parts {
		// Try matching the suffix starting at segment i.
		suffix := strings.Join(parts[i:], "/")
		if ok, _ := filepath.Match(pattern, suffix); ok {
			return true
		}
		// Try matching just this segment.
		if ok, _ := filepath.Match(pattern, parts[i]); ok {
			return true
		}
	}
	return false
}
