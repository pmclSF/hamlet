package depgraph

import (
	"testing"
)

func TestAnalyzeProfile_SmallGraph(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	cov := AnalyzeCoverage(g)
	dupes := DetectDuplicates(g)
	fanout := AnalyzeFanout(g, DefaultFanoutThreshold)
	insights := ProfileInsights{
		Coverage:   &cov,
		Duplicates: &dupes,
		Fanout:     &fanout,
	}

	profile := AnalyzeProfile(g, insights)

	// 4 tests → tiny.
	if profile.TestVolume != "tiny" {
		t.Errorf("expected tiny volume, got %s", profile.TestVolume)
	}
	if profile.CIPressure != "low" {
		t.Errorf("expected low CI pressure, got %s", profile.CIPressure)
	}
}

func TestAnalyzeProfile_NilInsights(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	profile := AnalyzeProfile(g, ProfileInsights{})

	if profile.CoverageConfidence != "low" {
		t.Errorf("expected low coverage confidence with nil insights, got %s", profile.CoverageConfidence)
	}
	if profile.RedundancyLevel != "low" {
		t.Errorf("expected low redundancy with nil insights, got %s", profile.RedundancyLevel)
	}
}

func TestDetectEdgeCases_FewTests(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	profile := RepoProfile{TestVolume: "tiny", CIPressure: "low"}
	cases := DetectEdgeCases(profile, g, ProfileInsights{})

	types := map[EdgeCaseType]bool{}
	for _, c := range cases {
		types[c.Type] = true
	}
	if !types[EdgeCaseFewTests] {
		t.Error("expected FEW_TESTS edge case")
	}
	if !types[EdgeCaseFastCIAlready] {
		t.Error("expected FAST_CI_ALREADY edge case")
	}
}

func TestApplyEdgeCasePolicy_FewTests(t *testing.T) {
	t.Parallel()
	cases := []EdgeCase{
		{Type: EdgeCaseFewTests, Severity: "critical"},
	}
	profile := RepoProfile{TestVolume: "tiny"}

	policy := ApplyEdgeCasePolicy(cases, profile)

	if !policy.OptimizationDisabled {
		t.Error("expected optimization disabled for FEW_TESTS")
	}
	if policy.FallbackLevel != FallbackFullSuite {
		t.Errorf("expected FullSuite fallback, got %s", policy.FallbackLevel)
	}
	if policy.ConfidenceAdjustment >= 1.0 {
		t.Errorf("expected confidence adjustment < 1.0, got %f", policy.ConfidenceAdjustment)
	}
	if len(policy.Recommendations) == 0 {
		t.Error("expected at least one recommendation")
	}
}

func TestApplyEdgeCasePolicy_NoEdgeCases(t *testing.T) {
	t.Parallel()
	profile := RepoProfile{TestVolume: "large"}

	policy := ApplyEdgeCasePolicy(nil, profile)

	if policy.OptimizationDisabled {
		t.Error("optimization should not be disabled with no edge cases")
	}
	if policy.ConfidenceAdjustment != 1.0 {
		t.Errorf("expected confidence 1.0 with no edge cases, got %f", policy.ConfidenceAdjustment)
	}
	if policy.FallbackLevel != FallbackDirectDeps {
		t.Errorf("expected DirectDeps fallback, got %s", policy.FallbackLevel)
	}
}

func TestDetectEdgeCases_ExternalServiceHeavy(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	profile := RepoProfile{}
	spd := SnapshotProfileData{ExternalServiceNodeCount: 6}
	cases := DetectEdgeCases(profile, g, ProfileInsights{Snapshot: spd})

	found := false
	for _, c := range cases {
		if c.Type == EdgeCaseExternalServiceHeavy {
			found = true
			if c.Severity != "caution" {
				t.Errorf("expected caution severity, got %s", c.Severity)
			}
		}
	}
	if !found {
		t.Error("expected EXTERNAL_SERVICE_HEAVY edge case when >5 external service nodes")
	}
}

func TestDetectEdgeCases_GeneratedArtifacts(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	profile := RepoProfile{}
	spd := SnapshotProfileData{GeneratedArtifactNodeCount: 1}
	cases := DetectEdgeCases(profile, g, ProfileInsights{Snapshot: spd})

	found := false
	for _, c := range cases {
		if c.Type == EdgeCaseGeneratedArtifacts {
			found = true
		}
	}
	if !found {
		t.Error("expected GENERATED_ARTIFACT_CHANGES edge case when generated artifacts present")
	}
}

