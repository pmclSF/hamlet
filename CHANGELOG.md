# Changelog

All notable changes to Terrain are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Added

- **The offline guarantee is now CI-enforced.** Every build runs Terrain's
  discovery, analysis, severity gate, and fix dry-run inside a network
  namespace with no interfaces, against a fixture that produces a real
  finding — so "no API key, no network" is a tested property, not a claim.
  Run the same assertions locally with `make offline-check`.

### Changed

- **AI-surface scans now honor `.gitignore`.** Prompt, schema, model-call-site,
  and dataset discovery — the no-args report, `analyze`, `fix`, and the AI
  detectors — no longer walk ignored directories, so vendored or locally
  generated content can't inflate the map or produce findings in code you
  don't own. Untracked files that are not ignored are still scanned.
- **`terrain report summary` reads like a report, not a metrics dump.**
  Dominant drivers and the recommended focus now use each check's human
  title (with the rule id alongside, for `terrain explain`), and unmeasured
  posture rows say what input they need instead of "Unknown".
- **`terrain --help` now lists every shipped command**, including `scaffold`,
  `inject`, `suppress`, `describe`, `accept-snapshot`, and `plugins`.
- **Unrecognized PR slash commands now reply with the verbs that work.**
  The placeholder `/terrain refresh` and `/terrain escalate` verbs were
  removed rather than acknowledging and doing nothing.
- **JSON Schema `$id` URIs now point at their canonical repository URLs**
  instead of an unregistered domain. Schema content is unchanged.

### Fixed

- **`terrain fix` now rejects rooted fix paths on Windows too.** The
  path-safety guard that keeps every fix repo-relative relied on
  `filepath.IsAbs`, which does not treat a leading-slash path like
  `/etc/passwd` as absolute on Windows, so such a path could slip past the
  check there. The guard now also rejects Windows volume names and any
  leading slash or backslash, so rooted paths are refused identically on
  every OS.

## [0.4.0] — Trust floor & first-run report

### Added

- **`terrain fix` applies the remediations Terrain can prove.** Today those are
  the correct-side prompt→schema drift corrections (a prompt reference rewritten
  to the schema field it should name) and the deps drift-risk caret-pinning fix
  (a floating `^`/`~` range pinned to a validated version). It is
  **dry-run by default** — it prints
  the diff of every ready fix and writes nothing; `--apply` writes them and
  prints `git diff` to review and a scoped `git restore` on the changed files
  to undo. Only fixes whose
  remediation is validated appear; a finding with no proven fix stays advisory.
- **`terrain` with no arguments is now a first-run report.** One command,
  no flags, gives you a map of what Terrain sees — your prompts, schemas,
  and evals — the few issues actually worth a look, and a health readout
  for prompt↔schema contracts and AI test coverage. It leads with what it
  understood about your system, not a wall of findings. On a repo with no
  AI surfaces it stays friendly and points you at `terrain analyze`.
- **`terrain analyze --no-trust-floor` / `terrain test --no-trust-floor`.**
  The opt-out for the new default gating behavior (below): restores
  severity-only gating so any `--fail-on` match blocks CI. Also settable as
  `trust_floor: false` in `terrain.yaml`.

### Changed

- **A "trust floor" now governs the `--fail-on` gate by default.** Heuristic
  AI findings fail the build only when Terrain can prove a fix for them, so CI
  never breaks on a low-confidence finding. Findings that must always gate —
  failing tests and regressions, security/safety leaks (secrets, PII, insecure
  deserialization), your own `policy.yaml` violations, and any Critical — still
  fail the build on severity regardless. When the trust floor holds a finding
  back, the run prints a one-line notice, so a passing build is never silent
  about it. To gate every finding on severity, pass `--no-trust-floor` (on
  `analyze`, `test`, and `report pr` / `check-runs`) or set `trust_floor: false`
  in `terrain.yaml`.
- **Terminal output now degrades cleanly everywhere.** Color follows your
  terminal's own theme (light or dark) via standard ANSI roles, and every
  glyph has an ASCII fallback for terminals that can't render Unicode
  (Windows conhost, CI logs, dumb terminals). Set `TERRAIN_ASCII=1` or
  `NO_COLOR` to force the plain path. The rendering system is documented in
  [docs/cli-design-tokens.md](docs/cli-design-tokens.md).

### Fixed

- **Fewer false positives in the default AI detectors.** The schema→prompt
  drift and model-deprecation checks were hardened to suppress
  structurally-benign matches — base-vs-subclass polymorphism, docstring
  examples, deprecation-map keys, and test files — so a finding is more likely
  to be a real issue worth acting on. The PII-in-eval and secrets-in-prompt
  checks (opt-in) gained the same class of synthetic-value / placeholder
  suppression.


## [0.3.1] — Homebrew tap auto-update

### Fixed

- **The Homebrew tap now updates automatically when a release is cut.**
  The tap-update workflow previously ran only on `release: published`,
  an event the release pipeline's token cannot emit, so the formula had
  to be bumped by hand after each release. It now also runs on `v*` tag
  pushes, so tagging a release updates the formula without manual steps.

## [0.3.0] — PR surface upgrade + adopter scaffolding

### Added

- **`.terrain/findings.json` is written after every `terrain analyze` run.**
  A canonical Finding artifact downstream consumers (`terrain mcp`, IDE
  plugins, third-party SARIF uploaders) can read without setting any
  flag. The MCP server now actually sees what the latest analyze run
  produced. The file is gitignored.
- **`/terrain scaffold accept` slash verb now materializes scaffolds for
  real.** Pass `schema:<path>` (and optionally `prompt:<path>`,
  `lang:python|typescript|json`) and the dispatcher runs the scaffold
  generator and returns the rendered test file in the reply. Path
  arguments are resolved under the repo root and rejected if they
  escape, so a hostile slash comment can't ask the server to read
  arbitrary files.
- **`terrain doctor` reports per-rule policy overrides.** Summarizes how
  many rules `.terrain/policy.yaml` turns off, demotes to warn, or
  promotes to block — so policy drift surfaces in the same place infra
  drift does.
- **`terrain init` now prints the path of the annotated
  `.terrain/policy.yaml.example`** alongside `policy.yaml`. Adopters
  no longer have to discover the reference file by inspecting the
  directory.
- **`terrain test --summary` flag help mentions `$GITHUB_STEP_SUMMARY`**;
  `docs/quickstart.md` has a new "Wire into CI" section with the
  canonical GitHub Actions snippet so reviewers see Terrain findings
  on the workflow run page.

### Fixed

- **`make docs-verify` no longer fails on the hand-authored
  `docs/rules/_template.md`.** The drift check now skips files starting
  with `_` so canonical templates stay as a reference without breaking
  the gate.
- **151 `gofmt` violations across the repo.** Added `gofmt -l` as the
  first step of `go-release-verify` so future drift fails the release
  gate before vet/test/build run.
