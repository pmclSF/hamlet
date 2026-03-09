package failure

import "testing"

func TestClassify_AssertionFailure(t *testing.T) {
	inputs := []FailureInput{
		{
			TestFilePath: "src/__tests__/math.test.js",
			TestName:     "adds numbers correctly",
			ErrorMessage: "expect(received).toBe(expected)\n\nExpected: 4\nReceived: 5",
		},
	}

	result := Classify(inputs)

	if result.TotalFailures != 1 {
		t.Fatalf("expected 1 failure, got %d", result.TotalFailures)
	}
	c := result.Classifications[0]
	if c.Category != CategoryAssertionFailure {
		t.Errorf("expected assertion_failure, got %s", c.Category)
	}
	if c.Confidence != ConfidenceInferred {
		t.Errorf("expected inferred confidence, got %s", c.Confidence)
	}
	if c.ConfidenceScore < 0.5 {
		t.Errorf("expected confidence score >= 0.5, got %f", c.ConfidenceScore)
	}
}

func TestClassify_AssertionWithAssertKeyword(t *testing.T) {
	inputs := []FailureInput{
		{
			TestFilePath: "test/utils.test.js",
			TestName:     "validates input",
			ErrorMessage: "AssertionError: expected true to be false",
		},
	}

	result := Classify(inputs)

	if result.Classifications[0].Category != CategoryAssertionFailure {
		t.Errorf("expected assertion_failure, got %s", result.Classifications[0].Category)
	}
}

func TestClassify_Timeout(t *testing.T) {
	inputs := []FailureInput{
		{
			TestFilePath: "e2e/login.spec.ts",
			TestName:     "should login successfully",
			ErrorMessage: "Timeout - Async callback was not invoked within 5000ms",
		},
	}

	result := Classify(inputs)

	if result.TotalFailures != 1 {
		t.Fatalf("expected 1 failure, got %d", result.TotalFailures)
	}
	c := result.Classifications[0]
	if c.Category != CategoryTimeout {
		t.Errorf("expected timeout, got %s", c.Category)
	}
	if c.Confidence != ConfidenceExact {
		t.Errorf("expected exact confidence, got %s", c.Confidence)
	}
}

func TestClassify_TimeoutTimedOut(t *testing.T) {
	inputs := []FailureInput{
		{
			TestFilePath: "tests/api.test.js",
			TestName:     "fetches data",
			ErrorMessage: "Error: ETIMEDOUT connect to api.example.com:443",
		},
	}

	result := Classify(inputs)

	if result.Classifications[0].Category != CategoryTimeout {
		t.Errorf("expected timeout, got %s", result.Classifications[0].Category)
	}
}

func TestClassify_SetupFailure(t *testing.T) {
	inputs := []FailureInput{
		{
			TestFilePath: "test/db.test.js",
			TestName:     "queries records",
			ErrorMessage: "Error in beforeEach: Failed to connect to test database",
		},
	}

	result := Classify(inputs)

	if result.TotalFailures != 1 {
		t.Fatalf("expected 1 failure, got %d", result.TotalFailures)
	}
	c := result.Classifications[0]
	if c.Category != CategorySetupFixture {
		t.Errorf("expected setup_or_fixture_failure, got %s", c.Category)
	}
	if c.Confidence != ConfidenceInferred {
		t.Errorf("expected inferred confidence, got %s", c.Confidence)
	}
}

func TestClassify_SetupBeforeAll(t *testing.T) {
	inputs := []FailureInput{
		{
			TestFilePath: "test/integration.test.js",
			TestName:     "runs suite",
			ErrorMessage: "beforeAll failed: Could not initialize server",
		},
	}

	result := Classify(inputs)

	if result.Classifications[0].Category != CategorySetupFixture {
		t.Errorf("expected setup_or_fixture_failure, got %s", result.Classifications[0].Category)
	}
}

func TestClassify_DependencyService(t *testing.T) {
	inputs := []FailureInput{
		{
			TestFilePath: "test/api.test.js",
			TestName:     "calls external API",
			ErrorMessage: "Error: connect ECONNREFUSED 127.0.0.1:5432",
		},
	}

	result := Classify(inputs)

	if result.TotalFailures != 1 {
		t.Fatalf("expected 1 failure, got %d", result.TotalFailures)
	}
	c := result.Classifications[0]
	if c.Category != CategoryDependencyService {
		t.Errorf("expected dependency_or_service_failure, got %s", c.Category)
	}
	if c.Confidence != ConfidenceExact {
		t.Errorf("expected exact confidence, got %s", c.Confidence)
	}
}

