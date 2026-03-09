package portfolio

import (
	"testing"

	"github.com/pmclSF/hamlet/internal/models"
)

// makeSnap creates a minimal snapshot for testing.
func makeSnap(opts ...func(*models.TestSuiteSnapshot)) *models.TestSuiteSnapshot {
	snap := &models.TestSuiteSnapshot{
		Repository: models.RepositoryMetadata{Name: "test-repo"},
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit},
			{Name: "cypress", Type: models.FrameworkTypeE2E},
		},
	}
	for _, opt := range opts {
		opt(snap)
	}
	return snap
}

func withTestFiles(files ...models.TestFile) func(*models.TestSuiteSnapshot) {
	return func(snap *models.TestSuiteSnapshot) {
		snap.TestFiles = files
	}
}

func withCodeUnits(units ...models.CodeUnit) func(*models.TestSuiteSnapshot) {
	return func(snap *models.TestSuiteSnapshot) {
		snap.CodeUnits = units
	}
}

func withOwnership(m map[string][]string) func(*models.TestSuiteSnapshot) {
	return func(snap *models.TestSuiteSnapshot) {
		snap.Ownership = m
	}
}

// --- BuildAssets tests ---

func TestBuildAssets_Empty(t *testing.T) {
	snap := makeSnap()
	assets := BuildAssets(snap)
	if len(assets) != 0 {
		t.Fatalf("expected 0 assets, got %d", len(assets))
	}
}

func TestBuildAssets_Basic(t *testing.T) {
	snap := makeSnap(
		withTestFiles(
			models.TestFile{Path: "test/a.test.js", Framework: "jest", TestCount: 3},
			models.TestFile{Path: "test/b.test.js", Framework: "cypress", TestCount: 5},
		),
	)
	assets := BuildAssets(snap)
	if len(assets) != 2 {
		t.Fatalf("expected 2 assets, got %d", len(assets))
	}

	// Sorted by path.
	if assets[0].Path != "test/a.test.js" {
		t.Errorf("expected first asset path test/a.test.js, got %s", assets[0].Path)
	}
	if assets[0].TestType != "unit" {
		t.Errorf("expected test type unit, got %s", assets[0].TestType)
	}
	if assets[1].TestType != "e2e" {
		t.Errorf("expected test type e2e, got %s", assets[1].TestType)
	}
}

func TestBuildAssets_WithRuntime(t *testing.T) {
	snap := makeSnap(
		withTestFiles(
			models.TestFile{
				Path:      "test/slow.test.js",
				Framework: "jest",
				TestCount: 2,
				RuntimeStats: &models.RuntimeStats{
					AvgRuntimeMs: 15000,
					PassRate:     0.8,
					RetryRate:    0.2,
				},
			},
		),
	)
	assets := BuildAssets(snap)
	if len(assets) != 1 {
		t.Fatalf("expected 1 asset, got %d", len(assets))
	}
	a := assets[0]
	if !a.HasRuntimeData {
		t.Error("expected HasRuntimeData=true")
	}
	if a.RuntimeMs != 15000 {
		t.Errorf("expected RuntimeMs=15000, got %f", a.RuntimeMs)
	}
	if a.CostClass != CostHigh {
		t.Errorf("expected CostClass=high, got %s", a.CostClass)
	}
}

func TestBuildAssets_WithCoverage(t *testing.T) {
	snap := makeSnap(
		withTestFiles(
			models.TestFile{
				Path:            "test/auth.test.js",
				Framework:       "jest",
				TestCount:       5,
				LinkedCodeUnits: []string{"AuthService", "UserService", "CacheManager"},
			},
		),
		withCodeUnits(
			models.CodeUnit{Name: "AuthService", Path: "src/auth/service.js", Exported: true, Owner: "team-a"},
			models.CodeUnit{Name: "UserService", Path: "src/user/service.js", Exported: true, Owner: "team-a"},
			models.CodeUnit{Name: "CacheManager", Path: "src/cache/manager.js", Exported: true, Owner: "team-b"},
		),
	)
	assets := BuildAssets(snap)
	if len(assets) != 1 {
		t.Fatalf("expected 1 asset, got %d", len(assets))
	}
	a := assets[0]
	if !a.HasCoverageData {
		t.Error("expected HasCoverageData=true")
	}
	if a.CoveredUnitCount != 3 {
		t.Errorf("expected CoveredUnitCount=3, got %d", a.CoveredUnitCount)
	}
	if a.ExportedUnitsCovered != 3 {
		t.Errorf("expected ExportedUnitsCovered=3, got %d", a.ExportedUnitsCovered)
	}
	if a.BreadthClass != BreadthBroad {
		t.Errorf("expected BreadthClass=broad, got %s", a.BreadthClass)
	}
}

