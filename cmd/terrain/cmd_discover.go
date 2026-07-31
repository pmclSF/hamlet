package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pmclSF/terrain/internal/aidetect"
	"github.com/pmclSF/terrain/internal/findings"
	"github.com/pmclSF/terrain/internal/gitignore"
	"github.com/pmclSF/terrain/internal/promptcontract"
	"github.com/pmclSF/terrain/internal/uitokens"
)

// runDiscover prints the no-args first-run report — the friendly first touch.
// It leads with comprehension, not a task list:
//
//   - MAPPED:  one line proving what Terrain parsed (prompts, schemas, evals).
//   - ISSUES:  the curated bug-class findings worth acting on (drift), or an
//     "all clear" reward state.
//   - HEALTH:  contracts (from the validated drift detector) and AI-surface
//     eval coverage as a score — not a flood of per-surface line-items.
//   - NEXT:    the commands to run from here.
//
// Rendered through internal/uitokens (see docs/cli-design-tokens.md), so it
// degrades cleanly to monochrome / ASCII / narrow terminals. On a repo with no
// AI surfaces it stays friendly and points at `terrain analyze`.
func runDiscover(root string) error {
	abs, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	// Discovery is strictly read-only — no scan writes .terrain/. The report
	// is AI-focused: what prompts/schemas/evals Terrain sees, and the drift
	// between them. Language/framework inventory lives in `terrain analyze`.
	aiResult := aidetect.Detect(abs)
	schemaFiles := detectSchemaFiles(abs)

	// The drift detector is the validated, LLM-free flagship — cheap enough
	// to run on a first touch and precise enough to lead with. Its inventory is
	// the comprehension proof (it reflects what the detector actually parsed).
	// Any error is non-fatal: the report still renders its map.
	inv, drift, _ := promptcontract.AnalyzeRepo(abs)
	// Surface issues only from the adopter's own code: drop drift whose prompt
	// lives in a borrowed-fixture directory, matching the same exclusion MAPPED
	// applies to the schema inventory (detectSchemaFiles). Without this, the
	// report flags issues in testdata/examples/ it never counted as mapped.
	drift = filterFixtureDrift(drift)

	fmt.Println()
	renderDiscoverMap(filepath.Base(abs), inv, aiResult, schemaFiles)

	// Drift can only fire when a prompt binds to a schema, so its presence is
	// itself proof of AI surfaces even if the coarser aidetect scan missed them.
	// Schemas count too: if MAPPED showed a schema, the report must not then
	// claim "no AI surfaces" — that contradiction is what a prompts-as-template-
	// files repo (a .txt prompt aidetect can't parse, plus a JSON schema) hits.
	hasAI := len(aiResult.PromptFiles) > 0 || len(aiResult.ModelFiles) > 0 ||
		len(aiResult.EvalConfigs) > 0 || inv.Prompts > 0 ||
		inv.Schemas > 0 || len(schemaFiles) > 0
	if !hasAI {
		fmt.Println()
		fmt.Printf("  %s  no AI surfaces here yet %s\n\n",
			uitokens.Muted(uitokens.GlyphMeterEmpty()),
			uitokens.Muted(uitokens.GlyphDash()+" run "+uitokens.Link("terrain analyze")+" for the full posture"))
		return nil
	}

	renderDiscoverIssues(drift)
	renderDiscoverHealth(aiResult, inv, drift)
	// Suggest `terrain fix` only when a drift actually carries a validated fix
	// (not merely when drift exists) — so the next-step is never a dead end.
	renderDiscoverNext(anyFixable(abs, drift))
	return nil
}

// anyFixable reports whether any drift finding carries a validated correct-side
// fix (reusing the already-computed drift; no second detector run).
func anyFixable(abs string, drift []promptcontract.Drift) bool {
	for _, s := range promptcontract.ToSignals(drift) {
		f := findings.FromSignal(s, "terrain/ai/prompt-schema-drift")
		if promptcontract.DriftFix(abs, f) != nil {
			return true
		}
	}
	return false
}

// renderDiscoverMap prints the MAPPED line — the comprehension proof: what
// Terrain parsed, in one scannable row. This is the trust moment.
func renderDiscoverMap(repo string, inv promptcontract.Inventory, ai *aidetect.DetectResult, schemaFiles []string) {
	var parts []string
	add := func(n int, singular string) {
		if n > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", n, plural(n, singular)))
		}
	}
	// Prefer the drift analyzer's inventory (it reflects what was actually
	// parsed); fall back to the coarser scan where it saw more.
	add(max(inv.PromptFiles, len(ai.PromptFiles)), "prompt")
	add(max(inv.Schemas, len(schemaFiles)), "schema")
	add(len(ai.EvalConfigs), "eval")
	add(len(ai.ModelFiles), "model call site")
	add(len(ai.DatasetFiles), "dataset")

	fmt.Printf("  %s   %s", uitokens.Muted("MAPPED"), uitokens.Bold(repo))
	if len(parts) > 0 {
		fmt.Printf("  %s  %s", uitokens.Muted(uitokens.GlyphDot()), uitokens.Muted(strings.Join(parts, " "+uitokens.GlyphDot()+" ")))
	}
	fmt.Println()
}

