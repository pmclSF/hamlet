package coverage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// IngestFile reads a coverage artifact file and returns normalized records.
// Supports LCOV and Istanbul JSON formats (auto-detected).
func IngestFile(artifactPath string, runLabel string) (*CoverageArtifact, error) {
	data, err := os.ReadFile(artifactPath)
	if err != nil {
		return nil, fmt.Errorf("reading coverage artifact: %w", err)
	}

	content := string(data)
	trimmed := strings.TrimSpace(content)

	var records []CoverageRecord
	var format string

	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		// Try Istanbul JSON.
		records, err = parseIstanbul(data)
		if err != nil {
			return nil, fmt.Errorf("parsing istanbul coverage: %w", err)
		}
		format = "istanbul"
	} else {
		// Try LCOV.
		records, err = parseLCOV(content)
		if err != nil {
			return nil, fmt.Errorf("parsing lcov coverage: %w", err)
		}
		format = "lcov"
	}

	return &CoverageArtifact{
		Records:  records,
		RunLabel: runLabel,
		Provenance: ArtifactProvenance{
			SourceFile: artifactPath,
			Format:     format,
			RunLabel:   runLabel,
		},
	}, nil
}

// IngestDirectory scans a directory for coverage artifacts and ingests them.
func IngestDirectory(dir string, runLabel string) ([]CoverageArtifact, error) {
	var artifacts []CoverageArtifact
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := strings.ToLower(e.Name())
		if isCoverageFile(name) {
			art, err := IngestFile(filepath.Join(dir, e.Name()), runLabel)
			if err != nil {
				continue // Skip unreadable artifacts.
			}
			artifacts = append(artifacts, *art)
		}
	}
	return artifacts, nil
}

func isCoverageFile(name string) bool {
	return name == "lcov.info" ||
		strings.HasSuffix(name, ".lcov") ||
		name == "coverage-final.json" ||
		name == "coverage-summary.json" ||
		(strings.HasPrefix(name, "coverage") && strings.HasSuffix(name, ".json"))
}

// --- LCOV parser ---

func parseLCOV(content string) ([]CoverageRecord, error) {
	var records []CoverageRecord
	var current *CoverageRecord

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "SF:") {
			filePath := strings.TrimPrefix(line, "SF:")
			filePath = normalizeCoveragePath(filePath)
			current = &CoverageRecord{
				FilePath:     filePath,
				LineHits:     map[int]int{},
				BranchHits:   map[string]int{},
				FunctionHits: map[string]int{},
			}
			continue
		}

		if current == nil {
			continue
		}

		switch {
		case strings.HasPrefix(line, "DA:"):
			// DA:line_number,hit_count
			parts := strings.SplitN(strings.TrimPrefix(line, "DA:"), ",", 2)
			if len(parts) == 2 {
				lineNum, _ := strconv.Atoi(parts[0])
				hits, _ := strconv.Atoi(parts[1])
				if lineNum > 0 {
					current.LineHits[lineNum] = hits
				}
			}

		case strings.HasPrefix(line, "FN:"):
			// FN:line_number,function_name
			parts := strings.SplitN(strings.TrimPrefix(line, "FN:"), ",", 2)
			if len(parts) == 2 {
				fnName := parts[1]
				if _, ok := current.FunctionHits[fnName]; !ok {
					current.FunctionHits[fnName] = 0
				}
			}

		case strings.HasPrefix(line, "FNDA:"):
			// FNDA:hit_count,function_name
			parts := strings.SplitN(strings.TrimPrefix(line, "FNDA:"), ",", 2)
			if len(parts) == 2 {
				hits, _ := strconv.Atoi(parts[0])
				fnName := parts[1]
				current.FunctionHits[fnName] = hits
			}

		case strings.HasPrefix(line, "BRDA:"):
			// BRDA:line,block,branch,hits
			parts := strings.SplitN(strings.TrimPrefix(line, "BRDA:"), ",", 4)
			if len(parts) == 4 {
				key := parts[0] + ":" + parts[1] + ":" + parts[2]
				hits := 0
				if parts[3] != "-" {
					hits, _ = strconv.Atoi(parts[3])
				}
				current.BranchHits[key] = hits
			}

		case line == "end_of_record":
			if current != nil {
				recomputeCounts(current)
				records = append(records, *current)
				current = nil
			}
		}
	}

	// Handle missing end_of_record.
	if current != nil {
		recomputeCounts(current)
		records = append(records, *current)
	}

	return records, nil
}

// --- Istanbul JSON parser ---

// istanbulEntry represents one file's coverage in Istanbul JSON format.
type istanbulEntry struct {
	Path          string                    `json:"path"`
	StatementMap  map[string]istanbulRange  `json:"statementMap"`
	S             map[string]int            `json:"s"`
	FnMap         map[string]istanbulFn     `json:"fnMap"`
	F             map[string]int            `json:"f"`
	BranchMap     map[string]istanbulBranch `json:"branchMap"`
	B             map[string][]int          `json:"b"`
}

type istanbulRange struct {
	Start istanbulPos `json:"start"`
	End   istanbulPos `json:"end"`
}

type istanbulPos struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

type istanbulFn struct {
	Name string        `json:"name"`
	Loc  istanbulRange `json:"loc"`
}

type istanbulBranch struct {
	Type      string          `json:"type"`
	Locations []istanbulRange `json:"locations"`
}

func parseIstanbul(data []byte) ([]CoverageRecord, error) {
	var entries map[string]istanbulEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	// Sort keys for deterministic output order.
	keys := make([]string, 0, len(entries))
	for k := range entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var records []CoverageRecord
	for _, k := range keys {
		entry := entries[k]
		rec := CoverageRecord{
			FilePath:     normalizeCoveragePath(entry.Path),
			LineHits:     map[int]int{},
			BranchHits:   map[string]int{},
			FunctionHits: map[string]int{},
		}

		// Statement coverage → line hits.
		for stmtID, hits := range entry.S {
			if loc, ok := entry.StatementMap[stmtID]; ok {
				rec.LineHits[loc.Start.Line] = hits
			}
		}

		// Function coverage.
		for fnID, hits := range entry.F {
			if fn, ok := entry.FnMap[fnID]; ok {
				name := fn.Name
				if name == "" {
					name = fmt.Sprintf("anonymous_%s", fnID)
				}
				rec.FunctionHits[name] = hits
			}
		}

		// Branch coverage.
		for brID, hits := range entry.B {
			for i, h := range hits {
				key := fmt.Sprintf("%s:%d", brID, i)
				rec.BranchHits[key] = h
			}
		}

		recomputeCounts(&rec)
		records = append(records, rec)
	}

	return records, nil
}

// normalizeCoveragePath normalizes a file path from coverage artifacts.
// Strips absolute path prefixes and normalizes separators.
func normalizeCoveragePath(p string) string {
	p = filepath.ToSlash(p)
	p = strings.ReplaceAll(p, "\\", "/")

	// Strip common absolute path prefixes.
	// Look for typical project root markers.
	markers := []string{"/src/", "/lib/", "/app/", "/packages/"}
	for _, m := range markers {
		if idx := strings.Index(p, m); idx > 0 {
			return p[idx+1:]
		}
	}

	// Strip leading / if still absolute.
	p = strings.TrimPrefix(p, "/")
	return p
}
