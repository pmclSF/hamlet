package ownership

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// CodeownersRule represents a parsed rule from a CODEOWNERS file.
type CodeownersRule struct {
	// Pattern is the original pattern string from the file.
	Pattern string

	// Owners are the normalized owner IDs for this rule.
	Owners []string

	// LineNumber is the line number in the CODEOWNERS file.
	LineNumber int
}

// CodeownersFile represents a fully parsed CODEOWNERS file.
type CodeownersFile struct {
	// Path is the repository-relative path to the CODEOWNERS file.
	Path string

	// Rules are the parsed rules in file order.
	// Per CODEOWNERS semantics, the last matching rule wins.
	Rules []CodeownersRule

	// Diagnostics holds any warnings from parsing.
	Diagnostics []Diagnostic
}

// codeownersLocations lists standard CODEOWNERS file locations in search order.
var codeownersLocations = []string{
	"CODEOWNERS",
	filepath.Join(".github", "CODEOWNERS"),
	filepath.Join("docs", "CODEOWNERS"),
}

// ParseCodeownersFile parses a CODEOWNERS file at the given absolute path
// and returns the structured result.
func ParseCodeownersFile(absPath, repoRelPath string) *CodeownersFile {
	cf := &CodeownersFile{Path: repoRelPath}

	f, err := os.Open(absPath)
	if err != nil {
		cf.Diagnostics = append(cf.Diagnostics, Diagnostic{
			Level:   "warning",
			Message: "could not open CODEOWNERS file: " + err.Error(),
			Source:  repoRelPath,
		})
		return cf
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip blank lines and comments.
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			cf.Diagnostics = append(cf.Diagnostics, Diagnostic{
				Level:   "warning",
				Message: "CODEOWNERS line has pattern but no owners",
				Source:  repoRelPath,
				Line:    lineNum,
			})
			continue
		}

		pattern := fields[0]

		// Check for unsupported patterns.
		if containsUnsupportedGlob(pattern) {
			cf.Diagnostics = append(cf.Diagnostics, Diagnostic{
				Level:   "info",
				Message: "pattern uses advanced glob syntax not fully supported: " + pattern,
				Source:  repoRelPath,
				Line:    lineNum,
			})
		}

		// Normalize all owner identifiers.
		owners := make([]string, 0, len(fields)-1)
		for _, raw := range fields[1:] {
			if strings.HasPrefix(raw, "#") {
				break // inline comment
			}
			owners = append(owners, NormalizeOwnerID(raw))
		}

		if len(owners) == 0 {
			cf.Diagnostics = append(cf.Diagnostics, Diagnostic{
				Level:   "warning",
				Message: "CODEOWNERS line has pattern but no valid owners after normalization",
				Source:  repoRelPath,
				Line:    lineNum,
			})
			continue
		}

		cf.Rules = append(cf.Rules, CodeownersRule{
			Pattern:    pattern,
			Owners:     owners,
			LineNumber: lineNum,
		})
	}

	return cf
}

// FindCodeownersFile locates the CODEOWNERS file in standard locations.
// Returns the absolute path, repo-relative path, and whether it was found.
func FindCodeownersFile(repoRoot string) (absPath, relPath string, found bool) {
	for _, loc := range codeownersLocations {
		abs := filepath.Join(repoRoot, loc)
		if _, err := os.Stat(abs); err == nil {
			return abs, loc, true
		}
	}
	return "", "", false
}

// MatchCodeowners finds the best matching rule for a file path.
// Returns the matched rule and true if found. Per CODEOWNERS semantics,
// the last matching rule wins.
func MatchCodeowners(rules []CodeownersRule, relPath string) (CodeownersRule, bool) {
	normalized := filepath.ToSlash(relPath)
	var lastMatch CodeownersRule
	matched := false

	for _, rule := range rules {
		if matchesCodeownersPattern(rule.Pattern, normalized) {
			lastMatch = rule
			matched = true
		}
	}

	return lastMatch, matched
}

// matchesCodeownersPattern checks if a file path matches a CODEOWNERS pattern.
//
// Supported patterns:
//   - Exact file: "path/to/file.js"
//   - Directory prefix: "src/auth/" or "/src/auth/"
//   - Wildcard extension: "*.js"
//   - Single-level wildcard: "docs/*"
//   - Double-star directory: "**/test/"
//   - Root-anchored: "/src/" (only matches from repo root)
func matchesCodeownersPattern(pattern, filePath string) bool {
	// Normalize pattern.
	p := filepath.ToSlash(pattern)

	// Bare "*" matches everything.
	if p == "*" {
		return true
	}

	// Handle wildcard extension patterns: "*.js"
	if strings.HasPrefix(p, "*.") {
		ext := p[1:] // ".js"
		return strings.HasSuffix(filePath, ext)
	}

	// Handle double-star patterns: "**/test/"
	if strings.HasPrefix(p, "**/") {
		suffix := strings.TrimPrefix(p, "**/")
		suffix = strings.TrimSuffix(suffix, "/")
		// Match if any path segment matches.
		parts := strings.Split(filePath, "/")
		for i, part := range parts {
			if part == suffix {
				_ = i
				return true
			}
			// Also match as a prefix of remaining path.
			remaining := strings.Join(parts[i:], "/")
			if strings.HasPrefix(remaining, suffix+"/") || remaining == suffix {
				return true
			}
		}
		return false
	}

	// Strip leading slash for root-anchored patterns.
	isRootAnchored := strings.HasPrefix(p, "/")
	p = strings.TrimPrefix(p, "/")

	// Handle single-level wildcard: "docs/*"
	if strings.HasSuffix(p, "/*") {
		dir := strings.TrimSuffix(p, "/*")
		if strings.HasPrefix(filePath, dir+"/") {
			// Only match one level deep.
			rest := strings.TrimPrefix(filePath, dir+"/")
			return !strings.Contains(rest, "/")
		}
		return false
	}

	// Directory prefix: pattern ends with "/"
	cleanPattern := strings.TrimSuffix(p, "/")

	// Check prefix match.
	if strings.HasPrefix(filePath, cleanPattern+"/") {
		return true
	}

	// Exact match.
	if filePath == cleanPattern {
		return true
	}

	// Non-root-anchored patterns also match as a path component anywhere.
	if !isRootAnchored && !strings.Contains(p, "/") {
		// Bare name matches as directory component.
		parts := strings.Split(filePath, "/")
		for _, part := range parts {
			if part == cleanPattern {
				return true
			}
		}
	}

	return false
}

// containsUnsupportedGlob checks for glob patterns we don't fully handle.
func containsUnsupportedGlob(pattern string) bool {
	// We support *, **, and *.ext but not complex patterns like [abc] or {a,b}.
	if strings.ContainsAny(pattern, "[]{},?") {
		return true
	}
	// Patterns with * in the middle of a name segment (not *.ext or */)
	// are partially supported.
	return false
}

// ToAssignment converts a matched CODEOWNERS rule into an OwnershipAssignment.
func (r CodeownersRule) ToAssignment(sourceFile string) OwnershipAssignment {
	owners := make([]Owner, len(r.Owners))
	for i, id := range r.Owners {
		owners[i] = Owner{ID: id}
	}
	return OwnershipAssignment{
		Owners:      owners,
		Source:      SourceCodeowners,
		Confidence:  ConfidenceHigh,
		Inheritance: InheritanceDirect,
		MatchedRule: r.Pattern,
		SourceFile:  sourceFile,
	}
}