// --- classifyCost tests ---

func TestClassifyCost_HighRuntime(t *testing.T) {
	a := TestAsset{HasRuntimeData: true, RuntimeMs: 12000}
	if got := classifyCost(a); got != CostHigh {
		t.Errorf("expected high, got %s", got)
	}
}

func TestClassifyCost_ModerateRuntime(t *testing.T) {
	a := TestAsset{HasRuntimeData: true, RuntimeMs: 5000}
	if got := classifyCost(a); got != CostModerate {
		t.Errorf("expected moderate, got %s", got)
	}
}

func TestClassifyCost_LowRuntime(t *testing.T) {
	a := TestAsset{HasRuntimeData: true, RuntimeMs: 500}
	if got := classifyCost(a); got != CostLow {
		t.Errorf("expected low, got %s", got)
	}
}

func TestClassifyCost_HighRetryRate(t *testing.T) {
	a := TestAsset{HasRuntimeData: true, RuntimeMs: 100, RetryRate: 0.35}
	if got := classifyCost(a); got != CostHigh {
		t.Errorf("expected high, got %s", got)
	}
}

func TestClassifyCost_NoRuntime_E2E(t *testing.T) {
	a := TestAsset{TestType: "e2e"}
	if got := classifyCost(a); got != CostHigh {
		t.Errorf("expected high for e2e without runtime, got %s", got)
	}
}

func TestClassifyCost_NoRuntime_Unit(t *testing.T) {
	a := TestAsset{TestType: "unit"}
	if got := classifyCost(a); got != CostUnknown {
		t.Errorf("expected unknown for unit without runtime, got %s", got)
	}
}

// --- Redundancy detection tests ---

func TestDetectRedundancy_NoOverlap(t *testing.T) {
	assets := []TestAsset{
		{Path: "a.test.js", HasCoverageData: true, CoveredUnitCount: 2, CoveredModules: []string{"mod1"}},
		{Path: "b.test.js", HasCoverageData: true, CoveredUnitCount: 2, CoveredModules: []string{"mod2"}},
	}
	findings := detectRedundancy(assets)
	if len(findings) != 0 {
		t.Errorf("expected 0 redundancy findings, got %d", len(findings))
	}
}

func TestDetectRedundancy_HighOverlap(t *testing.T) {
	assets := []TestAsset{
		{Path: "a.test.js", HasCoverageData: true, CoveredUnitCount: 3, CoveredModules: []string{"mod1", "mod2", "mod3"}},
		{Path: "b.test.js", HasCoverageData: true, CoveredUnitCount: 3, CoveredModules: []string{"mod1", "mod2", "mod3"}},
	}
	findings := detectRedundancy(assets)
	if len(findings) != 1 {
		t.Fatalf("expected 1 redundancy finding, got %d", len(findings))
	}
	if findings[0].Type != FindingRedundancyCandidate {
		t.Errorf("expected type redundancy_candidate, got %s", findings[0].Type)
	}
	if findings[0].Confidence != ConfidenceHigh {
		t.Errorf("expected high confidence for 100%% overlap, got %s", findings[0].Confidence)
	}
}

func TestDetectRedundancy_NoCoverageData(t *testing.T) {
	assets := []TestAsset{
		{Path: "a.test.js", HasCoverageData: false},
		{Path: "b.test.js", HasCoverageData: false},
	}
	findings := detectRedundancy(assets)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings without coverage data, got %d", len(findings))
	}
}

