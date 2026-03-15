package depgraph

import (
	"encoding/json"
	"sort"
)

// Graph is an in-memory directed graph with typed nodes and edges.
//
// After construction via Build(), the graph is read-only and safe for
// concurrent reads. All query methods return deterministically ordered
// results.
type Graph struct {
	nodes map[string]*Node  // id → node
	edges []*Edge           // all edges
	adj   map[string][]*Edge // outgoing: from → edges
	radj  map[string][]*Edge // incoming: to → edges
}

// NewGraph creates an empty graph.
func NewGraph() *Graph {
	return &Graph{
		nodes: make(map[string]*Node),
		adj:   make(map[string][]*Edge),
		radj:  make(map[string][]*Edge),
	}
}

// AddNode adds a node to the graph. If a node with the same ID already
// exists, it is replaced.
func (g *Graph) AddNode(n *Node) {
	g.nodes[n.ID] = n
}

// AddEdge adds a directed edge. Both endpoints should already exist as
// nodes, but this is not enforced to allow incremental construction.
func (g *Graph) AddEdge(e *Edge) {
	g.edges = append(g.edges, e)
	g.adj[e.From] = append(g.adj[e.From], e)
	g.radj[e.To] = append(g.radj[e.To], e)
}

// Node returns the node with the given ID, or nil.
func (g *Graph) Node(id string) *Node {
	return g.nodes[id]
}

// NodeCount returns the total number of nodes.
func (g *Graph) NodeCount() int {
	return len(g.nodes)
}

// EdgeCount returns the total number of edges.
func (g *Graph) EdgeCount() int {
	return len(g.edges)
}

