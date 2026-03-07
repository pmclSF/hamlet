package measurement

import (
	"testing"

	"github.com/pmclSF/hamlet/internal/models"
	"github.com/pmclSF/hamlet/internal/signals"
)

// --- helper builders ---

func makeSnap(testFiles int, sigs ...models.Signal) *models.TestSuiteSnapshot {
	snap := &models.TestSuiteSnapshot{}
	for i := 0; i < testFiles; i++ {
		snap.TestFiles = append(snap.TestFiles, models.TestFile{
			Path:      "test/file_" + string(rune('a'+i)) + ".test.js",
			Framework: "jest",
		})
	}
	snap.Signals = sigs
	return snap
}

func sig(t models.SignalType) models.Signal {
	return models.Signal{
		Type:     t,
		Location: models.SignalLocation{File: "test/file_a.test.js"},
	}
}

func sigInFile(t models.SignalType, file string) models.Signal {
	return models.Signal{
		Type:     t,
		Location: models.SignalLocation{File: file},
	}
}

// --- helpers tests ---

func TestCountSignals(t *testing.T) {
	snap := makeSnap(3,
		sig(signals.SignalFlakyTest),
		sig(signals.SignalFlakyTest),
		sig(signals.SignalSlowTest),
	)

	if got := countSignals(snap, signals.SignalFlakyTest); got != 2 {
		t.Errorf("countSignals(flakyTest) = %d, want 2", got)
	}
	if got := countSignals(snap, signals.SignalFlakyTest, signals.SignalSlowTest); got != 3 {
		t.Errorf("countSignals(flakyTest, slowTest) = %d, want 3", got)
	}
	if got := countSignals(snap, signals.SignalDeadTest); got != 0 {
		t.Errorf("countSignals(deadTest) = %d, want 0", got)
	}
}

func TestRatioToBand(t *testing.T) {
	tests := []struct {
		ratio float64
		want  string
	}{
		{0.0, "strong"},
		{0.05, "strong"},
		{0.06, "moderate"},
		{0.15, "moderate"},
		{0.16, "weak"},
		{0.30, "weak"},
		{0.31, "critical"},
	}
	for _, tt := range tests {
		got := ratioToBand(tt.ratio, 0.05, 0.15, 0.30)
		if got != tt.want {
			t.Errorf("ratioToBand(%v) = %q, want %q", tt.ratio, got, tt.want)
		}
	}
}

func TestRuntimeEvidence(t *testing.T) {
	// No runtime data.
	snap := makeSnap(2)
	if got := runtimeEvidence(snap); got != EvidenceWeak {
		t.Errorf("runtimeEvidence (no data) = %q, want %q", got, EvidenceWeak)
	}

	// With runtime data.
	snap.TestFiles[0].RuntimeStats = &models.RuntimeStats{AvgRuntimeMs: 100}
	if got := runtimeEvidence(snap); got != EvidenceStrong {
		t.Errorf("runtimeEvidence (with data) = %q, want %q", got, EvidenceStrong)
	}
}

func TestEvidenceLimitations(t *testing.T) {
	if lim := evidenceLimitations(EvidenceStrong); lim != nil {
		t.Errorf("evidenceLimitations(strong) = %v, want nil", lim)
	}
	if lim := evidenceLimitations(EvidenceWeak); len(lim) == 0 {
		t.Error("evidenceLimitations(weak) should return limitations")
	}
	if lim := evidenceLimitations(EvidenceNone); len(lim) == 0 {
		t.Error("evidenceLimitations(none) should return limitations")
	}
}

// --- registry tests ---

func TestRegistry_RegisterAndLen(t *testing.T) {
	r := NewRegistry()
	r.Register(Definition{ID: "test.one", Dimension: DimensionHealth})
	r.Register(Definition{ID: "test.two", Dimension: DimensionHealth})

	if r.Len() != 2 {
		t.Errorf("Len() = %d, want 2", r.Len())
	}
}

func TestRegistry_DuplicatePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on duplicate ID")
		}
	}()
	r := NewRegistry()
	r.Register(Definition{ID: "test.dup", Dimension: DimensionHealth})
	r.Register(Definition{ID: "test.dup", Dimension: DimensionHealth})
}