// --- Overbroad detection tests ---

func TestDetectOverbroad_NotBroad(t *testing.T) {
	assets := []TestAsset{
		{Path: "a.test.js", HasCoverageData: true, BreadthClass: BreadthNarrow, CoveredModules: []string{"mod1"}},
	}
	findings := detectOverbroad(assets)
	if len(findings) != 0 {
		t.Errorf("expected 0 overbroad findings for narrow test, got %d", len(findings))
	}
}

func TestDetectOverbroad_BroadMultiModule(t *testing.T) {
	assets := []TestAsset{
		{
			Path:           "a.test.js",
			HasCoverageData: true,
			BreadthClass:   BreadthBroad,
			CoveredModules: []string{"mod1", "mod2", "mod3", "mod4", "mod5"},
			OwnersCovered:  []string{"team-a", "team-b", "team-c"},
		},
	}
	findings := detectOverbroad(assets)
	if len(findings) != 1 {
		t.Fatalf("expected 1 overbroad finding, got %d", len(findings))
	}
	if findings[0].Confidence != ConfidenceHigh {
		t.Errorf("expected high confidence for 5 modules, got %s", findings[0].Confidence)
	}
}

// --- Low-value-high-cost detection tests ---

func TestDetectLowValueHighCost_SlowAndNarrow(t *testing.T) {
	assets := []TestAsset{
		{
			Path:             "slow.test.js",
			HasRuntimeData:   true,
			HasCoverageData:  true,
			RuntimeMs:        15000,
			RetryRate:        0.1,
			CostClass:        CostHigh,
			BreadthClass:     BreadthNarrow,
			CoveredUnitCount: 1,
		},
	}
	findings := detectLowValueHighCost(assets)
	if len(findings) != 1 {
		t.Fatalf("expected 1 low-value finding, got %d", len(findings))
	}
	if findings[0].Confidence != ConfidenceHigh {
		t.Errorf("expected high confidence with both data, got %s", findings[0].Confidence)
	}
}

func TestDetectLowValueHighCost_NotTriggered(t *testing.T) {
	assets := []TestAsset{
		{
			Path:             "good.test.js",
			HasRuntimeData:   true,
			RuntimeMs:        500,
			CostClass:        CostLow,
			BreadthClass:     BreadthModerate,
			CoveredUnitCount: 5,
		},
	}
	findings := detectLowValueHighCost(assets)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for low-cost test, got %d", len(findings))
	}
}

// --- High-leverage detection tests ---

func TestDetectHighLeverage_Efficient(t *testing.T) {
	assets := []TestAsset{
		{
			Path:                 "efficient.test.js",
			HasCoverageData:      true,
			HasRuntimeData:       true,
			CostClass:            CostLow,
			BreadthClass:         BreadthModerate,
			ExportedUnitsCovered: 5,
			CoveredModules:       []string{"mod1", "mod2"},
			RuntimeMs:            200,
		},
	}
	findings := detectHighLeverage(assets)
	if len(findings) != 1 {
		t.Fatalf("expected 1 high-leverage finding, got %d", len(findings))
	}
	if findings[0].Confidence != ConfidenceHigh {
		t.Errorf("expected high confidence with both data, got %s", findings[0].Confidence)
	}
}

func TestDetectHighLeverage_HighCostExcluded(t *testing.T) {
	assets := []TestAsset{
		{
			Path:                 "expensive.test.js",
			HasCoverageData:      true,
			CostClass:            CostHigh,
			BreadthClass:         BreadthBroad,
			ExportedUnitsCovered: 10,
		},
	}
	findings := detectHighLeverage(assets)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for high-cost test, got %d", len(findings))
	}
}

func TestDetectHighLeverage_TooFewExports(t *testing.T) {
	assets := []TestAsset{
		{
			Path:                 "small.test.js",
			HasCoverageData:      true,
			CostClass:            CostLow,
			BreadthClass:         BreadthModerate,
			ExportedUnitsCovered: 1,
		},
	}
	findings := detectHighLeverage(assets)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for <3 exported units, got %d", len(findings))
	}
}

