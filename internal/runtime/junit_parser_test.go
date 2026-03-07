package runtime

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseJUnitXML_Basic(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<testsuites>
  <testsuite name="auth" tests="3" failures="1" errors="0" skipped="1" time="2.5">
    <testcase name="login succeeds" classname="auth.LoginTest" time="1.2"/>
    <testcase name="login fails on bad password" classname="auth.LoginTest" time="0.8">
      <failure message="expected 401">assertion error</failure>
    </testcase>
    <testcase name="signup disabled" classname="auth.SignupTest" time="0.0">
      <skipped message="feature flag off"/>
    </testcase>
  </testsuite>
</testsuites>`

	path := writeTempFile(t, "results.xml", xml)
	result, err := ParseJUnitXML(path)
	if err != nil {
		t.Fatalf("ParseJUnitXML failed: %v", err)
	}

	if result.Format != "junit-xml" {
		t.Errorf("format = %q, want junit-xml", result.Format)
	}
	if len(result.Results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(result.Results))
	}

	// Check passed test.
	r0 := result.Results[0]
	if r0.Name != "login succeeds" {
		t.Errorf("name = %q", r0.Name)
	}
	if r0.Status != StatusPassed {
		t.Errorf("status = %q, want passed", r0.Status)
	}
	if r0.DurationMs != 1200 {
		t.Errorf("durationMs = %f, want 1200", r0.DurationMs)
	}
	if r0.Suite != "auth" {
		t.Errorf("suite = %q, want auth", r0.Suite)
	}

	// Check failed test.
	r1 := result.Results[1]
	if r1.Status != StatusFailed {
		t.Errorf("status = %q, want failed", r1.Status)
	}
	if r1.Message != "expected 401" {
		t.Errorf("message = %q", r1.Message)
	}

	// Check skipped test.
	r2 := result.Results[2]
	if r2.Status != StatusSkipped {
		t.Errorf("status = %q, want skipped", r2.Status)
	}
}

func TestParseJUnitXML_BareSuite(t *testing.T) {
	xml := `<testsuite name="bare" tests="1">
  <testcase name="it works" time="0.5"/>
</testsuite>`

	path := writeTempFile(t, "bare.xml", xml)
	result, err := ParseJUnitXML(path)
	if err != nil {
		t.Fatalf("ParseJUnitXML failed: %v", err)
	}
	if len(result.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Results))
	}
	if result.Results[0].DurationMs != 500 {
		t.Errorf("durationMs = %f, want 500", result.Results[0].DurationMs)
	}
}

func TestParseJUnitXML_RetryDetection(t *testing.T) {
	xml := `<testsuites>
  <testsuite name="retry" tests="3">
    <testcase name="flaky test" classname="Retry" time="0.1">
      <failure message="timeout"/>
    </testcase>
    <testcase name="flaky test" classname="Retry" time="0.2"/>
    <testcase name="stable test" classname="Retry" time="0.1"/>
  </testsuite>
</testsuites>`

	path := writeTempFile(t, "retry.xml", xml)
	result, err := ParseJUnitXML(path)
	if err != nil {
		t.Fatalf("ParseJUnitXML failed: %v", err)
	}

	if len(result.Results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(result.Results))
	}

	// First occurrence should not be marked as retried.
	if result.Results[0].Retried {
		t.Error("first occurrence should not be retried")
	}
	// Second occurrence of same name should be marked as retried.
	if !result.Results[1].Retried {
		t.Error("second occurrence should be marked as retried")
	}
	if result.Results[1].RetryAttempt != 1 {
		t.Errorf("retryAttempt = %d, want 1", result.Results[1].RetryAttempt)
	}
}

func writeTempFile(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
