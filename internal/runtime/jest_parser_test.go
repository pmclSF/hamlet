package runtime

import (
	"testing"
)

func TestParseJestJSON_Basic(t *testing.T) {
	json := `{
  "numTotalTests": 3,
  "numPassedTests": 2,
  "numFailedTests": 1,
  "success": false,
  "testResults": [
    {
      "testFilePath": "/app/src/auth.test.js",
      "numPassingTests": 1,
      "numFailingTests": 1,
      "numPendingTests": 0,
      "assertionResults": [
        {
          "fullName": "auth > login succeeds",
          "title": "login succeeds",
          "ancestorTitles": ["auth"],
          "status": "passed",
          "duration": 150,
          "failureMessages": []
        },
        {
          "fullName": "auth > login fails gracefully",
          "title": "login fails gracefully",
          "ancestorTitles": ["auth"],
          "status": "failed",
          "duration": 200,
          "failureMessages": ["Expected 401 but got 500"]
        }
      ]
    },
    {
      "testFilePath": "/app/src/utils.test.js",
      "numPassingTests": 1,
      "numFailingTests": 0,
      "numPendingTests": 0,
      "assertionResults": [
        {
          "fullName": "utils > format date",
          "title": "format date",
          "ancestorTitles": ["utils"],
          "status": "passed",
          "duration": 5
        }
      ]
    }
  ]
}`

	path := writeTempFile(t, "results.json", json)
	result, err := ParseJestJSON(path)
	if err != nil {
		t.Fatalf("ParseJestJSON failed: %v", err)
	}

	if result.Format != "jest-json" {
		t.Errorf("format = %q, want jest-json", result.Format)
	}
	if len(result.Results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(result.Results))
	}

	// Check passed test.
	r0 := result.Results[0]
	if r0.Name != "auth > login succeeds" {
		t.Errorf("name = %q", r0.Name)
	}
	if r0.Status != StatusPassed {
		t.Errorf("status = %q", r0.Status)
	}
	if r0.DurationMs != 150 {
		t.Errorf("durationMs = %f, want 150", r0.DurationMs)
	}
	if r0.Suite != "auth" {
		t.Errorf("suite = %q, want auth", r0.Suite)
	}

	// Check failed test.
	r1 := result.Results[1]
	if r1.Status != StatusFailed {
		t.Errorf("status = %q, want failed", r1.Status)
	}
	if r1.Message == "" {
		t.Error("expected failure message")
	}
}

func TestParseJestJSON_PendingSkipped(t *testing.T) {
	json := `{
  "numTotalTests": 1,
  "testResults": [
    {
      "testFilePath": "skip.test.js",
      "assertionResults": [
        {
          "fullName": "skipped test",
          "title": "skipped test",
          "status": "pending",
          "failureMessages": []
        }
      ]
    }
  ]
}`

	path := writeTempFile(t, "pending.json", json)
	result, err := ParseJestJSON(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Results))
	}
	if result.Results[0].Status != StatusSkipped {
		t.Errorf("status = %q, want skipped", result.Results[0].Status)
	}
}

func TestParseJestJSON_VitestRetry(t *testing.T) {
	json := `{
  "numTotalTests": 1,
  "testResults": [
    {
      "testFilePath": "flaky.test.ts",
      "assertionResults": [
        {
          "fullName": "flaky operation",
          "title": "flaky operation",
          "status": "passed",
          "duration": 300,
          "retryCount": 2,
          "failureMessages": []
        }
      ]
    }
  ]
}`

	path := writeTempFile(t, "vitest.json", json)
	result, err := ParseJestJSON(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Results) != 1 {
		t.Fatal("expected 1 result")
	}
	if !result.Results[0].Retried {
		t.Error("expected retried=true for retryCount > 0")
	}
	if result.Results[0].RetryAttempt != 2 {
		t.Errorf("retryAttempt = %d, want 2", result.Results[0].RetryAttempt)
	}
}
