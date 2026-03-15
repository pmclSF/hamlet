# Graph Traversal Algorithms

> **Status:** Implemented
> **Purpose:** Specify the BFS/DFS traversal algorithms that power impact analysis, coverage analysis, fanout detection, and path explainability.
> **Key decisions:**
> - Dual adjacency indexes (forward and reverse) maintained at construction time for O(1) neighbor lookup
> - Confidence decays by 0.85 per hop, making long transitive chains explicitly less certain
> - Cycle-safe traversal via visited-set tracking — cycles are not errors, they terminate traversal at the revisited node
> - All algorithms are O(V+E) and complete in milliseconds for typical repositories

**See also:** [02-graph-schema.md](02-graph-schema.md), [05-insight-engine-framework.md](05-insight-engine-framework.md), [14-evidence-scoring-and-confidence-model.md](14-evidence-scoring-and-confidence-model.md)

The insight engines rely on graph traversal to answer questions about the test system. This document describes the core algorithms.

## Data Structures

### Adjacency Index

The graph maintains two indexes for efficient traversal:

- **Forward index** (outgoing edges): `nodeId → Edge[]` — edges where `nodeId` is the source
- **Reverse index** (incoming edges): `nodeId → Edge[]` — edges where `nodeId` is the target

Both indexes are built during graph construction and updated on edge insertion.

## Algorithms

### Forward Reachability (Impact Analysis)

**Question:** Starting from a changed file, what tests are reachable?

**Algorithm:** BFS from the changed node, following edges in reverse direction (since edges point in the dependency direction: `test → source`).

```
function findImpactedTests(graph, changedFileId):
  queue = [changedFileId]
  visited = {}
  impacted = []

  while queue is not empty:
    current = queue.dequeue()
    for edge in graph.getIncoming(current):
      if edge.from not in visited:
        visited[edge.from] = true
        if isTestNode(edge.from):
          impacted.push(edge.from)
        queue.enqueue(edge.from)

  return impacted
```

**Confidence decay:** Each hop reduces confidence by a factor of 0.85. The confidence of the final path is:

```
confidence = 0.85 ^ pathLength
```

**Fanout penalty:** If an intermediate node has transitive fanout exceeding the threshold, confidence is further reduced proportionally.

### Reverse Coverage (Coverage Analysis)

**Question:** For a source file, which tests cover it?

**Algorithm:** Same as forward reachability but starting from source file nodes. The traversal follows incoming edges to find all test file nodes, then resolves individual test IDs within those files.

**Test resolution:** Test files contain suites and tests in a hierarchy:

```
TestFile
  └── Suite (via TEST_DEFINED_IN_FILE)
        └── Suite (via TEST_DEFINED_IN_FILE)
              └── Test (via TEST_DEFINED_IN_FILE)
```

The `resolveTests` function recursively walks this hierarchy to collect all leaf test nodes:

```
function collectTests(index, nodeId, results):
  for edge in getIncoming(index, nodeId, TEST_DEFINED_IN_FILE):
    if edge.from starts with "test:":
      results.push(edge.from)
    else if edge.from starts with "suite:":
      collectTests(index, edge.from, results)
```

### Transitive Fanout (Fanout Analysis)

**Question:** How many nodes are transitively reachable from a given node?

**Algorithm:** BFS following forward edges (outgoing), counting all reachable nodes.

```
function computeTransitiveFanout(graph, nodeId):
  queue = [nodeId]
  visited = {nodeId}

  while queue is not empty:
    current = queue.dequeue()
    for edge in graph.getEdges(current):
      if edge.to not in visited:
        visited.add(edge.to)
        queue.enqueue(edge.to)

  return visited.size - 1  // exclude the starting node
```

### Path Tracing (Explainability)

**Question:** What are all the dependency paths from a test to its leaves?

**Algorithm:** DFS from the test node, following forward edges, recording complete paths.

```
function tracePaths(graph, startId):
  paths = []

  function dfs(current, path):
    edges = graph.getEdges(current)
    if edges is empty:
      paths.push(copy(path))
      return

    for edge in edges:
      path.push({node: edge.to, edge: edge})
      dfs(edge.to, path)
      path.pop()

  dfs(startId, [{node: startId}])
  return paths
```

Each path includes the edge type and confidence at each step, enabling human-readable explanations like:

```
test:login.test.ts:5:validates → (TEST_DEFINED_IN_FILE) →
file:login.test.ts → (TEST_USES_FIXTURE) →
file:fixtures/auth.ts → (FIXTURE_IMPORTS_SOURCE) →
file:src/auth/login.ts
```

## Performance Characteristics

| Algorithm | Time Complexity | Space Complexity |
|-----------|----------------|------------------|
| Forward/Reverse reachability | O(V + E) | O(V) |
| Transitive fanout | O(V + E) | O(V) |
| Path tracing (all paths) | O(V * paths) | O(depth * paths) |
| Duplicate detection | O(T^2) where T = tests | O(T) |

For most repositories, V (nodes) is in the hundreds to low thousands and E (edges) is 2-5x V. All algorithms complete in milliseconds.

## Cycle Handling

The graph may contain cycles (e.g., circular imports). All traversal algorithms track visited nodes to prevent infinite loops. Cycles are not errors — they represent real dependency relationships — but they terminate traversal at the revisited node.
