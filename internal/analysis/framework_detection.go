package analysis

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pmclSF/hamlet/internal/models"
)

// FrameworkResult holds the result of framework detection for a single file.
type FrameworkResult struct {
	Framework  string
	Confidence float64
	Source     string // "import", "config", "fallback"
}

const frameworkProbeBytes = 64 * 1024

// detectFrameworkWithContext detects framework with optional project-level context.
// When projectCtx is provided and per-file detection yields "unknown", the project
// default is used as a fallback.
func detectFrameworkWithContext(relPath string, absPath string, projectCtx *ProjectContext) FrameworkResult {
	ext := strings.ToLower(filepath.Ext(relPath))

	var result FrameworkResult
	switch {
	case isJSExt(ext):
		result = detectJSFrameworkResult(absPath)
	case ext == ".go":
		result = FrameworkResult{Framework: "go-testing", Confidence: 0.99, Source: "convention"}
	case ext == ".py":
		result = detectPythonFrameworkResult(absPath)
	case ext == ".java":
		result = detectJavaFrameworkResult(absPath)
	default:
		result = FrameworkResult{Framework: "unknown", Confidence: 0, Source: ""}
	}

	// Layer 2: project-level fallback for unknown frameworks.
	if result.Framework == "unknown" && projectCtx != nil {
		lang := extToLanguage(ext)
		if fallback := projectCtx.DefaultFramework(lang); fallback != "" {
			result = FrameworkResult{
				Framework:  fallback,
				Confidence: 0.4,
				Source:     "project-fallback",
			}
		}
	}

	return result
}

// extToLanguage maps file extensions to language identifiers for framework fallback.
func extToLanguage(ext string) string {
	switch {
	case isJSExt(ext):
		return "javascript"
	case ext == ".go":
		return "go"
	case ext == ".py":
		return "python"
	case ext == ".java":
		return "java"
	default:
		return ""
	}
}

func isJSExt(ext string) bool {
	switch ext {
	case ".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs", ".mts", ".cts":
		return true
	}
	return false
}

// detectJSFrameworkResult returns a full FrameworkResult with confidence and source.
func detectJSFrameworkResult(absPath string) FrameworkResult {
	content := readHead(absPath, frameworkProbeBytes)
	if content == "" {
		return FrameworkResult{Framework: "unknown", Confidence: 0, Source: ""}
	}

	// E2E frameworks — check first as they are more specific
	if strings.Contains(content, "playwright") || strings.Contains(content, "@playwright/test") {
		return FrameworkResult{Framework: "playwright", Confidence: 0.9, Source: "import"}
	}
	if strings.Contains(content, "cypress") || strings.Contains(content, "cy.") {
		return FrameworkResult{Framework: "cypress", Confidence: 0.9, Source: "import"}
	}
	if strings.Contains(content, "puppeteer") {
		return FrameworkResult{Framework: "puppeteer", Confidence: 0.9, Source: "import"}
	}
	if strings.Contains(content, "webdriverio") || strings.Contains(content, "wdio") {
		return FrameworkResult{Framework: "webdriverio", Confidence: 0.9, Source: "import"}
	}
	if strings.Contains(content, "testcafe") || strings.Contains(content, "TestCafe") {
		return FrameworkResult{Framework: "testcafe", Confidence: 0.9, Source: "import"}
	}

	// Node.js built-in test runner
	if strings.Contains(content, "node:test") {
		return FrameworkResult{Framework: "node-test", Confidence: 0.95, Source: "import"}
	}

	// Unit test frameworks
	if strings.Contains(content, "vitest") || strings.Contains(content, "from 'vitest'") || strings.Contains(content, "from \"vitest\"") {
		return FrameworkResult{Framework: "vitest", Confidence: 0.9, Source: "import"}
	}
	if strings.Contains(content, "from '@jest") || strings.Contains(content, "jest.") || strings.Contains(content, "from 'jest") {
		return FrameworkResult{Framework: "jest", Confidence: 0.85, Source: "import"}
	}
	if strings.Contains(content, "from 'mocha'") || strings.Contains(content, "require('mocha')") {
		return FrameworkResult{Framework: "mocha", Confidence: 0.9, Source: "import"}
	}
	if strings.Contains(content, "from 'jasmine'") || strings.Contains(content, "jasmine.") {
		return FrameworkResult{Framework: "jasmine", Confidence: 0.85, Source: "import"}
	}

	// Mocha detection: done-callback pattern or node assert/supertest usage
	// without explicit jest imports suggests mocha (common in Express, Koa, etc.)
	if strings.Contains(content, "describe(") || strings.Contains(content, "it(") {
		if hasMochaIndicators(content) {
			return FrameworkResult{Framework: "mocha", Confidence: 0.7, Source: "import"}
		}
	}

	// Fallback: common test globals suggest jest (most common JS test framework)
	if strings.Contains(content, "describe(") && strings.Contains(content, "expect(") {
		return FrameworkResult{Framework: "jest", Confidence: 0.5, Source: "import"}
	}

	return FrameworkResult{Framework: "unknown", Confidence: 0, Source: ""}
}

