package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/pmclSF/terrain/internal/explain"
	"github.com/pmclSF/terrain/internal/identity"
	"github.com/pmclSF/terrain/internal/impact"
	"github.com/pmclSF/terrain/internal/models"
	"github.com/pmclSF/terrain/internal/reporting"
	"github.com/pmclSF/terrain/internal/signals"
	"github.com/pmclSF/terrain/internal/uitokens"
)

// jsonOut writes v to stdout as indented JSON.
func jsonOut(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func printShowUsage() {
	fmt.Fprintln(os.Stderr, "Usage: terrain show <test|unit|codeunit|owner|finding|rule> <id-or-path> [--root PATH] [--json]")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Examples:")
	fmt.Fprintln(os.Stderr, "  terrain show test src/auth/login.test.js")
	fmt.Fprintln(os.Stderr, "  terrain show codeunit src/auth/login.ts:authenticate --json")
	fmt.Fprintln(os.Stderr, "  terrain show owner platform")
}

func runExplain(target, root, baseRef string, jsonOutput, verbose bool) error {
	// Rule-id / signal-type fast path: when the target matches a
	// manifest entry, render the manifest + rule doc without running
	// the analyze pipeline. Covers `/terrain explain aiPromptSchemaDrift`,
	// `terrain explain terrain/ai/prompt-schema-drift`, and the bare
	// `terrain explain ai/prompt-schema-drift` shape. Without this,
	// the slash hint emitted by every prtemplates entry fell through
	// to the entity lookup and errored with "entity not found."
	if entry, ok := lookupManifestEntry(target); ok {
		return renderRuleExplanation(entry, root, jsonOutput)
	}

	result, err := runPipelineWithSignals(root, defaultPipelineOptionsWithProgress(jsonOutput))
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}
	snap := result.Snapshot

	// Compute impact result for structured explanation.
	impactResult, impactErr := computeImpactForExplain(root, baseRef, snap)
	if impactErr == nil {
		applyImpactPolicy(impactResult, result)
	}

	// "selection" mode: explain overall test selection strategy.
	if target == "selection" {
		if impactErr != nil {
			return fmt.Errorf("impact analysis required for selection explanation: %w", impactErr)
		}
		sel, err := explain.ExplainSelection(impactResult)
		if err != nil {
			return err
		}
		if jsonOutput {
			return jsonOut(sel)
		}
		reporting.RenderSelectionExplanation(os.Stdout, sel, verbose)
		return nil
	}

	// Try structured test explanation first (if impact data available).
	if impactErr == nil {
		te, err := explain.ExplainTest(target, impactResult)
		if err == nil {
			if jsonOutput {
				return jsonOut(te)
			}
			reporting.RenderTestExplanation(os.Stdout, te, verbose)
			return nil
		}
	}

	// Fall back to legacy entity lookup for non-test targets.

	// Try test file.
	for _, tf := range snap.TestFiles {
		if tf.Path == target {
			if jsonOutput {
				return jsonOut(tf)
			}
			renderTestDetail(tf, snap)
			return nil
		}
	}

	// Try test case by ID or canonical identity.
	for _, tc := range snap.TestCases {
		if tc.TestID == target || tc.CanonicalIdentity == target {
			if jsonOutput {
				return jsonOut(tc)
			}
			renderTestCaseDetail(tc, snap)
			return nil
		}
	}

	// Try code unit.
	for _, cu := range snap.CodeUnits {
		unitID := cu.Path + ":" + cu.Name
		if unitID == target || cu.Name == target || cu.Path == target {
			if jsonOutput {
				return jsonOut(cu)
			}
			renderCodeUnitDetail(cu, snap)
			return nil
		}
	}

	// Try owner.
	ownerID := strings.ToLower(target)
	ownerFound := false
	if snap.Ownership != nil {
		for _, owners := range snap.Ownership {
			for _, o := range owners {
				if strings.ToLower(o) == ownerID {
					ownerFound = true
					break
				}
			}
			if ownerFound {
				break
			}
		}
	}
	if ownerFound {
		return showOwner(target, snap, jsonOutput)
	}

	// Try finding.
	if snap.Portfolio != nil {
		for i, f := range snap.Portfolio.Findings {
			findingID := fmt.Sprintf("%d", i)
			if findingID == target || f.Type == target {
				return showFinding(target, snap, jsonOutput)
			}
		}
	}

	// Try scenario — first with rich explain (impact + snapshot), then fallback.
	if impactErr == nil {
		se, seErr := explain.ExplainEvalRich(target, impactResult, snap)
		if seErr == nil {
			if jsonOutput {
				return jsonOut(se)
			}
			renderScenarioExplanation(se, verbose)
			return nil
		}
	}

	// Fallback: show scenario metadata from snapshot even without impact data.
	for _, sc := range snap.Evals {
		if sc.EvalID == target || sc.Name == target {
			if jsonOutput {
				return jsonOut(sc)
			}
			fmt.Printf("Scenario: %s\n", sc.Name)
			fmt.Printf("ID: %s\n", sc.EvalID)
			if sc.Capability != "" {
				fmt.Printf("Capability: %s\n", sc.Capability)
			}
			fmt.Printf("Category: %s\n", sc.Category)
			if sc.Framework != "" {
				fmt.Printf("Framework: %s\n", sc.Framework)
			}
			if sc.Owner != "" {
				fmt.Printf("Owner: %s\n", sc.Owner)
			}
			if sc.Path != "" {
				fmt.Printf("Path: %s\n", sc.Path)
			}
			if len(sc.CoveredSurfaceIDs) > 0 {
				fmt.Printf("Covered surfaces (%d):\n", len(sc.CoveredSurfaceIDs))
				for _, sid := range sc.CoveredSurfaceIDs {
					fmt.Printf("  %s\n", sid)
				}
			}
			fmt.Println()
			fmt.Println("Next steps:")
			fmt.Println("  terrain impact --base main    see if this scenario is impacted by changes")
			fmt.Println("  terrain ai list               list all scenarios and surfaces")
			return nil
		}
	}

	// Try as a stable finding ID (e.g.
	// "weakAssertion@internal/auth/login_test.go:TestLogin#a1b2c3d4").
	// `terrain explain finding <id>` — round-trip a
	// finding ID back to its evidence + suggest a suppression command.
	if _, _, _, _, ok := identity.ParseFindingID(target); ok {
		if sig, found := lookupSignalByFindingID(snap, target); found {
			if jsonOutput {
				return jsonOut(sig)
			}
			renderFindingExplanation(sig, target, snap)
			return nil
		}
		// Looks like a finding ID but not in this snapshot — distinct
		// from "garbage input": tell the user it parsed correctly but
		// didn't resolve. Common cause: stale ID after a refactor.
		return cliExitError{
			code: exitNotFound,
			message: fmt.Sprintf(
				"finding ID parses but is not in the current snapshot: %s\n\n"+
					"Common causes:\n"+
					"  - the underlying signal moved (file rename, symbol rename, line drift without symbol)\n"+
					"  - the suppression file already drops it — check `.terrain/suppressions.yaml`\n"+
					"  - the snapshot is from a different run — re-run `terrain analyze`",
				target,
			),
		}
	}

	return cliExitError{
		code:    exitNotFound,
		message: fmt.Sprintf("entity not found: %s\n\nTry: a test file path, test ID, scenario ID, finding ID, or 'selection'", target),
	}
}

