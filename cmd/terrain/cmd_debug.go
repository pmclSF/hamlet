package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/pmclSF/terrain/internal/analysis"
	"github.com/pmclSF/terrain/internal/depgraph"
	"github.com/pmclSF/terrain/internal/metrics"
)

func runDepgraph(root string, jsonOutput bool, show string, changed string) error {
	analyzer := analysis.New(root)
	snapshot, err := analyzer.Analyze()
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	// Build the dependency graph from the snapshot.
	dg := depgraph.Build(snapshot)
	stats := dg.Stats()

	// Run impact if changed files specified.
	var impactResult *depgraph.ImpactResult
	computeImpact := func() {
		if changed == "" || impactResult != nil {
			return
		}
		files := strings.Split(changed, ",")
		ir := depgraph.AnalyzeImpact(dg, files)
		impactResult = &ir
	}

	var (
		coverageComputed   bool
		coverageResult     depgraph.CoverageResult
		duplicatesComputed bool
		duplicatesResult   depgraph.DuplicateResult
		fanoutComputed     bool
		fanoutResult       depgraph.FanoutResult
		profileComputed    bool
		profileResult      depgraph.RepoProfile
		edgeCasesResult    []depgraph.EdgeCase
		policyResult       depgraph.Policy
	)
	computeCoverage := func() depgraph.CoverageResult {
		if !coverageComputed {
			coverageResult = depgraph.AnalyzeCoverage(dg)
			coverageComputed = true
		}
		return coverageResult
	}
	computeDuplicates := func() depgraph.DuplicateResult {
		if !duplicatesComputed {
			duplicatesResult = depgraph.DetectDuplicates(dg)
			duplicatesComputed = true
		}
		return duplicatesResult
	}
	computeFanout := func() depgraph.FanoutResult {
		if !fanoutComputed {
			fanoutResult = depgraph.AnalyzeFanout(dg, depgraph.DefaultFanoutThreshold)
			fanoutComputed = true
		}
		return fanoutResult
	}
	computeProfile := func() (depgraph.RepoProfile, []depgraph.EdgeCase, depgraph.Policy) {
		if !profileComputed {
			coverage := computeCoverage()
			duplicates := computeDuplicates()
			fanout := computeFanout()
			insights := depgraph.ProfileInsights{
				Coverage:   &coverage,
				Duplicates: &duplicates,
				Fanout:     &fanout,
			}
			profileResult = depgraph.AnalyzeProfile(dg, insights)
			dgMetrics := metrics.Derive(snapshot)
			depgraph.EnrichProfileWithHealthRatios(&profileResult, dgMetrics.Health.SkippedTestRatio, dgMetrics.Health.FlakyTestRatio)
			edgeCasesResult = depgraph.DetectEdgeCases(profileResult, dg, insights)
			policyResult = depgraph.ApplyEdgeCasePolicy(edgeCasesResult, profileResult)
			profileComputed = true
		}
		return profileResult, edgeCasesResult, policyResult
	}

	if show != "" {
		switch show {
		case "stats":
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(stats)
			}
			fmt.Println("Terrain Dependency Graph Stats")
			fmt.Println(strings.Repeat("─", 50))
			fmt.Printf("  Nodes: %d    Edges: %d    Density: %.4f\n", stats.NodeCount, stats.EdgeCount, stats.Density)
			fmt.Println()
			fmt.Println("Node Types:")
			for _, nt := range sortedMapKeys(stats.NodesByType) {
				fmt.Printf("  %-20s %d\n", nt, stats.NodesByType[nt])
			}
			return nil
		case "coverage":
			coverage := computeCoverage()
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(coverage)
			}
			fmt.Println("Coverage (structural):")
			fmt.Printf("  Sources: %d   High: %d   Medium: %d   Low: %d\n",
				coverage.SourceCount,
				coverage.BandCounts[depgraph.CoverageBandHigh],
				coverage.BandCounts[depgraph.CoverageBandMedium],
				coverage.BandCounts[depgraph.CoverageBandLow])
			return nil
		case "duplicates":
			duplicates := computeDuplicates()
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(duplicates)
			}
			fmt.Println("Duplicates:")
			fmt.Printf("  Tests analyzed: %d   Duplicates: %d   Clusters: %d\n",
				duplicates.TestsAnalyzed, duplicates.DuplicateCount, len(duplicates.Clusters))
			if duplicates.Skipped {
				fmt.Printf("  Note: %s\n", duplicates.SkipReason)
			}
			if len(duplicates.Clusters) > 0 {
				top := duplicates.Clusters[0]
				fmt.Printf("  Top cluster: %d tests, similarity %.2f\n", len(top.Tests), top.Similarity)
			}
			return nil
		case "fanout":
			fanout := computeFanout()
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(fanout)
			}
			fmt.Println("Fanout:")
			fmt.Printf("  Nodes: %d   Flagged: %d   Threshold: %d\n",
				fanout.NodeCount, fanout.FlaggedCount, fanout.Threshold)
			if fanout.Skipped {
				fmt.Printf("  Note: %s\n", fanout.SkipReason)
			}
			if len(fanout.Entries) > 0 {
				top := fanout.Entries[0]
				fmt.Printf("  Highest: %s (transitive: %d)\n", top.NodeID, top.TransitiveFanout)
			}
			return nil
		case "impact":
			computeImpact()
			if impactResult == nil {
				return fmt.Errorf("impact view requires --changed")
			}
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(impactResult)
			}
			fmt.Println("Impact:")
			fmt.Printf("  Changed files: %d   Impacted tests: %d\n",
				len(impactResult.ChangedFiles), len(impactResult.Tests))
			fmt.Printf("  High: %d   Medium: %d   Low: %d\n",
				impactResult.LevelCounts["high"],
				impactResult.LevelCounts["medium"],
				impactResult.LevelCounts["low"])
			return nil
		case "profile":
			profile, edgeCases, pol := computeProfile()
			if jsonOutput {
				out := map[string]any{
					"profile":   profile,
					"edgeCases": edgeCases,
					"policy":    pol,
				}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(out)
			}
			fmt.Println("Repository Profile:")
			fmt.Printf("  Test Volume:          %s\n", profile.TestVolume)
			fmt.Printf("  CI Pressure:          %s\n", profile.CIPressure)
			fmt.Printf("  Coverage Confidence:  %s\n", profile.CoverageConfidence)
			fmt.Printf("  Redundancy Level:     %s\n", profile.RedundancyLevel)
			fmt.Printf("  Fanout Burden:        %s\n", profile.FanoutBurden)
			fmt.Println()
			if len(edgeCases) > 0 {
				fmt.Println("Edge Cases:")
				for _, ec := range edgeCases {
					fmt.Printf("  [%s] %s: %s\n", ec.Severity, ec.Type, ec.Description)
				}
				fmt.Println()
			}
			if len(pol.Recommendations) > 0 {
				fmt.Println("Recommendations:")
				for _, r := range pol.Recommendations {
					fmt.Printf("  • %s\n", r)
				}
				fmt.Println()
			}
			fmt.Printf("Policy: fallback=%s  confidence=%.2f  optimization=%s\n",
				pol.FallbackLevel, pol.ConfidenceAdjustment,
				map[bool]string{true: "disabled", false: "enabled"}[pol.OptimizationDisabled])
			return nil
		default:
			return fmt.Errorf("unknown view: %s (available: stats, coverage, duplicates, fanout, impact, profile)", show)
		}
	}

	// Full report mode (all engines).
	coverage := computeCoverage()
	duplicates := computeDuplicates()
	fanout := computeFanout()
	profile, edgeCases, pol := computeProfile()
	computeImpact()

	// JSON output.
	if jsonOutput {
		out := map[string]any{
			"stats":      stats,
			"coverage":   coverage,
			"duplicates": duplicates,
			"fanout":     fanout,
			"profile":    profile,
			"edgeCases":  edgeCases,
			"policy":     pol,
		}
		if impactResult != nil {
			out["impact"] = impactResult
		}

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	}

	// Text output.
	fmt.Println("Terrain Dependency Graph")
	fmt.Println(strings.Repeat("─", 50))
	fmt.Printf("  Nodes: %d    Edges: %d    Density: %.4f\n", stats.NodeCount, stats.EdgeCount, stats.Density)
	fmt.Println()

	// Node breakdown.
	fmt.Println("Node Types:")
	for _, nt := range sortedMapKeys(stats.NodesByType) {
		fmt.Printf("  %-20s %d\n", nt, stats.NodesByType[nt])
	}
	fmt.Println()

	// Coverage summary.
	fmt.Println("Coverage (structural):")
	fmt.Printf("  Sources: %d   High: %d   Medium: %d   Low: %d\n",
		coverage.SourceCount,
		coverage.BandCounts[depgraph.CoverageBandHigh],
		coverage.BandCounts[depgraph.CoverageBandMedium],
		coverage.BandCounts[depgraph.CoverageBandLow])
	fmt.Println()

	// Duplicates summary.
	fmt.Println("Duplicates:")
	fmt.Printf("  Tests analyzed: %d   Duplicates: %d   Clusters: %d\n",
		duplicates.TestsAnalyzed, duplicates.DuplicateCount, len(duplicates.Clusters))
	if duplicates.Skipped {
		fmt.Printf("  Note: %s\n", duplicates.SkipReason)
	}
	if len(duplicates.Clusters) > 0 {
		top := duplicates.Clusters[0]
		fmt.Printf("  Top cluster: %d tests, similarity %.2f\n", len(top.Tests), top.Similarity)
	}
	fmt.Println()

	// Fanout summary.
	fmt.Println("Fanout:")
	fmt.Printf("  Nodes: %d   Flagged: %d   Threshold: %d\n",
		fanout.NodeCount, fanout.FlaggedCount, fanout.Threshold)
	if fanout.Skipped {
		fmt.Printf("  Note: %s\n", fanout.SkipReason)
	}
	if len(fanout.Entries) > 0 {
		top := fanout.Entries[0]
		fmt.Printf("  Highest: %s (transitive: %d)\n", top.NodeID, top.TransitiveFanout)
	}
	fmt.Println()

	// Impact (if requested).
	if impactResult != nil {
		fmt.Println("Impact:")
		fmt.Printf("  Changed files: %d   Impacted tests: %d\n",
			len(impactResult.ChangedFiles), len(impactResult.Tests))
		fmt.Printf("  High: %d   Medium: %d   Low: %d\n",
			impactResult.LevelCounts["high"],
			impactResult.LevelCounts["medium"],
			impactResult.LevelCounts["low"])
		fmt.Println()
	}

	// Profile.
	fmt.Println("Repository Profile:")
	fmt.Printf("  Test Volume:          %s\n", profile.TestVolume)
	fmt.Printf("  CI Pressure:          %s\n", profile.CIPressure)
	fmt.Printf("  Coverage Confidence:  %s\n", profile.CoverageConfidence)
	fmt.Printf("  Redundancy Level:     %s\n", profile.RedundancyLevel)
	fmt.Printf("  Fanout Burden:        %s\n", profile.FanoutBurden)
	fmt.Println()

	// Edge cases and policy.
	if len(edgeCases) > 0 {
		fmt.Println("Edge Cases:")
		for _, ec := range edgeCases {
			fmt.Printf("  [%s] %s: %s\n", ec.Severity, ec.Type, ec.Description)
		}
		fmt.Println()
	}

	if len(pol.Recommendations) > 0 {
		fmt.Println("Recommendations:")
		for _, r := range pol.Recommendations {
			fmt.Printf("  • %s\n", r)
		}
		fmt.Println()
	}

	fmt.Printf("Policy: fallback=%s  confidence=%.2f  optimization=%s\n",
		pol.FallbackLevel, pol.ConfidenceAdjustment,
		map[bool]string{true: "disabled", false: "enabled"}[pol.OptimizationDisabled])

	return nil
}

func sortedMapKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
