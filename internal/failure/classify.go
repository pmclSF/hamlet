package failure

import (
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// categoryPattern defines a pattern that matches a failure category.
type categoryPattern struct {
	category   FailureCategory
	patterns   []string
	confidence ClassificationConfidence
	score      float64
	// priority controls precedence when multiple categories match.
	// Lower values take priority.
	priority int
}

// classificationPatterns defines the pattern rules ordered by priority.
// More specific categories (snapshot, selector) are checked before
// broader ones (assertion) to avoid misclassification.
var classificationPatterns = []categoryPattern{
	{
		category: CategorySnapshotMismatch,
		patterns: []string{
			"snapshot",
			"tomatchsnapshot",
			"tomatchinlinesnapshot",
			"inline snapshot",
			"snapshot mismatch",
			"snapshot changed",
			"obsolete snapshot",
		},
		confidence: ConfidenceExact,
		score:      0.95,
		priority:   1,
	},
	{
		category: CategorySelectorUI,
		patterns: []string{
			"selector",
			"element not found",
			"not visible",
			"detached",
			"stale element",
			"stale element reference",
			"nosuchelement",
			"element is not attached",
			"element not interactable",
			"elementnotinteractableexception",
			"waitfortimeout",
			"waiting for selector",
			"cy.get()",
			"page.$()",
			"queryselector",
		},
		confidence: ConfidenceExact,
		score:      0.90,
		priority:   2,
	},
	{
		category: CategoryInfraEnvironment,
		patterns: []string{
			"enomem",
			"out of memory",
			"enospc",
			"disk space",
			"permission denied",
			"eacces",
			"segmentation fault",
			"sigsegv",
			"sigkill",
			"oom killed",
			"cannot allocate memory",
			"no space left on device",
			"eperm",
		},
		confidence: ConfidenceExact,
		score:      0.95,
		priority:   3,
	},
	{
		category: CategoryDependencyService,
		patterns: []string{
			"econnrefused",
			"connection refused",
			"503",
			"502",
			"service unavailable",
			"enotfound",
			"econnreset",
			"econnaborted",
			"socket hang up",
			"dns resolution failed",
			"getaddrinfo",
			"network error",
			"fetch failed",
			"request failed with status",
			"ehostunreach",
		},
		confidence: ConfidenceExact,
		score:      0.90,
		priority:   4,
	},
	{
		category: CategoryTimeout,
		patterns: []string{
			"timeout",
			"timed out",
			"exceeded",
			"etimedout",
			"deadline exceeded",
			"async callback was not invoked within",
			"jest.settimeout",
			"exceeded timeout",
			"operation timed out",
		},
		confidence: ConfidenceExact,
		score:      0.90,
		priority:   5,
	},
	{
		category: CategorySetupFixture,
		patterns: []string{
			"beforeeach",
			"beforeall",
			"setup",
			"setup failed",
			"fixture",
			"aftereach",
			"afterall",
			"teardown",
			"@before",
			"@beforeclass",
			"setupmodule",
			"setupfilesafterframework",
			"test setup",
			"initialization failed",
		},
		confidence: ConfidenceInferred,
		score:      0.75,
		priority:   6,
	},
	{
		category: CategoryAssertionFailure,
		patterns: []string{
			"expect(",
			"assert",
			"tobe",
			"toequal",
			"should",
			"expected",
			"tobetruthy",
			"tobefalsy",
			"tocontain",
			"tohavebeencalled",
			"assertionerror",
			"assertequals",
			"assertthat",
			"asserttrue",
			"assertfalse",
			"received:",
			"expected:",
			"to equal",
			"to be",
			"to match",
			"deepstrictequal",
		},
		confidence: ConfidenceInferred,
		score:      0.80,
		priority:   7,
	},
}

// Classify takes a slice of failure inputs and produces a TaxonomyResult
// with all failures classified into categories.
func Classify(inputs []FailureInput) *TaxonomyResult {
	result := &TaxonomyResult{
		ByCategory: make(map[FailureCategory]int),
	}

	for _, input := range inputs {
		c := classifyOne(input)
		result.Classifications = append(result.Classifications, c)
		result.ByCategory[c.Category]++
	}

	result.TotalFailures = len(result.Classifications)
	result.DominantCategory = dominantCategory(result.ByCategory)

	// Sort for determinism: by file path, then test name.
	sort.Slice(result.Classifications, func(i, j int) bool {
		ci, cj := result.Classifications[i], result.Classifications[j]
		if ci.TestFilePath != cj.TestFilePath {
			return ci.TestFilePath < cj.TestFilePath
		}
		return ci.TestName < cj.TestName
	})

	return result
}

// classifyOne classifies a single failure input.
func classifyOne(input FailureInput) FailureClassification {
	c := FailureClassification{
		TestFilePath: input.TestFilePath,
		TestName:     input.TestName,
		ErrorMessage: input.ErrorMessage,
		StackTrace:   input.StackTrace,
	}

	// Combine error message and stack trace for pattern matching.
	combined := strings.ToLower(input.ErrorMessage + " " + input.StackTrace)

	if combined == " " || strings.TrimSpace(combined) == "" {
		c.Category = CategoryUnknown
		c.Confidence = ConfidenceWeak
		c.ConfidenceScore = 0.1
		c.Explanation = "no error message or stack trace available"
		return c
	}

	// Try each pattern set in priority order.
	var bestMatch *categoryPattern
	for i := range classificationPatterns {
		cp := &classificationPatterns[i]
		if matchesAny(combined, cp.patterns) {
			if bestMatch == nil || cp.priority < bestMatch.priority {
				bestMatch = cp
			}
		}
	}

	if bestMatch != nil {
		c.Category = bestMatch.category
		c.Confidence = bestMatch.confidence
		c.ConfidenceScore = bestMatch.score
		c.Explanation = explanationFor(bestMatch.category, input.ErrorMessage)
		return c
	}

	// No pattern matched.
	c.Category = CategoryUnknown
	c.Confidence = ConfidenceWeak
	c.ConfidenceScore = 0.2
	c.Explanation = "error message did not match any known failure pattern"
	return c
}

// matchesAny returns true if the text contains any of the patterns.
func matchesAny(text string, patterns []string) bool {
	for _, p := range patterns {
		if matchPattern(text, p) {
			return true
		}
	}
	return false
}

func matchPattern(text, pattern string) bool {
	if pattern == "" {
		return false
	}
	if !isWordPattern(pattern) {
		return strings.Contains(text, pattern)
	}

	start := 0
	for {
		i := strings.Index(text[start:], pattern)
		if i < 0 {
			return false
		}
		i += start
		j := i + len(pattern)
		beforeBoundary := i == 0 || !isWordByte(text[i-1])
		afterBoundary := j >= len(text) || !isWordByte(text[j])
		if beforeBoundary && afterBoundary {
			return true
		}
		start = i + len(pattern)
		if start >= len(text) {
			return false
		}
	}
}

func isWordPattern(p string) bool {
	for _, r := range p {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			continue
		}
		return false
	}
	return p != ""
}

