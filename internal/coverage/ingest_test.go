package coverage

import (
	"os"
	"path/filepath"
	"testing"
)

const sampleLCOV = `SF:src/utils.js
FN:1,formatDate
FN:10,parseDate
FNDA:5,formatDate
FNDA:0,parseDate
DA:1,5
DA:2,5
DA:3,5
DA:10,0
DA:11,0
BRDA:2,0,0,5
BRDA:2,0,1,0
LF:5
LH:3
end_of_record
SF:src/api.js
FN:1,fetchData
FNDA:3,fetchData
DA:1,3
DA:2,3
DA:3,0
LF:3
LH:2
end_of_record
`

func TestParseLCOV(t *testing.T) {
	t.Parallel()
	records, err := parseLCOV(sampleLCOV)
	if err != nil {
		t.Fatalf("parseLCOV error: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}

	// First record: src/utils.js
	r := records[0]
	if r.FilePath != "src/utils.js" {
		t.Errorf("path = %q, want src/utils.js", r.FilePath)
	}
	if r.LineCoveredCount != 3 {
		t.Errorf("LineCoveredCount = %d, want 3", r.LineCoveredCount)
	}
	if r.LineTotalCount != 5 {
		t.Errorf("LineTotalCount = %d, want 5", r.LineTotalCount)
	}
	if r.FunctionHits["formatDate"] != 5 {
		t.Errorf("formatDate hits = %d, want 5", r.FunctionHits["formatDate"])
	}
	if r.FunctionHits["parseDate"] != 0 {
		t.Errorf("parseDate hits = %d, want 0", r.FunctionHits["parseDate"])
	}
	if r.FunctionCoveredCount != 1 {
		t.Errorf("FunctionCoveredCount = %d, want 1", r.FunctionCoveredCount)
	}
	if r.BranchTotalCount != 2 {
		t.Errorf("BranchTotalCount = %d, want 2", r.BranchTotalCount)
	}
	if r.BranchCoveredCount != 1 {
		t.Errorf("BranchCoveredCount = %d, want 1", r.BranchCoveredCount)
	}
}

const sampleIstanbul = `{
  "/project/src/math.js": {
    "path": "/project/src/math.js",
    "statementMap": {
      "0": {"start": {"line": 1, "column": 0}, "end": {"line": 1, "column": 30}},
      "1": {"start": {"line": 2, "column": 0}, "end": {"line": 2, "column": 20}},
      "2": {"start": {"line": 5, "column": 0}, "end": {"line": 5, "column": 30}}
    },
    "s": {"0": 5, "1": 5, "2": 0},
    "fnMap": {
      "0": {"name": "add", "loc": {"start": {"line": 1, "column": 0}, "end": {"line": 3, "column": 1}}},
      "1": {"name": "subtract", "loc": {"start": {"line": 5, "column": 0}, "end": {"line": 7, "column": 1}}}
    },
    "f": {"0": 5, "1": 0},
    "branchMap": {
      "0": {"type": "if", "locations": [
        {"start": {"line": 2, "column": 0}, "end": {"line": 2, "column": 10}},
        {"start": {"line": 2, "column": 11}, "end": {"line": 2, "column": 20}}
      ]}
    },
    "b": {"0": [5, 0]}
  }
}`

func TestParseIstanbul(t *testing.T) {
	t.Parallel()
	records, err := parseIstanbul([]byte(sampleIstanbul))
	if err != nil {
		t.Fatalf("parseIstanbul error: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	r := records[0]
	if r.FilePath != "src/math.js" {
		t.Errorf("path = %q, want src/math.js", r.FilePath)
	}
	if r.FunctionHits["add"] != 5 {
		t.Errorf("add hits = %d, want 5", r.FunctionHits["add"])
	}
	if r.FunctionHits["subtract"] != 0 {
		t.Errorf("subtract hits = %d, want 0", r.FunctionHits["subtract"])
	}
	if r.FunctionCoveredCount != 1 {
		t.Errorf("FunctionCoveredCount = %d, want 1", r.FunctionCoveredCount)
	}
	if r.BranchTotalCount != 2 {
		t.Errorf("BranchTotalCount = %d, want 2", r.BranchTotalCount)
	}
}

func TestIngestFile_LCOV(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	lcovPath := filepath.Join(dir, "lcov.info")
	if err := os.WriteFile(lcovPath, []byte(sampleLCOV), 0644); err != nil {
		t.Fatalf("write lcov.info: %v", err)
	}

	art, err := IngestFile(lcovPath, "unit")
	if err != nil {
		t.Fatalf("IngestFile error: %v", err)
	}
	if art.Provenance.Format != "lcov" {
		t.Errorf("format = %q, want lcov", art.Provenance.Format)
	}
	if art.RunLabel != "unit" {
		t.Errorf("runLabel = %q, want unit", art.RunLabel)
	}
	if len(art.Records) != 2 {
		t.Errorf("expected 2 records, got %d", len(art.Records))
	}
}

func TestIngestFile_Istanbul(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "coverage-final.json")
	if err := os.WriteFile(p, []byte(sampleIstanbul), 0644); err != nil {
		t.Fatalf("write coverage-final.json: %v", err)
	}

	art, err := IngestFile(p, "e2e")
	if err != nil {
		t.Fatalf("IngestFile error: %v", err)
	}
	if art.Provenance.Format != "istanbul" {
		t.Errorf("format = %q, want istanbul", art.Provenance.Format)
	}
	if art.RunLabel != "e2e" {
		t.Errorf("runLabel = %q, want e2e", art.RunLabel)
	}
}

func TestMerge(t *testing.T) {
	t.Parallel()
	a1 := CoverageArtifact{
		Records: []CoverageRecord{
			{FilePath: "src/a.js", LineHits: map[int]int{1: 3, 2: 0}, LineTotalCount: 2, LineCoveredCount: 1},
		},
		Provenance: ArtifactProvenance{Format: "lcov", RunLabel: "unit"},
	}
	a2 := CoverageArtifact{
		Records: []CoverageRecord{
			{FilePath: "src/a.js", LineHits: map[int]int{1: 0, 2: 5}, LineTotalCount: 2, LineCoveredCount: 1},
		},
		Provenance: ArtifactProvenance{Format: "lcov", RunLabel: "e2e"},
	}

	merged := Merge([]CoverageArtifact{a1, a2})
	rec := merged.ByFile["src/a.js"]
	if rec == nil {
		t.Fatal("expected merged record for src/a.js")
	}
	if rec.LineHits[1] != 3 {
		t.Errorf("line 1 hits = %d, want 3", rec.LineHits[1])
	}
	if rec.LineHits[2] != 5 {
		t.Errorf("line 2 hits = %d, want 5", rec.LineHits[2])
	}
	if rec.LineCoveredCount != 2 {
		t.Errorf("LineCoveredCount = %d, want 2", rec.LineCoveredCount)
	}
}

func TestNormalizeCoveragePath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  string
	}{
		{"/project/src/utils.js", "src/utils.js"},
		{"/home/user/repo/lib/api.js", "lib/api.js"},
		{"src/utils.js", "src/utils.js"},
		{"/absolute/path/without/marker.js", "absolute/path/without/marker.js"},
	}
	for _, tt := range tests {
		got := normalizeCoveragePath(tt.input)
		if got != tt.want {
			t.Errorf("normalizeCoveragePath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestIngestDirectory_PartialWarning(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "lcov.info"), []byte(sampleLCOV), 0o644); err != nil {
		t.Fatalf("write lcov.info: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "coverage-summary.json"), []byte("{not-json"), 0o644); err != nil {
		t.Fatalf("write invalid json: %v", err)
	}

	arts, err := IngestDirectory(dir, "unit")
	if len(arts) == 0 {
		t.Fatal("expected at least one ingested artifact")
	}
	if err == nil {
		t.Fatal("expected partial ingest warning, got nil")
	}
	if _, ok := err.(*IngestWarning); !ok {
		t.Fatalf("expected *IngestWarning, got %T (%v)", err, err)
	}
}

func TestIngestDirectory_AllInvalidReturnsError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "coverage-final.json"), []byte("{bad-json"), 0o644); err != nil {
		t.Fatalf("write invalid coverage json: %v", err)
	}

	arts, err := IngestDirectory(dir, "unit")
	if err == nil {
		t.Fatal("expected error when no artifacts could be ingested")
	}
	if len(arts) != 0 {
		t.Fatalf("expected 0 artifacts when all inputs invalid, got %d", len(arts))
	}
}