func TestRegistry_ByDimension(t *testing.T) {
	r := NewRegistry()
	r.Register(Definition{ID: "h.one", Dimension: DimensionHealth})
	r.Register(Definition{ID: "s.one", Dimension: DimensionStructuralRisk})
	r.Register(Definition{ID: "h.two", Dimension: DimensionHealth})

	health := r.ByDimension(DimensionHealth)
	if len(health) != 2 {
		t.Errorf("ByDimension(health) = %d, want 2", len(health))
	}
	structural := r.ByDimension(DimensionStructuralRisk)
	if len(structural) != 1 {
		t.Errorf("ByDimension(structural) = %d, want 1", len(structural))
	}
}

func TestRegistry_Run(t *testing.T) {
	r := NewRegistry()
	r.Register(Definition{
		ID:        "test.constant",
		Dimension: DimensionHealth,
		Compute: func(_ *models.TestSuiteSnapshot) Result {
			return Result{ID: "test.constant", Value: 0.42, Band: "moderate"}
		},
	})

	results := r.Run(makeSnap(1))
	if len(results) != 1 {
		t.Fatalf("Run() returned %d results, want 1", len(results))
	}
	if results[0].Value != 0.42 {
		t.Errorf("result value = %v, want 0.42", results[0].Value)
	}
}

// --- default registry tests ---

func TestDefaultRegistry_AllDimensionsCovered(t *testing.T) {
	r := DefaultRegistry()

	dims := []Dimension{
		DimensionHealth,
		DimensionCoverageDepth,
		DimensionCoverageDiversity,
		DimensionStructuralRisk,
		DimensionOperationalRisk,
	}

	for _, dim := range dims {
		defs := r.ByDimension(dim)
		if len(defs) == 0 {
			t.Errorf("no measurements registered for dimension %q", dim)
		}
	}
}

func TestDefaultRegistry_NoDuplicateIDs(t *testing.T) {
	// DefaultRegistry panics on duplicate IDs; if this runs, no duplicates.
	r := DefaultRegistry()
	if r.Len() == 0 {
		t.Error("default registry has no measurements")
	}
}

// --- health measurement tests ---

func TestHealth_FlakyShare_NoFiles(t *testing.T) {
	snap := makeSnap(0)
	r := computeFlakyShare(snap)
	if r.Value != 0 || r.Band != "strong" || r.Evidence != EvidenceNone {
		t.Errorf("unexpected result for empty snap: %+v", r)
	}
}

func TestHealth_FlakyShare_WithSignals(t *testing.T) {
	snap := makeSnap(10,
		sig(signals.SignalFlakyTest),
		sig(signals.SignalFlakyTest),
		sig(signals.SignalUnstableSuite),
	)
	r := computeFlakyShare(snap)
	if r.Value != 0.3 {
		t.Errorf("flaky_share value = %v, want 0.3", r.Value)
	}
	if r.Band != "weak" {
		t.Errorf("flaky_share band = %q, want 'weak'", r.Band)
	}
}

func TestHealth_SkipDensity(t *testing.T) {
	snap := makeSnap(20,
		sig(signals.SignalSkippedTest),
	)
	r := computeSkipDensity(snap)
	if r.Value != 0.05 {
		t.Errorf("skip_density value = %v, want 0.05", r.Value)
	}
	if r.Band != "strong" {
		t.Errorf("skip_density band = %q, want 'strong'", r.Band)
	}
	if r.Evidence != EvidenceStrong {
		t.Errorf("skip_density evidence = %q, want 'strong'", r.Evidence)
	}
}

func TestHealth_DeadTestShare(t *testing.T) {
	snap := makeSnap(10,
		sig(signals.SignalDeadTest),
		sig(signals.SignalDeadTest),
	)
	r := computeDeadTestShare(snap)
	if r.Value != 0.2 {
		t.Errorf("dead_test_share value = %v, want 0.2", r.Value)
	}
	if r.Band != "weak" {
		t.Errorf("dead_test_share band = %q, want 'weak'", r.Band)
	}
}

