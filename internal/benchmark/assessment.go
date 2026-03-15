package benchmark

import (
	"encoding/json"
	"fmt"
	"strings"
)

// CommandAssessment scores a single command execution.
type CommandAssessment struct {
	Repo             string   `json:"repo"`
	Command          string   `json:"command"`
	Success          bool     `json:"success"`
	RuntimeMs        int64    `json:"runtimeMs"`
	OutputNonEmpty   bool     `json:"outputNonEmpty"`
	ParsedJSON       bool     `json:"parsedJson"`
	ExpectedSections []string `json:"expectedSections"`
	MissingSections  []string `json:"missingSections"`
	WarningFlags     []string `json:"warningFlags"`
	CredibilityScore int      `json:"credibilityScore"`
	Notes            []string `json:"notes"`
}

// RepoAssessment holds all assessments for one repo.
type RepoAssessment struct {
	Repo         RepoMeta            `json:"repo"`
	Assessments  []CommandAssessment `json:"assessments"`
	OverallScore int                 `json:"overallScore"`
}

// sectionCheck defines a named pattern check for command output.
type sectionCheck struct {
	name     string
	patterns []string
}

// expectedSections defines what output sections to look for per command.
var expectedSections = map[string][]sectionCheck{
	"analyze": {
		{name: "tests detected", patterns: []string{"testcases", "testfiles", "testcount", "test_count", "suites"}},
		{name: "repo profile", patterns: []string{"repoprofile", "repo_profile", "volume"}},
		{name: "coverage confidence", patterns: []string{"coverageconfidence", "coverage_confidence"}},
		{name: "duplicates", patterns: []string{"duplicatetests", "duplicatecount", "redundan"}},
		{name: "fanout", patterns: []string{"highfanout", "high-fanout", "fan-out"}},
		{name: "weak coverage", patterns: []string{"weakcoverage", "weak_coverage", "uncovered"}},
	},
	"impact": {
		{name: "changed files", patterns: []string{"changedfiles", "changed_files"}},
		{name: "impacted tests", patterns: []string{"impactedtests", "selectedtests", "\"impacted\""}},
		{name: "coverage confidence", patterns: []string{"confidence", "coverageconfidence"}},
		{name: "risk signal", patterns: []string{"risklevel", "risk_level", "severity", "\"risk\""}},
		{name: "dependency reasoning", patterns: []string{"reasonchain", "reason_chain", "\"reason\"", "dependency"}},
	},
	"insights": {
		{name: "duplicate clusters", patterns: []string{"duplicatecluster", "\"cluster\"", "redundan"}},
		{name: "high-fanout nodes", patterns: []string{"highfanout", "fan-out", "high-fanout"}},
		{name: "weak coverage", patterns: []string{"weakcoverage", "weak_coverage", "uncovered"}},
		{name: "recommendations", patterns: []string{"recommend", "suggestion", "action", "insight"}},
	},
	"explain": {
		{name: "test identifier", patterns: []string{"testid", "testId", "canonicalidentity", "canonicalIdentity"}},
		{name: "test location", patterns: []string{"filepath", "filePath", "line"}},
		{name: "confidence", patterns: []string{"confidence"}},
		{name: "framework", patterns: []string{"framework"}},
	},
	"depgraph:stats": {
		{name: "node count", patterns: []string{"nodecount", "nodesByType"}},
		{name: "edge count", patterns: []string{"edgecount", "edgesByType"}},
		{name: "graph density", patterns: []string{"density"}},
	},
	"depgraph:coverage": {
		{name: "coverage sources", patterns: []string{"sources", "sourcecount", "sourceCount"}},
		{name: "coverage bands", patterns: []string{"band", "bandcount", "bandCounts"}},
	},
	"depgraph:fanout": {
		{name: "fanout entries", patterns: []string{"entries", "nodecount", "nodeCount"}},
		{name: "fanout threshold", patterns: []string{"threshold", "flagged", "flaggedcount", "flaggedCount"}},
	},
	"depgraph:duplicates": {
		{name: "duplicate clusters", patterns: []string{"clusters", "duplicatecount", "duplicateCount"}},
		{name: "tests analyzed", patterns: []string{"testsanalyzed", "testsAnalyzed"}},
	},
	"debug:graph": {
		{name: "node count", patterns: []string{"nodecount", "nodesByType"}},
		{name: "edge count", patterns: []string{"edgecount", "edgesByType"}},
		{name: "graph density", patterns: []string{"density"}},
	},
	"debug:coverage": {
		{name: "coverage sources", patterns: []string{"sources", "sourcecount", "sourceCount"}},
		{name: "coverage bands", patterns: []string{"band", "bandcount", "bandCounts"}},
	},
	"debug:fanout": {
		{name: "fanout entries", patterns: []string{"entries", "nodecount", "nodeCount"}},
		{name: "fanout threshold", patterns: []string{"threshold", "flagged", "flaggedcount", "flaggedCount"}},
	},
	"debug:duplicates": {
		{name: "duplicate clusters", patterns: []string{"clusters", "duplicatecount", "duplicateCount"}},
		{name: "tests analyzed", patterns: []string{"testsanalyzed", "testsAnalyzed"}},
	},
}

