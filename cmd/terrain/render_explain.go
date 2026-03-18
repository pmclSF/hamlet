package main

import (
	"fmt"
	"strings"

	"github.com/pmclSF/terrain/internal/explain"
	"github.com/pmclSF/terrain/internal/models"
)

func renderScenarioExplanation(se *explain.ScenarioExplanation, verboseOpt ...bool) {
	isVerbose := len(verboseOpt) > 0 && verboseOpt[0]
	fmt.Println("Terrain Explain — Scenario")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()
	fmt.Printf("Scenario: %s\n", se.Name)
	if se.Capability != "" {
		fmt.Printf("Capability: %s\n", se.Capability)
	}
	if se.Category != "" {
		fmt.Printf("Category: %s\n", se.Category)
	}
	if se.Framework != "" {
		fmt.Printf("Framework: %s\n", se.Framework)
	}
	fmt.Printf("Confidence: %s\n", se.Confidence)
	fmt.Println()
	fmt.Printf("Verdict: %s\n", se.Verdict)
	if se.Relevance != "" {
		fmt.Printf("Reason: %s\n", se.Relevance)
	}
	fmt.Println()

	// Rich surface breakdown.
	if se.RelatedSurfaces != nil {
		rs := se.RelatedSurfaces
		for _, group := range []struct {
			label string
			items []explain.SurfaceRef
		}{
			{"Prompts", rs.Prompts},
			{"Contexts", rs.Contexts},
			{"Datasets", rs.Datasets},
			{"Tool Definitions", rs.ToolDefs},
			{"Retrieval / RAG", rs.Retrievals},
			{"Agent / Orchestration", rs.Agents},
			{"Eval Definitions", rs.EvalDefs},
		} {
			if len(group.items) == 0 {
				continue
			}
			fmt.Printf("%s:\n", group.label)
			for _, ref := range group.items {
				marker := " "
				if ref.Changed {
					marker = "*"
				}
				fmt.Printf(" %s %-30s %s\n", marker, ref.Name, ref.Path)
				if isVerbose {
					if ref.DetectionTier != "" {
						fmt.Printf("     tier: %-12s confidence: %.0f%%\n", ref.DetectionTier, ref.Confidence*100)
					}
				}
			}
		}
		fmt.Println()
	} else if len(se.ChangedSurfaces) > 0 {
		fmt.Printf("Changed surfaces (%d):\n", len(se.ChangedSurfaces))
		for _, s := range se.ChangedSurfaces {
			fmt.Printf("  %s\n", s)
		}
		fmt.Println()
	}

	// Signals.
	if len(se.Signals) > 0 {
		fmt.Printf("Signals (%d):\n", len(se.Signals))
		for _, s := range se.Signals {
			fmt.Printf("  [%s] %s: %s\n", s.Severity, s.Type, s.Explanation)
		}
		fmt.Println()
	}

	// Policy decision.
	if se.PolicyDecision != "" && se.PolicyDecision != "pass" {
		fmt.Printf("Policy: %s\n", se.PolicyDecision)
		fmt.Println()
	}

	fmt.Println("Next steps:")
	fmt.Println("  terrain ai list              view all detected scenarios")
	fmt.Println("  terrain impact --json         machine-readable impact data")
	fmt.Println()
}

func renderTestDetail(tf models.TestFile, snap *models.TestSuiteSnapshot) {
	fmt.Printf("Test File: %s\n", tf.Path)
	fmt.Printf("Framework: %s\n", tf.Framework)
	if tf.Owner != "" {
		fmt.Printf("Owner: %s\n", tf.Owner)
	}
	fmt.Printf("Tests: %d    Assertions: %d\n", tf.TestCount, tf.AssertionCount)
	if tf.MockCount > 0 {
		fmt.Printf("Mocks: %d\n", tf.MockCount)
	}
	if tf.RuntimeStats != nil {
		fmt.Printf("Runtime: %.0fms    Pass rate: %.0f%%    Retry rate: %.0f%%\n",
			tf.RuntimeStats.AvgRuntimeMs,
			tf.RuntimeStats.PassRate*100,
			tf.RuntimeStats.RetryRate*100)
	}
	if len(tf.LinkedCodeUnits) > 0 {
		fmt.Printf("Covers: %s\n", strings.Join(tf.LinkedCodeUnits, ", "))
	}

	// Show signals for this file.
	var fileSignals []models.Signal
	for _, sig := range snap.Signals {
		if sig.Location.File == tf.Path {
			fileSignals = append(fileSignals, sig)
		}
	}
	if len(fileSignals) > 0 {
		fmt.Printf("\nSignals (%d):\n", len(fileSignals))
		for _, sig := range fileSignals {
			fmt.Printf("  [%s] %s: %s\n", sig.Severity, sig.Type, sig.Explanation)
		}
	}
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  terrain explain %s       why this test is selected for changes\n", tf.Path)
	fmt.Println("  terrain impact --show tests   see all impacted tests")
	fmt.Println("  terrain insights              find improvement opportunities")
}

func renderTestCaseDetail(tc models.TestCase, snap *models.TestSuiteSnapshot) {
	fmt.Printf("Test: %s\n", tc.TestName)
	fmt.Printf("ID: %s\n", tc.TestID)
	fmt.Printf("File: %s\n", tc.FilePath)
	if len(tc.SuiteHierarchy) > 0 {
		fmt.Printf("Suite: %s\n", strings.Join(tc.SuiteHierarchy, " > "))
	}
	fmt.Printf("Framework: %s\n", tc.Framework)
	if tc.TestType != "" {
		fmt.Printf("Type: %s (confidence: %.0f%%)\n", tc.TestType, tc.TestTypeConfidence*100)
	}
	fmt.Printf("Extraction: %s (confidence: %.0f%%)\n", tc.ExtractionKind, tc.Confidence*100)
}

func renderCodeUnitDetail(cu models.CodeUnit, snap *models.TestSuiteSnapshot) {
	fmt.Printf("Code Unit: %s\n", cu.Name)
	fmt.Printf("Path: %s\n", cu.Path)
	fmt.Printf("Kind: %s\n", cu.Kind)
	fmt.Printf("Exported: %v\n", cu.Exported)
	if cu.Owner != "" {
		fmt.Printf("Owner: %s\n", cu.Owner)
	}

	// Find covering tests.
	unitID := cu.Path + ":" + cu.Name
	allowNameOnly := isUniqueCodeUnitName(snap, cu.Name)
	var coveringTests []string
	for _, tf := range snap.TestFiles {
		for _, linked := range tf.LinkedCodeUnits {
			if linked == unitID || (allowNameOnly && linked == cu.Name) {
				coveringTests = append(coveringTests, tf.Path)
				break
			}
		}
	}
	if len(coveringTests) > 0 {
		fmt.Printf("\nCovering tests (%d):\n", len(coveringTests))
		for _, t := range coveringTests {
			fmt.Printf("  %s\n", t)
		}
	} else {
		fmt.Println("\nNo covering tests detected.")
	}
	fmt.Println("\nNext: terrain show test <path>   drill into a covering test")
}
