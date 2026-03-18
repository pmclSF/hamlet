package changescope

import (
	"sort"
	"strings"
)

// DeduplicateFindings removes duplicate findings based on a composite key of
// type + severity + path + normalized explanation. Keeps the first occurrence.
func DeduplicateFindings(findings []ChangeScopedFinding) []ChangeScopedFinding {
	seen := map[string]bool{}
	var out []ChangeScopedFinding
	for _, f := range findings {
		key := findingKey(f)
		if !seen[key] {
			seen[key] = true
			out = append(out, f)
		}
	}
	return out
}

// findingKey produces a deduplication key from the finding's identifying fields.
func findingKey(f ChangeScopedFinding) string {
	// Normalize explanation: lowercase, strip trailing punctuation and whitespace.
	norm := strings.ToLower(strings.TrimSpace(f.Explanation))
	norm = strings.TrimRight(norm, ".")

	return f.Type + "|" + f.Severity + "|" + f.Path + "|" + norm
}

// ClassifyFindings separates findings into three categories:
//   - directRisk: protection gaps on directly changed files
//   - indirectRisk: protection gaps on transitively impacted files
//   - existing: pre-existing signals on changed files
func ClassifyFindings(findings []ChangeScopedFinding) (newRisk, existing []ChangeScopedFinding) {
	for _, f := range findings {
		if f.Type == "existing_signal" {
			existing = append(existing, f)
		} else {
			newRisk = append(newRisk, f)
		}
	}
	return
}

// ClassifyFindingsDetailed provides a 3-way classification for the markdown renderer.
func ClassifyFindingsDetailed(findings []ChangeScopedFinding) (directRisk, indirectRisk, existing []ChangeScopedFinding) {
	for _, f := range findings {
		switch {
		case f.Type == "existing_signal":
			existing = append(existing, f)
		case f.Scope == "indirect":
			indirectRisk = append(indirectRisk, f)
		default:
			directRisk = append(directRisk, f)
		}
	}
	return
}

// GroupTestsByPackage groups test file paths by their parent directory (package).
// Returns groups sorted by count descending.
func GroupTestsByPackage(paths []string) []TestGroup {
	groups := map[string][]string{}
	for _, p := range paths {
		pkg := parentDir(p)
		groups[pkg] = append(groups[pkg], p)
	}

	var out []TestGroup
	for pkg, files := range groups {
		sort.Strings(files)
		out = append(out, TestGroup{Package: pkg, Files: files, Count: len(files)})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Package < out[j].Package
	})
	return out
}

// TestGroup is a group of test files in the same package/directory.
type TestGroup struct {
	Package string
	Files   []string
	Count   int
}

func parentDir(path string) string {
	idx := strings.LastIndex(path, "/")
	if idx < 0 {
		return "."
	}
	return path[:idx]
}

// SummarizeFindingsBySeverity counts findings by severity.
func SummarizeFindingsBySeverity(findings []ChangeScopedFinding) map[string]int {
	counts := map[string]int{}
	for _, f := range findings {
		counts[f.Severity]++
	}
	return counts
}

// MergeRecommendation computes a human-readable merge recommendation.
func MergeRecommendation(postureBand string, findings []ChangeScopedFinding) (recommendation, explanation string) {
	highCount := 0
	for _, f := range findings {
		if f.Severity == "high" {
			highCount++
		}
	}

	switch {
	case postureBand == "well_protected" && highCount == 0:
		return "Safe to merge", "All changed code is well protected by existing tests."
	case postureBand == "evidence_limited":
		return "Informational only", "Insufficient data to assess change risk confidently."
	case highCount > 0:
		return "Merge with caution", "High-severity gaps found in changed code."
	case postureBand == "weakly_protected" || postureBand == "high_risk":
		return "Merge blocked", "Significant protection gaps in changed code require attention."
	default:
		return "Merge with caution", "Some protection gaps found. Review findings before merging."
	}
}
