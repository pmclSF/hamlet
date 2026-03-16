package benchmark

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// WriteResults writes benchmark results and assessments to the output directory.
//
// Output files:
//   - benchmark-results.json — raw command execution data (canonical name)
//   - benchmark-report.md — human-readable markdown report (canonical name)
//   - cli-benchmark-results.json — raw results (legacy alias)
//   - cli-benchmark-assessment.json — credibility scores
//   - cli-benchmark-summary.md — markdown summary (legacy alias)
func WriteResults(outputDir string, results []BenchResult, assessments []RepoAssessment) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("creating output dir: %w", err)
	}

	// Write raw results.
	if err := writeJSON(filepath.Join(outputDir, "benchmark-results.json"), results); err != nil {
		return err
	}
	// Legacy alias.
	if err := writeJSON(filepath.Join(outputDir, "cli-benchmark-results.json"), results); err != nil {
		return err
	}

	// Write assessment.
	if err := writeJSON(filepath.Join(outputDir, "cli-benchmark-assessment.json"), assessments); err != nil {
		return err
	}

	// Write markdown report.
	summary := GenerateSummary(assessments)
	for _, name := range []string{"benchmark-report.md", "cli-benchmark-summary.md"} {
		path := filepath.Join(outputDir, name)
		if err := os.WriteFile(path, []byte(summary), 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", name, err)
		}
	}

	return nil
}

func writeJSON(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON for %s: %w", path, err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}

// GenerateSummary produces a markdown summary from benchmark assessments.
func GenerateSummary(assessments []RepoAssessment) string {
	var sb strings.Builder

	sb.WriteString("# Terrain CLI Benchmark Summary\n\n")

	// Per-repo sections.
	for _, ra := range assessments {
		sb.WriteString(fmt.Sprintf("## Repo: %s\n\n", ra.Repo.Name))

		if ra.Repo.Description != "" {
			sb.WriteString(fmt.Sprintf("> %s\n\n", ra.Repo.Description))
		}

		sb.WriteString("| Property | Value |\n|----------|-------|\n")
		sb.WriteString(fmt.Sprintf("| Type | %s |\n", ra.Repo.Type))
		sb.WriteString(fmt.Sprintf("| Git repo | %v |\n", ra.Repo.IsGitRepo))
		if len(ra.Repo.Languages) > 0 {
			sb.WriteString(fmt.Sprintf("| Languages | %s |\n", strings.Join(ra.Repo.Languages, ", ")))
		}
		sb.WriteString(fmt.Sprintf("| Overall score | **%d** |\n", ra.OverallScore))
		sb.WriteString("\n")

		for _, a := range ra.Assessments {
			sb.WriteString(fmt.Sprintf("### %s\n", a.Command))
			sb.WriteString(fmt.Sprintf("- **success:** %v\n", a.Success))
			sb.WriteString(fmt.Sprintf("- **runtime:** %.1fs\n", float64(a.RuntimeMs)/1000.0))
			sb.WriteString(fmt.Sprintf("- **credibility:** %d\n", a.CredibilityScore))

			if a.ParsedJSON {
				sb.WriteString("- **parsed JSON:** yes\n")
			}

			if len(a.Notes) > 0 {
				sb.WriteString("- **notes:**\n")
				for _, note := range a.Notes {
					sb.WriteString(fmt.Sprintf("  - %s\n", note))
				}
			}

			if len(a.WarningFlags) > 0 {
				sb.WriteString("- **warnings:**\n")
				for _, w := range a.WarningFlags {
					sb.WriteString(fmt.Sprintf("  - %s\n", w))
				}
			}

			sb.WriteString("\n")
		}

		sb.WriteString("---\n\n")
	}

	// Cross-repo summary.
	sb.WriteString("## Cross-Repo Summary\n\n")

	commandScores := map[string][]int{}
	commandFailures := map[string]int{}
	commandMissing := map[string]map[string]int{}

	for _, ra := range assessments {
		for _, a := range ra.Assessments {
			commandScores[a.Command] = append(commandScores[a.Command], a.CredibilityScore)
			if !a.Success {
				commandFailures[a.Command]++
			}
			for _, section := range a.MissingSections {
				if commandMissing[a.Command] == nil {
					commandMissing[a.Command] = map[string]int{}
				}
				commandMissing[a.Command][section]++
			}
		}
	}

	// Command strength ranking.
	type cmdAvg struct {
		name string
		avg  float64
	}
	var cmdAvgs []cmdAvg
	for name, scores := range commandScores {
		total := 0
		for _, s := range scores {
			total += s
		}
		cmdAvgs = append(cmdAvgs, cmdAvg{name: name, avg: float64(total) / float64(len(scores))})
	}
	sort.Slice(cmdAvgs, func(i, j int) bool {
		return cmdAvgs[i].avg > cmdAvgs[j].avg
	})

	sb.WriteString("### Command Strength Ranking\n\n")
	sb.WriteString("| Command | Avg Credibility | Failure Rate |\n|---------|-----------------|-------------|\n")
	for _, ca := range cmdAvgs {
		runs := len(commandScores[ca.name])
		failures := commandFailures[ca.name]
		failRate := "0%"
		if failures > 0 {
			failRate = fmt.Sprintf("%.0f%%", float64(failures)/float64(runs)*100)
		}
		sb.WriteString(fmt.Sprintf("| %s | %.0f | %s (%d/%d) |\n", ca.name, ca.avg, failRate, failures, runs))
	}
	sb.WriteString("\n")

	// Frequently missing sections.
	sb.WriteString("### Frequently Missing Sections\n\n")
	type missingEntry struct {
		command string
		section string
		count   int
	}
	var missingList []missingEntry
	for cmd, sections := range commandMissing {
		for section, count := range sections {
			missingList = append(missingList, missingEntry{cmd, section, count})
		}
	}
	sort.Slice(missingList, func(i, j int) bool {
		return missingList[i].count > missingList[j].count
	})

	if len(missingList) == 0 {
		sb.WriteString("No missing sections detected.\n\n")
	} else {
		sb.WriteString("| Command | Section | Missing in N repos |\n|---------|---------|-------------------|\n")
		for _, m := range missingList {
			sb.WriteString(fmt.Sprintf("| %s | %s | %d |\n", m.command, m.section, m.count))
		}
		sb.WriteString("\n")
	}

	// Repo quality ranking.
	sb.WriteString("### Repo Quality Ranking\n\n")
	type repoScore struct {
		name  string
		score int
	}
	var repoScores []repoScore
	for _, ra := range assessments {
		repoScores = append(repoScores, repoScore{ra.Repo.Name, ra.OverallScore})
	}
	sort.Slice(repoScores, func(i, j int) bool {
		return repoScores[i].score > repoScores[j].score
	})

	sb.WriteString("| Repo | Overall Score |\n|------|---------------|\n")
	for _, rs := range repoScores {
		sb.WriteString(fmt.Sprintf("| %s | %d |\n", rs.name, rs.score))
	}
	sb.WriteString("\n")

	return sb.String()
}
