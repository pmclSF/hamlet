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
//
// Performance characteristics:
//   - Node/edge lookup by ID: O(1)
//   - NodesByType/NodesByFamily: O(k) where k = matching nodes (indexed)
//   - Neighbors/ReverseNeighbors: O(1) after first call (cached)
//   - Nodes() sorted: O(1) after first call (cached)
type Graph struct {
	nodes map[string]*Node  // id → node
	edges []*Edge           // all edges (insertion order)
	adj   map[string][]*Edge // outgoing: from → edges
	radj  map[string][]*Edge // incoming: to → edges

	// Indexes built incrementally during construction.
	typeIndex   map[NodeType][]*Node   // type → nodes (unsorted during build)
	familyIndex map[NodeFamily][]*Node // family → nodes (unsorted during build)

	// Caches populated lazily after Build(). These are nil during
	// construction and populated on first read-path access.
	sortedNodes      []*Node            // cached sorted node list
	neighborCache    map[string][]string // cached deduplicated+sorted neighbor lists
	revNeighborCache map[string][]string // cached deduplicated+sorted reverse neighbors
	outDegreeCache   map[string]int      // cached out-degree per node
	sealed           bool               // true after Seal() — enables caching
}

// NewGraph creates an empty graph.
func NewGraph() *Graph {
	return &Graph{
		nodes:       make(map[string]*Node),
		adj:         make(map[string][]*Edge),
		radj:        make(map[string][]*Edge),
		typeIndex:   make(map[NodeType][]*Node),
		familyIndex: make(map[NodeFamily][]*Node),
	}
}

// AddNode adds a node to the graph. If a node with the same ID already
// exists, it is replaced (indexes are rebuilt for the replaced node).
func (g *Graph) AddNode(n *Node) {
	if old, exists := g.nodes[n.ID]; exists && old.Type != n.Type {
		// Remove from old type/family indexes on type change.
		g.removeFromIndexes(old)
	}
	if _, exists := g.nodes[n.ID]; !exists {
		// Only add to indexes for genuinely new nodes.
		g.typeIndex[n.Type] = append(g.typeIndex[n.Type], n)
		fam := NodeTypeFamily(n.Type)
		if fam != "" {
			g.familyIndex[fam] = append(g.familyIndex[fam], n)
		}
	}
	g.nodes[n.ID] = n
	g.sortedNodes = nil // invalidate cache
}

func (g *Graph) removeFromIndexes(n *Node) {
	if idx, ok := g.typeIndex[n.Type]; ok {
		for i, existing := range idx {
			if existing.ID == n.ID {
				g.typeIndex[n.Type] = append(idx[:i], idx[i+1:]...)
				break
			}
		}
	}
	fam := NodeTypeFamily(n.Type)
	if fam != "" {
		if idx, ok := g.familyIndex[fam]; ok {
			for i, existing := range idx {
				if existing.ID == n.ID {
					g.familyIndex[fam] = append(idx[:i], idx[i+1:]...)
					break
				}
			}
		}
	}
}

// AddEdge adds a directed edge. Both endpoints should already exist as
// nodes, but this is not enforced to allow incremental construction.
func (g *Graph) AddEdge(e *Edge) {
	g.edges = append(g.edges, e)
	g.adj[e.From] = append(g.adj[e.From], e)
	g.radj[e.To] = append(g.radj[e.To], e)
	// Invalidate neighbor caches.
	g.neighborCache = nil
	g.revNeighborCache = nil
	g.outDegreeCache = nil
}

