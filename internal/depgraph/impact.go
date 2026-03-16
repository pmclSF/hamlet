package depgraph

import (
	"math"
	"sort"
)

// ImpactResult contains the impact analysis for a set of changed files.
type ImpactResult struct {
	// Changed files that were analyzed.
	ChangedFiles []string `json:"changedFiles"`

	// All impacted tests, sorted by confidence descending.
	Tests []ImpactedTest `json:"tests"`

	// Unique test IDs selected for re-run.
	SelectedTests []string `json:"selectedTests"`

	// Summary counts by confidence level.
	LevelCounts map[string]int `json:"levelCounts"`
}

// ImpactedTest represents a test affected by a change.
type ImpactedTest struct {
	// Test node ID.
	TestID string `json:"testId"`

	// Composite confidence score (0–1).
	Confidence float64 `json:"confidence"`

	// Classification based on confidence.
	Level string `json:"level"`

	// The changed file that triggered this impact.
	ChangedFile string `json:"changedFile"`

	// Full chain of edges from changed file to this test.
	ReasonChain []ReasonStep `json:"reasonChain"`
}

// ReasonStep describes one hop in the impact propagation path.
type ReasonStep struct {
	From           string  `json:"from"`
	To             string  `json:"to"`
	EdgeType       string  `json:"edgeType"`
	EdgeConfidence float64 `json:"edgeConfidence"`
}

// Impact confidence thresholds.
const (
	highConfidence   = 0.7
	mediumConfidence = 0.4
	minConfidence    = 0.1
	lengthDecay      = 0.85
	fanoutThreshold  = 5
	maxImpactDepth   = 20
)

// pathState tracks BFS traversal state for impact propagation.
type pathState struct {
	nodeID     string
	confidence float64
	chain      []ReasonStep
	depth      int
}

// AnalyzeImpact computes which tests are impacted by changes to the given
// files, using BFS with confidence decay through the graph.
//
// Algorithm:
//  1. For each changed file, find its node in the graph
//  2. BFS via reverse edges (incoming), decaying confidence at each hop:
//     - Multiply by edge confidence
//     - Multiply by length decay (0.85^depth)
//     - Divide by log2(fanout+1) if fanout > 5
//  3. Stop at test nodes (record them) or when confidence < 0.1
//  4. Deduplicate tests, keeping the highest confidence path
func AnalyzeImpact(g *Graph, changedFiles []string) ImpactResult {
	result := ImpactResult{
		ChangedFiles: changedFiles,
		LevelCounts:  map[string]int{},
	}

	if len(changedFiles) == 0 {
		return result
	}

	// Out-degree is computed lazily via g.OutDegree() with caching.

	// Track best confidence per test ID.
	bestTests := map[string]*ImpactedTest{}

	for _, changedFile := range changedFiles {
		fileID := "file:" + changedFile
		if g.Node(fileID) == nil {
			continue
		}

		// BFS via reverse edges. Uses index-based queue for O(1) dequeue.
		visited := map[string]float64{}
		queue := []pathState{{
			nodeID:     fileID,
			confidence: 1.0,
			depth:      0,
		}}
		head := 0

		for head < len(queue) {
			current := queue[head]
			head++

			// Skip if we've seen this node with higher confidence.
			if prev, ok := visited[current.nodeID]; ok && prev >= current.confidence {
				continue
			}
			visited[current.nodeID] = current.confidence

			node := g.Node(current.nodeID)
			if node == nil {
				continue
			}

			// If this is a test node, record it.
			if node.Type == NodeTest {
				existing, exists := bestTests[current.nodeID]
				if !exists || current.confidence > existing.Confidence {
					bestTests[current.nodeID] = &ImpactedTest{
						TestID:      current.nodeID,
						Confidence:  current.confidence,
						Level:       classifyLevel(current.confidence),
						ChangedFile: changedFile,
						ReasonChain: current.chain,
					}
				}
				continue // Don't traverse past test nodes.
			}

			// If this is a test file, collect its tests.
			if node.Type == NodeTestFile {
				for _, e := range g.Incoming(current.nodeID) {
					if e.Type == EdgeTestDefinedInFile {
						testNode := g.Node(e.From)
						if testNode != nil && testNode.Type == NodeTest {
							chain := append(cloneChain(current.chain), ReasonStep{
								From:           current.nodeID,
								To:             e.From,
								EdgeType:       string(e.Type),
								EdgeConfidence: e.Confidence,
							})
							existing, exists := bestTests[e.From]
							if !exists || current.confidence > existing.Confidence {
								bestTests[e.From] = &ImpactedTest{
									TestID:      e.From,
									Confidence:  current.confidence,
									Level:       classifyLevel(current.confidence),
									ChangedFile: changedFile,
									ReasonChain: chain,
								}
							}
						}
					}
				}
			}

			// Depth cap: do not enqueue further nodes beyond maxImpactDepth.
			if current.depth >= maxImpactDepth {
				continue
			}

			// Traverse reverse edges to find dependents.
			for _, e := range g.Incoming(current.nodeID) {
				// Skip TestDefinedInFile edges already handled above.
				if e.Type == EdgeTestDefinedInFile {
					continue
				}

				newConf := current.confidence * e.Confidence
				newConf *= lengthDecay

				// Fanout penalty.
				if od := g.OutDegree(e.From); od > fanoutThreshold {
					newConf /= math.Log2(float64(od) + 1)
				}

				if newConf < minConfidence {
					continue
				}

				chain := append(cloneChain(current.chain), ReasonStep{
					From:           current.nodeID,
					To:             e.From,
					EdgeType:       string(e.Type),
					EdgeConfidence: e.Confidence,
				})

				queue = append(queue, pathState{
					nodeID:     e.From,
					confidence: newConf,
					chain:      chain,
					depth:      current.depth + 1,
				})
			}
		}
	}

	// Convert to sorted list.
	tests := make([]ImpactedTest, 0, len(bestTests))
	selected := make([]string, 0, len(bestTests))
	for _, t := range bestTests {
		tests = append(tests, *t)
		selected = append(selected, t.TestID)
	}

	sort.Slice(tests, func(i, j int) bool {
		if math.Abs(tests[i].Confidence-tests[j].Confidence) > 1e-9 {
			return tests[i].Confidence > tests[j].Confidence
		}
		return tests[i].TestID < tests[j].TestID
	})
	sort.Strings(selected)

	for _, t := range tests {
		result.LevelCounts[t.Level]++
	}

	result.Tests = tests
	result.SelectedTests = selected

	return result
}

func classifyLevel(confidence float64) string {
	if confidence >= highConfidence {
		return "high"
	}
	if confidence >= mediumConfidence {
		return "medium"
	}
	return "low"
}

func cloneChain(chain []ReasonStep) []ReasonStep {
	if len(chain) == 0 {
		return []ReasonStep{}
	}
	out := make([]ReasonStep, len(chain))
	copy(out, chain)
	return out
}
