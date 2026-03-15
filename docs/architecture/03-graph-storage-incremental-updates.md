# Graph Storage and Incremental Updates

> **Status:** Planned
> **Purpose:** Persistence format and incremental rebuild strategy for the dependency graph.
> **Key decisions:**
> - Graph persisted as serialized JSON in a local `.terrain/graph/` directory
> - Incremental updates rebuild only the subgraph affected by recent changes (detected via `git diff`)
> - Atomic writes (write to temp, rename) ensure consistency on failure
> - Stored commit SHA anchors the incremental base

## Storage Format

The dependency graph is persisted to the `.terrain/graph/` directory as serialized JSON. The stored graph includes:

- All nodes with their types and metadata
- All edges with confidence scores and evidence types
- Graph metadata: repository root, commit SHA, creation timestamp

```
.terrain/
  graph/
    graph.json          # Serialized graph (nodes + edges)
    metadata.json       # Commit SHA, timestamp, repo root
```

See [02-graph-schema.md](02-graph-schema.md) for the node and edge data model.

## Incremental Updates

Full graph rebuilds are expensive for large repositories. Terrain supports incremental updates that only rebuild the portion of the graph affected by recent changes.

### Process

1. **Detect changes** — compare the current working tree against the last stored commit SHA using `git diff`
2. **Identify affected files** — changed, added, and deleted files
3. **Remove stale nodes** — delete nodes and edges for deleted files
4. **Rebuild affected subgraph** — re-run test discovery and import analysis for changed and new files
5. **Merge** — integrate the rebuilt subgraph into the existing graph
6. **Persist** — save the updated graph with the new commit SHA

### Change Detection

Terrain detects changes through two mechanisms:

- **Committed changes** — `git diff` between the stored base SHA and current HEAD
- **Uncommitted changes** — staged modifications, unstaged modifications, and untracked files

Both are combined to produce the full set of affected files.

### Incremental Flag

```bash
terrain graph build --incremental
```

When `--incremental` is set, Terrain loads the existing graph and only processes changed files. Without it, the graph is rebuilt from scratch.

## Consistency Guarantees

- The graph is always written atomically (write to temp, rename)
- If incremental update fails, the previous graph is preserved
- The stored commit SHA ensures the incremental base is always known
