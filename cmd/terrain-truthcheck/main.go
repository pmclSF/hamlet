// terrain-truthcheck validates Terrain's analysis output against a ground
// truth specification. It runs the full pipeline on a repository and
// compares actual findings to expected findings documented in a YAML truth spec.
//
// Usage:
//
//	terrain-truthcheck --root tests/fixtures/terrain-world --truth tests/fixtures/terrain-world/tests/truth/terrain_truth.yaml
//	terrain-truthcheck --root tests/fixtures/terrain-world --truth tests/fixtures/terrain-world/tests/truth/terrain_truth.yaml --output benchmarks/output/truthcheck/
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pmclSF/terrain/internal/truthcheck"
)

func main() {
	var (
		repoRoot  string
		truthPath string
		outputDir string
		jsonOnly  bool
	)
	flag.StringVar(&repoRoot, "root", ".", "repository root to analyze")
	flag.StringVar(&truthPath, "truth", "", "path to truth spec YAML (required)")
	flag.StringVar(&outputDir, "output", "", "output directory for reports (optional)")
	flag.BoolVar(&jsonOnly, "json", false, "output JSON only")
	flag.Parse()

	if truthPath == "" {
		fmt.Fprintln(os.Stderr, "Usage: terrain-truthcheck --root <repo> --truth <truth.yaml> [--output <dir>] [--json]")
		os.Exit(2)
	}

	// Resolve paths.
	absRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	absTruth, err := filepath.Abs(truthPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Truth validation: %s\n", absRoot)
	fmt.Fprintf(os.Stderr, "Truth spec: %s\n", absTruth)

	report, err := truthcheck.Run(absRoot, absTruth)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Write to output directory if specified.
	if outputDir != "" {
		absOut, _ := filepath.Abs(outputDir)
		if err := truthcheck.WriteReport(absOut, report); err != nil {
			fmt.Fprintf(os.Stderr, "error writing report: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Reports written to %s/\n", absOut)
	}

	// Print to stdout.
	if jsonOnly {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	} else {
		printTextReport(report)
	}

	// Exit with non-zero if any category failed.
	if report.Summary.PassedCount < report.Summary.TotalCategories {
		os.Exit(1)
	}
}

func printTextReport(report *TruthCheckReport) {
	s := report.Summary
	fmt.Println("Truth Validation Report")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()
	fmt.Printf("Overall: %.0f%% (F1)   Precision: %.0f%%   Recall: %.0f%%\n",
		s.OverallScore*100, s.OverallPrecision*100, s.OverallRecall*100)
	fmt.Printf("Categories: %d/%d passed\n", s.PassedCount, s.TotalCategories)
	fmt.Println()

	for _, c := range report.Categories {
		pass := "PASS"
		if !c.Passed {
			pass = "FAIL"
		}
		fmt.Printf("[%s] %-15s score=%.0f%%  precision=%.0f%%  recall=%.0f%%  (%d/%d matched)\n",
			pass, c.Category, c.Score*100, c.Precision*100, c.Recall*100, c.Matched, c.Expected)

		if len(c.Missing) > 0 {
			for _, m := range c.Missing {
				fmt.Printf("  MISSING: %s\n", m)
			}
		}
		if len(c.Unexpected) > 0 {
			for _, u := range c.Unexpected {
				fmt.Printf("  UNEXPECTED: %s\n", u)
			}
		}
	}
	fmt.Println()
}

// Use the truthcheck type to avoid import cycle.
type TruthCheckReport = truthcheck.TruthCheckReport