// renderDiscoverIssues prints the curated bug-class findings — the value
// moment. Capped and scannable; the full list lives behind `terrain report`.
func renderDiscoverIssues(drift []promptcontract.Drift) {
	fmt.Println()
	if len(drift) == 0 {
		fmt.Printf("  %s  %s %s\n\n",
			uitokens.Ok(uitokens.GlyphOK()),
			uitokens.Bold("all clear"),
			uitokens.Muted(uitokens.GlyphDash()+" nothing needs your attention"))
		return
	}

	const cap = 3
	shown := drift
	if len(shown) > cap {
		shown = shown[:cap]
	}
	label := "THING WORTH A LOOK"
	if len(drift) > 1 {
		label = fmt.Sprintf("%d THINGS WORTH A LOOK", len(drift))
	}
	fmt.Printf("  %s\n\n", uitokens.Muted(label))

	for _, d := range shown {
		fmt.Printf("  %s a prompt references a field the schema doesn't declare   %s\n",
			uitokens.Alert(uitokens.GlyphFinding()), uitokens.Warn("[drift]"))
		fmt.Printf("    %s   %s\n",
			uitokens.Muted(fmt.Sprintf("%s:%d", d.PromptPath, d.PromptLine)),
			uitokens.Accent("{"+d.Variable+"}"))
		fmt.Printf("    %s\n\n", uitokens.Muted(d.Message))
	}
	if len(drift) > cap {
		fmt.Printf("  %s\n\n", uitokens.Muted(fmt.Sprintf("+ %d more %s %s",
			len(drift)-cap, uitokens.GlyphDot(), uitokens.Link("terrain report"))))
	}
}

// renderDiscoverHealth prints the HEALTH block — the coverage-as-score
// reframe: contracts as a real drift-derived status, AI-surface eval coverage
// as a real ratio meter. A precise scored posture comes from `terrain analyze`.
func renderDiscoverHealth(ai *aidetect.DetectResult, inv promptcontract.Inventory, drift []promptcontract.Drift) {
	fmt.Printf("  %s\n", uitokens.Muted("HEALTH"))

	// Contracts: derived from the validated drift detector. No denominator to
	// fake a percentage — show the real state.
	if inv.Schemas > 0 && inv.Prompts > 0 {
		contract := uitokens.Ok(healthMeter(1) + "  in sync")
		if len(drift) > 0 {
			contract = uitokens.Alert(fmt.Sprintf("%s  %d drifting", healthMeter(0.3), len(drift)))
		}
		fmt.Printf("  %s   %s\n", padLabel("prompt "+uitokens.GlyphRelates()+" schema contracts"), contract)
	}

	// Coverage: fraction of AI surfaces (prompts + model call sites) whose
	// top-level directory also holds an eval config — the same co-location
	// heuristic the analyze pipeline uses, computed cheaply here.
	surfaces := dedupePaths(append(append([]string{}, ai.PromptFiles...), ai.ModelFiles...))
	if len(surfaces) > 0 {
		covered := countCovered(surfaces, ai.EvalConfigs)
		ratio := float64(covered) / float64(len(surfaces))
		uncovered := len(surfaces) - covered
		pct := int(ratio*100 + 0.5)
		note := uitokens.Ok("covered")
		if uncovered > 0 {
			note = uitokens.Muted(fmt.Sprintf("%d untested", uncovered))
		}
		fmt.Printf("  %s   %s  %s%%  %s\n", padLabel("ai test coverage"), healthMeter(ratio), fmt.Sprintf("%3d", pct), note)
	}
}

// renderDiscoverNext prints the next-step commands using the link role.
func renderDiscoverNext(hasFix bool) {
	fmt.Println()
	cmds := []string{}
	if hasFix {
		cmds = append(cmds, uitokens.Link("terrain fix"))
	}
	cmds = append(cmds, uitokens.Link("terrain analyze"), uitokens.Link("terrain report"))
	fmt.Printf("  %s %s  %s\n\n",
		uitokens.Muted("next"), uitokens.Muted(uitokens.GlyphChevron()),
		strings.Join(cmds, "  "+uitokens.Muted(uitokens.GlyphDot())+"  "))
}

// healthMeter renders a 10-cell ●/○ meter for a ratio in [0,1], colored by
// band (≥0.8 green, ≥0.4 yellow, else red). Callers colorize the label; the
// filled cells here carry the band color, the empty cells stay muted.
func healthMeter(ratio float64) string {
	const width = 10
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	filled := int(ratio*float64(width) + 0.5)
	full := strings.Repeat(uitokens.GlyphMeterFull(), filled)
	empty := strings.Repeat(uitokens.GlyphMeterEmpty(), width-filled)
	var band func(string) string
	switch {
	case ratio >= 0.8:
		band = uitokens.Ok
	case ratio >= 0.4:
		band = uitokens.Warn
	default:
		band = uitokens.Alert
	}
	return band(full) + uitokens.Muted(empty)
}