// Seal marks the graph as read-only, enabling query result caching.
// Called automatically at the end of Build().
func (g *Graph) Seal() {
	g.sealed = true
	// Sort type and family indexes for deterministic iteration.
	for t := range g.typeIndex {
		sortNodesByID(g.typeIndex[t])
	}
	for f := range g.familyIndex {
		sortNodesByID(g.familyIndex[f])
	}
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
// Result is cached after first call on a sealed graph.
func (g *Graph) Nodes() []*Node {
	if g.sealed && g.sortedNodes != nil {
		return g.sortedNodes
	}
	out := make([]*Node, 0, len(g.nodes))
	for _, n := range g.nodes {
		out = append(out, n)
	}
	sortNodesByID(out)
	if g.sealed {
		g.sortedNodes = out
	}
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
// deduplicated and sorted. Result is cached on sealed graphs.
func (g *Graph) Neighbors(id string) []string {
	if g.sealed && g.neighborCache != nil {
		if cached, ok := g.neighborCache[id]; ok {
			return cached
		}
	}
	seen := map[string]bool{}
	for _, e := range g.adj[id] {
		seen[e.To] = true
	}
	out := make([]string, 0, len(seen))
	for nid := range seen {
		out = append(out, nid)
	}
	sort.Strings(out)
	if g.sealed {
		if g.neighborCache == nil {
			g.neighborCache = make(map[string][]string)
		}
		g.neighborCache[id] = out
	}
	return out
}

// ReverseNeighbors returns the IDs of nodes with edges pointing to
// the given node, deduplicated and sorted. Result is cached on sealed graphs.
func (g *Graph) ReverseNeighbors(id string) []string {
	if g.sealed && g.revNeighborCache != nil {
		if cached, ok := g.revNeighborCache[id]; ok {
			return cached
		}
	}
	seen := map[string]bool{}
	for _, e := range g.radj[id] {
		seen[e.From] = true
	}
	out := make([]string, 0, len(seen))
	for nid := range seen {
		out = append(out, nid)
	}
	sort.Strings(out)
	if g.sealed {
		if g.revNeighborCache == nil {
			g.revNeighborCache = make(map[string][]string)
		}
		g.revNeighborCache[id] = out
	}
	return out
}

// OutDegree returns the number of unique outgoing neighbors for a node.
// Result is cached on sealed graphs.
func (g *Graph) OutDegree(id string) int {
	if g.sealed && g.outDegreeCache != nil {
		if deg, ok := g.outDegreeCache[id]; ok {
			return deg
		}
	}
	deg := len(g.Neighbors(id))
	if g.sealed {
		if g.outDegreeCache == nil {
			g.outDegreeCache = make(map[string]int)
		}
		g.outDegreeCache[id] = deg
	}
	return deg
}

// NodesByType returns all nodes of the given type, sorted by ID.
// Uses the type index for O(k) performance where k = matching nodes.
func (g *Graph) NodesByType(t NodeType) []*Node {
	indexed := g.typeIndex[t]
	if g.sealed {
		// Index is pre-sorted after Seal().
		return indexed
	}
	// During construction, return a sorted copy.
	out := make([]*Node, len(indexed))
	copy(out, indexed)
	sortNodesByID(out)
	return out
}

// NodesByFamily returns all nodes belonging to the given family, sorted by ID.
// Uses the family index for O(k) performance where k = matching nodes.
func (g *Graph) NodesByFamily(f NodeFamily) []*Node {
	indexed := g.familyIndex[f]
	if g.sealed {
		return indexed
	}
	out := make([]*Node, len(indexed))
	copy(out, indexed)
	sortNodesByID(out)
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
	for t := range validationNodeTypes {
		out = append(out, g.NodesByType(t)...)
	}
	sortNodesByID(out)
	return out
}

// ValidationsForSurface returns all validation-bearing nodes that validate
// a given code surface or behavior surface, following
// EdgeCoversCodeSurface and EdgeManualCovers edges in reverse.
//
// This answers: "what tests, scenarios, and manual coverage exist for this
// surface?" — the fundamental coverage question.
func (g *Graph) ValidationsForSurface(surfaceID string) []*Node {
	var out []*Node
	seen := map[string]bool{}
	for _, e := range g.radj[surfaceID] {
		switch e.Type {
		case EdgeCoversCodeSurface, EdgeManualCovers:
			n := g.nodes[e.From]
			if n != nil && validationNodeTypes[n.Type] && !seen[n.ID] {
				seen[n.ID] = true
				out = append(out, n)
			}
		}
	}
	sortNodesByID(out)
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
	NodeCount     int            `json:"nodeCount"`
	EdgeCount     int            `json:"edgeCount"`
	NodesByType   map[string]int `json:"nodesByType"`
	EdgesByType   map[string]int `json:"edgesByType"`
	NodesByFamily map[string]int `json:"nodesByFamily,omitempty"`
	Density       float64        `json:"density"`
}

func (g *Graph) Stats() Stats {
	s := Stats{
		NodeCount:     len(g.nodes),
		EdgeCount:     len(g.edges),
		NodesByType:   make(map[string]int, len(g.typeIndex)),
		EdgesByType:   map[string]int{},
		NodesByFamily: make(map[string]int, len(g.familyIndex)),
	}
	// Use indexes instead of scanning all nodes.
	for t, nodes := range g.typeIndex {
		s.NodesByType[string(t)] = len(nodes)
	}
	for f, nodes := range g.familyIndex {
		s.NodesByFamily[string(f)] = len(nodes)
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

// --- Serialization ---

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

// UnmarshalJSON deserializes a graph from JSON, rebuilding all indexes.
func (g *Graph) UnmarshalJSON(data []byte) error {
	var sg serializedGraph
	if err := json.Unmarshal(data, &sg); err != nil {
		return err
	}

	g.nodes = make(map[string]*Node, len(sg.Nodes))
	g.adj = make(map[string][]*Edge)
	g.radj = make(map[string][]*Edge)
	g.typeIndex = make(map[NodeType][]*Node)
	g.familyIndex = make(map[NodeFamily][]*Node)
	g.edges = sg.Edges

	for _, n := range sg.Nodes {
		g.nodes[n.ID] = n
		g.typeIndex[n.Type] = append(g.typeIndex[n.Type], n)
		fam := NodeTypeFamily(n.Type)
		if fam != "" {
			g.familyIndex[fam] = append(g.familyIndex[fam], n)
		}
	}
	for _, e := range sg.Edges {
		g.adj[e.From] = append(g.adj[e.From], e)
		g.radj[e.To] = append(g.radj[e.To], e)
	}

	g.Seal()
	return nil
}

// --- Helpers ---

func sortNodesByID(nodes []*Node) {
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
}