// --- Analyze integration tests ---

func TestAnalyze_Empty(t *testing.T) {
	snap := makeSnap()
	summary := Analyze(snap)
	if summary.Aggregates.TotalAssets != 0 {
		t.Errorf("expected 0 assets, got %d", summary.Aggregates.TotalAssets)
	}
	if len(summary.Findings) != 0 {
		t.Errorf("expected 0 findings, got %d", len(summary.Findings))
	}
}

func TestAnalyze_Nil(t *testing.T) {
	summary := Analyze(nil)
	if summary == nil {
		t.Fatal("expected non-nil summary for nil snapshot")
	}
}

func TestAnalyze_WithAssets(t *testing.T) {
	snap := makeSnap(
		withTestFiles(
			models.TestFile{Path: "test/a.test.js", Framework: "jest", TestCount: 3},
			models.TestFile{Path: "test/b.test.js", Framework: "jest", TestCount: 5},
		),
	)
	summary := Analyze(snap)
	if summary.Aggregates.TotalAssets != 2 {
		t.Errorf("expected 2 assets, got %d", summary.Aggregates.TotalAssets)
	}
}

func TestAnalyze_FindingsSorted(t *testing.T) {
	snap := makeSnap(
		withTestFiles(
			models.TestFile{
				Path:            "test/a.test.js",
				Framework:       "jest",
				TestCount:       3,
				LinkedCodeUnits: []string{"Unit1", "Unit2", "Unit3"},
				RuntimeStats:    &models.RuntimeStats{AvgRuntimeMs: 15000, RetryRate: 0.4},
			},
		),
		withCodeUnits(
			models.CodeUnit{Name: "Unit1", Path: "src/mod1/unit1.js", Exported: true},
			models.CodeUnit{Name: "Unit2", Path: "src/mod2/unit2.js", Exported: true},
			models.CodeUnit{Name: "Unit3", Path: "src/mod3/unit3.js", Exported: true},
		),
	)
	summary := Analyze(snap)

	// Findings should be sorted by type then path.
	for i := 1; i < len(summary.Findings); i++ {
		prev := summary.Findings[i-1]
		curr := summary.Findings[i]
		if prev.Type > curr.Type || (prev.Type == curr.Type && prev.Path > curr.Path) {
			t.Errorf("findings not sorted: %s/%s before %s/%s", prev.Type, prev.Path, curr.Type, curr.Path)
		}
	}
}

// --- Aggregates tests ---

func TestComputeAggregates_RuntimeConcentration(t *testing.T) {
	assets := make([]TestAsset, 10)
	for i := range assets {
		assets[i] = TestAsset{
			Path:           "test.js",
			HasRuntimeData: true,
			RuntimeMs:      100,
		}
	}
	// Make the last 2 assets very expensive.
	assets[8].RuntimeMs = 5000
	assets[9].RuntimeMs = 10000

	agg := computeAggregates(assets, nil, makeSnap())
	if !agg.HasRuntimeData {
		t.Error("expected HasRuntimeData=true")
	}
	// Top 20% (2 assets) = 15000 out of 15800 total ≈ 0.95.
	if agg.RuntimeConcentration < 0.90 {
		t.Errorf("expected high runtime concentration, got %f", agg.RuntimeConcentration)
	}
}

func TestComputeAggregates_FindingCounts(t *testing.T) {
	findings := []Finding{
		{Type: FindingRedundancyCandidate},
		{Type: FindingRedundancyCandidate},
		{Type: FindingOverbroad},
		{Type: FindingLowValueHighCost},
		{Type: FindingHighLeverage},
		{Type: FindingHighLeverage},
		{Type: FindingHighLeverage},
	}
	agg := computeAggregates(nil, findings, makeSnap())
	if agg.RedundancyCandidateCount != 2 {
		t.Errorf("expected 2 redundancy, got %d", agg.RedundancyCandidateCount)
	}
	if agg.OverbroadCount != 1 {
		t.Errorf("expected 1 overbroad, got %d", agg.OverbroadCount)
	}
	if agg.LowValueHighCostCount != 1 {
		t.Errorf("expected 1 low-value, got %d", agg.LowValueHighCostCount)
	}
	if agg.HighLeverageCount != 3 {
		t.Errorf("expected 3 high-leverage, got %d", agg.HighLeverageCount)
	}
}

