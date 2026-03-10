package runtime

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestParseJUnitXML_Basic(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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

func TestParseJUnitXML_DuplicatePassedCasesNotMarkedRetry(t *testing.T) {
	t.Parallel()
	xml := `<testsuites>
  <testsuite name="parallel" tests="3">
    <testcase name="shared name" classname="ShardA" file="a_test.py" time="0.1"/>
    <testcase name="shared name" classname="ShardB" file="b_test.py" time="0.1"/>
    <testcase name="shared name" classname="ShardA" file="a_test.py" time="0.1"/>
  </testsuite>
</testsuites>`

	path := writeTempFile(t, "parallel.xml", xml)
	result, err := ParseJUnitXML(path)
	if err != nil {
		t.Fatalf("ParseJUnitXML failed: %v", err)
	}

	if len(result.Results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(result.Results))
	}

	if result.Results[1].Retried {
		t.Error("different file/class should not be marked as retried")
	}
	if result.Results[2].Retried {
		t.Error("duplicate all-pass record should not be marked as retried")
	}
	if result.Results[2].RetryAttempt != 1 {
		t.Errorf("retryAttempt = %d, want 1 for duplicate key occurrence", result.Results[2].RetryAttempt)
	}
}

func BenchmarkParseJUnitXML_LargeReport(b *testing.B) {
	var doc strings.Builder
	doc.WriteString(`<testsuites><testsuite name="bench" tests="3000">`)
	for i := 0; i < 3000; i++ {
		doc.WriteString(`<testcase name="case-`)
		doc.WriteString(strconv.Itoa(i))
		doc.WriteString(`" classname="bench.Suite" time="0.001"/>`)
	}
	doc.WriteString(`</testsuite></testsuites>`)

	path := writeTempFile(b, "junit-bench.xml", doc.String())

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := ParseJUnitXML(path); err != nil {
			b.Fatalf("ParseJUnitXML failed: %v", err)
		}
	}
}

func writeTempFile(tb testing.TB, name, content string) string {
	tb.Helper()
	dir := tb.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		tb.Fatal(err)
	}
	return path
}
