// Package ownership implements Terrain's normalized ownership subsystem.
//
// Ownership is a routing layer, not a blame layer. It exists to make
// findings actionable by connecting risk, health, quality, and migration
// data to the people and teams who can act on it.
//
// Resolution precedence (highest to lowest):
//  1. Explicit Terrain ownership config (.terrain/ownership.yaml)
//  2. CODEOWNERS file matching
//  3. Path-prefix mapping (.terrain/ownership.yaml path_mappings)
//  4. Git-history fallback (.terrain/ownership.yaml git_history or auto when CODEOWNERS is absent)
//  5. Directory-based fallback (top-level directory name)
//  6. "unknown" when nothing matches
//
// Each resolution level produces a full OwnershipAssignment with
// provenance, confidence, and inheritance metadata.
package ownership

import (
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

// PathMapping maps a path prefix to one or more owners.
type PathMapping struct {
	Prefix string   `yaml:"prefix"`
	Owners []string `yaml:"owners"`
}

// GitHistoryConfig controls optional git-history ownership fallback.
type GitHistoryConfig struct {
	Enabled              bool  `yaml:"enabled"`
	MaxCommits           int   `yaml:"max_commits"`
	AutoWhenNoCodeowners *bool `yaml:"auto_when_no_codeowners"`
}

// Config is the explicit ownership configuration from .terrain/ownership.yaml.
type Config struct {
	Rules        []Rule           `yaml:"rules"`
	PathMappings []PathMapping    `yaml:"path_mappings"`
	GitHistory   GitHistoryConfig `yaml:"git_history"`
}

// Resolver resolves file ownership using configured rules and CODEOWNERS.
//
// The resolver loads all available ownership sources at construction time
// and evaluates them in precedence order during resolution.
type Resolver struct {
	repoRoot          string
	explicitRules     []Rule
	pathMappings      []PathMapping
	codeowners        *CodeownersFile
	gitHistoryEnabled bool
	// gitHistoryAutoNoCodeowners enables automatic git-history fallback when
	// no CODEOWNERS file is present. This is enabled by default and can be
	// disabled via ownership.git_history.auto_when_no_codeowners: false.
	gitHistoryAutoNoCodeowners bool
	gitHistoryMaxLogs int
	gitHistoryOwners  map[string]string
	gitHistoryLoaded  bool
	diagnostics       []Diagnostic
}

// NewResolver creates a Resolver by loading ownership config and CODEOWNERS
// from the given repository root. Missing files are handled gracefully.
func NewResolver(repoRoot string) *Resolver {
	r := &Resolver{
		repoRoot:                    repoRoot,
		gitHistoryAutoNoCodeowners:  true,
	}
	r.loadExplicitConfig(repoRoot)
	r.loadCodeownersFile(repoRoot)
	return r
}

// Resolve returns the primary owner ID for a given repository-relative file path.
// This is the backward-compatible API used by existing callers.
//
// For full ownership metadata, use ResolveAssignment instead.
func (r *Resolver) Resolve(relPath string) string {
	a := r.ResolveAssignment(relPath)
	return a.PrimaryOwnerID()
}

// ResolveAssignment returns a full OwnershipAssignment for a file path,
// including all owners, provenance, confidence, and source metadata.
func (r *Resolver) ResolveAssignment(relPath string) OwnershipAssignment {
	// 1. Explicit config rules (longest prefix match, highest precedence).
	if a, ok := r.matchExplicitAssignment(relPath); ok {
		return a
	}

	// 2. CODEOWNERS (last matching entry wins, per GitHub convention).
	if a, ok := r.matchCodeownersAssignment(relPath); ok {
		return a
	}

	// 3. Path mappings from config.
	if a, ok := r.matchPathMapping(relPath); ok {
		return a
	}

	// 4. Git-history fallback.
	if a, ok := r.matchGitHistoryAssignment(relPath); ok {
		return a
	}

	// 5. Directory fallback: use top-level directory name.
	normalized := filepath.ToSlash(relPath)
	parts := strings.SplitN(normalized, "/", 2)
	if len(parts) > 1 && parts[0] != "" && parts[0] != "." {
		return OwnershipAssignment{
			Owners:      []Owner{{ID: parts[0]}},
			Source:      SourceDirectoryFallback,
			Confidence:  ConfidenceLow,
			Inheritance: InheritanceDirect,
		}
	}

	// 6. Unknown.
	return OwnershipAssignment{
		Source:     SourceUnknown,
		Confidence: ConfidenceNone,
	}
}

// ResolveAll resolves ownership for a list of file paths and returns
// a map of path to assignment.
func (r *Resolver) ResolveAll(paths []string) map[string]OwnershipAssignment {
	result := make(map[string]OwnershipAssignment, len(paths))
	for _, p := range paths {
		result[p] = r.ResolveAssignment(p)
	}
	return result
}

// InheritFrom creates an inherited assignment from a parent (file-level)
// assignment. The inheritance kind is set to Inherited and the source
// metadata is preserved.
func InheritFrom(parent OwnershipAssignment) OwnershipAssignment {
	return OwnershipAssignment{
		Owners:      parent.Owners,
		Source:      parent.Source,
		Confidence:  parent.Confidence,
		Inheritance: InheritanceInherited,
		MatchedRule: parent.MatchedRule,
		SourceFile:  parent.SourceFile,
	}
}

// Diagnostics returns any warnings or issues from loading ownership sources.
func (r *Resolver) Diagnostics() []Diagnostic {
	var diags []Diagnostic
	diags = append(diags, r.diagnostics...)
	if r.codeowners != nil {
		diags = append(diags, r.codeowners.Diagnostics...)
	}
	return diags
}

// SourcesUsed returns which ownership sources were loaded.
func (r *Resolver) SourcesUsed() []SourceType {
	var sources []SourceType
	if len(r.explicitRules) > 0 {
		sources = append(sources, SourceExplicitConfig)
	}
	if r.codeowners != nil && len(r.codeowners.Rules) > 0 {
		sources = append(sources, SourceCodeowners)
	}
	if len(r.pathMappings) > 0 {
		sources = append(sources, SourcePathMapping)
	}
	if r.gitHistoryEnabled || (r.gitHistoryAutoNoCodeowners && !r.HasCodeowners()) {
		sources = append(sources, SourceGitHistory)
	}
	return sources
}

// HasCodeowners returns true if a CODEOWNERS file was found and parsed.
func (r *Resolver) HasCodeowners() bool {
	return r.codeowners != nil && len(r.codeowners.Rules) > 0
}

func (r *Resolver) loadExplicitConfig(root string) {
	path := filepath.Join(root, ".terrain", "ownership.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var ownershipFile struct {
		Ownership Config `yaml:"ownership"`
	}
	if err := yaml.Unmarshal(data, &ownershipFile); err != nil {
		r.diagnostics = append(r.diagnostics, Diagnostic{
			Level:   "warning",
			Message: "failed to parse ownership config: " + err.Error(),
			Source:  ".terrain/ownership.yaml",
		})
		return
	}
	r.explicitRules = ownershipFile.Ownership.Rules
	r.pathMappings = ownershipFile.Ownership.PathMappings
	r.gitHistoryEnabled = ownershipFile.Ownership.GitHistory.Enabled
	r.gitHistoryMaxLogs = ownershipFile.Ownership.GitHistory.MaxCommits
	if ownershipFile.Ownership.GitHistory.AutoWhenNoCodeowners != nil {
		r.gitHistoryAutoNoCodeowners = *ownershipFile.Ownership.GitHistory.AutoWhenNoCodeowners
	}
	if r.gitHistoryEnabled && r.gitHistoryMaxLogs <= 0 {
		r.gitHistoryMaxLogs = 1000
	}
}

func (r *Resolver) loadCodeownersFile(root string) {
	absPath, relPath, found := FindCodeownersFile(root)
	if !found {
		return
	}
	r.codeowners = ParseCodeownersFile(absPath, relPath)
}

func (r *Resolver) matchExplicitAssignment(relPath string) (OwnershipAssignment, bool) {
	normalized := filepath.ToSlash(relPath)
	var bestRule *Rule
	var bestLen int

	for i := range r.explicitRules {
		rule := &r.explicitRules[i]
		prefix := filepath.ToSlash(rule.Path)
		prefix = strings.TrimSuffix(prefix, "/")
		if pathHasPrefix(normalized, prefix) && len(prefix) > bestLen {
			bestRule = rule
			bestLen = len(prefix)
		}
	}

	if bestRule == nil {
		return OwnershipAssignment{}, false
	}

	return OwnershipAssignment{
		Owners:      []Owner{{ID: NormalizeOwnerID(bestRule.Owner)}},
		Source:      SourceExplicitConfig,
		Confidence:  ConfidenceHigh,
		Inheritance: InheritanceDirect,
		MatchedRule: bestRule.Path,
		SourceFile:  ".terrain/ownership.yaml",
	}, true
}

func (r *Resolver) matchCodeownersAssignment(relPath string) (OwnershipAssignment, bool) {
	if r.codeowners == nil || len(r.codeowners.Rules) == 0 {
		return OwnershipAssignment{}, false
	}

	rule, matched := MatchCodeowners(r.codeowners.Rules, relPath)
	if !matched {
		return OwnershipAssignment{}, false
	}

	return rule.ToAssignment(r.codeowners.Path), true
}

func (r *Resolver) matchPathMapping(relPath string) (OwnershipAssignment, bool) {
	normalized := filepath.ToSlash(relPath)
	var bestMapping *PathMapping
	var bestLen int

	for i := range r.pathMappings {
		pm := &r.pathMappings[i]
		prefix := filepath.ToSlash(pm.Prefix)
		prefix = strings.TrimSuffix(prefix, "/")
		if pathHasPrefix(normalized, prefix) && len(prefix) > bestLen {
			bestMapping = pm
			bestLen = len(prefix)
		}
	}

	if bestMapping == nil {
		return OwnershipAssignment{}, false
	}

	owners := make([]Owner, len(bestMapping.Owners))
	for i, o := range bestMapping.Owners {
		owners[i] = Owner{ID: NormalizeOwnerID(o)}
	}

	return OwnershipAssignment{
		Owners:      owners,
		Source:      SourcePathMapping,
		Confidence:  ConfidenceMedium,
		Inheritance: InheritanceDirect,
		MatchedRule: bestMapping.Prefix,
		SourceFile:  ".terrain/ownership.yaml",
	}, true
}

func pathHasPrefix(path, prefix string) bool {
	if prefix == "" || prefix == "." {
		return false
	}

	normPath := filepath.ToSlash(filepath.Clean(path))
	normPrefix := filepath.ToSlash(filepath.Clean(prefix))
	if normPrefix == "." {
		return false
	}

	if normPath == normPrefix {
		return true
	}

	return strings.HasPrefix(normPath, normPrefix+"/")
}

func (r *Resolver) shouldTryGitHistory() bool {
	if r.gitHistoryEnabled {
		return true
	}
	if !r.gitHistoryAutoNoCodeowners {
		return false
	}
	return !r.HasCodeowners()
}
