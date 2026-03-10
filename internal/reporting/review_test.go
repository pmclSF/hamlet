package reporting

import (
	"bytes"
	"strings"
	"testing"

	"github.com/pmclSF/hamlet/internal/models"
)

func TestGroupSignalsByOwner(t *testing.T) {
	t.Parallel()
	signals := []models.Signal{
		{Type: "weakAssertion", Owner: "payments-team"},
		{Type: "mockHeavyTest", Owner: "payments-team"},
		{Type: "weakAssertion", Owner: "auth-team"},
		{Type: "weakAssertion", Owner: ""},
	}

	groups := GroupSignalsByOwner(signals)
	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(groups))
	}
	// Sorted by count descending — payments-team should be first
	if groups[0].Key != "payments-team" || groups[0].Count != 2 {
		t.Errorf("expected payments-team with 2, got %s with %d", groups[0].Key, groups[0].Count)
	}
}

func TestGroupSignalsByType(t *testing.T) {
	t.Parallel()
	signals := []models.Signal{
		{Type: "weakAssertion"},
		{Type: "weakAssertion"},
		{Type: "mockHeavyTest"},
	}

	groups := GroupSignalsByType(signals)
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	if groups[0].Key != "weakAssertion" || groups[0].Count != 2 {
		t.Errorf("expected weakAssertion with 2, got %s with %d", groups[0].Key, groups[0].Count)
	}
}

func TestGroupSignalsByDirectory(t *testing.T) {
	t.Parallel()
	signals := []models.Signal{
		{Type: "weakAssertion", Location: models.SignalLocation{File: "src/auth/login.test.js"}},
		{Type: "weakAssertion", Location: models.SignalLocation{File: "src/auth/signup.test.js"}},
		{Type: "mockHeavyTest", Location: models.SignalLocation{File: "src/payments/checkout.test.js"}},
		{Type: "weakAssertion", Location: models.SignalLocation{Repository: "repo"}},
	}

	groups := GroupSignalsByDirectory(signals)
	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(groups))
	}
	if groups[0].Key != "src/auth" || groups[0].Count != 2 {
		t.Errorf("expected src/auth with 2, got %s with %d", groups[0].Key, groups[0].Count)
	}
}

func TestMigrationBlockers(t *testing.T) {
	t.Parallel()
	signals := []models.Signal{
		{Type: "migrationBlocker"},
		{Type: "weakAssertion"},
		{Type: "deprecatedTestPattern"},
		{Type: "customMatcherRisk"},
		{Type: "flakyTest"},
	}

	blockers := MigrationBlockers(signals)
	if len(blockers) != 3 {
		t.Errorf("expected 3 migration blockers, got %d", len(blockers))
	}
}

func TestMigrationBlockers_None(t *testing.T) {
	t.Parallel()
	signals := []models.Signal{
		{Type: "weakAssertion"},
		{Type: "flakyTest"},
	}

	blockers := MigrationBlockers(signals)
	if len(blockers) != 0 {
		t.Errorf("expected 0 migration blockers, got %d", len(blockers))
	}
}

func TestRenderReviewSections_WithData(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		Signals: []models.Signal{
			{Type: "weakAssertion", Owner: "auth-team", Location: models.SignalLocation{File: "src/auth/login.test.js"}},
			{Type: "weakAssertion", Owner: "auth-team", Location: models.SignalLocation{File: "src/auth/signup.test.js"}},
			{Type: "mockHeavyTest", Owner: "payments-team", Location: models.SignalLocation{File: "src/pay/pay.test.js"}},
		},
		Risk: []models.RiskSurface{
			{Type: "change", Scope: "repository", ScopeName: "repo", Band: "medium"},
		},
	}

	var buf bytes.Buffer
	RenderReviewSections(&buf, snap)
	output := buf.String()

	if !strings.Contains(output, "Highest-Risk Areas") {
		t.Error("expected Highest-Risk Areas section")
	}
	if !strings.Contains(output, "Review by Owner") {
		t.Error("expected Review by Owner section")
	}
}

func TestRenderReviewSections_Empty(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{}

	var buf bytes.Buffer
	RenderReviewSections(&buf, snap)
	output := buf.String()

	if output != "" {
		t.Errorf("expected empty output for empty signals, got %q", output)
	}
}
