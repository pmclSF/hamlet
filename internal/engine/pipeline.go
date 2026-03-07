package engine

import (
	"github.com/pmclSF/hamlet/internal/analysis"
	"github.com/pmclSF/hamlet/internal/measurement"
	"github.com/pmclSF/hamlet/internal/models"
	"github.com/pmclSF/hamlet/internal/ownership"
	"github.com/pmclSF/hamlet/internal/policy"
	"github.com/pmclSF/hamlet/internal/scoring"
)

// PipelineResult holds the output of a full analysis pipeline run.
type PipelineResult struct {
	Snapshot  *models.TestSuiteSnapshot
	HasPolicy bool
}

// RunPipeline executes the full analysis pipeline:
//  1. Static analysis (file discovery, framework detection, code units)
//  2. Signal detection via the detector registry
//  3. Ownership resolution
//  4. Risk scoring
//
// This replaces the duplicated detector invocation across CLI commands.
func RunPipeline(root string) (*PipelineResult, error) {
	// Step 1: Static analysis.
	analyzer := analysis.New(root)
	snapshot, err := analyzer.Analyze()
	if err != nil {
		return nil, err
	}

	// Step 2: Load policy config (needed to configure governance detector).
	policyResult, _ := policy.Load(root)
	hasPolicy := policyResult != nil && policyResult.Found

	var policyCfg *policy.Config
	if hasPolicy {
		policyCfg = policyResult.Config
	}

	// Step 3: Build detector registry and run all detectors.
	registry := DefaultRegistry(Config{
		RepoRoot:     root,
		PolicyConfig: policyCfg,
	})
	registry.Run(snapshot)

	// Step 4: Propagate ownership to signals.
	resolver := ownership.NewResolver(root)
	for i := range snapshot.Signals {
		if snapshot.Signals[i].Owner == "" && snapshot.Signals[i].Location.File != "" {
			snapshot.Signals[i].Owner = resolver.Resolve(snapshot.Signals[i].Location.File)
		}
	}

	// Step 5: Compute risk surfaces from signals.
	snapshot.Risk = scoring.ComputeRisk(snapshot)

	// Step 6: Compute measurement-layer posture.
	measRegistry := measurement.DefaultRegistry()
	measSnap := measRegistry.ComputeSnapshot(snapshot)
	snapshot.Measurements = measSnap.ToModel()

	return &PipelineResult{
		Snapshot:  snapshot,
		HasPolicy: hasPolicy,
	}, nil
}