func TestClassify_ServiceUnavailable(t *testing.T) {
	inputs := []FailureInput{
		{
			TestFilePath: "test/integration.test.js",
			TestName:     "sends request",
			ErrorMessage: "Request failed with status 503 Service Unavailable",
		},
	}

	result := Classify(inputs)

	if result.Classifications[0].Category != CategoryDependencyService {
		t.Errorf("expected dependency_or_service_failure, got %s", result.Classifications[0].Category)
	}
}

func TestClassify_SnapshotMismatch(t *testing.T) {
	inputs := []FailureInput{
		{
			TestFilePath: "test/components/Button.test.jsx",
			TestName:     "renders correctly",
			ErrorMessage: "expect(received).toMatchSnapshot()\n\nSnapshot name: `Button renders correctly 1`\n\n- Snapshot\n+ Received",
		},
	}

	result := Classify(inputs)

	if result.TotalFailures != 1 {
		t.Fatalf("expected 1 failure, got %d", result.TotalFailures)
	}
	c := result.Classifications[0]
	if c.Category != CategorySnapshotMismatch {
		t.Errorf("expected snapshot_mismatch, got %s", c.Category)
	}
	if c.Confidence != ConfidenceExact {
		t.Errorf("expected exact confidence, got %s", c.Confidence)
	}
	if c.ConfidenceScore < 0.9 {
		t.Errorf("expected high confidence score, got %f", c.ConfidenceScore)
	}
}

func TestClassify_InlineSnapshotMismatch(t *testing.T) {
	inputs := []FailureInput{
		{
			TestFilePath: "test/render.test.js",
			TestName:     "renders header",
			ErrorMessage: "expect(received).toMatchInlineSnapshot()",
		},
	}

	result := Classify(inputs)

	if result.Classifications[0].Category != CategorySnapshotMismatch {
		t.Errorf("expected snapshot_mismatch, got %s", result.Classifications[0].Category)
	}
}

func TestClassify_SelectorUI(t *testing.T) {
	inputs := []FailureInput{
		{
			TestFilePath: "e2e/checkout.spec.ts",
			TestName:     "clicks submit button",
			ErrorMessage: "Error: Element not found: selector '#submit-btn' did not match any elements",
		},
	}

	result := Classify(inputs)

	if result.TotalFailures != 1 {
		t.Fatalf("expected 1 failure, got %d", result.TotalFailures)
	}
	c := result.Classifications[0]
	if c.Category != CategorySelectorUI {
		t.Errorf("expected selector_or_ui_fragility, got %s", c.Category)
	}
}

func TestClassify_StaleElement(t *testing.T) {
	inputs := []FailureInput{
		{
			TestFilePath: "e2e/dashboard.spec.ts",
			TestName:     "opens modal",
			ErrorMessage: "StaleElementReferenceException: stale element reference: element is not attached to the page document",
		},
	}

	result := Classify(inputs)

	if result.Classifications[0].Category != CategorySelectorUI {
		t.Errorf("expected selector_or_ui_fragility, got %s", result.Classifications[0].Category)
	}
}

func TestClassify_InfrastructureOOM(t *testing.T) {
	inputs := []FailureInput{
		{
			TestFilePath: "test/heavy.test.js",
			TestName:     "processes large dataset",
			ErrorMessage: "FATAL ERROR: CALL_AND_RETRY_LAST Allocation failed - JavaScript heap out of memory",
		},
	}

	result := Classify(inputs)

	if result.TotalFailures != 1 {
		t.Fatalf("expected 1 failure, got %d", result.TotalFailures)
	}
	c := result.Classifications[0]
	if c.Category != CategoryInfraEnvironment {
		t.Errorf("expected infrastructure_or_environment, got %s", c.Category)
	}
	if c.Confidence != ConfidenceExact {
		t.Errorf("expected exact confidence, got %s", c.Confidence)
	}
}

func TestClassify_PermissionDenied(t *testing.T) {
	inputs := []FailureInput{
		{
			TestFilePath: "test/fs.test.js",
			TestName:     "writes file",
			ErrorMessage: "Error: EACCES: permission denied, open '/etc/test.conf'",
		},
	}

	result := Classify(inputs)

	if result.Classifications[0].Category != CategoryInfraEnvironment {
		t.Errorf("expected infrastructure_or_environment, got %s", result.Classifications[0].Category)
	}
}

