package engine

import (
	"os"
	"path/filepath"
	"strings"
)

// ArtifactDiscovery holds the results of scanning for coverage and runtime
// artifacts at known locations within a repository. This enables zero-config
// analysis: when the user doesn't specify --coverage or --runtime, Terrain
// can still find and use common artifact locations automatically.
type ArtifactDiscovery struct {
	// CoveragePath is the discovered coverage artifact path, or empty.
	CoveragePath string `json:"coveragePath,omitempty"`

	// CoverageFormat describes the detected format (lcov, istanbul, go-cover).
	CoverageFormat string `json:"coverageFormat,omitempty"`

	// RuntimePaths are the discovered runtime artifact paths.
	RuntimePaths []string `json:"runtimePaths,omitempty"`

	// RuntimeFormats describes each discovered runtime format.
	RuntimeFormats []string `json:"runtimeFormats,omitempty"`

	// CoverageAutoDetected is true when coverage was found by auto-discovery
	// rather than explicit flag.
	CoverageAutoDetected bool `json:"coverageAutoDetected,omitempty"`

	// RuntimeAutoDetected is true when runtime artifacts were found by
	// auto-discovery rather than explicit flag.
	RuntimeAutoDetected bool `json:"runtimeAutoDetected,omitempty"`
}

// coverageCandidates lists known coverage artifact locations in priority order.
// First match wins. These cover the most common CI and local configurations.
var coverageCandidates = []struct {
	path   string
	format string
}{
	// JavaScript/TypeScript — Jest, Vitest, NYC/Istanbul
	{"coverage/lcov.info", "lcov"},
	{"coverage/coverage-final.json", "istanbul"},
	{"coverage/lcov-report/lcov.info", "lcov"},

	// Istanbul JSON alternatives
	{"coverage-final.json", "istanbul"},
	{".nyc_output/coverage-final.json", "istanbul"},

	// LCOV at root
	{"lcov.info", "lcov"},
	{"coverage.lcov", "lcov"},

	// Go coverage
	{"coverage.out", "go-cover"},
	{"cover.out", "go-cover"},
	{"coverage.txt", "go-cover"},

	// Python — pytest-cov, coverage.py
	{"htmlcov/lcov.info", "lcov"},
	{"coverage.xml", "cobertura"},
	{".coverage", "coverage-py"},

	// CI common locations
	{"test-coverage/lcov.info", "lcov"},
	{"reports/coverage/lcov.info", "lcov"},
}

// runtimeCandidates lists known runtime artifact locations in priority order.
var runtimeCandidates = []struct {
	path   string
	format string
}{
	// JUnit XML — produced by most CI systems, Jest (jest-junit), pytest, Go
	{"junit.xml", "junit-xml"},
	{"test-results.xml", "junit-xml"},
	{"reports/junit.xml", "junit-xml"},
	{"junit/junit.xml", "junit-xml"},
	{"test-results/junit.xml", "junit-xml"},
	{"build/test-results/junit.xml", "junit-xml"},

	// Jest/Vitest JSON
	{"test-results.json", "jest-json"},
	{"jest-results.json", "jest-json"},
	{"reports/test-results.json", "jest-json"},

	// Python pytest
	{"report.xml", "junit-xml"},
	{"pytest-results.xml", "junit-xml"},

	// Go test JSON
	{"test-output.json", "go-test-json"},
}

// DiscoverArtifacts scans the repository root for common coverage and runtime
// artifacts. The scan is lightweight — only os.Stat on known paths, no
// directory walking. Returns what was found.
func DiscoverArtifacts(root string) *ArtifactDiscovery {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return &ArtifactDiscovery{}
	}

	d := &ArtifactDiscovery{}

	// Coverage: first match wins.
	for _, c := range coverageCandidates {
		p := filepath.Join(absRoot, c.path)
		if info, err := os.Stat(p); err == nil && !info.IsDir() && info.Size() > 0 {
			d.CoveragePath = p
			d.CoverageFormat = c.format
			break
		}
	}

	// Runtime: collect all matches (user may have multiple artifact files).
	seen := map[string]bool{}
	for _, r := range runtimeCandidates {
		p := filepath.Join(absRoot, r.path)
		if info, err := os.Stat(p); err == nil && !info.IsDir() && info.Size() > 0 {
			if !seen[p] {
				seen[p] = true
				d.RuntimePaths = append(d.RuntimePaths, p)
				d.RuntimeFormats = append(d.RuntimeFormats, r.format)
			}
		}
	}

	return d
}

// ApplyDiscovery merges auto-discovered artifacts into PipelineOptions,
// respecting explicit flag precedence. Returns a description of what was
// applied for user-facing messaging.
//
// Rules:
//   - If opts.CoveragePath is already set (explicit flag), discovery is skipped.
//   - If opts.RuntimePaths is already set, discovery is skipped.
//   - Only high-confidence artifacts (known formats at known paths) are used.
func ApplyDiscovery(opts *PipelineOptions, discovery *ArtifactDiscovery) []string {
	if discovery == nil {
		return nil
	}

	var messages []string

	// Coverage: apply only if not explicitly set.
	if opts.CoveragePath == "" && discovery.CoveragePath != "" {
		opts.CoveragePath = discovery.CoveragePath
		discovery.CoverageAutoDetected = true
		messages = append(messages,
			"Auto-detected coverage: "+relativePath(discovery.CoveragePath)+" ("+discovery.CoverageFormat+")")
	}

	// Runtime: apply only if not explicitly set.
	if len(opts.RuntimePaths) == 0 && len(discovery.RuntimePaths) > 0 {
		// Use only the first runtime artifact to avoid noise.
		opts.RuntimePaths = discovery.RuntimePaths[:1]
		discovery.RuntimeAutoDetected = true
		messages = append(messages,
			"Auto-detected runtime: "+relativePath(discovery.RuntimePaths[0])+" ("+discovery.RuntimeFormats[0]+")")
	}

	return messages
}

// MissingArtifactHints returns user-facing hints for artifacts that were
// not found and not explicitly provided. Returns nil if everything is
// available or explicitly provided.
func MissingArtifactHints(opts *PipelineOptions, discovery *ArtifactDiscovery) []string {
	var hints []string

	if opts.CoveragePath == "" {
		hints = append(hints,
			"Coverage data not found. Provide with --coverage <path> to unlock coverage signals.")
	}

	if len(opts.RuntimePaths) == 0 {
		hints = append(hints,
			"Runtime data not found. Provide with --runtime <path> to unlock health signals.")
	}

	return hints
}

func relativePath(absPath string) string {
	if cwd, err := os.Getwd(); err == nil {
		if rel, err := filepath.Rel(cwd, absPath); err == nil && !strings.HasPrefix(rel, "..") {
			return rel
		}
	}
	return absPath
}