func TestHealth_SlowTestShare_RuntimeEvidence(t *testing.T) {
	snap := makeSnap(4,
		sig(signals.SignalSlowTest),
	)
	r := computeSlowTestShare(snap)
	if r.Evidence != EvidenceWeak {
		t.Errorf("slow_test_share evidence without runtime = %q, want 'weak'", r.Evidence)
	}

	snap.TestFiles[0].RuntimeStats = &models.RuntimeStats{AvgRuntimeMs: 500}
	r = computeSlowTestShare(snap)
	if r.Evidence != EvidenceStrong {
		t.Errorf("slow_test_share evidence with runtime = %q, want 'strong'", r.Evidence)
	}
}

// --- coverage depth tests ---

func TestCoverageDepth_UncoveredExports(t *testing.T) {
	snap := makeSnap(5)
	snap.CodeUnits = []models.CodeUnit{
		{Name: "Foo", Exported: true},
		{Name: "Bar", Exported: true},
		{Name: "baz", Exported: false},
	}
	snap.Signals = []models.Signal{
		sig(signals.SignalUntestedExport),
	}

	r := computeUncoveredExports(snap)
	// 1 untested out of 2 exported = 0.5
	if r.Value != 0.5 {
		t.Errorf("uncovered_exports value = %v, want 0.5", r.Value)
	}
}

func TestCoverageDepth_UncoveredExports_NoExports(t *testing.T) {
	snap := makeSnap(5)
	r := computeUncoveredExports(snap)
	if r.Value != 0 || r.Band != "strong" {
		t.Errorf("uncovered_exports with no exports: %+v", r)
	}
}

func TestCoverageDepth_WeakAssertionShare(t *testing.T) {
	snap := makeSnap(4,
		sig(signals.SignalWeakAssertion),
		sig(signals.SignalWeakAssertion),
	)
	r := computeWeakAssertionShare(snap)
	if r.Value != 0.5 {
		t.Errorf("weak_assertion_share value = %v, want 0.5", r.Value)
	}
}

func TestCoverageDepth_CoverageBreachShare_NoCovData(t *testing.T) {
	snap := makeSnap(5)
	r := computeCoverageBreachShare(snap)
	if r.Evidence != EvidenceWeak {
		t.Errorf("coverage_breach_share without coverage data evidence = %q, want 'weak'", r.Evidence)
	}
}

func TestCoverageDepth_CoverageBreachShare_WithBreaches(t *testing.T) {
	snap := makeSnap(10,
		sig(signals.SignalCoverageThresholdBreak),
		sig(signals.SignalCoverageThresholdBreak),
	)
	r := computeCoverageBreachShare(snap)
	if r.Value != 0.2 {
		t.Errorf("coverage_breach_share value = %v, want 0.2", r.Value)
	}
	if r.Evidence != EvidenceStrong {
		t.Errorf("coverage_breach_share evidence = %q, want 'strong'", r.Evidence)
	}
}

// --- coverage diversity tests ---

func TestCoverageDiversity_MockHeavyShare(t *testing.T) {
	snap := makeSnap(5,
		sig(signals.SignalMockHeavyTest),
		sig(signals.SignalMockHeavyTest),
	)
	r := computeMockHeavyShare(snap)
	if r.Value != 0.4 {
		t.Errorf("mock_heavy_share value = %v, want 0.4", r.Value)
	}
}

func TestCoverageDiversity_FrameworkFragmentation(t *testing.T) {
	snap := makeSnap(10)
	snap.Frameworks = []models.Framework{
		{Name: "jest", FileCount: 6},
		{Name: "mocha", FileCount: 3},
		{Name: "vitest", FileCount: 1},
	}

	r := computeFrameworkFragmentation(snap)
	if r.Value != 0.3 {
		t.Errorf("framework_fragmentation value = %v, want 0.3", r.Value)
	}
	if r.Band != "moderate" {
		t.Errorf("framework_fragmentation band = %q, want 'moderate'", r.Band)
	}
}

func TestCoverageDiversity_FrameworkFragmentation_NoFrameworks(t *testing.T) {
	snap := makeSnap(5)
	r := computeFrameworkFragmentation(snap)
	if r.Value != 0 || r.Evidence != EvidenceNone {
		t.Errorf("unexpected result for no frameworks: %+v", r)
	}
}

