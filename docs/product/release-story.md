# Hamlet: Signal-First Test Intelligence

Hamlet is an open-source CLI that analyzes your test suite structurally -- not by counting lines of code, but by examining what your tests actually protect, where risk concentrates, and what your testing strategy is missing.

Most test tooling stops at coverage percentages. Hamlet goes further. It maps your test files, code units, frameworks, and signals into a structured snapshot, then evaluates that snapshot across five posture dimensions: health, coverage depth, coverage diversity, structural risk, and operational risk. Each dimension is backed by concrete measurements with transparent evidence, so findings are traceable and actionable rather than opaque scores.

Hamlet is built for engineering teams, tech leads, and engineering managers who need to answer questions like: Where is our test risk concentrated? Are we actually testing our public API surface? Which tests are redundant? What blocks us from migrating frameworks? What changed since last week?

Run `hamlet analyze` on any repository with test files. Within seconds, you get a structural analysis covering framework detection, signal discovery, risk scoring, posture assessment, and portfolio intelligence. Run `hamlet summary` for a leadership-ready overview. Run `hamlet posture` to see the measurement evidence behind each dimension. Run `hamlet portfolio` to find redundant, overbroad, or high-leverage tests.

Hamlet is local-first. All analysis runs on your machine against your code. Static analysis is the foundation; optional coverage and runtime artifact ingestion enrich the picture. There is no hosted service, no telemetry, no account required. Privacy-safe benchmark exports let you compare anonymized posture data across repositories without exposing source code or file paths.

V3 introduces 18 measurements across 5 posture dimensions, portfolio intelligence, impact analysis, migration readiness assessment, snapshot-based trend tracking, and policy governance. It is a new engine written in Go, designed for speed and precision on repositories of any size.