// padLabel right-pads a health-row label to a fixed column so the meters
// line up. Operates on the plain text (labels here are not pre-colored).
// (plural lives in cmd_severity_gate.go — reused here.)
func padLabel(s string) string { return uitokens.PadRight(s, 28) }

// topDir returns the first path segment of a slash- or OS-separated path.
func topDir(p string) string {
	p = filepath.ToSlash(p)
	if i := strings.IndexByte(p, '/'); i >= 0 {
		return p[:i]
	}
	return p
}

// countCovered counts surfaces whose top-level directory also contains an
// eval config.
func countCovered(surfaces, evalConfigs []string) int {
	evalDirs := map[string]bool{}
	for _, e := range evalConfigs {
		evalDirs[topDir(e)] = true
	}
	n := 0
	for _, s := range surfaces {
		if evalDirs[topDir(s)] {
			n++
		}
	}
	return n
}

// dedupePaths returns the input with duplicates removed, order preserved.
func dedupePaths(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, p := range in {
		if !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	return out
}

// detectSchemaFiles surfaces likely schema/contract files: JSON Schema
// (.schema.json), Protocol Buffers (.proto), GraphQL SDL (.graphql /
// .gql), Pydantic model files (heuristic), and TypeScript type files
// (interface / type aliases in a dedicated types/ or schemas/ dir).
// Skips paths inside test-fixture / benchmark / vendor trees so the
// report doesn't flood on borrowed fixture content. Heuristic, not
// exhaustive — calibration happens in `terrain analyze`.
func detectSchemaFiles(root string) []string {
	var out []string
	// Honour the repo .gitignore, matching the shared scan walkers — a
	// schema inside a gitignored tree is not the adopter's own contract.
	ign := gitignore.Load(root)
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			base := filepath.Base(path)
			if shouldSkipDir(base) {
				return filepath.SkipDir
			}
			if rel, relErr := filepath.Rel(root, path); relErr == nil && ign.Match(filepath.ToSlash(rel), true) {
				return filepath.SkipDir
			}
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		if isFixturePath(rel) || ign.Match(rel, false) {
			return nil
		}
		name := d.Name()
		lower := strings.ToLower(name)

		// File-extension matches.
		if strings.HasSuffix(lower, ".proto") ||
			strings.HasSuffix(lower, ".graphql") ||
			strings.HasSuffix(lower, ".gql") ||
			strings.HasSuffix(lower, ".schema.json") {
			out = append(out, rel)
			return nil
		}

		// Path-based matches: files inside `schemas/` or `models/`
		// directories named with the conventional suffixes.
		relLower := strings.ToLower(rel)
		if (strings.Contains(relLower, "/schemas/") || strings.HasPrefix(relLower, "schemas/")) &&
			(strings.HasSuffix(lower, ".py") ||
				strings.HasSuffix(lower, ".ts") ||
				strings.HasSuffix(lower, ".json") ||
				strings.HasSuffix(lower, ".yaml") ||
				strings.HasSuffix(lower, ".yml")) {
			out = append(out, rel)
			return nil
		}

		return nil
	})
	sort.Strings(out)
	return out
}

// filterFixtureDrift drops drift whose prompt lives in a borrowed-fixture
// directory, so the discovery report's issues stay consistent with the
// fixture-excluded MAPPED inventory.
func filterFixtureDrift(drift []promptcontract.Drift) []promptcontract.Drift {
	out := drift[:0]
	for _, d := range drift {
		if isFixturePath(d.PromptPath) {
			continue
		}
		out = append(out, d)
	}
	return out
}

// isFixturePath returns true when the path lives inside a directory
// commonly used to hold borrowed test fixtures, benchmark inputs, or
// vendored examples. The discovery report should not surface those as
// part of the adopter's actual codebase.
func isFixturePath(rel string) bool {
	lower := strings.ToLower(rel)
	for _, seg := range []string{
		"/testdata/", "testdata/",
		"/fixtures/", "fixtures/",
		"/__fixtures__/", "__fixtures__/",
		"/benchmarks/", "benchmarks/",
		"/examples/", "examples/",
		"/vendor/", "vendor/",
		"/third_party/", "third_party/",
	} {
		if strings.HasPrefix(lower, seg) || strings.Contains(lower, seg) {
			return true
		}
	}
	return false
}

// shouldSkipDir returns true for common no-scan directories. Mirrors the
// skip set used by other walkers in the codebase.
func shouldSkipDir(base string) bool {
	switch base {
	case ".git", "node_modules", "vendor", ".venv", "venv", "dist", "build",
		".terrain", ".next", ".cache", "target", "__pycache__", ".pytest_cache":
		return true
	}
	return false
}

// _ ensures os import is used in code generated tests; safe to remove.
var _ = os.Stat
