package runtime

import "testing"

func TestApplyToTestFiles_PopulatesVariance(t *testing.T) {
	t.Parallel()
	results := []TestResult{
		{File: "a.test.js", DurationMs: 100, Status: StatusPassed},
		{File: "a.test.js", DurationMs: 200, Status: StatusPassed},
		{File: "a.test.js", DurationMs: 300, Status: StatusFailed, Retried: true},
	}
	updates := []TestFileUpdate{{Path: "a.test.js"}}

	ApplyToTestFiles(results, updates)
	u := updates[0]
	if u.AvgRuntimeMs != 200 {
		t.Fatalf("AvgRuntimeMs = %f, want 200", u.AvgRuntimeMs)
	}
	if u.RuntimeVariance <= 0 {
		t.Fatalf("RuntimeVariance = %f, want > 0", u.RuntimeVariance)
	}
	if u.RetryRate <= 0 {
		t.Fatalf("RetryRate = %f, want > 0", u.RetryRate)
	}
}

func TestVariance(t *testing.T) {
	t.Parallel()
	if variance(nil) != 0 {
		t.Fatalf("variance(nil) should be 0")
	}
	if variance([]float64{10}) != 0 {
		t.Fatalf("variance(single) should be 0")
	}
	v := variance([]float64{1, 2, 3})
	if v <= 0.6 || v >= 0.8 {
		t.Fatalf("variance([1,2,3]) = %f, want around 0.666", v)
	}
}