func TestClassify_UnknownFailure(t *testing.T) {
	inputs := []FailureInput{
		{
			TestFilePath: "test/misc.test.js",
			TestName:     "does something",
			ErrorMessage: "SyntaxError: invalid token in JSON at position 42",
		},
	}

	result := Classify(inputs)

	if result.TotalFailures != 1 {
		t.Fatalf("expected 1 failure, got %d", result.TotalFailures)
	}
	c := result.Classifications[0]
	if c.Category != CategoryUnknown {
		t.Errorf("expected unknown, got %s", c.Category)
	}
	if c.Confidence != ConfidenceWeak {
		t.Errorf("expected weak confidence, got %s", c.Confidence)
	}
}

func TestClassify_EmptyInput(t *testing.T) {
	result := Classify(nil)

	if result.TotalFailures != 0 {
		t.Errorf("expected 0 failures, got %d", result.TotalFailures)
	}
	if len(result.Classifications) != 0 {
		t.Errorf("expected 0 classifications, got %d", len(result.Classifications))
	}
	if result.DominantCategory != CategoryUnknown {
		t.Errorf("expected unknown dominant category, got %s", result.DominantCategory)
	}
}

func TestClassify_EmptyErrorMessage(t *testing.T) {
	inputs := []FailureInput{
		{
			TestFilePath: "test/empty.test.js",
			TestName:     "no message",
			ErrorMessage: "",
			StackTrace:   "",
		},
	}

	result := Classify(inputs)

	c := result.Classifications[0]
	if c.Category != CategoryUnknown {
		t.Errorf("expected unknown, got %s", c.Category)
	}
	if c.ConfidenceScore > 0.2 {
		t.Errorf("expected low confidence score, got %f", c.ConfidenceScore)
	}
}

func TestClassify_StackTraceOnly(t *testing.T) {
	inputs := []FailureInput{
		{
			TestFilePath: "test/stack.test.js",
			TestName:     "has stack",
			ErrorMessage: "",
			StackTrace:   "Error: connect ECONNREFUSED 127.0.0.1:3000\n    at TCPConnectWrap.afterConnect",
		},
	}

	result := Classify(inputs)

	c := result.Classifications[0]
	if c.Category != CategoryDependencyService {
		t.Errorf("expected dependency_or_service_failure from stack trace, got %s", c.Category)
	}
}

func TestClassify_DominantCategory(t *testing.T) {
	inputs := []FailureInput{
		{TestFilePath: "a.test.js", TestName: "t1", ErrorMessage: "expect(1).toBe(2)"},
		{TestFilePath: "b.test.js", TestName: "t2", ErrorMessage: "expect(3).toEqual(4)"},
		{TestFilePath: "c.test.js", TestName: "t3", ErrorMessage: "expected true to be false"},
		{TestFilePath: "d.test.js", TestName: "t4", ErrorMessage: "Timeout after 5000ms"},
	}

	result := Classify(inputs)

	if result.DominantCategory != CategoryAssertionFailure {
		t.Errorf("expected assertion_failure as dominant, got %s", result.DominantCategory)
	}
	if result.ByCategory[CategoryAssertionFailure] != 3 {
		t.Errorf("expected 3 assertion failures, got %d", result.ByCategory[CategoryAssertionFailure])
	}
	if result.ByCategory[CategoryTimeout] != 1 {
		t.Errorf("expected 1 timeout, got %d", result.ByCategory[CategoryTimeout])
	}
}

func TestClassify_DeterministicSort(t *testing.T) {
	inputs := []FailureInput{
		{TestFilePath: "z.test.js", TestName: "z", ErrorMessage: "expect(1).toBe(2)"},
		{TestFilePath: "a.test.js", TestName: "a", ErrorMessage: "expect(1).toBe(2)"},
		{TestFilePath: "m.test.js", TestName: "m", ErrorMessage: "expect(1).toBe(2)"},
	}

	result := Classify(inputs)

	if result.Classifications[0].TestFilePath != "a.test.js" {
		t.Errorf("expected first classification to be a.test.js, got %s", result.Classifications[0].TestFilePath)
	}
	if result.Classifications[1].TestFilePath != "m.test.js" {
		t.Errorf("expected second classification to be m.test.js, got %s", result.Classifications[1].TestFilePath)
	}
	if result.Classifications[2].TestFilePath != "z.test.js" {
		t.Errorf("expected third classification to be z.test.js, got %s", result.Classifications[2].TestFilePath)
	}
}

