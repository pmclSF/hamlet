package stability

import "sort"

// Classify assigns a stability class to each test based on its history.
func Classify(histories []TestHistory) *ClassificationResult {
	result := &ClassificationResult{
		ByClass: make(map[StabilityClass]int),
	}

	maxDepth := 0
	for _, h := range histories {
		if len(h.Observations) > maxDepth {
			maxDepth = len(h.Observations)
		}
	}
	result.HistoryDepth = maxDepth

	for _, h := range histories {
		c := classifyOne(h)
		result.Classifications = append(result.Classifications, c)
		result.ByClass[c.Class]++
	}

	// Sort for determinism.
	sort.Slice(result.Classifications, func(i, j int) bool {
		return result.Classifications[i].TestID < result.Classifications[j].TestID
	})

	return result
}

func classifyOne(h TestHistory) Classification {
	c := Classification{
		TestID:       h.TestID,
		TestName:     h.TestName,
		FilePath:     h.FilePath,
		Owner:        h.Owner,
		HistoryDepth: len(h.Observations),
	}

	presentObs := presentObservations(h.Observations)

	if len(presentObs) < MinHistoryDepth {
		c.Class = ClassDataInsufficient
		c.Confidence = 0.3
		c.Explanation = "fewer than 3 observations available"
		c.RecentTrend = "insufficient"
		return c
	}

	// Check for quarantined/suppressed pattern: mostly skipped.
	skipRate := skippedRate(presentObs)
	if skipRate >= 0.5 {
		c.Class = ClassQuarantinedSuppressed
		c.Confidence = 0.7 + 0.3*skipRate
		c.Explanation = "test skipped in majority of observations"
		c.RecentTrend = trend(presentObs)
		return c
	}

	// Compute failure and flake rates.
	failRate := failureRate(presentObs)
	flakyRate := flakySignalRate(presentObs)
	slowRate := slowSignalRate(presentObs)

	// Check improving: recent observations better than older ones.
	recentTrend := trend(presentObs)
	c.RecentTrend = recentTrend

	if recentTrend == "improving" && (failRate > 0 || flakyRate > 0) {
		// Was bad but getting better.
		c.Class = ClassImproving
		c.Confidence = 0.6
		c.Explanation = "recent observations show improvement over historical pattern"
		return c
	}

	// Chronically flaky: persistent flaky signals across history.
	if flakyRate >= 0.4 {
		c.Class = ClassChronicallyFlaky
		c.Confidence = 0.6 + 0.3*flakyRate
		c.Explanation = "flaky signals present in multiple observations"
		c.RecentTrend = recentTrend
		return c
	}

	// Newly unstable: recently started failing after being stable.
	if isNewlyUnstable(presentObs) {
		c.Class = ClassNewlyUnstable
		c.Confidence = 0.7
		c.Explanation = "test was stable but has recently started failing"
		return c
	}

	// Intermittently slow: slow signals in some but not all observations.
	if slowRate >= 0.3 && slowRate < 1.0 {
		c.Class = ClassIntermittentlySlow
		c.Confidence = 0.6 + 0.2*slowRate
		c.Explanation = "slow test signals present intermittently"
		return c
	}

	// Consistently stable: low failure rate, low flake rate.
	if failRate <= 0.1 && flakyRate <= 0.1 {
		c.Class = ClassConsistentlyStable
		c.Confidence = 0.7 + 0.3*(1-failRate)
		c.Explanation = "test has been consistently passing"
		return c
	}

	// Default: data insufficient for clear classification.
	c.Class = ClassDataInsufficient
	c.Confidence = 0.4
	c.Explanation = "mixed signals, cannot determine clear stability pattern"
	return c
}

func presentObservations(obs []Observation) []Observation {
	var result []Observation
	for _, o := range obs {
		if o.Present {
			result = append(result, o)
		}
	}
	return result
}

func failureRate(obs []Observation) float64 {
	if len(obs) == 0 {
		return 0
	}
	count := 0
	for _, o := range obs {
		if o.Failed {
			count++
		}
	}
	return float64(count) / float64(len(obs))
}

func skippedRate(obs []Observation) float64 {
	if len(obs) == 0 {
		return 0
	}
	count := 0
	for _, o := range obs {
		if o.Skipped {
			count++
		}
	}
	return float64(count) / float64(len(obs))
}

func flakySignalRate(obs []Observation) float64 {
	if len(obs) == 0 {
		return 0
	}
	count := 0
	for _, o := range obs {
		if o.FlakySignal {
			count++
		}
	}
	return float64(count) / float64(len(obs))
}

func slowSignalRate(obs []Observation) float64 {
	if len(obs) == 0 {
		return 0
	}
	count := 0
	for _, o := range obs {
		if o.SlowSignal {
			count++
		}
	}
	return float64(count) / float64(len(obs))
}

// trend determines if the test is improving, worsening, or stable
// by comparing the first half of observations to the second half.
func trend(obs []Observation) string {
	if len(obs) < MinHistoryDepth {
		return "insufficient"
	}

	mid := len(obs) / 2
	earlyFails := failureRate(obs[:mid])
	lateFails := failureRate(obs[mid:])
	earlyFlaky := flakySignalRate(obs[:mid])
	lateFlaky := flakySignalRate(obs[mid:])

	earlyProblems := earlyFails + earlyFlaky
	lateProblems := lateFails + lateFlaky

	diff := lateProblems - earlyProblems
	switch {
	case diff < -0.2:
		return "improving"
	case diff > 0.2:
		return "worsening"
	default:
		return "stable"
	}
}

// isNewlyUnstable returns true if the test was stable in earlier observations
// but has started failing recently.
func isNewlyUnstable(obs []Observation) bool {
	if len(obs) < MinHistoryDepth {
		return false
	}

	// Check that the first 60% were mostly passing.
	earlyEnd := len(obs) * 6 / 10
	if earlyEnd < 2 {
		earlyEnd = 2
	}
	earlyFails := failureRate(obs[:earlyEnd])
	if earlyFails > 0.1 {
		return false
	}

	// Check that the last 40% have failures.
	lateFails := failureRate(obs[earlyEnd:])
	return lateFails >= 0.3
}