// lookupSignalByFindingID searches the snapshot for a signal with the
// given FindingID. Returns the signal + ok=true on hit. Walks both
// top-level Signals and per-test-file Signals so any emission path
// resolves.
func lookupSignalByFindingID(snap *models.TestSuiteSnapshot, id string) (models.Signal, bool) {
	if snap == nil {
		return models.Signal{}, false
	}
	for _, s := range snap.Signals {
		if s.FindingID == id {
			return s, true
		}
	}
	for _, tf := range snap.TestFiles {
		for _, s := range tf.Signals {
			if s.FindingID == id {
				return s, true
			}
		}
	}
	return models.Signal{}, false
}

// renderFindingExplanation prints a finding's evidence in human-
// readable form. Mirrors the shape used by other explain renders:
// section header → finding metadata → next-step pointers (including
// the canonical `terrain suppress` invocation).
func renderFindingExplanation(s models.Signal, id string, snap *models.TestSuiteSnapshot) {
	rule := strings.Repeat("─", 60)
	fmt.Println("Terrain — finding explanation")
	fmt.Println(rule)
	fmt.Println()

	fmt.Printf("Finding ID:  %s\n", id)
	fmt.Printf("Detector:    %s\n", s.Type)
	fmt.Printf("Severity:    %s\n", strings.ToUpper(string(s.Severity)))
	if s.Category != "" {
		fmt.Printf("Category:    %s\n", s.Category)
	}
	fmt.Println()

	if s.Location.File != "" {
		loc := s.Location.File
		if s.Location.Symbol != "" {
			loc += " :: " + s.Location.Symbol
		}
		if s.Location.Line > 0 {
			loc += fmt.Sprintf(" (line %d)", s.Location.Line)
		}
		fmt.Printf("Location:    %s\n", loc)
	}
	if s.Owner != "" {
		fmt.Printf("Owner:       %s\n", s.Owner)
	}
	if s.EvidenceStrength != "" {
		fmt.Printf("Evidence:    %s", s.EvidenceStrength)
		if s.EvidenceSource != "" {
			fmt.Printf(" (%s)", s.EvidenceSource)
		}
		fmt.Println()
	}
	if s.RuleID != "" {
		fmt.Printf("Rule:        %s\n", s.RuleID)
	}
	fmt.Println()

	if s.Explanation != "" {
		fmt.Println("Why it matters:")
		fmt.Printf("  %s\n", s.Explanation)
		fmt.Println()
	}
	if s.SuggestedAction != "" {
		fmt.Println("What to do:")
		fmt.Printf("  %s\n", s.SuggestedAction)
		fmt.Println()
	}

	if ev := explain.DetectorEvidenceFor(string(s.Type)); ev != nil {
		if line := ev.FormatTrustLine(); line != "" {
			fmt.Println("Detector confidence:")
			fmt.Printf("  %s\n", line)
			fmt.Println()
		}
	}

	if stacked := relatedFindings(s, snap); len(stacked) > 0 {
		fmt.Println("Cross-detector evidence (same file or symbol):")
		for _, r := range stacked {
			loc := r.Location.File
			if r.Location.Line > 0 {
				loc += fmt.Sprintf(":%d", r.Location.Line)
			}
			if r.Location.Symbol != "" && r.Location.Symbol != s.Location.Symbol {
				loc += " :: " + r.Location.Symbol
			}
			fmt.Printf("  %s %s (%s) %s %s\n", uitokens.GlyphBullet(), r.Type, strings.ToUpper(string(r.Severity)), uitokens.GlyphDash(), loc)
		}
		fmt.Println()
	}

	if lineage, _ := explain.LookupLineage(".", s.Location.File, s.Location.Line); lineage != nil {
		age := lineage.FormatAge(time.Now())
		who := lineage.Author
		if who == "" {
			who = "?"
		}
		fmt.Println("Lineage:")
		if age != "" {
			fmt.Printf("  introduced %s by %s (commit %s)\n", age, who, lineage.ShortSHA)
		} else {
			fmt.Printf("  introduced by %s (commit %s)\n", who, lineage.ShortSHA)
		}
		if lineage.CommitsSince > 0 {
			fmt.Printf("  %d commits to this file since then\n", lineage.CommitsSince)
		}
		fmt.Println()
	}

	fmt.Println("Next steps:")
	fmt.Printf("  terrain suppress %q --reason \"<why>\"   waive this finding (with a reason)\n", id)
	if s.RuleURI != "" {
		fmt.Printf("  see %s for the full detector reference\n", s.RuleURI)
	}
}

