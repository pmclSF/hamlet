package ownership

import (
	"bufio"
	"os"
	"path"
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
		if reason := unsupportedGlobReason(pattern); reason != "" {
			cf.Diagnostics = append(cf.Diagnostics, Diagnostic{
				Level:   "warning",
				Message: "CODEOWNERS pattern uses unsupported or malformed glob syntax; matching will use best-effort fallback (" + reason + ")",
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
// Supported glob features:
//   - `*` and `**`
//   - `?`
//   - character classes like `[ab]`
//   - brace expansion like `{api,ui}`
//
// Directory patterns ending with `/` match files under that directory.
func matchesCodeownersPattern(pattern, filePath string) bool {
	p := filepath.ToSlash(strings.TrimSpace(pattern))
	fp := filepath.ToSlash(strings.TrimPrefix(filePath, "./"))
	if p == "" || fp == "" {
		return false
	}

	expanded := expandBracePatterns(p, 32)
	for _, candidate := range expanded {
		if matchExpandedCodeownersPattern(candidate, fp) {
			return true
		}
	}
	return false
}

func matchExpandedCodeownersPattern(pattern, filePath string) bool {
	p := pattern
	if strings.HasPrefix(p, "/") {
		p = strings.TrimPrefix(p, "/")
	}

	dirOnly := strings.HasSuffix(p, "/")
	p = strings.TrimSuffix(p, "/")
	if p == "" {
		return true
	}

	if !strings.Contains(p, "/") {
		parts := strings.Split(filePath, "/")
		limit := len(parts)
		if dirOnly && limit > 0 {
			limit--
		}
		for i := 0; i < limit; i++ {
			if matchSegmentGlob(p, parts[i]) {
				return true
			}
		}
		return false
	}

	if !dirOnly {
		return matchPathGlob(p, filePath)
	}

	// Directory rule: match any directory prefix in the path.
	parts := strings.Split(filePath, "/")
	for i := 1; i < len(parts); i++ {
		dir := strings.Join(parts[:i], "/")
		if matchPathGlob(p, dir) {
			return true
		}
	}
	return false
}

func matchPathGlob(pattern, target string) bool {
	patternSegments := splitPathSegments(pattern)
	targetSegments := splitPathSegments(target)
	return matchPathSegments(patternSegments, targetSegments)
}

func matchPathSegments(patternSegments, targetSegments []string) bool {
	if len(patternSegments) == 0 {
		return len(targetSegments) == 0
	}
	if patternSegments[0] == "**" {
		for len(patternSegments) > 1 && patternSegments[1] == "**" {
			patternSegments = patternSegments[1:]
		}
		if len(patternSegments) == 1 {
			return true
		}
		for i := 0; i <= len(targetSegments); i++ {
			if matchPathSegments(patternSegments[1:], targetSegments[i:]) {
				return true
			}
		}
		return false
	}
	if len(targetSegments) == 0 {
		return false
	}
	if !matchSegmentGlob(patternSegments[0], targetSegments[0]) {
		return false
	}
	return matchPathSegments(patternSegments[1:], targetSegments[1:])
}

func matchSegmentGlob(pattern, segment string) bool {
	ok, err := path.Match(pattern, segment)
	if err != nil {
		return pattern == segment
	}
	return ok
}

func splitPathSegments(v string) []string {
	if v == "" {
		return nil
	}
	raw := strings.Split(v, "/")
	segments := make([]string, 0, len(raw))
	for _, part := range raw {
		if part == "" {
			continue
		}
		segments = append(segments, part)
	}
	return segments
}

func unsupportedGlobReason(pattern string) string {
	p := strings.TrimSpace(pattern)
	if p == "" {
		return ""
	}
	if containsUnsupportedExtGlob(p) {
		return "extglob tokens (@(, +(, ?(, !(, *()) are not supported"
	}
	if strings.Count(p, "{") != strings.Count(p, "}") {
		return "unbalanced brace expansion"
	}

	expanded := expandBracePatterns(filepath.ToSlash(p), 32)
	for _, candidate := range expanded {
		for _, seg := range splitPathSegments(strings.TrimPrefix(candidate, "/")) {
			if seg == "**" {
				continue
			}
			if _, err := path.Match(seg, "x"); err != nil {
				return "invalid glob segment " + seg + ": " + err.Error()
			}
		}
	}
	return ""
}

func containsUnsupportedExtGlob(pattern string) bool {
	for _, marker := range []string{"@(", "+(", "?(", "!(", "*("} {
		if strings.Contains(pattern, marker) {
			return true
		}
	}
	return false
}

func expandBracePatterns(pattern string, maxVariants int) []string {
	if maxVariants <= 0 {
		maxVariants = 1
	}

	current := []string{pattern}
	for {
		changed := false
		next := make([]string, 0, len(current))
		for _, p := range current {
			start, end, options, ok := findBraceGroup(p)
			if !ok {
				next = append(next, p)
				continue
			}
			changed = true
			prefix := p[:start]
			suffix := p[end+1:]
			for _, opt := range options {
				next = append(next, prefix+opt+suffix)
				if len(next) >= maxVariants {
					return next
				}
			}
		}
		current = next
		if !changed {
			return current
		}
	}
}

func findBraceGroup(pattern string) (start int, end int, options []string, ok bool) {
	depth := 0
	groupStart := -1
	var optionStart int
	var opts []string

	for i, r := range pattern {
		switch r {
		case '{':
			if depth == 0 {
				groupStart = i
				optionStart = i + 1
			}
			depth++
		case '}':
			if depth == 0 {
				return 0, 0, nil, false
			}
			depth--
			if depth == 0 && groupStart >= 0 {
				opts = append(opts, pattern[optionStart:i])
				return groupStart, i, opts, true
			}
		case ',':
			if depth == 1 {
				opts = append(opts, pattern[optionStart:i])
				optionStart = i + 1
			}
		}
	}

	return 0, 0, nil, false
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
