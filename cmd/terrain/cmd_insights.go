package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pmclSF/terrain/internal/analyze"
	"github.com/pmclSF/terrain/internal/benchmark"
	"github.com/pmclSF/terrain/internal/comparison"
	"github.com/pmclSF/terrain/internal/depgraph"
	"github.com/pmclSF/terrain/internal/engine"
	"github.com/pmclSF/terrain/internal/graph"
	"github.com/pmclSF/terrain/internal/heatmap"
	"github.com/pmclSF/terrain/internal/insights"
	"github.com/pmclSF/terrain/internal/matrix"
	"github.com/pmclSF/terrain/internal/metrics"
	"github.com/pmclSF/terrain/internal/reporting"
	"github.com/pmclSF/terrain/internal/stability"
	"github.com/pmclSF/terrain/internal/summary"
)

func runPortfolio(root string, jsonOutput bool) error {
	result, err := engine.RunPipeline(root, defaultPipelineOptions())
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result.Snapshot.Portfolio)
	}

	reporting.RenderPortfolioReport(os.Stdout, result.Snapshot)
	return nil
}

// runPosture performs analysis and outputs a detailed posture breakdown.
func runPosture(root string, jsonOutput bool) error {
	result, err := engine.RunPipeline(root, defaultPipelineOptionsWithProgress(jsonOutput))
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result.Snapshot.Measurements)
	}

	reporting.RenderPostureReport(os.Stdout, result.Snapshot)
	return nil
}

// runMetrics performs analysis and outputs aggregate metrics.
func runMetrics(root string, jsonOutput bool) error {
	result, err := engine.RunPipeline(root, defaultPipelineOptions())
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	ms := metrics.Derive(result.Snapshot)

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(ms)
	}

	reporting.RenderMetricsReport(os.Stdout, ms)
	return nil
}

// runSummary performs analysis and outputs an executive summary with
// trend highlights (if prior snapshots exist) and benchmark readiness.
func runSummary(root string, jsonOutput bool) error {
	result, err := engine.RunPipeline(root, defaultPipelineOptionsWithProgress(jsonOutput))
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}
	snapshot := result.Snapshot

	// Build graph, heatmap (enriched with graph data), and derive metrics.
	g := graph.Build(snapshot)
	h := heatmap.BuildWithGraph(snapshot, g)
	ms := metrics.Derive(snapshot)

	// Attempt to load prior snapshot for trend comparison.
	var comp *comparison.SnapshotComparison
	absRoot, _ := filepath.Abs(root)
	snapDir := filepath.Join(absRoot, ".terrain", "snapshots")
	latest, previous, snapErr := findRecentSnapshots(snapDir)
	if snapErr == nil && latest != "" && previous != "" {
		fromSnap, err1 := loadSnapshot(previous)
		toSnap, err2 := loadSnapshot(latest)
		if err1 == nil && err2 == nil {
			comp = comparison.Compare(fromSnap, toSnap)
		}
	}

	// Build benchmark segment.
	seg := &benchmark.BuildExport(snapshot, ms, result.HasPolicy).Segment

	// Build executive summary.
	es := summary.Build(&summary.BuildInput{
		Snapshot:   snapshot,
		Heatmap:    h,
		Metrics:    ms,
		Comparison: comp,
		Segment:    seg,
		HasPolicy:  result.HasPolicy,
	})

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(es)
	}

	reporting.RenderExecutiveSummary(os.Stdout, es)
	return nil
}

// runFocus performs analysis and emits a compact action-first view.
func runFocus(root string, jsonOutput bool) error {
	result, err := engine.RunPipeline(root, defaultPipelineOptionsWithProgress(jsonOutput))
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}
	snapshot := result.Snapshot

	g := graph.Build(snapshot)
	h := heatmap.BuildWithGraph(snapshot, g)
	ms := metrics.Derive(snapshot)
	seg := &benchmark.BuildExport(snapshot, ms, result.HasPolicy).Segment

	es := summary.Build(&summary.BuildInput{
		Snapshot:  snapshot,
		Heatmap:   h,
		Metrics:   ms,
		Segment:   seg,
		HasPolicy: result.HasPolicy,
	})

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(map[string]any{
			"recommendedFocus": es.RecommendedFocus,
			"topRiskAreas":     es.TopRiskAreas,
			"recommendations":  es.Recommendations,
			"posture":          es.Posture,
		})
	}

	fmt.Println("Terrain Focus")
	fmt.Println()
	if es.RecommendedFocus != "" {
		fmt.Printf("Now: %s\n", es.RecommendedFocus)
	} else {
		fmt.Println("Now: No immediate focus area detected.")
	}

	if len(es.TopRiskAreas) > 0 {
		fmt.Println()
		fmt.Println("Top Risk Areas")
		for i, area := range es.TopRiskAreas {
			fmt.Printf("  %d. %s (%s)\n", i+1, area.Name, area.Band)
			if area.RiskType != "" {
				fmt.Printf("     risk: %s (%d signal(s))\n", area.RiskType, area.SignalCount)
			}
		}
	}

	if len(es.Recommendations) > 0 {
		fmt.Println()
		fmt.Println("Recommended Actions")
		for i, r := range es.Recommendations {
			fmt.Printf("  %d. %s\n", i+1, r.What)
			if r.Why != "" {
				fmt.Printf("     why: %s\n", r.Why)
			}
			if r.Where != "" {
				fmt.Printf("     where: %s\n", r.Where)
			}
		}
	}

	fmt.Println()
	fmt.Println("Next: terrain posture    see detailed evidence by dimension")
	return nil
}