func TestCoverageDiversity_E2EConcentration(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "a.test.js", Framework: "cypress"},
			{Path: "b.test.js", Framework: "cypress"},
			{Path: "c.test.js", Framework: "jest"},
			{Path: "d.test.js", Framework: "jest"},
		},
		Frameworks: []models.Framework{
			{Name: "cypress", Type: models.FrameworkTypeE2E, FileCount: 2},
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 2},
		},
	}

	r := computeE2EConcentration(snap)
	if r.Value != 0.5 {
		t.Errorf("e2e_concentration value = %v, want 0.5", r.Value)
	}
	if r.Band != "strong" {
		t.Errorf("e2e_concentration band = %q, want 'strong'", r.Band)
	}
}

func TestCoverageDiversity_E2EConcentration_HighRatio(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "a.test.js", Framework: "playwright"},
			{Path: "b.test.js", Framework: "playwright"},
			{Path: "c.test.js", Framework: "playwright"},
			{Path: "d.test.js", Framework: "playwright"},
			{Path: "e.test.js", Framework: "jest"},
		},
		Frameworks: []models.Framework{
			{Name: "playwright", Type: models.FrameworkTypeE2E, FileCount: 4},
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 1},
		},
	}

	r := computeE2EConcentration(snap)
	if r.Value != 0.8 {
		t.Errorf("e2e_concentration value = %v, want 0.8", r.Value)
	}
	if r.Band != "moderate" {
		t.Errorf("e2e_concentration band = %q, want 'moderate'", r.Band)
	}
}

// --- structural risk tests ---

func TestStructuralRisk_MigrationBlockerDensity(t *testing.T) {
	snap := makeSnap(10,
		sig(signals.SignalMigrationBlocker),
		sig(signals.SignalDeprecatedTestPattern),
	)
	r := computeMigrationBlockerDensity(snap)
	if r.Value != 0.2 {
		t.Errorf("migration_blocker_density value = %v, want 0.2", r.Value)
	}
}

func TestStructuralRisk_DeprecatedPatternShare(t *testing.T) {
	snap := makeSnap(20,
		sig(signals.SignalDeprecatedTestPattern),
		sig(signals.SignalDeprecatedTestPattern),
		sig(signals.SignalDeprecatedTestPattern),
	)
	r := computeDeprecatedPatternShare(snap)
	if r.Value != 0.15 {
		t.Errorf("deprecated_pattern_share value = %v, want 0.15", r.Value)
	}
	if r.Band != "moderate" {
		t.Errorf("deprecated_pattern_share band = %q, want 'moderate'", r.Band)
	}
}

func TestStructuralRisk_DynamicGenerationShare(t *testing.T) {
	snap := makeSnap(10,
		sig(signals.SignalDynamicTestGeneration),
	)
	r := computeDynamicGenerationShare(snap)
	if r.Value != 0.1 {
		t.Errorf("dynamic_generation_share value = %v, want 0.1", r.Value)
	}
	if r.Band != "moderate" {
		t.Errorf("dynamic_generation_share band = %q, want 'moderate'", r.Band)
	}
}

// --- operational risk tests ---

func TestOperationalRisk_PolicyViolationDensity(t *testing.T) {
	snap := makeSnap(10,
		sig(signals.SignalPolicyViolation),
		sig(signals.SignalPolicyViolation),
	)
	r := computePolicyViolationDensity(snap)
	if r.Value != 0.2 {
		t.Errorf("policy_violation_density value = %v, want 0.2", r.Value)
	}
}

func TestOperationalRisk_LegacyFrameworkShare(t *testing.T) {
	snap := makeSnap(10,
		sig(signals.SignalLegacyFrameworkUsage),
	)
	r := computeLegacyFrameworkShare(snap)
	if r.Value != 0.1 {
		t.Errorf("legacy_framework_share value = %v, want 0.1", r.Value)
	}
}

func TestOperationalRisk_RuntimeBudgetBreachShare(t *testing.T) {
	snap := makeSnap(5,
		sig(signals.SignalRuntimeBudgetExceeded),
	)
	r := computeRuntimeBudgetBreachShare(snap)
	if r.Value != 0.2 {
		t.Errorf("runtime_budget_breach_share value = %v, want 0.2", r.Value)
	}
	if r.Evidence != EvidenceWeak {
		t.Errorf("runtime_budget_breach_share evidence without runtime = %q, want 'weak'", r.Evidence)
	}
}

// --- posture computation tests ---

