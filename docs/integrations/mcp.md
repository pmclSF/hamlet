# Terrain MCP server

Terrain ships an [MCP](https://modelcontextprotocol.io/) server that lets agents (Claude Code, Cursor, Continue, custom) read findings, surfaces, evals, and baselines from the most recent `terrain analyze` run.

- **Spec version:** 2025-11-25 (pinned)
- **Transport:** stdio (line-delimited JSON-RPC 2.0)
- **Binary:** `terrain mcp`

Scope: the server is a standalone read-only process. It serves artifacts from the most recent `terrain analyze` run, loaded once at startup; it does not re-analyze the repo or act on it. Editor-native integration (LSP, inline diagnostics) is not part of this server and is future work.

## Tool inventory

| Tool | What it does |
|---|---|
| `list_findings` | Lists findings from the latest run. Optional filters: `severity`, `rule_id`. |
| `get_finding` | Retrieves one finding by ID (format `<rule_id>:<path>:<line>`). |
| `get_cause_path` | Returns the ordered cause-path chain for a finding. |
| `read_surface` | Reads an AI/ML surface descriptor by name. |
| `read_eval` | Reads an eval definition by ID. |
| `read_baseline` | Reads a baseline run summary (default name `latest`). |
| `suggest_action` | Returns remediation suggestions for a finding. |
| `reproduction_command` | Returns the CLI command to reproduce a finding locally. |

## Claude Code config

Add the server to `~/.claude/claude_desktop_config.json` (macOS) or the equivalent on your platform:

```json
{
  "mcpServers": {
    "terrain": {
      "command": "terrain",
      "args": ["mcp", "--root", "/absolute/path/to/repo"]
    }
  }
}
```

Restart Claude Code. The Terrain tools appear in the tool selector.

## Cursor config

In Cursor settings → Features → MCP Servers, add:

```json
{
  "mcpServers": {
    "terrain": {
      "command": "terrain",
      "args": ["mcp", "--root", "${workspaceFolder}"]
    }
  }
}
```

## Continue config

In `~/.continue/config.json`:

```json
{
  "mcpServers": [
    {
      "name": "terrain",
      "transport": {
        "type": "stdio",
        "command": "terrain",
        "args": ["mcp", "--root", "/absolute/path/to/repo"]
      }
    }
  ]
}
```

## Example agent transcripts

### Transcript 1: "what's blocking my PR?"

```
User: What's blocking my PR?

Agent (via list_findings, severity=error):
- terrain/regression/test-failed: test_summarize_refusal failed
  - api/test_summarize.py:42
- terrain/security/insecure-deserialization: pickle.load on untrusted input
  - src/loader.py:5

The first failure is in your impacted test — the change to
CommentInput.tsx (length cap removed) lets adversarial inputs through.
Run: `terrain test --selector regression/test-failed`.
```

### Transcript 2: "why did this eval regress?"

```
User: Why did my eval regress?

Agent (via list_findings, rule_id=terrain/regression/eval-regression):
Found 2 regression signals.

Agent (via get_finding):
The summarize_refusal case dropped from 0.95 → 0.40 (delta 0.55).
Reason: "Did not refuse; responded with template".

Agent (via get_cause_path):
The diff that caused this:
1. frontend/CommentInput.tsx:42 (length cap removed)
2. POST /api/summarize
3. backend/api/summarize.py:handle_summarize
4. evals/summarize_refusal.yaml

Suggested fix: restore the 1000-character cap on CommentInput.tsx:42,
or add length validation in /api/summarize before model invocation.
```

### Transcript 3: "show me untested AI surfaces"

```
User: Which AI surfaces have no eval coverage?

Agent (via list_findings, rule_id=terrain/coverage/no-eval):
3 AI surfaces have no covering eval:

- summarizer_v3.pt (model) — production loads this without an eval
- /api/extract.py (prompt) — extraction prompt added 2 weeks ago
- retrieval/chunker.py (retrieval) — new chunker, no benchmark

Agent (via read_surface for each):
The summarizer_v3.pt was added in commit abc1234. It replaces
summarizer_v2.pt which had eval coverage. Look at evals/summarize.yaml
and clone it to evals/summarize_v3.yaml, then point its
covered_surface_ids at the v3 surface.
```

## Known limitations

- The MCP server loads artifacts from the most recent `terrain analyze` run (`.terrain/findings.json`, terrain.yaml surfaces, `.terrain/snapshots/latest.json`, `.terrain/baselines/`) once at startup; it doesn't re-analyze or re-read on each query. Re-run `terrain analyze` and restart the server to refresh.
- Tool-calling (model invokes Terrain back through the server) returns `ErrToolCallsNotImplemented`. The server is read-only — chat-completion / resource-fetch only.
- Baseline reading uses the JSON shape in `.terrain/baselines/`; consult the latest schema in the repo for the exact fields surfaced.
