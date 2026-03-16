package truthcheck

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WriteReport writes the truth check report to the given output directory.
// Creates report.json and report.md.
func WriteReport(outputDir string, report *TruthCheckReport) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("creating output dir: %w", err)
	}

	// JSON report.
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling report: %w", err)
	}
	jsonPath := filepath.Join(outputDir, "report.json")
	if err := os.WriteFile(jsonPath, jsonData, 0o644); err != nil {
		return fmt.Errorf("writing JSON report: %w", err)
	}

	// Markdown report.
	md := generateMarkdown(report)
	mdPath := filepath.Join(outputDir, "report.md")
	if err := os.WriteFile(mdPath, []byte(md), 0o644); err != nil {
		return fmt.Errorf("writing markdown report: %w", err)
	}

	return nil
}

func generateMarkdown(report *TruthCheckReport) string {
	var sb strings.Builder

	sb.WriteString("# Truth Validation Report\n\n")
	sb.WriteString(fmt.Sprintf("**Repo:** `%s`\n", report.RepoRoot))
	sb.WriteString(fmt.Sprintf("**Truth spec:** `%s`\n\n", report.TruthFile))

	// Summary.
	s := report.Summary
	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("| Metric | Value |\n|--------|-------|\n"))
	sb.WriteString(fmt.Sprintf("| Categories | %d |\n", s.TotalCategories))
	sb.WriteString(fmt.Sprintf("| Passed | %d / %d |\n", s.PassedCount, s.TotalCategories))
	sb.WriteString(fmt.Sprintf("| Overall score (F1) | **%.0f%%** |\n", s.OverallScore*100))
	sb.WriteString(fmt.Sprintf("| Overall precision | %.0f%% |\n", s.OverallPrecision*100))
	sb.WriteString(fmt.Sprintf("| Overall recall | %.0f%% |\n", s.OverallRecall*100))
	sb.WriteString("\n")

	// Per-category results.
	sb.WriteString("## Categories\n\n")
	sb.WriteString("| Category | Score | Precision | Recall | Expected | Matched | Missing | Unexpected | Pass |\n")
	sb.WriteString("|----------|-------|-----------|--------|----------|---------|---------|------------|------|\n")
	for _, c := range report.Categories {
		pass := "PASS"
		if !c.Passed {
			pass = "FAIL"
		}
		sb.WriteString(fmt.Sprintf("| %s | %.0f%% | %.0f%% | %.0f%% | %d | %d | %d | %d | %s |\n",
			c.Category, c.Score*100, c.Precision*100, c.Recall*100,
			c.Expected, c.Matched, len(c.Missing), len(c.Unexpected), pass))
	}
	sb.WriteString("\n")

	// Detailed findings per category.
	for _, c := range report.Categories {
		sb.WriteString(fmt.Sprintf("### %s\n\n", c.Category))
		if c.Description != "" {
			sb.WriteString(fmt.Sprintf("> %s\n\n", c.Description))
		}

		if len(c.Missing) > 0 {
			sb.WriteString("**Missing (expected but not found):**\n")
			for _, m := range c.Missing {
				sb.WriteString(fmt.Sprintf("- %s\n", m))
			}
			sb.WriteString("\n")
		}

		if len(c.Unexpected) > 0 {
			sb.WriteString("**Unexpected (found but not expected):**\n")
			for _, u := range c.Unexpected {
				sb.WriteString(fmt.Sprintf("- %s\n", u))
			}
			sb.WriteString("\n")
		}

		if len(c.Details) > 0 {
			sb.WriteString("**Details:**\n")
			for _, d := range c.Details {
				sb.WriteString(fmt.Sprintf("- %s\n", d))
			}
			sb.WriteString("\n")
		}

		sb.WriteString("---\n\n")
	}

	return sb.String()
}