// --- Owner aggregation tests ---

func TestComputeOwnerAggregates(t *testing.T) {
	assets := []TestAsset{
		{Path: "a.test.js", Owner: "team-a", RuntimeMs: 100},
		{Path: "b.test.js", Owner: "team-a", RuntimeMs: 200},
		{Path: "c.test.js", Owner: "team-b", RuntimeMs: 300},
	}
	findings := []Finding{
		{Type: FindingRedundancyCandidate, Owner: "team-a"},
		{Type: FindingHighLeverage, Owner: "team-b"},
	}
	result := computeOwnerAggregates(assets, findings)
	if len(result) != 2 {
		t.Fatalf("expected 2 owners, got %d", len(result))
	}
	// Sorted by owner name.
	if result[0].Owner != "team-a" {
		t.Errorf("expected first owner team-a, got %s", result[0].Owner)
	}
	if result[0].AssetCount != 2 {
		t.Errorf("expected 2 assets for team-a, got %d", result[0].AssetCount)
	}
	if result[0].RedundancyCandidateCount != 1 {
		t.Errorf("expected 1 redundancy for team-a, got %d", result[0].RedundancyCandidateCount)
	}
	if result[1].HighLeverageCount != 1 {
		t.Errorf("expected 1 high-leverage for team-b, got %d", result[1].HighLeverageCount)
	}
}

// --- ToModel conversion tests ---

func TestToModel_Nil(t *testing.T) {
	var s *PortfolioSummary
	if s.ToModel() != nil {
		t.Error("expected nil model for nil summary")
	}
}

func TestToModel_Basic(t *testing.T) {
	summary := &PortfolioSummary{
		Assets: []TestAsset{
			{Path: "a.test.js", CostClass: CostLow, BreadthClass: BreadthNarrow},
		},
		Findings: []Finding{
			{Type: FindingHighLeverage, Path: "a.test.js", Confidence: ConfidenceHigh},
		},
		Aggregates: PortfolioAggregates{TotalAssets: 1, HighLeverageCount: 1},
	}
	model := summary.ToModel()
	if model == nil {
		t.Fatal("expected non-nil model")
	}
	if len(model.Assets) != 1 {
		t.Errorf("expected 1 asset, got %d", len(model.Assets))
	}
	if model.Assets[0].CostClass != "low" {
		t.Errorf("expected cost class low, got %s", model.Assets[0].CostClass)
	}
	if len(model.Findings) != 1 {
		t.Errorf("expected 1 finding, got %d", len(model.Findings))
	}
	if model.Findings[0].Confidence != "high" {
		t.Errorf("expected confidence high, got %s", model.Findings[0].Confidence)
	}
}

// --- Benchmark aggregate tests ---

func TestBuildBenchmarkAggregate_Nil(t *testing.T) {
	if BuildBenchmarkAggregate(nil) != nil {
		t.Error("expected nil for nil summary")
	}
}

func TestBuildBenchmarkAggregate_Empty(t *testing.T) {
	s := &PortfolioSummary{}
	if BuildBenchmarkAggregate(s) != nil {
		t.Error("expected nil for zero assets")
	}
}

func TestBuildBenchmarkAggregate_Basic(t *testing.T) {
	s := &PortfolioSummary{
		Aggregates: PortfolioAggregates{
			TotalAssets:              100,
			RuntimeConcentration:     0.75,
			RedundancyCandidateCount: 5,
			OverbroadCount:           3,
			LowValueHighCostCount:    2,
			HighLeverageCount:        15,
		},
	}
	ba := BuildBenchmarkAggregate(s)
	if ba == nil {
		t.Fatal("expected non-nil benchmark aggregate")
	}
	if ba.RuntimeConcentrationBand != "concentrated" {
		t.Errorf("expected concentrated, got %s", ba.RuntimeConcentrationBand)
	}
	if ba.RedundancyCandidateShareBand != "low" {
		t.Errorf("expected low, got %s", ba.RedundancyCandidateShareBand)
	}
	if ba.PortfolioPostureBand != "moderate" {
		t.Errorf("expected moderate posture, got %s", ba.PortfolioPostureBand)
	}
}