// AssessCommand scores a single command result.
func AssessCommand(cr CommandResult) CommandAssessment {
	a := CommandAssessment{
		Command:   cr.Command,
		Repo:      cr.RepoName,
		RuntimeMs: cr.RuntimeMs,
	}

	// Base: command succeeded.
	if cr.ExitCode == 0 && cr.Error == "" {
		a.Success = true
		a.CredibilityScore += 20
		a.Notes = append(a.Notes, "command succeeded")
	} else {
		a.Notes = append(a.Notes, fmt.Sprintf("command failed (exit %d)", cr.ExitCode))
		if cr.Error != "" {
			a.Notes = append(a.Notes, "error: "+cr.Error)
		}
		a.CredibilityScore -= 50
	}

	// Output non-empty.
	stdout := strings.TrimSpace(cr.Stdout)
	if len(stdout) > 0 {
		a.OutputNonEmpty = true
		a.CredibilityScore += 10
	} else {
		a.CredibilityScore -= 20
		a.Notes = append(a.Notes, "output is empty")
	}

	// Trivially short output.
	if len(stdout) > 0 && len(stdout) < 20 {
		a.WarningFlags = append(a.WarningFlags, "trivially short output")
		a.CredibilityScore -= 10
	}
	if cr.StdoutTruncated {
		a.WarningFlags = append(a.WarningFlags, fmt.Sprintf("stdout truncated (%d bytes)", cr.StdoutBytes))
		a.Notes = append(a.Notes, "stdout was truncated for storage")
	}
	if cr.StderrTruncated {
		a.WarningFlags = append(a.WarningFlags, fmt.Sprintf("stderr truncated (%d bytes)", cr.StderrBytes))
		a.Notes = append(a.Notes, "stderr was truncated for storage")
	}
	if cr.TimedOut {
		a.WarningFlags = append(a.WarningFlags, "command timed out")
		a.Notes = append(a.Notes, "command exceeded timeout")
		a.CredibilityScore -= 20
	}

	// Try to parse as JSON.
	if len(stdout) > 0 {
		var parsed interface{}
		if err := json.Unmarshal([]byte(stdout), &parsed); err == nil {
			a.ParsedJSON = true
			a.CredibilityScore += 10
		}
	}

	// Check stderr for warnings.
	if strings.TrimSpace(cr.Stderr) != "" {
		stderrLines := strings.Split(strings.TrimSpace(cr.Stderr), "\n")
		for _, line := range stderrLines {
			line = strings.TrimSpace(line)
			if line != "" {
				a.WarningFlags = append(a.WarningFlags, "stderr: "+truncate(line, 100))
				a.CredibilityScore -= 5
			}
		}
	}

	// Check expected sections.
	checks, ok := expectedSections[cr.Command]
	if !ok {
		return clamp(&a)
	}

	lowerOutput := strings.ToLower(stdout)
	for _, check := range checks {
		found := false
		for _, pattern := range check.patterns {
			if strings.Contains(lowerOutput, strings.ToLower(pattern)) {
				found = true
				break
			}
		}
		if found {
			a.ExpectedSections = append(a.ExpectedSections, check.name)
			a.CredibilityScore += 10
			a.Notes = append(a.Notes, check.name+" present")
		} else {
			a.MissingSections = append(a.MissingSections, check.name)
			a.Notes = append(a.Notes, check.name+" missing")
		}
	}

	return clamp(&a)
}

// AssessResults produces assessments for all command results in a benchmark.
func AssessResults(br BenchResult) RepoAssessment {
	ra := RepoAssessment{
		Repo: br.Repo,
	}

	totalScore := 0
	for _, cr := range br.Commands {
		assessment := AssessCommand(cr)
		ra.Assessments = append(ra.Assessments, assessment)
		totalScore += assessment.CredibilityScore
	}

	if len(ra.Assessments) > 0 {
		ra.OverallScore = totalScore / len(ra.Assessments)
	}

	return ra
}

func clamp(a *CommandAssessment) CommandAssessment {
	if a.CredibilityScore < 0 {
		a.CredibilityScore = 0
	}
	if a.CredibilityScore > 100 {
		a.CredibilityScore = 100
	}
	return *a
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