func TestComputeSnapshot_AllDimensionsPresent(t *testing.T) {
	r := DefaultRegistry()
	snap := makeSnap(10,
		sig(signals.SignalFlakyTest),
		sig(signals.SignalWeakAssertion),
		sig(signals.SignalMigrationBlocker),
		sig(signals.SignalPolicyViolation),
	)

	ms := r.ComputeSnapshot(snap)
	if len(ms.Posture) != 5 {
		t.Errorf("expected 5 posture dimensions, got %d", len(ms.Posture))
	}
	if len(ms.Measurements) != r.Len() {
		t.Errorf("expected %d measurements, got %d", r.Len(), len(ms.Measurements))
	}
}

func TestPosture_StrongWhenClean(t *testing.T) {
	r := DefaultRegistry()
	snap := makeSnap(10)

	ms := r.ComputeSnapshot(snap)
	for _, p := range ms.Posture {
		// Coverage dimensions may be unknown if no coverage data, others should be strong.
		if p.Dimension == DimensionHealth {
			// Health with no runtime data → evidence is weak → capped at moderate.
			if p.Band != PostureModerate && p.Band != PostureStrong {
				t.Errorf("health posture = %q, want strong or moderate", p.Band)
			}
		}
	}
}

func TestPosture_WeakWhenManyIssues(t *testing.T) {
	sigs := make([]models.Signal, 0)
	for i := 0; i < 8; i++ {
		sigs = append(sigs, sig(signals.SignalFlakyTest))
	}
	snap := makeSnap(10, sigs...)

	r := DefaultRegistry()
	ms := r.ComputeSnapshot(snap)

	for _, p := range ms.Posture {
		if p.Dimension == DimensionHealth {
			if p.Band == PostureStrong {
				t.Errorf("health posture should not be strong with 80%% flaky: %q", p.Band)
			}
		}
	}
}

func TestResolvePostureBand(t *testing.T) {
	tests := []struct {
		name  string
		bands []string
		want  PostureBand
	}{
		{"empty", nil, PostureUnknown},
		{"all strong", []string{"strong", "strong"}, PostureStrong},
		{"one weak", []string{"strong", "weak"}, PostureWeak},
		{"majority weak escalates", []string{"weak", "weak", "strong"}, PostureWeak},
		{"critical dominates", []string{"strong", "critical"}, PostureCritical},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolvePostureBand(tt.bands)
			if got != tt.want {
				t.Errorf("resolvePostureBand(%v) = %q, want %q", tt.bands, got, tt.want)
			}
		})
	}
}

// --- integration test ---

func TestFullPipeline_EndToEnd(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "src/auth/auth.test.js", Framework: "jest", RuntimeStats: &models.RuntimeStats{AvgRuntimeMs: 150}},
			{Path: "src/api/api.test.js", Framework: "jest"},
			{Path: "e2e/login.spec.js", Framework: "cypress"},
			{Path: "e2e/checkout.spec.js", Framework: "cypress"},
		},
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 2},
			{Name: "cypress", Type: models.FrameworkTypeE2E, FileCount: 2},
		},
		CodeUnits: []models.CodeUnit{
			{Name: "AuthService", Exported: true},
			{Name: "ApiClient", Exported: true},
		},
		Signals: []models.Signal{
			sigInFile(signals.SignalWeakAssertion, "src/api/api.test.js"),
			sigInFile(signals.SignalMockHeavyTest, "src/auth/auth.test.js"),
			sigInFile(signals.SignalSlowTest, "e2e/login.spec.js"),
			sigInFile(signals.SignalUntestedExport, "src/api/api.test.js"),
		},
	}

	r := DefaultRegistry()
	ms := r.ComputeSnapshot(snap)

	// Every measurement should have a non-empty ID.
	for _, m := range ms.Measurements {
		if m.ID == "" {
			t.Error("measurement has empty ID")
		}
		if m.Explanation == "" {
			t.Error("measurement has empty explanation")
		}
	}

	// Every posture dimension should be present.
	if len(ms.Posture) != 5 {
		t.Errorf("expected 5 posture dimensions, got %d", len(ms.Posture))
	}

	for _, p := range ms.Posture {
		if p.Band == "" {
			t.Errorf("posture dimension %q has empty band", p.Dimension)
		}
		if p.Explanation == "" {
			t.Errorf("posture dimension %q has empty explanation", p.Dimension)
		}
	}
}