func isWordByte(b byte) bool {
	r, _ := utf8.DecodeRune([]byte{b})
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

// explanationFor returns a human-readable explanation for a category.
func explanationFor(cat FailureCategory, errorMsg string) string {
	msg := strings.TrimSpace(errorMsg)
	if msg == "" {
		msg = "(no error message)"
	}

	switch cat {
	case CategoryAssertionFailure:
		return "error message contains assertion keywords: " + msg
	case CategoryTimeout:
		return "error message indicates a timeout: " + msg
	case CategorySetupFixture:
		return "error message indicates a setup or fixture failure: " + msg
	case CategoryDependencyService:
		return "error message indicates a dependency or service failure: " + msg
	case CategorySnapshotMismatch:
		return "error message indicates a snapshot mismatch: " + msg
	case CategorySelectorUI:
		return "error message indicates selector or UI fragility: " + msg
	case CategoryInfraEnvironment:
		return "error message indicates an infrastructure or environment issue: " + msg
	default:
		return "could not determine failure category: " + msg
	}
}

// dominantCategory returns the category with the highest count.
func dominantCategory(byCat map[FailureCategory]int) FailureCategory {
	if len(byCat) == 0 {
		return CategoryUnknown
	}

	best := CategoryUnknown
	bestCount := 0
	for cat, count := range byCat {
		if count > bestCount {
			best = cat
			bestCount = count
		}
	}
	return best
}