- **CLI typo-suggestion list completeness.** `terrain plugin`,
  `terrain injec`, `terrain scaffol`, `terrain webhok` now suggest the
  correct command — previously they silently fell through.
- **MCP `ServerVersion` bumped to `0.3.0`** so the handshake reflects
  the current binary.
- **Removed internal-only wording from public source comments.**

### Earlier in this cycle

- **`terrain scaffold --schema <path>`** generates a runnable mutation-test
  scaffold from a JSON Schema. Each declared property produces a
  parametrized test exercising boundary cases (empty, max-length,
  unicode-edge, SQL/XSS/path-traversal shapes, null-byte, INT32 bounds,
  near-double-limits, etc.). Emits `python`, `typescript`, or `json`;
  default Python output parses cleanly. The `/terrain scaffold accept`
  slash verb now points adopters at the equivalent CLI invocation
  (production deployments override `DefaultDispatcher` to execute the
  generator and post the scaffold inline).
- **`terrain inject --prompt <path>`** matches a prompt body against a
  cataloged set of jailbreak/injection patterns (DAN-style, instruction
  leak, system-prompt fishing, role confusion, indirect-via-retrieval)
  and emits a runnable test scaffold so adopters can assert their
  prompt pipeline degrades safely. LLM-free: the assertion is the
  adopter's; terrain never invokes the model.
- **`terrain plugins manifest <path>`** validates third-party plugin
  manifests against `schema_version: 1`. Plugins must declare an
  allowed `mechanism_class` (structural-ast | import-graph |
  receiver-type | manifest-schema); literal-string and regex
  primitives are explicitly forbidden, so every plugin rule clears a
  class, not a cell. `terrain plugins add/remove` is reserved for
  a future release once the runtime ships.
- **`terrain` (no-args) discovery report.** A friendly first-touch summary —
  detected frameworks, AI surfaces, schema files, trace logs — followed by
  three copy-pasteable next-step commands. Strictly read-only: no
  `.terrain/` directory is created until you opt in via `terrain init`.
- **`--mechanisms.<name>=on|off|shadow` runtime override** for per-detector
  shadow/live experiments. Internal mechanism inspection verbs are
  maintainer-only and gated behind `TERRAIN_DEV=1`.
- **`terrain mcp`** is now a top-level command that starts the MCP server
  on stdio and loads the most-recent `terrain analyze` artifacts from
  `.terrain/`. (The server existed previously; it just wasn't reachable
  from the CLI.)
- **`terrain doctor` infra checks** for the detector registry, the
  signal-type alias registry, and whether `.gitignore` covers
  `.terrain/`. Additional maintainer-only checks (per-state detector
  breakdown, shadow-sink presence + byte count) surface only when
  `TERRAIN_DEV=1` is set, so the default doctor output stays tight.
- **`.terrain/shadow-report.jsonl`** sink: an append-only log of
  would-suppress / would-add events emitted by mechanisms in shadow state.
  Buffered writes; flushed on pipeline exit. `.terrain/.gitignore` is
  generated by `terrain init` to exclude it from commits.
- **`TERRAIN_ASCII=1`** environment variable forces ASCII separators in
  the discovery report for non-UTF-8 terminals. Defaults to auto-detect
  via `LC_ALL` / `LC_CTYPE` / `LANG`.
- **`TERRAIN_QUIET=1`** is now the umbrella stderr-quiet flag; it
  silences both engine alias-migration `[NOTE]`s and legacy-command
  deprecation hints.
- **Rule splits with alias machinery:** `staticSkippedTest`,
  `depsDriftRisk`, and `aiHardcodedAPIKey` each split into structurally
  distinct halves; the original rule_ids remain valid via the alias
  registry and expand into every replacement automatically.
- **`make regression-gate`** validates regression-suite and recall-harness
  YAML files under `harness/`. Wired into `make release-gate` so
  release verification fails on a malformed suite.
- **PR-comment vocabulary:** the PR-surface labels are now
  `BLOCK / GATE / WATCH / NOTE`, decoupling the user-visible badge
  from the internal `CRIT / HIGH / MED / LOW` severity ladder. The
  internal severity rubric is unchanged; the user-facing layer maps
  through `uitokens.PRLabel(tier, severity)`.
- **Per-detector visibility floor on PR comments.** Each detector
  always gets at least one inline card per PR, even when changescope
  dedup would have collapsed everything else. Extra findings collapse
  into `+N more — see terrain explain <rule>`. Adopters never lose
  visibility into a detector firing.
- **PR-comment templates** (`internal/prtemplates`) provide a stable
  Title / Summary / Action / SlashHints quartet per rule. The
  renderer reads from this registry instead of inlining detector
  prose, so future copy edits don't sprawl across the codebase.
- **Suppressions schema v2** adds `scope` (instance / file /
  directory / repo), `content_hash` (SHA-256 of a 5-line normalized
  context window), and per-scope default expiry. v1 entries continue
  to load; analyzer migrates them on first write.
- **Alias-aware suppression matching:** `ApplyWithAliases` lets a
  v1-era rule_id continue to suppress its post-split halves without
  user action.
- **Two GitHub Checks-API check runs** per PR:
  `terrain (gate)` is required and fails when an undismissed
  gate-tier finding fires; `terrain (observability)` is informational
  and always neutral. Emit both from one analyze pass with
  `terrain report check-runs --head-sha=<sha>`. The bundle is split
  + posted via `gh api` by the workflow — terrain itself remains
  network-free.
- **Slash-command grammar** (`/dismiss`, `/terrain explain`,
  `/terrain show`, `/terrain commands`, plus five reserved verbs)
  with deterministic, LLM-free parsing. `terrain webhook` (gated
  behind `TERRAIN_DEV=1`) serves the receiver; signature validation
  uses HMAC-SHA256 against `TERRAIN_WEBHOOK_SECRET`.
- **Per-repo finding-history learning.** After three un-dismissed
  fires of the same (rule, file), the PR-comment renderer demotes
  that card from `[GATE]` to `[WATCH]`. State lives in
  `.terrain/finding-history.yaml`; `/dismiss` resets the counter
  through the same store.
- **`aiPromptSchemaDrift` detector.** When `terrain analyze --base
  <ref>` runs, Terrain compares JSON Schemas at HEAD against the
  same files at `<ref>` and flags prompt templates that reference
  fields the PR removed or whose type changed. The finding includes
  a side-by-side rendering of the template before and after, with
  values synthesized from each schema's property types and a
  `MISSING(<name>)` marker where a referenced field is no longer
  present. LLM-free, deterministic, no network. Ships at observability
  tier: it emits findings but never blocks CI.
- **Prompt-template renderer** (`internal/prompttemplate`). Pure
  substitution with two flavors: mustache (`{{name}}`, used in
  `.md`/`.markdown` template files) and Python f-string (`{name}`,
  used inline in Python sources — caller extracts the literal).
  Both honor escape sequences (`{{{{x}}}}` → literal `{{x}}` for
  mustache; `{{` / `}}` → literal braces for f-string) and trim
  whitespace inside placeholders. Returns a typed `MissingVarError`
  with the template path when a referenced variable is absent.