func TestDetectEdgeCases_MigrationOverlap(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	profile := RepoProfile{}
	spd := SnapshotProfileData{MigrationSignalCount: 15}
	cases := DetectEdgeCases(profile, g, ProfileInsights{Snapshot: spd})

	found := false
	for _, c := range cases {
		if c.Type == EdgeCaseMigrationOverlap {
			found = true
			if c.Severity != "caution" {
				t.Errorf("expected caution severity, got %s", c.Severity)
			}
		}
	}
	if !found {
		t.Error("expected MIGRATION_OVERLAP edge case when migration signals > 10")
	}
}

func TestDetectEdgeCases_SnapshotHeavy(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	profile := RepoProfile{}
	spd := SnapshotProfileData{
		SnapshotAssertionCount: 50,
		TotalAssertionCount:    100,
	}
	cases := DetectEdgeCases(profile, g, ProfileInsights{Snapshot: spd})

	found := false
	for _, c := range cases {
		if c.Type == EdgeCaseSnapshotHeavy {
			found = true
			if c.Severity != "warning" {
				t.Errorf("expected warning severity, got %s", c.Severity)
			}
		}
	}
	if !found {
		t.Error("expected SNAPSHOT_HEAVY_SUITE edge case when >40% snapshot assertions")
	}
}

func TestDetectEdgeCases_SnapshotHeavy_BelowThreshold(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	profile := RepoProfile{}
	// 30% snapshot ratio — below 40% threshold.
	spd := SnapshotProfileData{
		SnapshotAssertionCount: 30,
		TotalAssertionCount:    100,
	}
	cases := DetectEdgeCases(profile, g, ProfileInsights{Snapshot: spd})

	for _, c := range cases {
		if c.Type == EdgeCaseSnapshotHeavy {
			t.Error("did not expect SNAPSHOT_HEAVY_SUITE edge case at 30% snapshot ratio")
		}
	}
}

func TestDetectEdgeCases_LegacyZone(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	profile := RepoProfile{}
	spd := SnapshotProfileData{LegacyFrameworkSignalCount: 8}
	cases := DetectEdgeCases(profile, g, ProfileInsights{Snapshot: spd})

	found := false
	for _, c := range cases {
		if c.Type == EdgeCaseLegacyZone {
			found = true
		}
	}
	if !found {
		t.Error("expected LEGACY_ZONE edge case when legacy signals > 5")
	}
}

func TestDetectEdgeCases_MixedTestCultures(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	profile := RepoProfile{}
	spd := SnapshotProfileData{
		FrameworkCount: 4,
		FrameworkTypes: []string{"unit", "integration", "e2e", "contract"},
	}
	cases := DetectEdgeCases(profile, g, ProfileInsights{Snapshot: spd})

	found := false
	for _, c := range cases {
		if c.Type == EdgeCaseMixedTestCultures {
			found = true
		}
	}
	if !found {
		t.Error("expected MIXED_TEST_CULTURES edge case when >=4 frameworks")
	}
}

func TestDetectEdgeCases_LargeManualSuite(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	profile := RepoProfile{ManualCoveragePresence: "significant"}
	cases := DetectEdgeCases(profile, g, ProfileInsights{})

	found := false
	for _, c := range cases {
		if c.Type == EdgeCaseLargeManualSuite {
			found = true
		}
	}
	if !found {
		t.Error("expected LARGE_MANUAL_SUITE edge case when manual coverage is significant")
	}
}

func TestDetectEdgeCases_RedundantTestSuite(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	profile := RepoProfile{RedundancyLevel: "high"}
	cases := DetectEdgeCases(profile, g, ProfileInsights{})

	found := false
	for _, c := range cases {
		if c.Type == EdgeCaseRedundantSuite {
			found = true
			if c.Severity != "caution" {
				t.Errorf("expected caution severity, got %s", c.Severity)
			}
		}
	}
	if !found {
		t.Error("expected REDUNDANT_TEST_SUITE edge case when redundancy is high")
	}
}

func TestDetectEdgeCases_HighSkipBurden(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	profile := RepoProfile{SkipBurden: "high"}
	cases := DetectEdgeCases(profile, g, ProfileInsights{})

	found := false
	for _, c := range cases {
		if c.Type == EdgeCaseHighSkipBurden {
			found = true
			if c.Severity != "caution" {
				t.Errorf("expected caution severity, got %s", c.Severity)
			}
		}
	}
	if !found {
		t.Error("expected HIGH_SKIP_BURDEN edge case when skip burden is high")
	}
}

