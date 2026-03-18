package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pmclSF/terrain/internal/analyze"
	"github.com/pmclSF/terrain/internal/changescope"
	"github.com/pmclSF/terrain/internal/depgraph"
	"github.com/pmclSF/terrain/internal/engine"
	"github.com/pmclSF/terrain/internal/impact"
	"github.com/pmclSF/terrain/internal/metrics"
	"github.com/pmclSF/terrain/internal/reporting"
)

func runImpact(root, baseRef string, jsonOutput bool, show, ownerFilter string) error {
	result, err := engine.RunPipeline(root, defaultPipelineOptionsWithProgress(jsonOutput))
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	cs, err := impact.ChangeSetFromGitDiff(absRoot, baseRef)
	if err != nil {
		return fmt.Errorf("failed to determine changed files: %w", err)
	}

	impactResult := impact.AnalyzeChangeSet(cs, result.Snapshot)

	// Apply edge-case policy to adjust confidence and add warnings.
	snapshot := result.Snapshot
	dg := depgraph.Build(snapshot)
	dgCov := depgraph.AnalyzeCoverage(dg)
	dgDupes := depgraph.DetectDuplicates(dg)
	dgFanout := depgraph.AnalyzeFanout(dg, depgraph.DefaultFanoutThreshold)
	ms := metrics.Derive(snapshot)
	pi := depgraph.ProfileInsights{
		Coverage:   &dgCov,
		Duplicates: &dgDupes,
		Fanout:     &dgFanout,
		Snapshot:   analyze.BuildSnapshotProfileData(snapshot),
	}
	dgProfile := depgraph.AnalyzeProfile(dg, pi)
	depgraph.EnrichProfileWithHealthRatios(&dgProfile, ms.Health.SkippedTestRatio, ms.Health.FlakyTestRatio)
	dgEdgeCases := depgraph.DetectEdgeCases(dgProfile, dg, pi)
	if len(dgEdgeCases) > 0 {
		dgPolicy := depgraph.ApplyEdgeCasePolicy(dgEdgeCases, dgProfile)
		impactResult.ApplyEdgeCasePolicy(dgPolicy.ConfidenceAdjustment, dgPolicy.RiskElevated, dgPolicy.Recommendations)
	}

	// Apply manual coverage overlay to annotate protection gaps.
	if len(snapshot.ManualCoverage) > 0 {
		impactResult.ApplyManualCoverageOverlay(snapshot.ManualCoverage)
	}

	// Apply owner filter if specified.
	if ownerFilter != "" {
		impactResult = impact.FilterByOwner(impactResult, ownerFilter)
	}

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(impactResult)
	}

	switch show {
	case "units":
		reporting.RenderImpactUnits(os.Stdout, impactResult)
	case "gaps":
		reporting.RenderImpactGaps(os.Stdout, impactResult)
	case "tests":
		reporting.RenderImpactTests(os.Stdout, impactResult)
	case "owners":
		reporting.RenderImpactOwners(os.Stdout, impactResult)
	case "graph":
		reporting.RenderImpactGraph(os.Stdout, impactResult)
	case "selected":
		reporting.RenderProtectiveSet(os.Stdout, impactResult)
	case "":
		reporting.RenderImpactReport(os.Stdout, impactResult)
	default:
		return fmt.Errorf("unknown --show value: %q (valid: units, gaps, tests, owners, graph, selected)", show)
	}
	return nil
}

// runSelectTests performs impact analysis and outputs the protective test set.
func runSelectTests(root, baseRef string, jsonOutput bool) error {
	result, err := engine.RunPipeline(root, defaultPipelineOptions())
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	cs, err := impact.ChangeSetFromGitDiff(absRoot, baseRef)
	if err != nil {
		return fmt.Errorf("failed to determine changed files: %w", err)
	}

	impactResult := impact.AnalyzeChangeSet(cs, result.Snapshot)

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(impactResult.ProtectiveSet)
	}

	reporting.RenderProtectiveSet(os.Stdout, impactResult)
	return nil
}


func runPR(root, baseRef string, jsonOutput bool, format string) error {
	result, err := engine.RunPipeline(root, defaultPipelineOptions())
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	cs, err := impact.ChangeSetFromGitDiff(absRoot, baseRef)
	if err != nil {
		return fmt.Errorf("failed to determine changed files: %w", err)
	}

	pr := changescope.AnalyzePRFromChangeSet(cs, result.Snapshot)

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(pr)
	}

	switch format {
	case "markdown", "md":
		changescope.RenderPRSummaryMarkdown(os.Stdout, pr)
	case "comment":
		changescope.RenderPRCommentConcise(os.Stdout, pr)
	case "annotation", "ci":
		changescope.RenderCIAnnotation(os.Stdout, pr)
	default:
		changescope.RenderChangeScopedReport(os.Stdout, pr)
	}
	return nil
}