// hasMochaIndicators checks for patterns that suggest mocha rather than jest.
// Mocha projects commonly use done callbacks, node's assert module, or supertest,
// none of which are typical in jest projects.
func hasMochaIndicators(content string) bool {
	// done callback: function(done) or function (done) — idiomatic mocha, discouraged in jest
	if strings.Contains(content, "function(done)") || strings.Contains(content, "function (done)") {
		return true
	}
	// supertest: commonly paired with mocha for HTTP testing
	if strings.Contains(content, "require('supertest')") || strings.Contains(content, "require(\"supertest\")") {
		return true
	}
	// node assert module: common in mocha, not used in jest
	if strings.Contains(content, "require('assert')") || strings.Contains(content, "require(\"assert\")") {
		return true
	}
	// chai: assertion library commonly paired with mocha
	if strings.Contains(content, "require('chai')") || strings.Contains(content, "require(\"chai\")") ||
		strings.Contains(content, "from 'chai'") || strings.Contains(content, "from \"chai\"") {
		return true
	}
	return false
}

// detectPythonFrameworkResult returns a full FrameworkResult.
func detectPythonFrameworkResult(absPath string) FrameworkResult {
	content := readHead(absPath, frameworkProbeBytes)
	if content == "" {
		return FrameworkResult{Framework: "unknown", Confidence: 0, Source: ""}
	}
	if strings.Contains(content, "import pytest") || strings.Contains(content, "from pytest") {
		return FrameworkResult{Framework: "pytest", Confidence: 0.9, Source: "import"}
	}
	if strings.Contains(content, "import unittest") || strings.Contains(content, "from unittest") {
		return FrameworkResult{Framework: "unittest", Confidence: 0.9, Source: "import"}
	}
	if strings.Contains(content, "import nose") || strings.Contains(content, "from nose") {
		return FrameworkResult{Framework: "nose2", Confidence: 0.9, Source: "import"}
	}
	// pytest is conventional default for Python test files.
	return FrameworkResult{Framework: "pytest", Confidence: 0.5, Source: "convention"}
}

// detectJavaFrameworkResult returns a full FrameworkResult.
func detectJavaFrameworkResult(absPath string) FrameworkResult {
	content := readHead(absPath, frameworkProbeBytes)
	if content == "" {
		return FrameworkResult{Framework: "unknown", Confidence: 0, Source: ""}
	}
	if strings.Contains(content, "org.testng") {
		return FrameworkResult{Framework: "testng", Confidence: 0.9, Source: "import"}
	}
	if strings.Contains(content, "org.junit.jupiter") {
		return FrameworkResult{Framework: "junit5", Confidence: 0.9, Source: "import"}
	}
	if strings.Contains(content, "org.junit.Test") {
		return FrameworkResult{Framework: "junit4", Confidence: 0.9, Source: "import"}
	}
	// junit5 is conventional default for Java test files.
	return FrameworkResult{Framework: "junit5", Confidence: 0.5, Source: "convention"}
}

// readHead reads the first n bytes of a file for content-based detection.
func readHead(path string, n int) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	buf := make([]byte, n)
	nr, _ := f.Read(buf)
	if nr == 0 {
		return ""
	}
	return string(buf[:nr])
}

// buildFrameworkInventory aggregates test files into a framework summary.
func buildFrameworkInventory(testFiles []models.TestFile) []models.Framework {
	counts := map[string]int{}
	for _, tf := range testFiles {
		if tf.Framework != "" && tf.Framework != "unknown" {
			counts[tf.Framework]++
		}
	}

	frameworks := make([]models.Framework, 0, len(counts))
	for name, count := range counts {
		frameworks = append(frameworks, models.Framework{
			Name:      name,
			Type:      inferFrameworkType(name),
			FileCount: count,
		})
	}

	// Sort by file count descending, then by name for deterministic output.
	sort.Slice(frameworks, func(i, j int) bool {
		if frameworks[i].FileCount != frameworks[j].FileCount {
			return frameworks[i].FileCount > frameworks[j].FileCount
		}
		return frameworks[i].Name < frameworks[j].Name
	})

	return frameworks
}

// inferFrameworkType maps framework names to their broad category.
func inferFrameworkType(name string) models.FrameworkType {
	switch name {
	case "jest", "vitest", "mocha", "jasmine", "go-testing", "pytest", "unittest", "nose2", "junit4", "junit5", "testng", "node-test":
		return models.FrameworkTypeUnit
	case "playwright", "cypress", "puppeteer", "webdriverio", "testcafe", "selenium":
		return models.FrameworkTypeE2E
	default:
		return models.FrameworkTypeUnknown
	}
}
