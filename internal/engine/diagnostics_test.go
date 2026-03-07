package engine

import (
	"bytes"
	"strings"
	"testing"
)

func TestPipelineDiagnostics_Enabled(t *testing.T) {
	result, err := RunPipeline("../analysis/testdata/sample-repo", PipelineOptions{
		CollectDiagnostics: true,
	})
	if err != nil {
		t.Fatalf("RunPipeline failed: %v", err)
	}

	diag := result.Diagnostics
	if diag == nil {
		t.Fatal("expected diagnostics when CollectDiagnostics=true")
	}

	if len(diag.Steps) < 4 {
		t.Errorf("expected at least 4 steps, got %d", len(diag.Steps))
	}

	if diag.Total <= 0 {
		t.Error("expected positive total duration")
	}

	// Verify expected step names are present.
	names := map[string]bool{}
	for _, s := range diag.Steps {
		names[s.Name] = true
	}
	for _, expected := range []string{"static-analysis", "signal-detection", "risk-scoring", "measurement"} {
		if !names[expected] {
			t.Errorf("missing diagnostics step: %s", expected)
		}
	}
}

func TestPipelineDiagnostics_Disabled(t *testing.T) {
	result, err := RunPipeline("../analysis/testdata/sample-repo")
	if err != nil {
		t.Fatalf("RunPipeline failed: %v", err)
	}

	if result.Diagnostics != nil {
		t.Error("expected nil diagnostics when CollectDiagnostics=false")
	}
}

func TestPipelineDiagnostics_Render(t *testing.T) {
	result, err := RunPipeline("../analysis/testdata/sample-repo", PipelineOptions{
		CollectDiagnostics: true,
	})
	if err != nil {
		t.Fatalf("RunPipeline failed: %v", err)
	}

	var buf bytes.Buffer
	result.Diagnostics.Render(&buf)
	output := buf.String()

	if !strings.Contains(output, "Pipeline Diagnostics") {
		t.Error("expected header in rendered output")
	}
	if !strings.Contains(output, "static-analysis") {
		t.Error("expected static-analysis step in rendered output")
	}
	if !strings.Contains(output, "TOTAL") {
		t.Error("expected TOTAL line in rendered output")
	}
}