- **JSON Schema diff** (`internal/schemadiff`).
  `DiffJSONSchema(oldDoc, newDoc)` returns field-level changes
  (added / removed / type-changed) for flat JSON Schemas. Results
  sorted by field name for deterministic downstream consumers.
- **Prompt-template / schema discovery** (`internal/promptflow`).
  Walks a repository for prompt templates (mustache `.md`/
  `.markdown`) and JSON Schema files (positive signal: `$schema`
  URI matching json-schema.org, or `"type": "object"` alongside
  `properties`). Skips noise directories (`node_modules`,
  `vendor`, `.git`, `dist`, `build`, `.terrain`). Includes a git
  helper `DiscoverFromGit(root, baseRef)` that uses `git show
  <baseRef>:<path>` to retrieve schema content at the base ref.

### Changed

- **`terrain analyze --base <ref>`** is now a real flag (was
  previously redirected to `terrain report pr` / `terrain report
  impact`). With it, analyze runs the prompt-template / schema
  drift detector against the provided git ref. Without it, that
  detector simply doesn't fire — analyze is unchanged otherwise.
- **`prompttemplate.Template` gains a `Path` field.** Optional; when
  set, it's surfaced in `MissingVarError.Error()` so render failures
  name the template that produced them.
- **`Report.GateRelevantSummary`** (new field on the analyze report)
  drives `--fail-on=<severity>` decisions. Mirrors `SignalSummary` but
  excludes findings from detectors that ship at observability tier, so
  observability-tier findings stay informational and never block CI.
- **Alias expansion in suppressions:** a `.terrain/suppressions.yaml`
  entry against `aiHardcodedAPIKey` (or any other split rule_id) now
  expands to suppress every replacement rule_id too. Previously it
  silently stopped suppressing once the split halves shipped.
- **Tier-aware AI blocking signals:** observability-tier AI findings
  no longer count as "blocking" in PR-comment summaries even when
  declared severity is Critical / High.

### Fixed

- **`terrain` no-args used to write `.terrain/policy.yaml` as a side
  effect** of its discovery scan. On read-only directories the scan
  crashed with a permission error. The discovery path now uses a
  read-only `engine.ScanRepo`; explicit `terrain init` still writes
  defaults.
- **Template-literal interpolation** in scope walkers
  (`internal/triggergate`, `internal/looppredicate`) no longer drifts
  brace depth when test files use \`it(\${name}, () => { ... })\`.
- **Several backward-compat doc paths** repaired in help / hint text:
  `terrain ai --help`, `terrain migrate`, `terrain analyze` next-step
  hint.
- **Same-named helpers across packages** now contribute to assertion
  counts (previously only the first body indexed by
  `internal/deffollowing` was credited).
- **Performance:** shadow sink no longer fsyncs per record (buffered
  writer + periodic flush); `internal/surfacelit` caches comment-
  stripped file content per (path, mtime, size); `internal/runtimeconfig`
  caches the per-repo loader-presence check; `internal/barrelresolver`
  defers its repo-wide `package.json` and `__init__.py` walks until
  first call.

### Internal

- Maintenance pass on documentation tone and source-comment text.
- Source-comment and struct-string privacy review across `internal/`,
  including the auto-generated rule-doc surfaces.

## [0.2.0] — AI-feature pre-flight ships

### Headline

- **Twelve new AI detectors** with substantial precision lift over a path-based baseline on representative app-shape repositories. Reach them via `terrain ai findings`.
- **Verdict engine** — typed-evidence pipeline with cross-file context and per-cohort weighting.
- **CLI compression** — 35 commands → 11 canonical (legacy aliases still work).
- **No LLM API key required, ever** — every detector is structural; zero outbound network calls in default config.

### Added

- 12 AI detectors with a recall-regression gate over the bundled fixture set
- Verdict engine reachable from `terrain ai findings`
- `terrain report pr / impact` with `--fail-on / --new-findings-only / --timeout`
- Suppressions file `.terrain/suppressions.yaml`
- Stable finding IDs + `terrain explain finding <id>`
- One recommended GitHub Action template

### Changed

- CLI surface compressed 35 → 11 canonical commands (legacy aliases preserved)
- Recall-regression runner is now a load-bearing release gate

### Release notes

Per-capability status: [`docs/release/feature-status.md`](docs/release/feature-status.md). Full release notes: [`docs/release/RELEASE-NOTES-0.2.0.md`](docs/release/RELEASE-NOTES-0.2.0.md).

### What's stable in 0.2

Read this before adopting 0.2 in CI. Source of truth for per-feature
detail is `docs/release/feature-status.md`.

**Stable** — covered by tests, documented behavior, won't change shape
in 0.2.x:

- repository scan + framework detection (Tier-1 frameworks)
- snapshot generation + schema versioning
- signal registry + manifest export
- AI surface inventory (prompt/agent/tool/context/eval/model/scenario)
- Promptfoo / DeepEval / Ragas / Great Expectations eval-artifact ingestion (plus gauntlet via JSON-compatible ingestion)
- recall regression gate over the bundled fixture set
- 10 of 12 new AI detectors marked `[stable]`
- canonical 11-command CLI shape (legacy aliases still work)

**Experimental** — useful but not yet hardened; expect signal/UX
changes:

- `aiPromptInjectionRisk` and `aiFewShotContamination` detectors
  (regex-based)
- `terrain serve` local HTTP server (no auth model, localhost-only)
- `terrain portfolio` multi-repo analysis
- portfolio-level scoring thresholds
- AI surface inference *precision* (recall is regression-gated)

### AI detector batch (12 new)

10 ship `[stable]`, 2 ship `[experimental]`. Most detectors carry a
recall-anchor fixture and pass the regression gate on the bundled
fixture set. `aiHardcodedAPIKey` ships without a synthetic fixture
(constructing a real-shaped key would risk repository secret-scanner
alerts); it is covered by unit tests only.

- **`aiHardcodedAPIKey`** `[stable]` — config files leaking provider API
  keys. *Covered by unit tests only.*
- **`aiNonDeterministicEval`** `[stable]` — eval configs declaring a model
  without pinning `temperature: 0`. Per-provider scoping (multi-provider
  configs emit one verdict per provider entry, not one for the whole
  file). Accepts `.yaml`, `.yml`, `.json`, `.toml`.
- **`aiModelDeprecationRisk`** `[stable]` — floating model tags
  (`gpt-4`, `claude-3-opus`, etc.) and sunset variants
  (`text-davinci-003`, `code-davinci-001/002`, `claude-2`). Severity by
  category: deprecated → High, floating → Medium. Comment-prefix
  detection covers SQL `--`, INI `;`, HTML `<!--`, Markdown bullet/
  blockquote, RST `..`, VB `'`.