func TestClassify_SnapshotTakesPriorityOverAssertion(t *testing.T) {
	// "toMatchSnapshot" contains both snapshot and assertion-like keywords.
	// Snapshot should win because it's more specific.
	inputs := []FailureInput{
		{
			TestFilePath: "test/component.test.js",
			TestName:     "renders",
			ErrorMessage: "expect(received).toMatchSnapshot()\n\nExpected: ...\nReceived: ...",
		},
	}

	result := Classify(inputs)

	if result.Classifications[0].Category != CategorySnapshotMismatch {
		t.Errorf("expected snapshot_mismatch (higher priority), got %s", result.Classifications[0].Category)
	}
}

func TestClassify_SelectorTakesPriorityOverTimeout(t *testing.T) {
	// "waiting for selector" contains both selector and timeout-like context.
	inputs := []FailureInput{
		{
			TestFilePath: "e2e/page.spec.ts",
			TestName:     "loads page",
			ErrorMessage: "Timeout exceeded while waiting for selector '#main-content'",
		},
	}

	result := Classify(inputs)

	// Selector has higher priority than timeout.
	if result.Classifications[0].Category != CategorySelectorUI {
		t.Errorf("expected selector_or_ui_fragility (higher priority), got %s", result.Classifications[0].Category)
	}
}

func TestClassify_PreservesOriginalErrorAndStack(t *testing.T) {
	errMsg := "expect(received).toBe(expected)"
	stack := "at Object.<anonymous> (test.js:10:5)"

	inputs := []FailureInput{
		{
			TestFilePath: "test.js",
			TestName:     "t",
			ErrorMessage: errMsg,
			StackTrace:   stack,
		},
	}

	result := Classify(inputs)

	c := result.Classifications[0]
	if c.ErrorMessage != errMsg {
		t.Errorf("expected original error message preserved, got %s", c.ErrorMessage)
	}
	if c.StackTrace != stack {
		t.Errorf("expected original stack trace preserved, got %s", c.StackTrace)
	}
}

func TestClassify_MultipleCategories(t *testing.T) {
	inputs := []FailureInput{
		{TestFilePath: "a.test.js", TestName: "t1", ErrorMessage: "expect(1).toBe(2)"},
		{TestFilePath: "b.test.js", TestName: "t2", ErrorMessage: "Timeout after 5000ms"},
		{TestFilePath: "c.test.js", TestName: "t3", ErrorMessage: "connect ECONNREFUSED 127.0.0.1:5432"},
		{TestFilePath: "d.test.js", TestName: "t4", ErrorMessage: "toMatchSnapshot() mismatch"},
		{TestFilePath: "e.test.js", TestName: "t5", ErrorMessage: "element not found: #btn"},
		{TestFilePath: "f.test.js", TestName: "t6", ErrorMessage: "ENOMEM cannot allocate memory"},
		{TestFilePath: "g.test.js", TestName: "t7", ErrorMessage: "beforeEach hook failed"},
	}

	result := Classify(inputs)

	if result.TotalFailures != 7 {
		t.Fatalf("expected 7 failures, got %d", result.TotalFailures)
	}

	expectedCategories := map[FailureCategory]int{
		CategoryAssertionFailure:  1,
		CategoryTimeout:           1,
		CategoryDependencyService: 1,
		CategorySnapshotMismatch:  1,
		CategorySelectorUI:        1,
		CategoryInfraEnvironment:  1,
		CategorySetupFixture:      1,
	}

	for cat, expected := range expectedCategories {
		if result.ByCategory[cat] != expected {
			t.Errorf("expected %d %s, got %d", expected, cat, result.ByCategory[cat])
		}
	}
}

func TestClassify_ExplanationNotEmpty(t *testing.T) {
	inputs := []FailureInput{
		{TestFilePath: "test.js", TestName: "t", ErrorMessage: "expect(1).toBe(2)"},
	}

	result := Classify(inputs)

	if result.Classifications[0].Explanation == "" {
		t.Error("expected non-empty explanation")
	}
}

func TestClassify_LongErrorMessageTruncatedInExplanation(t *testing.T) {
	longMsg := ""
	for i := 0; i < 200; i++ {
		longMsg += "x"
	}
	longMsg += " expect(1).toBe(2)"

	inputs := []FailureInput{
		{TestFilePath: "test.js", TestName: "t", ErrorMessage: longMsg},
	}

	result := Classify(inputs)

	c := result.Classifications[0]
	// Original error should be preserved in full.
	if c.ErrorMessage != longMsg {
		t.Error("expected full error message preserved")
	}
	// Explanation should be truncated.
	if len(c.Explanation) > 300 {
		t.Errorf("expected explanation to be truncated, got length %d", len(c.Explanation))
	}
}
