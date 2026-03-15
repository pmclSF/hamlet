package reasoning

import (
	"path/filepath"
	"strings"

	"github.com/pmclSF/terrain/internal/depgraph"
)

// FallbackStrategy describes which fallback expansion method was used.
type FallbackStrategy string

const (
	FallbackNone      FallbackStrategy = "none"
	FallbackPackage   FallbackStrategy = "package"
	FallbackDirectory FallbackStrategy = "directory"
	FallbackFamily    FallbackStrategy = "family"
	FallbackAll       FallbackStrategy = "all"
)

// FallbackResult contains the expanded node set from a fallback strategy.
type FallbackResult struct {
	// Strategy is the fallback method that was used.
	Strategy FallbackStrategy

	// NodeIDs are the nodes selected by the fallback.
	NodeIDs []string

	// Explanation describes why fallback was triggered and what was selected.
	Explanation string
}

// FallbackConfig controls fallback expansion behavior.
type FallbackConfig struct {
	// MinResults is the minimum number of results before triggering fallback.
	// Default: 1.
	MinResults int

	// Strategies is the ordered list of fallback strategies to try.
	// Default: [package, directory].
	Strategies []FallbackStrategy
}

// DefaultFallbackConfig returns standard fallback parameters.
func DefaultFallbackConfig() FallbackConfig {
	return FallbackConfig{
		MinResults: 1,
		Strategies: []FallbackStrategy{FallbackPackage, FallbackDirectory},
	}
}

// ExpandFallback tries fallback strategies when primary analysis yields
// insufficient results. It returns the first strategy that produces results.
//
// This is used when impact analysis finds no tests for a change, or when
// coverage analysis finds no covering tests for a file. The fallback
// progressively widens the search scope.
func ExpandFallback(g *depgraph.Graph, seedNodeIDs []string, currentResults []string, cfg FallbackConfig) FallbackResult {
	if cfg.MinResults <= 0 {
		cfg.MinResults = 1
	}
	if len(cfg.Strategies) == 0 {
		cfg.Strategies = []FallbackStrategy{FallbackPackage, FallbackDirectory}
	}

	// If we already have enough results, no fallback needed.
	if len(currentResults) >= cfg.MinResults {
		return FallbackResult{Strategy: FallbackNone}
	}

	// Collect seed node metadata for fallback lookups.
	var seedPaths []string
	var seedPackages []string
	for _, id := range seedNodeIDs {
		node := g.Node(id)
		if node == nil {
			continue
		}
		if node.Path != "" {
			seedPaths = append(seedPaths, node.Path)
		}
		if node.Package != "" {
			seedPackages = append(seedPackages, node.Package)
		}
	}

	exclude := map[string]bool{}
	for _, id := range seedNodeIDs {
		exclude[id] = true
	}
	for _, id := range currentResults {
		exclude[id] = true
	}

	for _, strategy := range cfg.Strategies {
		var result FallbackResult
		switch strategy {
		case FallbackPackage:
			result = fallbackByPackage(g, seedPackages, exclude)
		case FallbackDirectory:
			result = fallbackByDirectory(g, seedPaths, exclude)
		case FallbackFamily:
			result = fallbackByFamily(g, seedNodeIDs, exclude)
		case FallbackAll:
			result = fallbackAll(g, exclude)
		default:
			continue
		}
		if len(result.NodeIDs) > 0 {
			return result
		}
	}

	return FallbackResult{
		Strategy:    FallbackNone,
		Explanation: "no fallback strategy produced results",
	}
}

// fallbackByPackage finds validation nodes in the same package(s) as the seeds.
func fallbackByPackage(g *depgraph.Graph, packages []string, exclude map[string]bool) FallbackResult {
	if len(packages) == 0 {
		return FallbackResult{Strategy: FallbackPackage}
	}

	pkgSet := map[string]bool{}
	for _, p := range packages {
		pkgSet[p] = true
	}

	var nodeIDs []string
	for _, n := range g.NodesByType(depgraph.NodeTest) {
		if exclude[n.ID] {
			continue
		}
		if pkgSet[n.Package] {
			nodeIDs = append(nodeIDs, n.ID)
		}
	}

	if len(nodeIDs) == 0 {
		return FallbackResult{Strategy: FallbackPackage}
	}

	return FallbackResult{
		Strategy:    FallbackPackage,
		NodeIDs:     nodeIDs,
		Explanation: "expanded to tests in same package(s): " + strings.Join(packages, ", "),
	}
}

// fallbackByDirectory finds validation nodes in the same directory as the seeds.
func fallbackByDirectory(g *depgraph.Graph, paths []string, exclude map[string]bool) FallbackResult {
	if len(paths) == 0 {
		return FallbackResult{Strategy: FallbackDirectory}
	}

	dirs := map[string]bool{}
	for _, p := range paths {
		dirs[filepath.Dir(p)] = true
	}

	var nodeIDs []string
	for _, n := range g.NodesByType(depgraph.NodeTest) {
		if exclude[n.ID] {
			continue
		}
		if n.Path != "" && dirs[filepath.Dir(n.Path)] {
			nodeIDs = append(nodeIDs, n.ID)
		}
	}

	if len(nodeIDs) == 0 {
		return FallbackResult{Strategy: FallbackDirectory}
	}

	dirList := make([]string, 0, len(dirs))
	for d := range dirs {
		dirList = append(dirList, d)
	}

	return FallbackResult{
		Strategy:    FallbackDirectory,
		NodeIDs:     nodeIDs,
		Explanation: "expanded to tests in same directory: " + strings.Join(dirList, ", "),
	}
}

// fallbackByFamily finds validation nodes in the same node family as the seeds.
func fallbackByFamily(g *depgraph.Graph, seedNodeIDs []string, exclude map[string]bool) FallbackResult {
	families := map[depgraph.NodeFamily]bool{}
	for _, id := range seedNodeIDs {
		node := g.Node(id)
		if node != nil {
			families[depgraph.NodeTypeFamily(node.Type)] = true
		}
	}

	var nodeIDs []string
	for fam := range families {
		for _, n := range g.NodesByFamily(fam) {
			if !exclude[n.ID] && depgraph.IsValidationNode(n.Type) {
				nodeIDs = append(nodeIDs, n.ID)
			}
		}
	}

	if len(nodeIDs) == 0 {
		return FallbackResult{Strategy: FallbackFamily}
	}

	return FallbackResult{
		Strategy:    FallbackFamily,
		NodeIDs:     nodeIDs,
		Explanation: "expanded to validation nodes in same family",
	}
}

// fallbackAll returns all validation nodes as a last resort.
func fallbackAll(g *depgraph.Graph, exclude map[string]bool) FallbackResult {
	targets := g.ValidationTargets()
	var nodeIDs []string
	for _, n := range targets {
		if !exclude[n.ID] {
			nodeIDs = append(nodeIDs, n.ID)
		}
	}

	if len(nodeIDs) == 0 {
		return FallbackResult{Strategy: FallbackAll}
	}

	return FallbackResult{
		Strategy:    FallbackAll,
		NodeIDs:     nodeIDs,
		Explanation: "expanded to all validation targets (full test suite fallback)",
	}
}
