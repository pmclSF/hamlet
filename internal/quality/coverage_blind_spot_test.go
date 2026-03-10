package quality

import (
	"testing"

	"github.com/pmclSF/hamlet/internal/models"
)

func TestCoverageBlindSpotDetector_FromCoverageInsights(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		CoverageInsights: []models.CoverageInsight{
			{
				Type:            "only_e2e_coverage",
				Severity:        "medium",
				Description:     "3 code unit(s) are covered only by e2e tests.",
				SuggestedAction: "Add unit tests.",
			},
			{
				Type:            "only_e2e_unit",
				Severity:        "low",
				Description:     "AuthService is covered only by e2e tests.",
				Path:            "src/auth.js",
				UnitID:          "src/auth.js:AuthService",
				SuggestedAction: "Add unit tests for AuthService.",
			},
			{
				Type:            "partially_covered",
				Severity:        "low",
				Description:     "Ignored by blind-spot detector.",
				SuggestedAction: "n/a",
			},
			{
				Type:            "uncovered_exported",
				Severity:        "high",
				Description:     "1 exported/public function has no coverage.",
				SuggestedAction: "Prioritize public API tests.",
			},
		},
	}

	d := &CoverageBlindSpotDetector{}
	signals := d.Detect(snap)
	if len(signals) != 3 {
		t.Fatalf("expected 3 coverageBlindSpot signals, got %d", len(signals))
	}

	var repoLevel, fileLevel, high bool
	for _, s := range signals {
		if s.Type != "coverageBlindSpot" {
			t.Fatalf("unexpected signal type %q", s.Type)
		}
		if s.Location.File == "" && s.Location.Repository == "coverage" {
			repoLevel = true
		}
		if s.Location.File == "src/auth.js" {
			fileLevel = true
		}
		if s.Severity == models.SeverityHigh {
			high = true
		}
	}

	if !repoLevel {
		t.Fatal("expected at least one repository-level blind-spot signal")
	}
	if !fileLevel {
		t.Fatal("expected file-level blind-spot signal for src/auth.js")
	}
	if !high {
		t.Fatal("expected high-severity blind-spot signal for uncovered exported insight")
	}
}
