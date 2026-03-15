package depgraph

import (
	"fmt"
	"sort"
)

// FanoutResult contains the fanout analysis for all nodes in a graph.
type FanoutResult struct {
	// Entries for all analyzed nodes, sorted by TransitiveFanout descending.
	Entries []FanoutEntry `json:"entries"`

	// Total nodes analyzed.
	NodeCount int `json:"nodeCount"`

	// Number of nodes exceeding the threshold.
	FlaggedCount int `json:"flaggedCount"`

	// Threshold used for flagging.
	Threshold int `json:"threshold"`

	// Indicates transitive fanout analysis was skipped for scale-safety.
	Skipped bool `json:"skipped,omitempty"`

	// Human-readable reason when transitive fanout is skipped.
	SkipReason string `json:"skipReason,omitempty"`
}

// FanoutEntry holds fanout metrics for a single node.
type FanoutEntry struct {
	// Node ID.
	NodeID string `json:"nodeId"`

	// Node type.
	NodeType string `json:"nodeType"`

	// File path (if available).
	Path string `json:"path,omitempty"`

	// Direct outgoing dependency count.
	Fanout int `json:"fanout"`

	// Transitive outgoing dependency count (BFS).
	TransitiveFanout int `json:"transitiveFanout"`

	// Whether this node exceeds the threshold.
	Flagged bool `json:"flagged"`
}

// DefaultFanoutThreshold is the default threshold for flagging excessive fanout.
const DefaultFanoutThreshold = 10

// maxFanoutNodes is the safety threshold for transitive fanout analysis.
// Above this size, exact transitive analysis is too expensive for interactive
// CLI use; we fall back to direct-fanout summary counts.
const maxFanoutNodes = 150000

// AnalyzeFanout computes direct and transitive fanout for every node.
//
// Direct fanout counts unique outgoing neighbors. Transitive fanout counts all
// reachable nodes. Uses reverse-topological traversal to compute transitive
// reachability for all nodes in a single O(n+e) pass instead of per-node BFS.
func AnalyzeFanout(g *Graph, threshold int) FanoutResult {
	if g == nil {
		return FanoutResult{}
	}
	if threshold <= 0 {
		threshold = DefaultFanoutThreshold
	}

	nodes := g.Nodes()

	// Build adjacency index: node → unique outgoing neighbors.
	adjIndex := make(map[string][]string, len(nodes))
	for _, n := range nodes {
		adjIndex[n.ID] = g.Neighbors(n.ID)
	}

	if len(nodes) > maxFanoutNodes {
		flagged := 0
		for _, n := range nodes {
			if len(adjIndex[n.ID]) >= threshold {
				flagged++
			}
		}
		return FanoutResult{
			NodeCount:    len(nodes),
			FlaggedCount: flagged,
			Threshold:    threshold,
			Skipped:      true,
			SkipReason:   fmt.Sprintf("skipped transitive fanout for %d nodes (limit %d); using direct fanout summary", len(nodes), maxFanoutNodes),
		}
	}

	// Compute transitive reachability using reverse-topological ordering.
	// For each node, its reachable set is the union of its neighbors' reachable
	// sets plus the neighbors themselves. Processing in reverse-topo order
	// ensures all successors are computed before their predecessors.
	reachable := computeTransitiveReachability(nodes, adjIndex)

	flagged := 0
	entries := make([]FanoutEntry, 0, len(nodes))
	for _, n := range nodes {
		direct := len(adjIndex[n.ID])
		transitive := reachable[n.ID]

		entry := FanoutEntry{
			NodeID:           n.ID,
			NodeType:         string(n.Type),
			Path:             n.Path,
			Fanout:           direct,
			TransitiveFanout: transitive,
			Flagged:          transitive >= threshold,
		}
		if entry.Flagged {
			flagged++
		}
		entries = append(entries, entry)
	}

	// Sort by transitive fanout descending, then by ID for determinism.
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].TransitiveFanout != entries[j].TransitiveFanout {
			return entries[i].TransitiveFanout > entries[j].TransitiveFanout
		}
		return entries[i].NodeID < entries[j].NodeID
	})

	return FanoutResult{
		Entries:      entries,
		NodeCount:    len(nodes),
		FlaggedCount: flagged,
		Threshold:    threshold,
	}
}