// --- Portfolio posture tests ---

func TestComputePortfolioPosture_Strong(t *testing.T) {
	s := &PortfolioSummary{
		Aggregates: PortfolioAggregates{
			TotalAssets:              100,
			RedundancyCandidateCount: 1,
			OverbroadCount:           1,
			LowValueHighCostCount:    0,
		},
	}
	if got := computePortfolioPosture(s); got != "strong" {
		t.Errorf("expected strong, got %s", got)
	}
}

func TestComputePortfolioPosture_Critical(t *testing.T) {
	s := &PortfolioSummary{
		Aggregates: PortfolioAggregates{
			TotalAssets:              10,
			RedundancyCandidateCount: 3,
			OverbroadCount:           2,
			LowValueHighCostCount:    2,
		},
	}
	if got := computePortfolioPosture(s); got != "critical" {
		t.Errorf("expected critical, got %s", got)
	}
}

// --- Band helper tests ---

func TestConcentrationBand(t *testing.T) {
	tests := []struct {
		ratio float64
		want  string
	}{
		{0, "unknown"},
		{0.40, "balanced"},
		{0.60, "moderate"},
		{0.80, "concentrated"},
		{0.95, "highly_concentrated"},
	}
	for _, tt := range tests {
		got := concentrationBand(tt.ratio)
		if got != tt.want {
			t.Errorf("concentrationBand(%f) = %s, want %s", tt.ratio, got, tt.want)
		}
	}
}

func TestShareBand(t *testing.T) {
	tests := []struct {
		ratio float64
		want  string
	}{
		{0.01, "minimal"},
		{0.05, "low"},
		{0.15, "moderate"},
		{0.50, "high"},
	}
	for _, tt := range tests {
		got := shareBand(tt.ratio)
		if got != tt.want {
			t.Errorf("shareBand(%f) = %s, want %s", tt.ratio, got, tt.want)
		}
	}
}

// --- Owner resolution tests ---

func TestResolveOwner_FromTestFile(t *testing.T) {
	tf := models.TestFile{Path: "test.js", Owner: "team-direct"}
	got := resolveOwner(tf, map[string]string{"test.js": "team-fallback"})
	if got != "team-direct" {
		t.Errorf("expected team-direct, got %s", got)
	}
}

func TestResolveOwner_FromOwnership(t *testing.T) {
	tf := models.TestFile{Path: "test.js"}
	got := resolveOwner(tf, map[string]string{"test.js": "team-ownership"})
	if got != "team-ownership" {
		t.Errorf("expected team-ownership, got %s", got)
	}
}

func TestResolveOwner_Unknown(t *testing.T) {
	tf := models.TestFile{Path: "test.js"}
	got := resolveOwner(tf, map[string]string{})
	if got != "unknown" {
		t.Errorf("expected unknown, got %s", got)
	}
}

// --- inferTestType tests ---

func TestInferTestType(t *testing.T) {
	fwTypes := map[string]models.FrameworkType{
		"jest":       models.FrameworkTypeUnit,
		"cypress":    models.FrameworkTypeE2E,
		"supertest":  models.FrameworkTypeIntegration,
	}
	tests := []struct {
		framework string
		want      string
	}{
		{"jest", "unit"},
		{"cypress", "e2e"},
		{"supertest", "integration"},
		{"unknown-fw", "unknown"},
	}
	for _, tt := range tests {
		tf := models.TestFile{Framework: tt.framework}
		got := inferTestType(tf, fwTypes)
		if got != tt.want {
			t.Errorf("inferTestType(%s) = %s, want %s", tt.framework, got, tt.want)
		}
	}
}