// Nodes returns all nodes, sorted by ID for determinism.
func (g *Graph) Nodes() []*Node {
	out := make([]*Node, 0, len(g.nodes))
	for _, n := range g.nodes {
		out = append(out, n)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// Edges returns all edges.
func (g *Graph) Edges() []*Edge {
	return g.edges
}

// Outgoing returns edges originating from the given node ID.
func (g *Graph) Outgoing(id string) []*Edge {
	return g.adj[id]
}

// Incoming returns edges targeting the given node ID.
func (g *Graph) Incoming(id string) []*Edge {
	return g.radj[id]
}

// Neighbors returns the IDs of nodes reachable via outgoing edges,
// deduplicated and sorted.
func (g *Graph) Neighbors(id string) []string {
	seen := map[string]bool{}
	for _, e := range g.adj[id] {
		seen[e.To] = true
	}
	out := make([]string, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

// ReverseNeighbors returns the IDs of nodes with edges pointing to
// the given node, deduplicated and sorted.
func (g *Graph) ReverseNeighbors(id string) []string {
	seen := map[string]bool{}
	for _, e := range g.radj[id] {
		seen[e.From] = true
	}
	out := make([]string, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

// NodesByType returns all nodes of the given type, sorted by ID.
func (g *Graph) NodesByType(t NodeType) []*Node {
	var out []*Node
	for _, n := range g.nodes {
		if n.Type == t {
			out = append(out, n)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// NodesByFamily returns all nodes belonging to the given family, sorted by ID.
func (g *Graph) NodesByFamily(f NodeFamily) []*Node {
	var out []*Node
	for _, n := range g.nodes {
		if NodeTypeFamily(n.Type) == f {
			out = append(out, n)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// validationNodeTypes lists all node types that represent validation-bearing
// entities. This is the canonical definition of "what counts as a validation
// target" in the graph — it keeps the abstraction in one place.
var validationNodeTypes = map[NodeType]bool{
	NodeTest:           true,
	NodeScenario:       true,
	NodeManualCoverage: true,
}

// IsValidationNode returns true if the given node type represents a
// validation-bearing entity (test, scenario, or manual coverage artifact).
func IsValidationNode(t NodeType) bool {
	return validationNodeTypes[t]
}

// ValidationTargets returns all validation-bearing nodes in the graph —
// tests, scenarios, and manual coverage artifacts — sorted by ID.
//
// This is the primary query method for code that needs to reason over
// "all things that validate behavior" without caring about the concrete type.
func (g *Graph) ValidationTargets() []*Node {
	var out []*Node
	for _, n := range g.nodes {
		if validationNodeTypes[n.Type] {
			out = append(out, n)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// ValidationsForSurface returns all validation-bearing nodes that validate
// a given code surface or behavior surface, following EdgeCoversCodeSurface,
// EdgeTestExercises, EdgeManualCovers, and EdgeValidates edges in reverse.
//
// This answers: "what tests, scenarios, and manual coverage exist for this
// surface?" — the fundamental coverage question.
func (g *Graph) ValidationsForSurface(surfaceID string) []*Node {
	var out []*Node
	seen := map[string]bool{}
	for _, e := range g.radj[surfaceID] {
		switch e.Type {
		case EdgeCoversCodeSurface, EdgeTestExercises, EdgeManualCovers, EdgeValidates:
			n := g.nodes[e.From]
			if n != nil && validationNodeTypes[n.Type] && !seen[n.ID] {
				seen[n.ID] = true
				out = append(out, n)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// EdgesByType returns all edges of the given type.
func (g *Graph) EdgesByType(t EdgeType) []*Edge {
	var out []*Edge
	for _, e := range g.edges {
		if e.Type == t {
			out = append(out, e)
		}
	}
	return out
}

// Stats returns summary statistics about the graph.
type Stats struct {
	NodeCount       int            `json:"nodeCount"`
	EdgeCount       int            `json:"edgeCount"`
	NodesByType     map[string]int `json:"nodesByType"`
	EdgesByType     map[string]int `json:"edgesByType"`
	NodesByFamily   map[string]int `json:"nodesByFamily,omitempty"`
	Density         float64        `json:"density"`
}

func (g *Graph) Stats() Stats {
	s := Stats{
		NodeCount:     len(g.nodes),
		EdgeCount:     len(g.edges),
		NodesByType:   map[string]int{},
		EdgesByType:   map[string]int{},
		NodesByFamily: map[string]int{},
	}
	for _, n := range g.nodes {
		s.NodesByType[string(n.Type)]++
		if fam := NodeTypeFamily(n.Type); fam != "" {
			s.NodesByFamily[string(fam)]++
		}
	}
	for _, e := range g.edges {
		s.EdgesByType[string(e.Type)]++
	}
	n := float64(s.NodeCount)
	if n > 1 {
		s.Density = float64(s.EdgeCount) / (n * (n - 1))
	}
	return s
}

// serializedGraph is the JSON-serializable representation of a Graph.
type serializedGraph struct {
	Version string  `json:"version"`
	Nodes   []*Node `json:"nodes"`
	Edges   []*Edge `json:"edges"`
}

// MarshalJSON serializes the graph to JSON. Nodes are sorted by ID and
// edges preserve insertion order for deterministic output.
func (g *Graph) MarshalJSON() ([]byte, error) {
	sg := serializedGraph{
		Version: "1.0.0",
		Nodes:   g.Nodes(),
		Edges:   g.edges,
	}
	if sg.Nodes == nil {
		sg.Nodes = []*Node{}
	}
	if sg.Edges == nil {
		sg.Edges = []*Edge{}
	}
	return json.Marshal(sg)
}

// UnmarshalJSON deserializes a graph from JSON, rebuilding adjacency indexes.
func (g *Graph) UnmarshalJSON(data []byte) error {
	var sg serializedGraph
	if err := json.Unmarshal(data, &sg); err != nil {
		return err
	}

	g.nodes = make(map[string]*Node, len(sg.Nodes))
	g.adj = make(map[string][]*Edge)
	g.radj = make(map[string][]*Edge)
	g.edges = sg.Edges

	for _, n := range sg.Nodes {
		g.nodes[n.ID] = n
	}
	for _, e := range sg.Edges {
		g.adj[e.From] = append(g.adj[e.From], e)
		g.radj[e.To] = append(g.radj[e.To], e)
	}

	return nil
}
