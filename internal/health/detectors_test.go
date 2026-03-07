package health

import (
	"testing"

	"github.com/pmclSF/hamlet/internal/runtime"
)

func TestSlowTestDetector_OverThreshold(t *testing.T) {
	d := &SlowTestDetector{ThresholdMs: 1000}
	results := []runtime.TestResult{
		{Name: "fast test", File: "fast.test.js", DurationMs: 500, Status: runtime.StatusPassed},
		{Name: "slow test", File: "slow.test.js", DurationMs: 3000, Status: runtime.StatusPassed},
		{Name: "very slow test", File: "very-slow.test.js", DurationMs: 12000, Status: runtime.StatusPassed},
	}

	signals := d.Detect(results)

	if len(signals) != 2 {
		t.Fatalf("expected 2 slow signals, got %d", len(signals))
	}

	// First signal should be for 3000ms test (medium severity, 3x threshold).
	if signals[0].Location.File != "slow.test.js" {
		t.Errorf("file = %q, want slow.test.js", signals[0].Location.File)
	}
	if signals[0].Severity != "medium" {
		t.Errorf("severity = %q, want medium", signals[0].Severity)
	}

	// Second signal should be for 12000ms test (high severity, 12x threshold).
	if signals[1].Severity != "high" {
		t.Errorf("severity = %q, want high", signals[1].Severity)
	}
}

func TestSlowTestDetector_UnderThreshold(t *testing.T) {
	d := &SlowTestDetector{ThresholdMs: 5000}
	results := []runtime.TestResult{
		{Name: "fast", DurationMs: 100, Status: runtime.StatusPassed},
		{Name: "normal", DurationMs: 2000, Status: runtime.StatusPassed},
	}

	signals := d.Detect(results)
	if len(signals) != 0 {
		t.Errorf("expected 0 signals for under-threshold tests, got %d", len(signals))
	}
}

func TestSlowTestDetector_SkipsSkipped(t *testing.T) {
	d := &SlowTestDetector{ThresholdMs: 100}
	results := []runtime.TestResult{
		{Name: "skipped slow", DurationMs: 99999, Status: runtime.StatusSkipped},
	}

	signals := d.Detect(results)
	if len(signals) != 0 {
		t.Errorf("should not flag skipped tests as slow, got %d", len(signals))
	}
}

func TestFlakyTestDetector_RetryEvidence(t *testing.T) {
	d := &FlakyTestDetector{}
	results := []runtime.TestResult{
		{Name: "stable", File: "a.test.js", Status: runtime.StatusPassed},
		{Name: "flaky", File: "b.test.js", Status: runtime.StatusPassed, Retried: true, RetryAttempt: 1},
	}

	signals := d.Detect(results)

	found := false
	for _, s := range signals {
		if s.Type == "flakyTest" && s.Location.Symbol == "flaky" {
			found = true
		}
	}
	if !found {
		t.Error("expected flakyTest signal for retried test")
	}
}

func TestFlakyTestDetector_MixedOutcomes(t *testing.T) {
	d := &FlakyTestDetector{}
	results := []runtime.TestResult{
		{Name: "intermittent", Suite: "suite", File: "c.test.js", Status: runtime.StatusFailed},
		{Name: "intermittent", Suite: "suite", File: "c.test.js", Status: runtime.StatusPassed},
	}

	signals := d.Detect(results)

	found := false
	for _, s := range signals {
		if s.Type == "flakyTest" {
			found = true
		}
	}
	if !found {
		t.Error("expected flakyTest signal for mixed pass/fail outcomes")
	}
}

func TestFlakyTestDetector_NoEvidence(t *testing.T) {
	d := &FlakyTestDetector{}
	results := []runtime.TestResult{
		{Name: "stable1", Status: runtime.StatusPassed},
		{Name: "stable2", Status: runtime.StatusPassed},
	}

	signals := d.Detect(results)
	if len(signals) != 0 {
		t.Errorf("expected 0 flaky signals without evidence, got %d", len(signals))
	}
}

func TestSkippedTestDetector_SomeSkipped(t *testing.T) {
	d := &SkippedTestDetector{}
	results := []runtime.TestResult{
		{Name: "t1", Status: runtime.StatusPassed},
		{Name: "t2", Status: runtime.StatusPassed},
		{Name: "t3", Status: runtime.StatusSkipped},
		{Name: "t4", Status: runtime.StatusPassed},
		{Name: "t5", Status: runtime.StatusSkipped},
	}

	signals := d.Detect(results)
	if len(signals) != 1 {
		t.Fatalf("expected 1 skipped signal, got %d", len(signals))
	}
	if signals[0].Type != "skippedTest" {
		t.Errorf("type = %q", signals[0].Type)
	}
	// 2/5 = 40% → medium severity
	if signals[0].Severity != "medium" {
		t.Errorf("severity = %q, want medium for 40%% skip rate", signals[0].Severity)
	}
}

func TestSkippedTestDetector_NoneSkipped(t *testing.T) {
	d := &SkippedTestDetector{}
	results := []runtime.TestResult{
		{Name: "t1", Status: runtime.StatusPassed},
		{Name: "t2", Status: runtime.StatusFailed},
	}

	signals := d.Detect(results)
	if len(signals) != 0 {
		t.Errorf("expected 0 signals when no skips, got %d", len(signals))
	}
}