// computeTransitiveReachability computes the number of transitively reachable
// nodes for every node using a single pass in reverse-topological order.
//
// For DAGs this is exact. For graphs with cycles, it falls back to per-node
// BFS only for cycle members, keeping the fast path for the majority of nodes.
func computeTransitiveReachability(nodes []*Node, adj map[string][]string) map[string]int {
	// Compute in-degree for topological sort.
	inDegree := make(map[string]int, len(nodes))
	for _, n := range nodes {
		if _, ok := inDegree[n.ID]; !ok {
			inDegree[n.ID] = 0
		}
		for _, neighbor := range adj[n.ID] {
			inDegree[neighbor]++
		}
	}

	// Kahn's algorithm for topological sort.
	queue := make([]string, 0, len(nodes))
	head := 0
	for _, n := range nodes {
		if inDegree[n.ID] == 0 {
			queue = append(queue, n.ID)
		}
	}

	var topoOrder []string
	for head < len(queue) {
		cur := queue[head]
		head++
		topoOrder = append(topoOrder, cur)
		for _, neighbor := range adj[cur] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// If we got all nodes in topo order, the graph is a DAG.
	// Process in reverse topo order: for each node, its reachable count is
	// the size of the union of (neighbor + neighbor's reachable set).
	//
	// For exact counts we propagate reachable sets using bitmaps for small
	// graphs or bounded counting for large ones. For large graphs, we use
	// a conservative approach: sum neighbor reachable counts + direct neighbors,
	// capped by total node count. This may overcount due to shared descendants,
	// but matches the original BFS semantics for flagging purposes.
	//
	// For cycle nodes (not in topo order), fall back to BFS.
	result := make(map[string]int, len(nodes))

	if len(topoOrder) == len(nodes) {
		// Pure DAG — process in reverse topo order.
		// For exact counts, we track reachable sets. For very large graphs,
		// use BFS-equivalent counting via set propagation with size tracking.
		reachSets := make(map[string]map[string]bool, len(nodes))

		for i := len(topoOrder) - 1; i >= 0; i-- {
			nodeID := topoOrder[i]
			neighbors := adj[nodeID]

			if len(neighbors) == 0 {
				reachSets[nodeID] = map[string]bool{}
				result[nodeID] = 0
				continue
			}

			// For nodes with small reachable sets, propagate exactly.
			// For large sets, use BFS fallback to avoid excessive memory.
			totalReachable := 0
			for _, nb := range neighbors {
				totalReachable += result[nb] + 1
			}

			if totalReachable <= 1000 {
				// Check if any neighbor used BFS fallback (nil set).
				// If so, we can't propagate exactly and must use BFS too.
				needBFS := false
				for _, nb := range neighbors {
					if reachSets[nb] == nil {
						needBFS = true
						break
					}
				}
				if needBFS {
					result[nodeID] = bfsReachable(nodeID, adj)
					reachSets[nodeID] = nil
				} else {
					// Exact set propagation.
					rset := make(map[string]bool, totalReachable)
					for _, nb := range neighbors {
						rset[nb] = true
						for r := range reachSets[nb] {
							rset[r] = true
						}
					}
					reachSets[nodeID] = rset
					result[nodeID] = len(rset)
				}
			} else {
				// Large reachable set — use BFS for this node to get exact count.
				result[nodeID] = bfsReachable(nodeID, adj)
				reachSets[nodeID] = nil
			}
		}
	} else {
		// Graph has cycles — fall back to BFS for cycle members.
		inTopo := make(map[string]bool, len(topoOrder))
		for _, id := range topoOrder {
			inTopo[id] = true
		}

		// Process DAG portion in reverse topo order.
		reachSets := make(map[string]map[string]bool, len(topoOrder))
		for i := len(topoOrder) - 1; i >= 0; i-- {
			nodeID := topoOrder[i]
			neighbors := adj[nodeID]
			rset := make(map[string]bool)
			for _, nb := range neighbors {
				if inTopo[nb] {
					rset[nb] = true
					for r := range reachSets[nb] {
						rset[r] = true
					}
				}
			}
			reachSets[nodeID] = rset
			result[nodeID] = len(rset)
		}

		// BFS for cycle members.
		for _, n := range nodes {
			if !inTopo[n.ID] {
				result[n.ID] = bfsReachable(n.ID, adj)
			}
		}
	}

	return result
}

// bfsReachable counts the number of nodes reachable from start via BFS,
// excluding start itself. Uses index-based queue for O(1) dequeue.
func bfsReachable(start string, adj map[string][]string) int {
	visited := map[string]bool{start: true}
	queue := []string{start}
	head := 0

	for head < len(queue) {
		current := queue[head]
		head++
		for _, neighbor := range adj[current] {
			if !visited[neighbor] {
				visited[neighbor] = true
				queue = append(queue, neighbor)
			}
		}
	}

	return len(visited) - 1
}
