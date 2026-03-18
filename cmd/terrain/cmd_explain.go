package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pmclSF/terrain/internal/engine"
	"github.com/pmclSF/terrain/internal/explain"
	"github.com/pmclSF/terrain/internal/impact"
	"github.com/pmclSF/terrain/internal/models"
	"github.com/pmclSF/terrain/internal/reporting"
)

func printShowUsage() {
	fmt.Fprintln(os.Stderr, "Usage: terrain show <test|unit|codeunit|owner|finding> <id-or-path> [--root PATH] [--json]")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Examples:")
	fmt.Fprintln(os.Stderr, "  terrain show test src/auth/login.test.js")
	fmt.Fprintln(os.Stderr, "  terrain show codeunit src/auth/login.ts:authenticate --json")
	fmt.Fprintln(os.Stderr, "  terrain show owner platform")
}

func runExplain(target, root, baseRef string, jsonOutput, verbose bool) error {
	result, err := engine.RunPipeline(root, defaultPipelineOptionsWithProgress(jsonOutput))
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}
	snap := result.Snapshot

	// Compute impact result for structured explanation.
	impactResult, impactErr := computeImpactForExplain(root, baseRef, snap)

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
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(sel)
		}
		reporting.RenderSelectionExplanation(os.Stdout, sel, verbose)
		return nil
	}

	// Try structured test explanation first (if impact data available).
	if impactErr == nil {
		te, err := explain.ExplainTest(target, impactResult)
		if err == nil {
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(te)
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
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(tf)
			}
			renderTestDetail(tf, snap)
			return nil
		}
	}

	// Try test case by ID or canonical identity.
	for _, tc := range snap.TestCases {
		if tc.TestID == target || tc.CanonicalIdentity == target {
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(tc)
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
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(cu)
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
		se, seErr := explain.ExplainScenarioRich(target, impactResult, snap)
		if seErr == nil {
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(se)
			}
			renderScenarioExplanation(se, verbose)
			return nil
		}
	}

	// Fallback: show scenario metadata from snapshot even without impact data.
	for _, sc := range snap.Scenarios {
		if sc.ScenarioID == target || sc.Name == target {
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(sc)
			}
			fmt.Printf("Scenario: %s\n", sc.Name)
			fmt.Printf("ID: %s\n", sc.ScenarioID)
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

	return fmt.Errorf("entity not found: %s\n\nTry: a test file path, test ID, scenario ID, or 'selection'", target)
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
	default:
		return fmt.Errorf("unknown entity type: %q (valid: test, unit, codeunit, owner, finding)", entity)
	}
	if id == "" {
		return fmt.Errorf("missing ID for show %q", entity)
	}

	result, err := engine.RunPipeline(root, defaultPipelineOptions())
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
	}
	return nil
}

func showTest(id string, snap *models.TestSuiteSnapshot, jsonOutput bool) error {
	// Search by test ID or file path.
	for _, tf := range snap.TestFiles {
		if tf.Path == id {
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(tf)
			}
			renderTestDetail(tf, snap)
			return nil
		}
	}
	// Search test cases by ID.
	for _, tc := range snap.TestCases {
		if tc.TestID == id || tc.CanonicalIdentity == id {
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(tc)
			}
			renderTestCaseDetail(tc, snap)
			return nil
		}
	}
	return fmt.Errorf("test not found: %s", id)
}

func showCodeUnit(id string, snap *models.TestSuiteSnapshot, jsonOutput bool) error {
	for _, cu := range snap.CodeUnits {
		unitID := cu.Path + ":" + cu.Name
		if unitID == id || cu.Name == id || cu.Path == id {
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(cu)
			}
			renderCodeUnitDetail(cu, snap)
			return nil
		}
	}
	return fmt.Errorf("code unit not found: %s", id)
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

	for _, sig := range snap.Signals {
		if strings.ToLower(sig.Owner) == ownerID {
			data.SignalCount++
			if len(data.Signals) < 10 {
				data.Signals = append(data.Signals, sig)
			}
		}
	}

	if len(data.OwnedFiles) == 0 && len(data.TestFiles) == 0 && data.SignalCount == 0 {
		return fmt.Errorf("owner not found: %s", id)
	}

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
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
					enc := json.NewEncoder(os.Stdout)
					enc.SetIndent("", "  ")
					return enc.Encode(f)
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
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(sig)
			}
			fmt.Printf("Signal: %s\n", sig.Type)
			fmt.Printf("Category: %s\n", sig.Category)
			fmt.Printf("Severity: %s\n", sig.Severity)
			fmt.Printf("File: %s\n", sig.Location.File)
			fmt.Printf("Explanation: %s\n", sig.Explanation)
			return nil
		}
	}
	return fmt.Errorf("finding not found: %s", id)
}

func isUniqueCodeUnitName(snap *models.TestSuiteSnapshot, name string) bool {
	if name == "" {
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

