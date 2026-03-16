package stability

import (
	"fmt"
	"sort"

	"github.com/pmclSF/terrain/internal/depgraph"
	"github.com/pmclSF/terrain/internal/models"
)

// CauseKind classifies the type of shared dependency that likely causes
// instability in a cluster of tests.
type CauseKind string

const (
	CauseFixture         CauseKind = "shared_fixture"
	CauseHelper          CauseKind = "shared_helper"
	CauseEnvironment     CauseKind = "shared_environment"
	CauseExternalService CauseKind = "external_service"
	CauseSourceFile      CauseKind = "shared_source"
)

// Cluster represents a group of unstable tests that share a common
// dependency, suggesting a shared root cause for their instability.
type Cluster struct {
	// ID is a deterministic identifier for this cluster.
	ID string `json:"id"`

	// CauseKind classifies the type of shared dependency.
	CauseKind CauseKind `json:"causeKind"`

	// CauseNodeID is the graph node ID of the shared dependency.
	CauseNodeID string `json:"causeNodeId"`

	// CauseName is a human-readable name for the shared dependency.
	CauseName string `json:"causeName"`

	// CausePath is the file path of the shared dependency, if any.
	CausePath string `json:"causePath,omitempty"`

	// Members are the test node IDs in this cluster.
	Members []string `json:"members"`

	// MemberNames are the human-readable test names.
	MemberNames []string `json:"memberNames"`

	// Confidence is the likelihood that this shared dependency is
	// the root cause (0.0–1.0).
	Confidence float64 `json:"confidence"`

	// Remediation is a suggested action for stabilizing this cluster.
	Remediation string `json:"remediation"`
}

// ClusterResult holds all stability clusters found in the graph.
type ClusterResult struct {
	// Clusters are the detected instability clusters, sorted by
	// size (largest first), then confidence (highest first).
	Clusters []Cluster `json:"clusters"`

	// UnstableTestCount is the total number of unstable test nodes.
	UnstableTestCount int `json:"unstableTestCount"`

	// ClusteredTestCount is how many unstable tests belong to at
	// least one cluster (i.e., share a dependency with another
	// unstable test).
	ClusteredTestCount int `json:"clusteredTestCount"`
}

// causeCategory maps graph node types to CauseKind and ranks their
// priority for root cause attribution. Lower rank = more likely cause.
var causeCategory = map[depgraph.NodeType]struct {
	Kind CauseKind
	Rank int
}{
	depgraph.NodeEnvironment: {CauseEnvironment, 1},
	depgraph.NodeSourceFile:  {CauseSourceFile, 2},
}

// edgeToSharedDep maps edge types to the dependency direction.
// These are edges originating FROM a test/test-file node.
var edgeToSharedDep = map[depgraph.EdgeType]bool{
	depgraph.EdgeImportsModule:      true,
	depgraph.EdgeTargetsEnvironment: true,
}

// DetectClusters identifies groups of unstable tests that share common
// dependencies in the graph, suggesting shared root causes for instability.
//
// The algorithm:
//  1. Collect all test node IDs that have flaky or unstable signals.
//  2. For each unstable test, walk outgoing edges to find shared
//     dependencies (fixtures, helpers, environments, services, resources).
//  3. Group unstable tests by shared dependency — if ≥2 unstable tests
//     share a dependency, that forms a cluster.
//  4. Rank clusters by size and confidence. Deduplicate tests that appear
//     in multiple clusters by keeping the highest-confidence assignment.
func DetectClusters(g *depgraph.Graph, signals []models.Signal) *ClusterResult {
	result := &ClusterResult{}

	if g == nil {
		return result
	}

	// Step 1: Identify unstable test node IDs from signals.
	unstableFiles := collectUnstableFiles(signals)
	if len(unstableFiles) == 0 {
		return result
	}

	// Map file paths to test node IDs in the graph.
	unstableTestIDs := resolveUnstableTests(g, unstableFiles)
	result.UnstableTestCount = len(unstableTestIDs)
	if result.UnstableTestCount < 2 {
		return result
	}

	// Step 2: For each unstable test, collect shared dependencies.
	// depToTests maps dependency node ID → set of unstable test node IDs.
	depToTests := map[string]map[string]bool{}
	for testID := range unstableTestIDs {
		deps := collectSharedDeps(g, testID)
		for depID := range deps {
			if depToTests[depID] == nil {
				depToTests[depID] = map[string]bool{}
			}
			depToTests[depID][testID] = true
		}
	}

	// Step 3: Build clusters from dependencies shared by ≥2 unstable tests.
	var clusters []Cluster
	for depID, testIDs := range depToTests {
		if len(testIDs) < 2 {
			continue
		}

		node := g.Node(depID)
		if node == nil {
			continue
		}

		cat, ok := causeCategory[node.Type]
		if !ok {
			continue
		}

		members := sortedKeys(testIDs)
		memberNames := resolveNames(g, members)

		cluster := Cluster{
			ID:          fmt.Sprintf("stability:%s:%s", cat.Kind, depID),
			CauseKind:   cat.Kind,
			CauseNodeID: depID,
			CauseName:   nodeName(node),
			CausePath:   node.Path,
			Members:     members,
			MemberNames: memberNames,
			Confidence:  clusterConfidence(cat.Kind, len(members), result.UnstableTestCount),
			Remediation: clusterRemediation(cat.Kind, nodeName(node)),
		}
		clusters = append(clusters, cluster)
	}

	// Step 4: Sort by size desc, then confidence desc, then ID for determinism.
	sort.Slice(clusters, func(i, j int) bool {
		if len(clusters[i].Members) != len(clusters[j].Members) {
			return len(clusters[i].Members) > len(clusters[j].Members)
		}
		if clusters[i].Confidence != clusters[j].Confidence {
			return clusters[i].Confidence > clusters[j].Confidence
		}
		return clusters[i].ID < clusters[j].ID
	})

	result.Clusters = clusters

	// Count unique clustered tests.
	clustered := map[string]bool{}
	for _, c := range clusters {
		for _, m := range c.Members {
			clustered[m] = true
		}
	}
	result.ClusteredTestCount = len(clustered)

	return result
}