// runInsights aggregates all insight engines into a single actionable report.
// It combines executive summary, depgraph profile, and portfolio findings.
func runInsights(root string, jsonOutput bool) error {
	result, err := engine.RunPipeline(root, defaultPipelineOptionsWithProgress(jsonOutput))
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}
	snapshot := result.Snapshot

	// Derive metrics for health ratios.
	ms := metrics.Derive(snapshot)

	// Build depgraph insights with a preflight guard for very large repos.
	const maxInsightsGraphNodes = 150000
	estimatedGraphNodes := len(snapshot.TestCases) + len(snapshot.CodeUnits) + len(snapshot.TestFiles)

	input := &insights.BuildInput{
		Snapshot:  snapshot,
		HasPolicy: result.HasPolicy,
	}

	if estimatedGraphNodes > maxInsightsGraphNodes {
		input.DepgraphSkipped = true
		input.DepgraphSkipReason = fmt.Sprintf(
			"depgraph analysis skipped for estimated graph size %d (limit %d)",
			estimatedGraphNodes, maxInsightsGraphNodes,
		)
		input.Duplicates = depgraph.DuplicateResult{
			Skipped:       true,
			SkipReason:    input.DepgraphSkipReason,
			TestsAnalyzed: len(snapshot.TestCases),
		}
		input.Fanout = depgraph.FanoutResult{
			Skipped:      true,
			SkipReason:   input.DepgraphSkipReason,
			NodeCount:    estimatedGraphNodes,
			Threshold:    depgraph.DefaultFanoutThreshold,
			FlaggedCount: 0,
		}
		input.Profile = depgraph.RepoProfile{
			TestVolume:         "large",
			CIPressure:         "high",
			CoverageConfidence: "low",
			RedundancyLevel:    "low",
			FanoutBurden:       "low",
			SkipBurden:         "low",
			FlakeBurden:        "low",
		}
		input.Policy = depgraph.Policy{
			FallbackLevel:        depgraph.FallbackSmokeRegression,
			ConfidenceAdjustment: 0.6,
			RiskElevated:         true,
			Recommendations: []string{
				"Depgraph analysis skipped due to repository scale; narrow scope for full dependency insights.",
			},
		}
		input.EdgeCases = []depgraph.EdgeCase{{
			Type:        depgraph.EdgeCaseLowGraphVisibility,
			Severity:    "warning",
			Description: input.DepgraphSkipReason,
		}}
		input.Coverage = depgraph.CoverageResult{BandCounts: map[depgraph.CoverageBand]int{}}
	} else {
		dg := depgraph.Build(snapshot)
		input.Coverage = depgraph.AnalyzeCoverage(dg)
		input.Duplicates = depgraph.DetectDuplicates(dg)
		input.Fanout = depgraph.AnalyzeFanout(dg, depgraph.DefaultFanoutThreshold)
		dgInsights := depgraph.ProfileInsights{
			Coverage:   &input.Coverage,
			Duplicates: &input.Duplicates,
			Fanout:     &input.Fanout,
			Snapshot:   analyze.BuildSnapshotProfileData(snapshot),
		}
		input.Profile = depgraph.AnalyzeProfile(dg, dgInsights)
		depgraph.EnrichProfileWithHealthRatios(&input.Profile, ms.Health.SkippedTestRatio, ms.Health.FlakyTestRatio)
		input.EdgeCases = depgraph.DetectEdgeCases(input.Profile, dg, dgInsights)
		input.Policy = depgraph.ApplyEdgeCasePolicy(input.EdgeCases, input.Profile)
		dgRedundancy := depgraph.AnalyzeRedundancy(dg)
		if len(dgRedundancy.Clusters) > 0 {
			input.BehaviorRedundancy = &dgRedundancy
		}
		input.StabilityClusters = stability.DetectClusters(dg, snapshot.Signals)
		matrixResult := matrix.Analyze(dg)
		if len(matrixResult.Classes) > 0 {
			input.MatrixCoverage = matrixResult
		}
	}

	report := insights.Build(input)

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(report)
	}

	reporting.RenderInsightsReport(os.Stdout, report)
	return nil
}