// relatedFindings returns up to 5 other signals from snap that touch
// the same file or symbol as s, excluding s itself. Cross-detector
// evidence stacking: when multiple unrelated rules flag the same
// surface, the underlying issue is much more likely to be real (and
// usually deeper than any single rule conveys). Reviewing the cluster
// is faster than triaging each rule in isolation.
//
// Heuristic: match on (a) exact symbol or (b) same file. Same-line
// matches are a stronger hit than same-file matches; we surface
// stronger hits first.
func relatedFindings(s models.Signal, snap *models.TestSuiteSnapshot) []models.Signal {
	if snap == nil || s.Location.File == "" {
		return nil
	}
	type scored struct {
		sig   models.Signal
		score int
	}
	var hits []scored
	for _, other := range snap.Signals {
		if other.FindingID == s.FindingID && s.FindingID != "" {
			continue
		}
		if other.Type == s.Type && other.Location.File == s.Location.File && other.Location.Line == s.Location.Line {
			continue // exact duplicate
		}
		score := 0
		if other.Location.Symbol != "" && other.Location.Symbol == s.Location.Symbol {
			score += 10
		}
		if other.Location.File == s.Location.File {
			score += 5
			if other.Location.Line > 0 && s.Location.Line > 0 &&
				abs(other.Location.Line-s.Location.Line) <= 10 {
				score += 5
			}
		}
		if score >= 5 && other.Type != s.Type {
			hits = append(hits, scored{other, score})
		}
	}
	sort.Slice(hits, func(i, j int) bool { return hits[i].score > hits[j].score })
	if len(hits) > 5 {
		hits = hits[:5]
	}
	out := make([]models.Signal, 0, len(hits))
	for _, h := range hits {
		out = append(out, h.sig)
	}
	return out
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// computeImpactForExplain runs impact analysis using git diff to detect changes.
func computeImpactForExplain(root, baseRef string, snap *models.TestSuiteSnapshot) (*impact.ImpactResult, error) {
	if baseRef == "" {
		baseRef = "HEAD~1"
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	cs, err := impact.ChangeSetFromGitDiff(absRoot, baseRef)
	if err != nil {
		return nil, fmt.Errorf("change detection failed: %w", err)
	}

	return impact.AnalyzeChangeSet(cs, snap), nil
}

func runShow(entity, id, root string, jsonOutput bool) error {
	entity = strings.TrimSpace(entity)
	id = strings.TrimSpace(id)
	switch entity {
	case "test", "unit", "codeunit", "owner", "finding":
	case "rule":
		// Rule lookup goes through the manifest, not the snapshot.
		// The deprecation hint emitted on every aliased rule_id tells
		// users to "Run `terrain show rule <id>`" — without this case
		// the suggestion errored with "unknown entity type."
		if id == "" {
			return fmt.Errorf("missing rule id for show rule")
		}
		entry, ok := lookupManifestEntry(id)
		if !ok {
			return fmt.Errorf("rule %q not found in manifest (try the bare SignalType, the full RuleID, or `<category>/<rule>` form)", id)
		}
		return renderRuleExplanation(entry, root, jsonOutput)
	default:
		return fmt.Errorf("unknown entity type: %q (valid: test, unit, codeunit, owner, finding, rule)", entity)
	}
	if id == "" {
		return fmt.Errorf("missing ID for show %q", entity)
	}

	result, err := runPipelineWithSignals(root, defaultPipelineOptions())
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}
	snap := result.Snapshot

	switch entity {
	case "test":
		return showTest(id, snap, jsonOutput)
	case "unit", "codeunit":
		return showCodeUnit(id, snap, jsonOutput)
	case "owner":
		return showOwner(id, snap, jsonOutput)
	case "finding":
		return showFinding(id, snap, jsonOutput)
	default:
		return fmt.Errorf("unhandled entity type: %q", entity)
	}
}

func showTest(id string, snap *models.TestSuiteSnapshot, jsonOutput bool) error {
	// Search by test ID or file path.
	for _, tf := range snap.TestFiles {
		if tf.Path == id {
			if jsonOutput {
				return jsonOut(tf)
			}
			renderTestDetail(tf, snap)
			return nil
		}
	}
	// Search test cases by ID.
	for _, tc := range snap.TestCases {
		if tc.TestID == id || tc.CanonicalIdentity == id {
			if jsonOutput {
				return jsonOut(tc)
			}
			renderTestCaseDetail(tc, snap)
			return nil
		}
	}
	return cliExitError{code: exitNotFound, message: fmt.Sprintf("test not found: %s", id)}
}

func showCodeUnit(id string, snap *models.TestSuiteSnapshot, jsonOutput bool) error {
	for _, cu := range snap.CodeUnits {
		unitID := cu.Path + ":" + cu.Name
		if unitID == id || cu.Name == id || cu.Path == id {
			if jsonOutput {
				return jsonOut(cu)
			}
			renderCodeUnitDetail(cu, snap)
			return nil
		}
	}
	return cliExitError{code: exitNotFound, message: fmt.Sprintf("code unit not found: %s", id)}
}

func showOwner(id string, snap *models.TestSuiteSnapshot, jsonOutput bool) error {
	ownerID := strings.ToLower(id)

	// Collect owner's files, signals, test files.
	type ownerData struct {
		Owner       string          `json:"owner"`
		OwnedFiles  []string        `json:"ownedFiles"`
		TestFiles   []string        `json:"testFiles"`
		SignalCount int             `json:"signalCount"`
		Signals     []models.Signal `json:"signals,omitempty"`
	}

	data := ownerData{Owner: id}

	if snap.Ownership != nil {
		for path, owners := range snap.Ownership {
			for _, o := range owners {
				if strings.ToLower(o) == ownerID {
					data.OwnedFiles = append(data.OwnedFiles, path)
				}
			}
		}
	}
	sort.Strings(data.OwnedFiles)

	for _, tf := range snap.TestFiles {
		if strings.ToLower(tf.Owner) == ownerID {
			data.TestFiles = append(data.TestFiles, tf.Path)
		}
	}
	sort.Strings(data.TestFiles)

	for _, sig := range snap.Signals {
		if strings.ToLower(sig.Owner) == ownerID {
			data.SignalCount++
			if len(data.Signals) < 10 {
				data.Signals = append(data.Signals, sig)
			}
		}
	}

	if len(data.OwnedFiles) == 0 && len(data.TestFiles) == 0 && data.SignalCount == 0 {
		return cliExitError{code: exitNotFound, message: fmt.Sprintf("owner not found: %s", id)}
	}

	if jsonOutput {
		return jsonOut(data)
	}

	fmt.Printf("Owner: %s\n", data.Owner)
	fmt.Printf("Owned files: %d\n", len(data.OwnedFiles))
	fmt.Printf("Test files: %d\n", len(data.TestFiles))
	fmt.Printf("Signals: %d\n", data.SignalCount)
	if len(data.OwnedFiles) > 0 {
		fmt.Println("\nOwned files:")
		limit := 10
		if len(data.OwnedFiles) < limit {
			limit = len(data.OwnedFiles)
		}
		for _, f := range data.OwnedFiles[:limit] {
			fmt.Printf("  %s\n", f)
		}
		if len(data.OwnedFiles) > 10 {
			fmt.Printf("  ... and %d more\n", len(data.OwnedFiles)-10)
		}
	}
	if data.SignalCount > 0 {
		fmt.Println("\nTop signals:")
		for _, sig := range data.Signals {
			fmt.Printf("  [%s] %s — %s\n", sig.Severity, sig.Type, sig.Location.File)
		}
	}
	fmt.Println("\nNext: terrain show test <path>   drill into a specific test file")
	return nil
}

func showFinding(id string, snap *models.TestSuiteSnapshot, jsonOutput bool) error {
	// Findings are identified by index or type.
	if snap.Portfolio != nil {
		for i, f := range snap.Portfolio.Findings {
			findingID := fmt.Sprintf("%d", i)
			if findingID == id || f.Type == id {
				if jsonOutput {
					return jsonOut(f)
				}
				fmt.Printf("Finding: %s\n", f.Type)
				fmt.Printf("Path: %s\n", f.Path)
				fmt.Printf("Confidence: %s\n", f.Confidence)
				fmt.Printf("Explanation: %s\n", f.Explanation)
				if f.SuggestedAction != "" {
					fmt.Printf("Action: %s\n", f.SuggestedAction)
				}
				return nil
			}
		}
	}
	// Also search signals.
	for i, sig := range snap.Signals {
		sigID := fmt.Sprintf("s%d", i)
		if sigID == id || string(sig.Type) == id {
			if jsonOutput {
				return jsonOut(sig)
			}
			fmt.Printf("Signal: %s\n", sig.Type)
			fmt.Printf("Category: %s\n", sig.Category)
			fmt.Printf("Severity: %s\n", sig.Severity)
			fmt.Printf("File: %s\n", sig.Location.File)
			fmt.Printf("Explanation: %s\n", sig.Explanation)
			return nil
		}
	}
	// Distinct from the stable-finding-ID path above (which gives a
	// detailed "ID parses but didn't resolve" diagnostic). This branch
	// runs when the user passed a numeric index or a type string and
	// neither matched. Help them figure out what to try next.
	return cliExitError{
		code: exitNotFound,
		message: fmt.Sprintf(
			"finding not found: %s\n\n"+
				"`terrain explain finding <id>` accepts:\n"+
				"  - a stable finding ID  (e.g. `weakAssertion@src/auth_test.go:TestLogin#a1b2c3d4`)\n"+
				"  - a portfolio index    (e.g. `0`, `1`, `2` — see `terrain analyze --json`)\n"+
				"  - a signal type        (e.g. `weakAssertion`)\n\n"+
				"If you copied this ID from an older run, re-run `terrain analyze` —\n"+
				"file renames, symbol renames, or line drift can produce a new ID.",
			id,
		),
	}
}

func isUniqueCodeUnitName(snap *models.TestSuiteSnapshot, name string) bool {
	if snap == nil || name == "" {
		return false
	}
	count := 0
	for _, cu := range snap.CodeUnits {
		if cu.Name == name {
			count++
			if count > 1 {
				return false
			}
		}
	}
	return count == 1
}

// lookupManifestEntry resolves a target string to a manifest entry. The
// target may be a SignalType (`aiPromptSchemaDrift`), a full RuleID
// (`terrain/ai/prompt-schema-drift`), or a bare path (`ai/prompt-schema-drift`).
// Returns the matching entry and true on success.
func lookupManifestEntry(target string) (signals.ManifestEntry, bool) {
	target = strings.TrimSpace(target)
	if target == "" {
		return signals.ManifestEntry{}, false
	}
	normalized := strings.TrimPrefix(target, "terrain/")
	for _, entry := range signals.Manifest() {
		if string(entry.Type) == target {
			return entry, true
		}
		if entry.RuleID == target {
			return entry, true
		}
		if strings.TrimPrefix(entry.RuleID, "terrain/") == normalized {
			return entry, true
		}
	}
	return signals.ManifestEntry{}, false
}

// renderRuleExplanation prints the manifest entry's metadata plus the
// hand-authored rule doc (when present on disk). Output is intentionally
// plain text so the slash-dispatcher captureRun path forwards it as a
// PR-comment reply without HTML conversion.
func renderRuleExplanation(entry signals.ManifestEntry, root string, jsonOutput bool) error {
	if jsonOutput {
		return jsonOut(entry)
	}
	fmt.Println(entry.RuleID)
	fmt.Println(strings.Repeat("─", len(entry.RuleID)))
	fmt.Println()
	fmt.Printf("Title:    %s\n", entry.Title)
	fmt.Printf("Type:     %s\n", entry.Type)
	fmt.Printf("Domain:   %s\n", entry.Domain)
	fmt.Printf("Severity: %s (default)\n", entry.DefaultSeverity)
	fmt.Printf("Status:   %s\n", entry.Status)
	fmt.Printf("Tier:     %s\n", entry.Tier)
	if entry.DisabledByDefault {
		fmt.Println("          (disabled by default — opt in via .terrain/policy.yaml)")
	}
	fmt.Println()
	if entry.Description != "" {
		fmt.Println("Description")
		fmt.Println(entry.Description)
		fmt.Println()
	}
	if entry.Remediation != "" {
		fmt.Println("Remediation")
		fmt.Println(entry.Remediation)
		fmt.Println()
	}
	if entry.RuleURI != "" {
		// First try to load the doc from the adopter's repo (works
		// when terrain is invoked from a checkout of the terrain repo
		// itself — e.g. during development). Otherwise just print the
		// canonical doc URL; adopters don't have terrain's rule docs
		// on their disk, so a missing-file warning would be noise.
		docPath := filepath.Join(root, entry.RuleURI)
		if body, err := os.ReadFile(docPath); err == nil && len(body) > 0 {
			fmt.Println("Rule documentation")
			fmt.Println(strings.Repeat("─", 18))
			fmt.Println(strings.TrimSpace(string(body)))
		} else if entry.Status == signals.StatusPlanned {
			fmt.Println("Rule documentation: deferred (this rule is planned; the detector hasn't landed yet, so no doc is shipped).")
		} else {
			fmt.Printf("Rule documentation: https://github.com/pmclSF/terrain/blob/main/%s\n", entry.RuleURI)
		}
	}
	return nil
}