// collectUnstableFiles extracts file paths from flaky/unstable signals.
func collectUnstableFiles(signals []models.Signal) map[string]bool {
	files := map[string]bool{}
	for _, sig := range signals {
		if sig.Type == "flakyTest" || sig.Type == "unstableSuite" {
			if sig.Location.File != "" {
				files[sig.Location.File] = true
			}
		}
	}
	return files
}

// resolveUnstableTests maps unstable file paths to test node IDs in the graph.
func resolveUnstableTests(g *depgraph.Graph, unstableFiles map[string]bool) map[string]bool {
	testIDs := map[string]bool{}
	for _, n := range g.NodesByType(depgraph.NodeTest) {
		if unstableFiles[n.Path] {
			testIDs[n.ID] = true
		}
	}
	// Also check test file nodes — if a test file is flagged, include
	// all tests defined in that file.
	for _, n := range g.NodesByType(depgraph.NodeTestFile) {
		if unstableFiles[n.Path] {
			for _, e := range g.Incoming(n.ID) {
				if e.Type == depgraph.EdgeTestDefinedInFile {
					testIDs[e.From] = true
				}
			}
		}
	}
	return testIDs
}

// collectSharedDeps walks outgoing edges from a test (and its test file)
// to find shared infrastructure nodes.
func collectSharedDeps(g *depgraph.Graph, testID string) map[string]bool {
	deps := map[string]bool{}
	visited := map[string]bool{testID: true}
	queue := []string{testID}

	// Also include the test's file node as a starting point.
	testNode := g.Node(testID)
	if testNode != nil {
		for _, e := range g.Outgoing(testID) {
			if e.Type == depgraph.EdgeTestDefinedInFile {
				if !visited[e.To] {
					visited[e.To] = true
					queue = append(queue, e.To)
				}
			}
		}
	}

	// BFS one hop from test and test file nodes.
	for _, startID := range queue {
		for _, e := range g.Outgoing(startID) {
			if !edgeToSharedDep[e.Type] {
				continue
			}
			target := g.Node(e.To)
			if target == nil {
				continue
			}
			if _, ok := causeCategory[target.Type]; ok {
				deps[e.To] = true
			}
		}
	}

	return deps
}

// clusterConfidence computes confidence that a shared dependency is the
// root cause of instability. Higher when:
//   - The cause kind is more likely to introduce instability (fixtures > source files)
//   - More unstable tests share the dependency (higher concentration)
func clusterConfidence(kind CauseKind, clusterSize, totalUnstable int) float64 {
	// Base confidence by cause kind.
	var base float64
	switch kind {
	case CauseFixture:
		base = 0.8
	case CauseHelper:
		base = 0.7
	case CauseExternalService:
		base = 0.85
	case CauseEnvironment:
		base = 0.75
	case CauseSourceFile:
		base = 0.5
	default:
		base = 0.4
	}

	// Boost by concentration: what fraction of unstable tests share this dep?
	if totalUnstable > 0 {
		concentration := float64(clusterSize) / float64(totalUnstable)
		base += 0.15 * concentration
	}

	if base > 1.0 {
		base = 1.0
	}
	return base
}

// clusterRemediation returns a human-readable remediation suggestion.
func clusterRemediation(kind CauseKind, name string) string {
	switch kind {
	case CauseFixture:
		return fmt.Sprintf("Audit fixture %q for shared mutable state, timing dependencies, or teardown gaps.", name)
	case CauseHelper:
		return fmt.Sprintf("Review helper %q for side effects or non-deterministic behavior that may leak across tests.", name)
	case CauseExternalService:
		return fmt.Sprintf("External service dependency %q is a likely flake source. Consider stubbing or adding retry/circuit-breaker logic.", name)
	case CauseEnvironment:
		return fmt.Sprintf("Shared environment %q may introduce timing or resource contention. Consider test isolation or dedicated environments.", name)
	case CauseSourceFile:
		return fmt.Sprintf("Multiple unstable tests depend on %q. Investigate recent changes for introduced non-determinism.", name)
	default:
		return fmt.Sprintf("Investigate shared dependency %q for instability sources.", name)
	}
}

// nodeName returns a human-readable name for a node.
func nodeName(n *depgraph.Node) string {
	if n.Name != "" {
		return n.Name
	}
	if n.Path != "" {
		return n.Path
	}
	return n.ID
}

// resolveNames maps node IDs to human-readable names.
func resolveNames(g *depgraph.Graph, ids []string) []string {
	names := make([]string, len(ids))
	for i, id := range ids {
		if n := g.Node(id); n != nil {
			names[i] = nodeName(n)
		} else {
			names[i] = id
		}
	}
	return names
}

// sortedKeys returns the keys of a map sorted alphabetically.
func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