func TestDetectEdgeCases_HighFlakeBurden(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	profile := RepoProfile{FlakeBurden: "high"}
	cases := DetectEdgeCases(profile, g, ProfileInsights{})

	found := false
	for _, c := range cases {
		if c.Type == EdgeCaseHighFlakeBurden {
			found = true
			if c.Severity != "caution" {
				t.Errorf("expected caution severity, got %s", c.Severity)
			}
		}
	}
	if !found {
		t.Error("expected HIGH_FLAKE_BURDEN edge case when flake burden is high")
	}
}

func TestDetectEdgeCases_HighFanoutFixture(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	// Need a profile with high fanout burden.
	profile := RepoProfile{FanoutBurden: "high"}
	fanout := FanoutResult{
		NodeCount:    10,
		FlaggedCount: 4, // 40% > 30% threshold
		Threshold:    DefaultFanoutThreshold,
	}
	cases := DetectEdgeCases(profile, g, ProfileInsights{Fanout: &fanout})

	found := false
	for _, c := range cases {
		if c.Type == EdgeCaseHighFanoutFixture {
			found = true
		}
	}
	if !found {
		t.Error("expected HIGH_FANOUT_FIXTURE edge case when >30% nodes flagged")
	}
}

func TestDetectEdgeCases_LowGraphVisibility(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	profile := RepoProfile{CoverageConfidence: "low"}
	cases := DetectEdgeCases(profile, g, ProfileInsights{})

	found := false
	for _, c := range cases {
		if c.Type == EdgeCaseLowGraphVisibility {
			found = true
			if c.Severity != "warning" {
				t.Errorf("expected warning severity, got %s", c.Severity)
			}
		}
	}
	if !found {
		t.Error("expected LOW_GRAPH_VISIBILITY edge case when coverage confidence is low")
	}
}

func TestDetectEdgeCases_FastCIAlready(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	profile := RepoProfile{CIPressure: "low"}
	cases := DetectEdgeCases(profile, g, ProfileInsights{})

	found := false
	for _, c := range cases {
		if c.Type == EdgeCaseFastCIAlready {
			found = true
		}
	}
	if !found {
		t.Error("expected FAST_CI_ALREADY edge case when CI pressure is low")
	}
}

func TestApplyEdgeCasePolicy_MultipleEdgeCases(t *testing.T) {
	t.Parallel()
	cases := []EdgeCase{
		{Type: EdgeCaseHighFlakeBurden, Severity: "caution"},
		{Type: EdgeCaseHighFanoutFixture, Severity: "caution"},
		{Type: EdgeCaseExternalServiceHeavy, Severity: "caution"},
	}
	profile := RepoProfile{}

	policy := ApplyEdgeCasePolicy(cases, profile)

	if !policy.RiskElevated {
		t.Error("expected risk elevated with high flake burden")
	}
	// 1.0 * 0.75 * 0.7 * 0.85 = 0.44625
	if policy.ConfidenceAdjustment > 0.45 || policy.ConfidenceAdjustment < 0.44 {
		t.Errorf("expected confidence ~0.45, got %f", policy.ConfidenceAdjustment)
	}
	if policy.FallbackLevel < FallbackPackageTests {
		t.Errorf("expected at least PackageTests fallback, got %s", policy.FallbackLevel)
	}
	if len(policy.Recommendations) != 3 {
		t.Errorf("expected 3 recommendations, got %d", len(policy.Recommendations))
	}
}

func TestApplyEdgeCasePolicy_ConfidenceClamp(t *testing.T) {
	t.Parallel()
	// Stack many edge cases to drive confidence very low.
	cases := []EdgeCase{
		{Type: EdgeCaseFewTests, Severity: "critical"},
		{Type: EdgeCaseHighFlakeBurden, Severity: "caution"},
		{Type: EdgeCaseHighFanoutFixture, Severity: "caution"},
		{Type: EdgeCaseLowGraphVisibility, Severity: "warning"},
		{Type: EdgeCaseRedundantSuite, Severity: "caution"},
		{Type: EdgeCaseMigrationOverlap, Severity: "caution"},
	}
	profile := RepoProfile{}

	policy := ApplyEdgeCasePolicy(cases, profile)

	if policy.ConfidenceAdjustment < 0.1 {
		t.Errorf("confidence should be clamped at 0.1, got %f", policy.ConfidenceAdjustment)
	}
}
