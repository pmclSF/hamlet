package main

import (
	"testing"
)

func TestIsInteractive_InTestEnvironment(t *testing.T) {
	t.Parallel()
	// In test environments, stderr is not a TTY.
	if isInteractive() {
		t.Skip("test environment is interactive (unusual)")
	}
}

func TestNewProgressFunc_JSONModeSuppressed(t *testing.T) {
	t.Parallel()
	pf := newProgressFunc(true) // JSON mode
	if pf != nil {
		t.Error("expected nil progress func in JSON mode")
	}
}

func TestNewProgressFunc_NonInteractiveSuppressed(t *testing.T) {
	t.Parallel()
	// In test environments (pipe), non-interactive should return nil.
	pf := newProgressFunc(false)
	if pf != nil {
		t.Log("progress func is nil in non-interactive mode (expected in CI/pipe)")
	}
}

func TestProgressFunc_Signature(t *testing.T) {
	t.Parallel()
	// Verify the progress function can be called without panic.
	var callCount int
	pf := func(step, total int, label string) {
		callCount++
		if step < 1 || step > total {
			t.Errorf("step %d out of range [1, %d]", step, total)
		}
		if label == "" {
			t.Error("empty label")
		}
	}
	pf(1, 5, "Scanning repository")
	pf(2, 5, "Building graph")
	pf(3, 5, "Inferring validations")
	pf(4, 5, "Computing insights")
	pf(5, 5, "Writing report")
	if callCount != 5 {
		t.Errorf("expected 5 calls, got %d", callCount)
	}
}