- **`aiPromptInjectionRisk`** `[experimental]` — user-input concatenated
  into prompt-shaped variables without sanitisation. Multi-line
  concatenation supported (3-line window). User-input shapes cover
  Express/Koa, FastAPI typed-parameter constructs, Flask, Django,
  Pyramid, gRPC, and CLI-arg-driven input.
- **`aiToolWithoutSandbox`** `[stable]` — destructive agent tools without
  an approval gate, sandbox flag, or dry-run path. Structural
  key-name + truthy-value check (description fields excluded so
  adversarial bypass via prose doesn't suppress the finding).
  Benign-object whitelist (`delete_cache`, `purge_logs`, etc.) suppresses
  the bounded-blast-radius cases; always-high verbs (`exec`, `eval`,
  `send_payment`) keep firing regardless.
- **`aiSafetyEvalMissing`** `[stable]` — safety-critical AI surfaces
  (prompt / agent / tool / context) with no safety-shaped scenario
  coverage. Implicit path-based coverage when `CoveredSurfaceIDs` is
  empty (the default for auto-derived scenarios) so the detector
  doesn't flood false positives on the dominant scenario shape.
- **`aiHallucinationRate`** `[stable]` — eval runs with
  hallucination-shaped failure rate above the configured threshold.
  Denominator excludes errored cases (provider crash / timeout) via
  `caseIsScoreable` so infra noise doesn't dilute the rate.
  Keyword set covers 17 stems including "not in source", "no
  evidence", "unsupported", "outside scope", "off-topic".
- **`aiCostRegression`** `[stable]` — paired-case avg cost-per-case rising
  more than the configured threshold versus a baseline snapshot. Both
  relative AND absolute deltas must clear (default `MinAbsDelta` =
  $0.0005/case) so $0.0001 → $0.0002 noise doesn't fire. Confidence
  scales by paired-case count (0.5 at paired=1, plateau at 0.9 from
  paired≥20). Catastrophic regressions (≥2× cost) escalate to High
  via `sev-high-008`.
- **`aiRetrievalRegression`** `[stable]` — retrieval-quality named scores
  dropping versus baseline. Allowlist covers Ragas modern
  (`context_precision`, `context_recall`, `context_entity_recall`),
  Ragas legacy (`context_relevance`), `nDCG`, `coverage`, `faithfulness`,
  `answer_relevancy`, and LangSmith `relevance_score`. Confidence
  scales by paired-case count (shared helper with `aiCostRegression`).
- **`aiPromptVersioning`** `[stable]` — prompt-kind surfaces shipping
  without a recognisable version marker. Placeholder tokens
  (`version: TODO`, `version: TBD`, `version: ???`,
  `version: placeholder`, `version: none`, `version: unknown`) do NOT
  satisfy the requirement.
- **`aiFewShotContamination`** `[experimental]` — prompt few-shot examples
  overlapping verbatim with the inputs of eval scenarios that cover them.
  Implicit path-based coverage matches the dominant auto-derived
  scenario shape (empty `CoveredSurfaceIDs`).
- **`aiEmbeddingModelChange`** `[stable]` — repos referencing an embedding
  model in source without a retrieval-shaped eval scenario. Prefers
  structured RAG surfaces (EvidenceStrong) when present; falls back to
  file-scan (EvidenceModerate). Catches env-var-loaded models via
  framework constructor patterns (`OpenAIEmbeddings`,
  `SentenceTransformer`, `langchaingo.NewEmbeddings`, etc.).

### Recall-regression fixture suite

- **Bundled fixtures cover AI / quality / health / migration / structural /
  runtime signal types on real-shaped inputs.** The regression gate
  enforces that every labeled signal still fires after a detector change;
  unlabeled signals are silent (counted neither as positive nor negative).
- **Regression gate is now load-bearing.** Any unmatched expected label
  fails the build; empty-fixture-set bypass is closed.
- **Match-key precision improved.** Matcher key now includes `Symbol`
  in addition to `(Type, File)` so multi-symbol fixtures distinguish
  "fired per-symbol" from "fired once on the same line."
  `ExpectedAbsent` path matching uses the same normalization as the
  positive-match path, fixing eval-data detectors that stamp absolute
  paths.
- **Known gaps**: `aiHardcodedAPIKey` has no fixture
  (constructing a real-shaped key risks repo secret-scanner alerts);
  near-threshold fixtures for cost/retrieval/coverage detectors do not
  yet cover comparator-flip regressions.
- **Eval-data fixture authoring.** The regression runner auto-discovers
  per-fixture `eval-runs/{promptfoo,deepeval,ragas}.json` and
  `baseline.json`. Synthesises baseline snapshots from
  `baseline/eval-runs/` so regression-shaped fixtures are authored as two
  pairs of framework JSON files, not hand-written snapshot blobs.
- **`terrain.yaml` `scenarios.description` field.** Propagates onto
  `models.Scenario.Description` for detectors that compare scenario
  inputs to prompt content.

### Verdict engine — AI-surface findings with cross-file evidence

The headline 0.2 addition. A typed-evidence pipeline (`internal/aipipeline`) replaces regex-only AI surface detection and emits `Finding` records with a confidence score, severity, and full evidence chain.

**Precision lift over the path-based baseline:**

- Substantial precision lift over a path-based baseline on representative app-shape repositories.
- Per-cohort base rates and per-rule weight overrides tune confidence scoring per detector.

The pipeline emits typed evidence atoms; a weighted-log-odds composer produces a posture-thresholded confidence (observability vs gate). Each rule ships with a per-rule fix scaffold (Promptfoo eval YAML, DeepEval pytest, MLflow/W&B trackers).

**Production-context training detector.** `ai.train.missing_tracker` at face value flags every `sklearn.fit(X, y)` — and many of those are noise (tutorials, kaggle exports, research code that doesn't need tracking). The detector is gated on production-context signals (production ML SDKs, scheduling decorators, model-registry registration calls); without one of those signals, sklearn-shaped training code stays below the emission threshold.

**User-facing command.** The pipeline is reachable via the new `terrain ai findings` subcommand:

```
terrain ai findings [--root <dir>] [--json]
                    [--posture observability|gate]
                    [--rule <rule-id>]
```

Output renders one block per finding: path, rule, severity,
confidence, cohort, and the full evidence chain (atom rule ID,
weight, span). JSON mode is structured for CI consumption.

**Known limits.** Training-rule production-context atoms are architecturally correct but conservatively gated. `ai.train.missing_tracker` ships as **preview** — reachable via `terrain ai findings --rule ai.train.missing_tracker`, output carries a `[preview]` tag, and the rule is not in the default rule set.

`terrain ai list` (inventory), `terrain ai findings` (verdict engine), and the AI catalog detectors (via `terrain analyze` / `terrain report pr`) emit different shapes and answer different questions — run all three in CI.

### CLI restructure — canonical 11 commands + 33 legacy aliases

The canonical 11-command surface ships as non-breaking namespace
dispatchers (`terrain report`, `terrain migrate`, `terrain config`)
alongside the historical 32 top-level commands. The binary today
accepts ~43 top-level entries; the 11-command shape is the
*recommended* surface, not the only-reachable surface, and `terrain
--help` still lists the legacy commands. Legacy commands remain
through 0.2; in-band deprecation warnings and removal are future work.

```
1.  terrain init
2.  terrain analyze
3.  terrain report <verb>     # 9 read-side verbs (summary, insights,
                              #   metrics, explain, show, impact, pr,
                              #   posture, select-tests)
4.  terrain migrate <verb>    # 11 verbs (run/config/list/detect/
                              #   shorthands/estimate/status/checklist/
                              #   readiness/blockers/preview)
5.  terrain ai <verb>
6.  terrain portfolio <verb>
7.  terrain config <verb>     # feedback, telemetry
8.  terrain doctor
9.  terrain debug <verb>
10. terrain serve
11. terrain version
```

`terrain convert <file> --to <framework>` continues to work via the
per-file converter — the `convert` namespace dispatcher falls through
to `runConvertCLI` (single-file mode) for non-verb args, distinct from
the `migrate` namespace's directory-mode fall-through.

### Eval framework adapters

- **Promptfoo.** `internal/airun.ParsePromptfooJSON` reads `--output` JSON
  (v3 nested + v4 flat shapes). Wired through `--promptfoo-results` flag.
- **DeepEval.** `--deepeval-results` flag, same envelope shape.
- **Ragas.** `--ragas-results` flag, same envelope shape.
- **Baseline-snapshot mechanism.** `--baseline <path>` loads a previous
  `TestSuiteSnapshot` and attaches it to the current run for
  regression-aware detectors.

### RAG structured parser — Go + Java added

`ParseRAGStructured` was JS+Python only in 0.1; 0.2 adds langchaingo
(Go) and langchain4j (Java) parsers. Same six component kinds across
all four languages: embedding model, vector store, text splitter,
retriever, document loader, reranker. Config extraction
(`ChunkSize`, `TopK`, `ModelName`) maps to `RAGComponentConfig` so
the embedding-model-change detector gets the high-confidence
structured-surface path on Go and Java codebases too.

### SignalV2 schema

Nine new fields on `models.Signal`, all `omitempty`:
`SeverityClauses`, `Actionability`, `LifecycleStages`, `AIRelevance`,
`RuleID`, `RuleURI`, `DetectorVersion`, `RelatedSignals`,
`ConfidenceDetail`. **TestSuiteSnapshot schema** bumped from 1.0.0 →
1.1.0; **manifest export schema** stays at 1.0.0 (the two version
strings are independent — see `docs/schema/COMPAT.md`).

### Severity rubric

17 stable clauses (`sev-{critical,high,medium,low,info}-NNN`) named in
`internal/severity/rubric.go`, rendered to `docs/severity-rubric.md`
via `cmd/internal/terrain-docs-gen`. Each detector quotes the clauses it
exercises in its emitted signals.

### Auto-generated rule docs

`docs/rules/<domain>/<rule>.md` auto-generated for all 68 rules by
`cmd/internal/terrain-docs-gen`. Hand-authored prose below the
`<!-- docs-gen: end stub -->` marker is preserved across regenerations.
Drift fails `make docs-verify` (CI gate).

### Other infrastructure

- **Generated signal manifest.** Regenerated from `internal/signals.allSignalManifest` via `cmd/internal/terrain-docs-gen`. `make docs-gen` writes; `make docs-verify` diffs.
- **CI hard-fail gate** on `make docs-verify` (extended ubuntu runner).
- **Performance regression gate.** `make bench-gate` +
  `terrain-bench-gate` fail PRs that regress benchmarks >10%.
- **SLSA L2 build provenance.** `actions/attest-build-provenance@v3`
  emits a signed in-toto attestation per release archive.
- **Tree-sitter parser pool.** `sync.Pool` reuses parsers across calls.
- **Pytest fixture dependency graph.** `@pytest.fixture` parameter
  extraction feeds the import graph.
- **JUnit 5 `@Nested` + `@DisplayName` extraction.** Hierarchical test
  identification matches the framework's reporting model.
- **Hierarchical Go `t.Run` extraction.** Sub-test stack tracking.
- **Vitest in-source tests.** `if (import.meta.vitest)` blocks discovered
  alongside conventional spec files.
- **TSConfig path resolution.** `extends` chain + multi-target +
  `jsconfig.json` fallback.
- **`.terrain/conversion-history/` audit trail.** Every conversion writes
  a JSONL line.
- **Per-file conversion confidence.** Per-file scores expose where the
  converter was uncertain.
- **`terrain convert --preview`.** LCS-based unified diff.
- **AI surface detection expansion.** Datasets, pgvector cursor calls,
  MCP tool definitions, in-memory FAISS indexes.
- **Capability validation gap detector.** Pairs AI capabilities with
  eval scenarios; flags capabilities without validation.
- **`terrain ai run` captures eval framework output** to
  `.terrain/artifacts/`.
- **Cosign keyless signing + npm provenance + SLSA attestations** on
  every release archive. The npm postinstall verifier
  (`bin/terrain-installer.js`) requires cosign by default and
  hard-fails when it isn't on `PATH`. Two opt-out env vars are supported and documented in the failure message: `TERRAIN_INSTALLER_ALLOW_MISSING_COSIGN=1` for checksum-only verification, `TERRAIN_INSTALLER_SKIP_VERIFY=1` to skip verification entirely. **Known UX gap (future work):** `bin/postinstall.js` currently catches the verifier error and prints a warning rather than failing `npm install`, so a host without cosign gets a successful install + a deferred fetch retry on first run.

### Changed

- `package.json`, `extension/vscode/package.json`, `package-lock.json`
  at 0.2.0.

### Fixed

- Race detector failure on Ubuntu CI from `os.Stdout`-touching parallel
  tests; `runCaptured` wraps the previously-unprotected callers.
- `TestParallelForEachIndexCtx_CancelMidway` flaky on Ubuntu race
  runners; per-item sleep makes cancellation propagation visible.
- Fixture coverage file wasn't tracked
  (`.gitignore` filtered `coverage/`); exception added.
- `docs-verify.sh` lacked the executable bit in the git index.
- `aiModelDeprecationRisk` regex matched dot-versioned variants like
  `claude-2.1` and `gpt-3.5-turbo-0125` against their undated parents
  (`claude-2`, `gpt-3.5-turbo`) — guaranteed false positive on current
  pinned models. Trailing-boundary class now excludes `.`.
- `aiRetrievalRegression` allowlist missed Ragas's modern
  `context_precision`/`context_recall`/`context_entity_recall` keys;
  detector silently fired zero signals on real Ragas runs. Added.
- `terrain convert <file> --to <framework>` regressed during the CLI
  fold-in (routed to project-wide migrate runner). Restored by giving
  `convert` its own namespace dispatcher with `runConvertCLI` as the
  fall-through.

### Release-prep polish

Beyond the headline detector + CLI work, a release-prep review pass
closed the verified P0/P1 subset before tag:

- **Release infra**: `npm-release` job adds `setup-go` (would have
  crashed at first publish via `prepublishOnly → verify-pack.js → go
  build`); `supply-chain.md` drops a phantom `windows/arm64` artifact
  goreleaser doesn't build; SLSA L2 build-provenance via
  `actions/attest-build-provenance@v3` is documented; new
  `release-smoke` job downloads + verifies the published archive
  reports the tag's version.
- **Engine self-diagnostic**: `detectorPanic` added to
  `models.SignalCatalog` + manifest. Pre-fix `safeDetect`'s panic-
  recovery emitted a sentinel that `ValidateSnapshot` then rejected as
  unknown, dropping the whole snapshot — defeating the graceful-
  degradation promise. `RequiresGraph` mismatch now surfaces a
  detectorPanic-shaped diagnostic instead of silently dropping the
  registration.
- **Eval adapters**: Promptfoo errors-bucket wired through the row-
  derived stats fallback so provider-crash rows land in
  `Aggregates.Errors` (not `Failures`); per-case cost falls back to
  top-level `cost` field when `r.response.tokenUsage.cost` is zero;
  `createdAt` magnitude check (seconds vs millis) handles v4 CLI
  variants. DeepEval gains `runId` fallback (newer 1.x shape) and
  metric-name whitespace normalization. Ragas accepts
  `evaluation_results` (modern ≥0.1.0) and `scores` (DataFrame export)
  shapes alongside legacy `results`. Envelope `SourcePath` now
  repo-relative (forward-slash normalized) so SARIF output doesn't
  leak developer home directories.
- **CLI**: 14 legacy commands gain `legacyDeprecationNotice` calls so
  `TERRAIN_LEGACY_HINT=1` produces uniform migration prompts;
  `--read-only` on `terrain serve` promoted from no-op to actual HTTP
  405 enforcement; `terrain version --json` includes
  `schemaVersion`; `terrain show`/`explain` use a dedicated `exit 5
  (not found)` so CI scripts can branch on missing-entity vs analysis
  failure. `runDepgraph` routed through `AnalyzeContext` for Ctrl-C
  unwind.
- **Determinism**: `sortSignals` adds `Symbol` as a tiebreaker after
  `Line` and switches to `sort.SliceStable` so byte-identical snapshot
  output under `SOURCE_DATE_EPOCH` survives signals on the same
  (Type, File, Line) but different symbols.
- **Supply chain hardening**: every PR-triggered workflow gains
  `concurrency` + `cancel-in-progress` so force-pushes don't pile up
  runs; `timeout-minutes` on every job (15-45min); CodeQL Python
  matrix dropped (no production Python under analysis);
  `COSIGN_EXPERIMENTAL=1` removed from cosign 2.x invocations;
  installer redirect chain capped at 5; goreleaser archives ship
  `LICENSE` + `README.md`.
- **Documentation**: `CODE_OF_CONDUCT.md` (Contributor Covenant 2.1); three issue templates (bug-report, false-positive, feature-request); new `docs/glossary.md`, `docs/versioning.md`, `docs/compatibility.md`; per-framework integration guides under `docs/integrations/{promptfoo,deepeval,ragas}.md`.
- **CLI visual polish**: dropped a stray `file:` loader-
  prefix in `terrain insights` source paths; replaced `n thing(s)`
  pluralisation notation with proper plural forms across analyze /
  insights / summary / reporting (~19 sites); switched dimension
  display labels to sentence case (`Coverage Depth` →
  `Coverage depth`) for inline use; added polarity-aware band
  rendering so risk-shaped dimensions read naturally
  (`Structural risk: Strong` → `Structural risk: Low`); replaced
  band-only posture lines with concrete totals
  (`Health: Strong  (28 / 772 skipped)`) and dropped zero-valued
  measurements so the line shows what moved the band; added
  `debug <verb>` verb list to top-level help for parity with the
  other namespace dispatchers; `terrain export benchmark` now
  accepts `--json` (no-op; output is always JSON) for flag parity.

## [0.1.2] — Truth-up & foundation

The deliberate "boring" release. No new headline features; instead, every
gap between what Terrain marketed and what the code actually delivered is
either closed or explicitly tagged. Schemas, signal vocabulary, and
distribution surfaces are locked so 0.2 can ship features against a stable
foundation.

### Honest about what ships

- New: `docs/release/feature-status.md` is the canonical inventory of
  stable / experimental / planned features. Drift between marketing and
  code becomes a release blocker starting in 0.2.
- README: example CLI outputs are now framed explicitly as illustrative
  shape, not literal output. Three signals shown (`xfailAccumulation`,
  statistical ">10% failure rate" flaky detection, `0.91+` duplicate
  similarity) are explicitly tagged `[experimental]` or `[planned]`
  because the underlying detectors don't ship in 0.1.2.
- README: the "30 seconds" claim is now scoped to small-to-medium repos
  with realistic numbers for larger workspaces.
- `docs/legacy/`: every file now carries a strong **DEPRECATED — DO NOT
  USE FOR NEW WORK** banner pointing at current docs.
- `internal/convert/catalog.go`: 10 conversion directions tagged `GoNativeStateExperimental` (Java, Python, TestCafe, Selenium families). `terrain convert` warns to stderr when invoked on an experimental direction.

### Distribution

- Goreleaser now builds five platforms instead of one: darwin/amd64,
  darwin/arm64, linux/amd64, linux/arm64, windows/amd64. Each is built
  on a matching CI runner because go-tree-sitter requires CGO and
  cannot cross-compile cleanly.
- Release archives, SBOMs, and checksums are signed via Sigstore
  keyless cosign. Signatures and certificates are uploaded with each
  artifact.
- npm postinstall (`bin/terrain-installer.js`) gains a best-effort
  cosign verifier: in 0.1.2 it warns on missing cosign, missing
  signature artifacts, or verification failure but does not block
  install. 0.2 makes this hard-fail unless
  `TERRAIN_INSTALLER_SKIP_VERIFY=1` is set.
- `.github/dependabot.yml`: gomod, github-actions, and the VS Code
  extension package are now tracked alongside the existing root-npm
  ecosystem. Tree-sitter grammar updates surface as PRs automatically.

### Schema & signal vocabulary

- `internal/signals/manifest.go` (new): single source of truth for all
  56 signal types. Status (stable / experimental / planned), default
  severity, confidence range, evidence sources, RuleID, RuleURI, and
  promotion plan are recorded for every entry.
  `TestManifest_MatchesSignalTypes` makes constant↔manifest drift a
  build failure.
- `internal/models/MaxSupportedMajorSchema = 1`. Snapshot reads now
  reject majors above the current binary's understanding via
  `ValidateSchemaVersion`.
- `docs/schema/COMPAT.md` (new): the public compatibility contract.
  Documents what is allowed at minor steps, what requires a major bump,
  and how the manifest's drift gates fit in.
- Risk-band assignment and Health Grade derivation now use named constants throughout (no magic numbers in the implementation).

### Correctness & durability fixes

- `.gitignore` is now honored during repository scanning. Vendored
  trees and generated artefacts the user has explicitly excluded are
  no longer walked.
- File cache is bounded: per-file 8 MB, total 256 MB. Files past the
  cap stream from disk on every read instead of failing the process.
- Worker-pool sizing capped at `min(GOMAXPROCS, 16)`.
- Framework detection probe size raised from 64 KB to 256 KB.
- `internal/metrics/metrics.go:Derive`, `internal/analyze/analyze.go:Build`,
  and `internal/insights/insights.go:Build` are now nil-safe; the
  adversarial test that previously swallowed their panics with
  `t.Logf("acceptable")` is now a strict contract test that fails on
  panic.

### CLI ergonomics

- `NO_COLOR`, `TERM=dumb`, and every common CI provider
  (GitHub Actions, GitLab, CircleCI, Buildkite, Jenkins, Azure
  Pipelines) now suppress progress output. Logs no longer get
  carriage-return garbage in CI.
- Did-you-mean suggestions on unknown commands. Levenshtein distance
  ≤2 gets you up to three suggestions; in-tree implementation, no new
  dependency.
- Exit codes documented as a 5-level scheme. `exitPolicyViolation`
  remains 2 for back-compat in 0.1.2; 0.2 splits it cleanly.
- `terrain doctor` and `terrain ai doctor` consolidation deferred to
  0.2 (the larger CLI restructure).

### Security & privacy

- `--base` git refs are validated against an allow-listed regex
  before being passed to `git diff`. Shell-injection payloads,
  reflog selectors (`@{-1}`), `--upload-pack=evil`, and whitespace
  are all rejected.
- Telemetry config and event log now ship 0o600; the parent
  `~/.terrain` directory ships 0o700.
- SARIF emission gains `--redact-paths`; absolute paths inside the
  repo are rewritten relative, paths outside collapse to bare
  basenames.
- `terrain serve` ships a security middleware: CSP, X-Frame-Options
  DENY, X-Content-Type-Options nosniff, Referrer-Policy no-referrer
  on every response. Origin/Referer validation rejects browser-driven
  cross-origin attacks against localhost. New `--host` flag warns
  when bound to a non-localhost address.

### CI & governance

- Multi-OS test matrix: ubuntu-latest, macos-latest, windows-latest.
  ubuntu remains the canonical runner with the race detector and full
  fixture suite; macos and windows run unit tests to catch
  platform-specific regressions before binaries ship.
- Determinism gate (`make test-determinism`) now runs in CI on every
  PR.
- New: `.github/CODEOWNERS`, `.github/pull_request_template.md`,
  `.husky/pre-commit` (blocks files >5 MB and binary-only extensions).
- `.nvmrc` strict-pinned to `22.11.0`.

### Removed

- `internal/plugin/` package (extension-point interfaces that were
  never wired into the engine). The only adopters were tests in the
  package itself. The current detector registry pattern is documented
  in the contributor docs.

### Versioning

- npm package, `extension/vscode/package.json`, and
  `package-lock.json` all bumped to `0.1.2`. Git-tag/package.json
  drift is now a release-gate failure.

## [0.1.0] — Test System Intelligence Platform

Terrain 0.1.0 is the first public release of the Terrain test intelligence
platform. A ground-up rewrite of the analysis engine in Go, the legacy
JavaScript converter becomes one subsystem within a signal-first intelligence
platform that maps test suites, surfaces risk, and drives CI optimization —
all from a single statically-linked binary with zero runtime dependencies.

**83k lines of Go across 47 internal packages. 210 test files. 48 test
packages, all passing. Zero `go vet` warnings. Zero `gofmt` issues.**

### Core Analysis Pipeline

- 10-step deterministic pipeline: scan, policy, signals, ownership, runtime, risk, coverage, measurement, portfolio, snapshot
- Repository scanning with framework detection (17 frameworks across Go, JS/TS, Python, Java)
- Test file discovery, code unit extraction, import graph construction
- Signal-first architecture: every finding is a structured Signal with type, severity, confidence, evidence, and location
- Code surface inference: prompts, contexts, datasets, tool definitions, retrieval/RAG, agents, eval definitions
- Behavior surface derivation from API routes, event handlers, and state transitions
- Environment/device matrix analysis from CI configs and framework settings

### 18 Measurements Across 5 Posture Dimensions

| Dimension | Measurements |
|-----------|-------------|
| Health | flaky share, skip density, dead test share, slow test share |
| Coverage Depth | uncovered exports, weak assertion share, coverage breach share |
| Coverage Diversity | mock-heavy share, framework fragmentation, E2E concentration, E2E-only units, unit test coverage |
| Structural Risk | migration blocker density, deprecated pattern share, dynamic generation share |
| Operational Risk | policy violation density, legacy framework share, runtime budget breach share |

### Signal Detectors

- **Quality**: weak assertions, mock-heavy tests, untested exports, assertion-free tests, orphaned tests
- **Health**: slow tests, flaky tests, skipped tests, dead tests, unstable suites (runtime-backed)
- **Migration**: deprecated patterns, dynamic test generation, custom matchers, unsupported setup, framework fragmentation
- **Governance**: policy violations, legacy framework usage, runtime budget exceeded, AI safety
- **Structural**: phantom eval scenarios, blast-radius hotspots, coverage gap clusters

### CLI Commands (30+)

**Primary commands:**
- `terrain analyze` — full test system analysis with key findings, repo profile, risk posture
- `terrain insights` — prioritized health report with categorized findings and recommendations
- `terrain impact` — change-scope analysis: impacted units, tests, protection gaps, owners
- `terrain explain` — structured reasoning chains for any entity (test, unit, owner, scenario, selection)

**Supporting commands:**
- `terrain init` — detect data files and generate recommended analyze command
- `terrain summary` — executive summary with risk, trends, benchmark readiness
- `terrain focus` — prioritized next actions with top risk areas
- `terrain posture` — detailed posture breakdown with measurement evidence
- `terrain portfolio` — portfolio intelligence: cost, breadth, leverage, redundancy
- `terrain metrics` — aggregate metrics scorecard
- `terrain compare` — snapshot-to-snapshot trend tracking
- `terrain select-tests` — protective test set for a change
- `terrain pr` — PR/change-scoped analysis (markdown, comment, annotation output)
- `terrain show <entity> <id>` — drill into test, unit, owner, or finding
- `terrain migration <sub>` — readiness, blockers, or preview
- `terrain policy check` — evaluate local policy rules (exit 0/1/2 for CI)
- `terrain export benchmark` — privacy-safe JSON export
- `terrain serve` — local HTTP server with HTML report and JSON API

**AI / eval:**
- `terrain ai list` — list detected scenarios, prompts, datasets, eval files
- `terrain ai run` — execute eval scenarios with impact-based selection
- `terrain ai replay` — replay and verify a previous run artifact
- `terrain ai record` — save eval results as baseline
- `terrain ai baseline` — manage eval baselines (show, compare)
- `terrain ai doctor` — validate AI/eval setup

**Conversion / migration:**
- `terrain convert` — Go-native source test conversion (25 directions)
- `terrain convert-config` — framework config file conversion
- `terrain migrate` — project-wide migration with state tracking
- `terrain estimate` — migration complexity estimation
- `terrain status` / `terrain checklist` / `terrain doctor` / `terrain reset` — migration workflow
- `terrain list-conversions` / `terrain shorthands` / `terrain detect` — catalog and detection
- 50 shorthand aliases (e.g., `terrain cy2pw`, `terrain jest2vt`)

**Debug:**
- `terrain debug graph|coverage|fanout|duplicates|depgraph` — internal analysis inspection

### AI / Regular Test Parity

AI surfaces receive the same CI treatment as regular tests:

- Discovery: prompts, contexts, datasets, tool definitions, RAG pipelines, agents, eval definitions
- Impact selection: `terrain ai run --base main` selects only impacted eval scenarios
- Protection gaps: changed AI surfaces without eval coverage appear in `terrain impact` and `terrain pr`
- Policy enforcement: 7 AI-specific policy rules (`block_on_safety_failure`, `block_on_uncovered_context`, etc.)
- PR comments: AI Risk Review section in `terrain pr` output (markdown + text)
- GitHub Action: `terrain-ai.yml` template for AI CI gates
- Health insights: uncovered AI surfaces appear in `terrain insights`

### Structural Intelligence

Three features that use the dependency graph and surface model to produce recommendations no individual tool can generate:

- **"What to test next"**: ranks untested source files by import graph dependency count — files with more dependents create larger blind spots for change-scoped test selection
- **AI behavior impact chains**: detects files with multiple AI surface types where some are covered and others aren't — a change to the untested surface can alter downstream AI behavior undetected
- **Capability gap detection**: identifies AI capabilities with only positive/accuracy scenarios but no adversarial, safety, or robustness scenarios

### Impact Analysis

- Change-scope analysis against git diff with structural dependency tracing
- Protective test set selection with confidence scoring and reason chains
- Edge-case policy: fallback strategies, confidence adjustments, risk elevation
- Drill-down views: units, gaps, tests, owners, graph, selected
- Manual coverage overlay for untestable paths
- PR-scoped output: markdown, CI comment, GitHub annotations
- AI protection gaps: changed AI surfaces without eval coverage

### Dependency Graph Engine

- 5 reasoning engines: coverage, duplicates, fanout, redundancy, profile
- Edge-case detection (14 types) with policy recommendations
- Stability clustering for shared root-cause detection
- Environment/device matrix coverage analysis
- Language-aware fanout threshold (25, tuned across Go/Python/JS/Java)

### Go-Native Conversion Runtime

- 25 conversion directions across 4 categories (E2E, unit JS, unit Java, unit Python)
- AST-based converters using tree-sitter for structural accuracy
- Semantic validation of converted output
- Config file conversion (Jest, Vitest, Cypress, Playwright, WebdriverIO, Mocha)
- Project-wide migration with dependency ordering, state tracking, resume/retry
- Confidence scoring per converted file

### Artifact Ingestion

- Runtime: JUnit XML and Jest JSON parsers with file-level metric aggregation
- Coverage: LCOV and Istanbul JSON parsers with code unit attribution
- Coverage by type: unit, integration, e2e run labeling
- Per-test coverage mapping
- Gauntlet AI eval artifact ingestion
- Auto-discovery of common artifact paths

### Reporting

- 14 report renderers (analyze, impact, insights, posture, metrics, portfolio, summary, focus, migration, policy, comparison, explain, impact drilldown, executive)
- HTML report with embedded charts
- SARIF output for IDE integration
- GitHub annotation output for CI
- Markdown PR comment output

### Snapshot and Comparison

- `terrain analyze --write-snapshot` — persist snapshots for trend tracking
- `terrain compare` — snapshot-to-snapshot comparison with signal deltas and risk band changes
- Automatic trend loading in summary and insights commands

### Ownership

- CODEOWNERS file parsing with glob pattern matching
- terrain.yaml ownership configuration
- Git history-based ownership inference
- Owner-scoped health and quality summaries
- Owner-filtered impact analysis

### Policy and Governance

- `.terrain/policy.yaml` rule definitions
- AI policy: block on safety failure, block on signal types
- Framework allowlists and denylists
- Runtime budget enforcement
- CI-friendly exit codes (0 = pass, 2 = violations)

### Packaging

- goreleaser config for multi-platform binaries (macOS, Linux, Windows; amd64, arm64)
- SBOM generation (CycloneDX, SPDX)
- Sigstore signing
- Homebrew tap (`pmclSF/terrain/mapterrain`)
- npm package (`mapterrain`) with platform-specific binary installation
- VS Code extension with sidebar views and commands
- Opt-in privacy-respecting telemetry (local only, no network)

---

## [0.0.1] — Signal-First Foundation

Internal milestone. Initial Go-native analysis engine with signal-first
architecture, replacing the V2 JavaScript converter.

### Core Analysis
- Repository scanning with framework detection (17 JS/TS/Java/Python/Go frameworks)
- Test file discovery and code unit extraction
- Signal-first architecture: every finding is a structured Signal with type, severity, evidence, and location
- Evidence model with strength (strong/moderate/weak), source, and confidence scoring

### Signal Detectors
- **Quality**: weak assertions, mock-heavy tests, untested exports, coverage threshold breaks
- **Migration**: deprecated patterns, dynamic test generation, custom matchers, unsupported setup, framework fragmentation
- **Governance**: policy violations, legacy framework usage, runtime budget exceeded
- **Health**: slow tests, flaky tests, skipped tests (runtime-backed)

### Risk Modeling
- Explainable risk engine with reliability, change, and speed dimensions
- Risk surfaces by file, directory, owner, and repository scope
- Heatmap model with directory and owner hotspots

### Migration Intelligence
- `terrain migration readiness` — readiness assessment with quality factors and area assessments
- `terrain migration blockers` — blockers by type and area with representative examples
- `terrain migration preview` — file-level and scope-level migration difficulty preview

### VS Code Extension
- Sidebar views: Overview, Health, Quality, Migration, Review
- TreeDataProvider implementations over CLI JSON output

### Packaging
- goreleaser config for multi-platform binaries
- `terrain version` with build metadata
