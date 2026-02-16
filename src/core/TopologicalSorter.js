/**
 * Topologically sorts a dependency graph using Kahn's algorithm (BFS-based).
 *
 * Files with no dependencies come first (helpers, fixtures, utilities).
 * Handles cycles gracefully (doesn't infinite loop, returns best-effort order).
 */

export class TopologicalSorter {
  /**
   * Sort a dependency graph topologically (leaves first).
   *
   * @param {{nodes: string[], edges: Map<string, string[]>}} graph
   *   `edges` maps each node to the list of nodes it depends on.
   * @returns {string[]} Ordered array of node identifiers (leaves first)
   */
  sort(graph) {
    const { nodes, edges } = graph;

    if (nodes.length === 0) return [];

    // Build in-degree map and reverse adjacency list
    const inDegree = new Map();
    const dependents = new Map(); // node -> nodes that depend on it

    for (const node of nodes) {
      inDegree.set(node, 0);
      dependents.set(node, []);
    }

    for (const node of nodes) {
      const deps = edges.get(node) || [];
      for (const dep of deps) {
        // Only count edges to nodes within the graph
        if (inDegree.has(dep)) {
          inDegree.set(node, inDegree.get(node) + 1);
          dependents.get(dep).push(node);
        }
      }
    }

    // Kahn's algorithm: start with nodes that have zero in-degree (leaves)
    const queue = [];
    for (const [node, degree] of inDegree) {
      if (degree === 0) {
        queue.push(node);
      }
    }

    const sorted = [];

    while (queue.length > 0) {
      const node = queue.shift();
      sorted.push(node);

      for (const dependent of dependents.get(node) || []) {
        const newDegree = inDegree.get(dependent) - 1;
        inDegree.set(dependent, newDegree);
        if (newDegree === 0) {
          queue.push(dependent);
        }
      }
    }

    // If there are remaining nodes (due to cycles), append them
    if (sorted.length < nodes.length) {
      for (const node of nodes) {
        if (!sorted.includes(node)) {
          sorted.push(node);
        }
      }
    }

    return sorted;
  }
}
