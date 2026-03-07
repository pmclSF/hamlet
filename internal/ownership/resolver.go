// Package ownership implements Hamlet's local ownership resolution.
//
// Ownership is used for grouping and visibility, not surveillance.
// It helps teams understand which areas of the codebase are affected
// by test quality, health, and migration findings.
//
// Resolution precedence:
//  1. Explicit Hamlet ownership config (.hamlet/ownership.yaml)
//  2. CODEOWNERS file matching
//  3. Directory-based fallback (top-level directory name)
//  4. "unknown" when nothing matches
//
// Future expansion points:
//   - git blame heuristics (not implemented yet)
//   - org-level ownership registry
//   - team identity resolution
package ownership

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const unknownOwner = "unknown"

// Rule maps a path prefix to an owner label.
type Rule struct {
	Path  string `yaml:"path"`
	Owner string `yaml:"owner"`
}

// Config is the explicit ownership configuration from .hamlet/ownership.yaml.
type Config struct {
	Rules []Rule `yaml:"rules"`
}

// Resolver resolves file ownership using configured rules and CODEOWNERS.
type Resolver struct {
	explicitRules []Rule
	codeowners    []codeownersEntry
}

type codeownersEntry struct {
	pattern string
	owner   string
}

// NewResolver creates a Resolver by loading ownership config and CODEOWNERS
// from the given repository root. Missing files are handled gracefully.
func NewResolver(repoRoot string) *Resolver {
	r := &Resolver{}
	r.loadExplicitConfig(repoRoot)
	r.loadCodeowners(repoRoot)
	return r
}

// Resolve returns the owner for a given repository-relative file path.
//
// Precedence:
//  1. Explicit Hamlet ownership config
//  2. CODEOWNERS
//  3. Top-level directory fallback
//  4. "unknown"
func (r *Resolver) Resolve(relPath string) string {
	// 1. Explicit config rules (longest prefix match)
	if owner := r.matchExplicit(relPath); owner != "" {
		return owner
	}

	// 2. CODEOWNERS (last matching entry wins, per GitHub convention)
	if owner := r.matchCodeowners(relPath); owner != "" {
		return owner
	}

	// 3. Directory fallback: use top-level directory name
	parts := strings.SplitN(filepath.ToSlash(relPath), "/", 2)
	if len(parts) > 1 && parts[0] != "" && parts[0] != "." {
		return parts[0]
	}

	// 4. Unknown
	return unknownOwner
}

func (r *Resolver) loadExplicitConfig(root string) {
	path := filepath.Join(root, ".hamlet", "ownership.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var ownershipFile struct {
		Ownership Config `yaml:"ownership"`
	}
	if err := yaml.Unmarshal(data, &ownershipFile); err != nil {
		return
	}
	r.explicitRules = ownershipFile.Ownership.Rules
}

func (r *Resolver) loadCodeowners(root string) {
	// Check standard CODEOWNERS locations
	candidates := []string{
		filepath.Join(root, "CODEOWNERS"),
		filepath.Join(root, ".github", "CODEOWNERS"),
		filepath.Join(root, "docs", "CODEOWNERS"),
	}

	for _, path := range candidates {
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				r.codeowners = append(r.codeowners, codeownersEntry{
					pattern: fields[0],
					owner:   fields[1],
				})
			}
		}
		return // Use first found CODEOWNERS file
	}
}

func (r *Resolver) matchExplicit(relPath string) string {
	normalized := filepath.ToSlash(relPath)
	var bestMatch string
	var bestLen int

	for _, rule := range r.explicitRules {
		prefix := filepath.ToSlash(rule.Path)
		// Ensure prefix matching works with or without trailing slash
		prefix = strings.TrimSuffix(prefix, "/")
		if strings.HasPrefix(normalized, prefix) && len(prefix) > bestLen {
			bestMatch = rule.Owner
			bestLen = len(prefix)
		}
	}
	return bestMatch
}

func (r *Resolver) matchCodeowners(relPath string) string {
	normalized := filepath.ToSlash(relPath)
	var lastMatch string

	for _, entry := range r.codeowners {
		pattern := entry.pattern
		// Simple prefix matching (covers most practical CODEOWNERS patterns)
		// Full glob matching could be added later if needed.
		cleanPattern := strings.TrimSuffix(strings.TrimPrefix(pattern, "/"), "/")

		if strings.HasPrefix(normalized, cleanPattern) {
			lastMatch = entry.owner
		}
		// Also handle wildcard suffix patterns like "*.js"
		if strings.HasPrefix(pattern, "*") {
			ext := strings.TrimPrefix(pattern, "*")
			if strings.HasSuffix(normalized, ext) {
				lastMatch = entry.owner
			}
		}
	}
	return lastMatch
}
