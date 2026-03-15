package reasoning

import (
	"github.com/pmclSF/terrain/internal/depgraph"
)

// TraversalConfig controls BFS behavior.
type TraversalConfig struct {
	// MaxDepth caps the number of hops. Default: 20.
	MaxDepth int

	// MinConfidence prunes paths below this threshold. Default: 0.1.
	MinConfidence float64

	// LengthDecay is the per-hop decay factor. Default: 0.85.
	LengthDecay float64

	// FanoutThreshold triggers fanout penalty when a node's out-degree
	// exceeds this value. Default: 5.
	FanoutThreshold int

	// Direction controls traversal direction.
	// "reverse" follows incoming edges (dependents → source).
	// "forward" follows outgoing edges (source → dependents).
	Direction string

	// StopAt is a predicate that stops traversal at matching nodes.
	// The node is still recorded in results but not traversed further.
	// Nil means no stop condition.
	StopAt func(n *depgraph.Node) bool

	// EdgeFilter selects which edges to traverse. Nil means all edges.
	EdgeFilter func(e *depgraph.Edge) bool
}

// DefaultTraversalConfig returns a TraversalConfig with standard defaults
// matching the existing impact analysis parameters.
func DefaultTraversalConfig() TraversalConfig {
	return TraversalConfig{
		MaxDepth:        20,
		MinConfidence:   0.1,
		LengthDecay:     0.85,
		FanoutThreshold: 5,
		Direction:       "reverse",
	}
}

// ReachResult is a node discovered during traversal.
type ReachResult struct {
	// NodeID is the discovered node's ID.
	NodeID string

	// Confidence is the composite confidence at this node.
	Confidence float64

	// Depth is the number of hops from the origin.
	Depth int

	// Chain is the sequence of edges traversed to reach this node.
	Chain []Step

	// Origin is the starting node ID that led to this result.
	Origin string
}

// Step describes one hop in a traversal path.
type Step struct {
	From           string  `json:"from"`
	To             string  `json:"to"`
	EdgeType       string  `json:"edgeType"`
	EdgeConfidence float64 `json:"edgeConfidence"`
}

// traversalState tracks BFS queue entries.
type traversalState struct {
	nodeID     string
	confidence float64
	chain      []Step
	depth      int
	origin     string
}

// Reachable performs BFS from the given start nodes through the graph,
// applying confidence decay at each hop. Returns all reachable nodes
// with their confidence scores and reason chains.
//
// This is the core traversal primitive used by impact analysis, coverage
// queries, and any other analysis that needs to propagate through the graph
// with confidence tracking.
func Reachable(g *depgraph.Graph, startNodes []string, cfg TraversalConfig) []ReachResult {
	if g == nil || len(startNodes) == 0 {
		return nil
	}

	cfg = applyDefaults(cfg)

	// Compute out-degree for fanout penalty.
	outDegree := map[string]int{}
	for _, n := range g.Nodes() {
		outDegree[n.ID] = len(g.Neighbors(n.ID))
	}

	// Best confidence seen per node.
	best := map[string]*ReachResult{}

	// Initialize queue with start nodes.
	queue := make([]traversalState, 0, len(startNodes))
	for _, id := range startNodes {
		if g.Node(id) == nil {
			continue
		}
		queue = append(queue, traversalState{
			nodeID:     id,
			confidence: 1.0,
			depth:      0,
			origin:     id,
		})
	}

	head := 0
	visited := map[string]float64{}

	for head < len(queue) {
		cur := queue[head]
		head++

		// Skip if seen at higher confidence.
		if prev, ok := visited[cur.nodeID]; ok && prev >= cur.confidence {
			continue
		}
		visited[cur.nodeID] = cur.confidence

		node := g.Node(cur.nodeID)
		if node == nil {
			continue
		}

		// Record this node (overwrite if higher confidence).
		if existing, ok := best[cur.nodeID]; !ok || cur.confidence > existing.Confidence {
			best[cur.nodeID] = &ReachResult{
				NodeID:     cur.nodeID,
				Confidence: cur.confidence,
				Depth:      cur.depth,
				Chain:      cloneSteps(cur.chain),
				Origin:     cur.origin,
			}
		}

		// Stop conditions: don't traverse past this node.
		if cfg.StopAt != nil && cfg.StopAt(node) {
			continue
		}

		// Depth cap.
		if cur.depth >= cfg.MaxDepth {
			continue
		}

		// Get edges based on direction.
		var edges []*depgraph.Edge
		if cfg.Direction == "forward" {
			edges = g.Outgoing(cur.nodeID)
		} else {
			edges = g.Incoming(cur.nodeID)
		}

		for _, e := range edges {
			// Apply edge filter if configured.
			if cfg.EdgeFilter != nil && !cfg.EdgeFilter(e) {
				continue
			}

			// Compute next node ID based on direction.
			nextID := e.From
			if cfg.Direction == "forward" {
				nextID = e.To
			}

			// Score the hop.
			newConf := ScoreHop(cur.confidence, e.Confidence, cur.depth, outDegree[nextID], cfg.LengthDecay, cfg.FanoutThreshold)

			if newConf < cfg.MinConfidence {
				continue
			}

			step := Step{
				From:           cur.nodeID,
				To:             nextID,
				EdgeType:       string(e.Type),
				EdgeConfidence: e.Confidence,
			}

			queue = append(queue, traversalState{
				nodeID:     nextID,
				confidence: newConf,
				chain:      appendStep(cur.chain, step),
				depth:      cur.depth + 1,
				origin:     cur.origin,
			})
		}
	}

	// Convert to sorted results (by confidence descending, then ID).
	results := make([]ReachResult, 0, len(best))
	for _, r := range best {
		// Exclude start nodes from results.
		isStart := false
		for _, s := range startNodes {
			if r.NodeID == s {
				isStart = true
				break
			}
		}
		if !isStart {
			results = append(results, *r)
		}
	}

	sortResults(results)
	return results
}

// ReachableNodes returns just the node IDs reachable from the start nodes,
// without confidence tracking. Useful for simple reachability queries.
func ReachableNodes(g *depgraph.Graph, startNodes []string, cfg TraversalConfig) []string {
	// For simple reachability, use a high min confidence to avoid unnecessary work,
	// but still respect the caller's setting.
	results := Reachable(g, startNodes, cfg)
	ids := make([]string, len(results))
	for i, r := range results {
		ids[i] = r.NodeID
	}
	return ids
}

func applyDefaults(cfg TraversalConfig) TraversalConfig {
	if cfg.MaxDepth <= 0 {
		cfg.MaxDepth = 20
	}
	if cfg.MinConfidence <= 0 {
		cfg.MinConfidence = 0.1
	}
	if cfg.LengthDecay <= 0 {
		cfg.LengthDecay = 0.85
	}
	if cfg.FanoutThreshold <= 0 {
		cfg.FanoutThreshold = 5
	}
	if cfg.Direction == "" {
		cfg.Direction = "reverse"
	}
	return cfg
}

func cloneSteps(steps []Step) []Step {
	if len(steps) == 0 {
		return nil
	}
	out := make([]Step, len(steps))
	copy(out, steps)
	return out
}

func appendStep(steps []Step, s Step) []Step {
	out := make([]Step, len(steps)+1)
	copy(out, steps)
	out[len(steps)] = s
	return out
}
