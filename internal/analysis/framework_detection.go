package analysis

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pmclSF/hamlet/internal/models"
)

// detectFramework infers the test framework for a single file.
//
// This is a first-pass heuristic detector. It uses:
//   - file extension to determine language
//   - simple content patterns for JS/TS files
//   - naming conventions for Go, Python, Java
//
// Limitations:
//   - Does not perform full AST analysis.
//   - May misidentify framework if multiple frameworks are used in one file.
//   - Config-based detection (e.g. reading jest.config.js) is a future enhancement.
func detectFramework(relPath string, absPath string) string {
	ext := strings.ToLower(filepath.Ext(relPath))

	switch {
	case isJSExt(ext):
		return detectJSFramework(absPath)
	case ext == ".go":
		return "go-testing"
	case ext == ".py":
		return detectPythonFramework(absPath)
	case ext == ".java":
		return detectJavaFramework(absPath)
	default:
		return "unknown"
	}
}

func isJSExt(ext string) bool {
	switch ext {
	case ".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs", ".mts", ".cts":
		return true
	}
	return false
}

// detectJSFramework reads the first portion of a JS/TS file and looks for
// import/require patterns that indicate a framework.
//
// Priority order: more specific frameworks first.
func detectJSFramework(absPath string) string {
	content := readHead(absPath, 4096)
	if content == "" {
		return "unknown"
	}

	// E2E frameworks — check first as they are more specific
	if strings.Contains(content, "playwright") || strings.Contains(content, "@playwright/test") {
		return "playwright"
	}
	if strings.Contains(content, "cypress") || strings.Contains(content, "cy.") {
		return "cypress"
	}
	if strings.Contains(content, "puppeteer") {
		return "puppeteer"
	}
	if strings.Contains(content, "webdriverio") || strings.Contains(content, "wdio") {
		return "webdriverio"
	}
	if strings.Contains(content, "testcafe") || strings.Contains(content, "TestCafe") {
		return "testcafe"
	}

	// Unit test frameworks
	if strings.Contains(content, "vitest") || strings.Contains(content, "from 'vitest'") || strings.Contains(content, "from \"vitest\"") {
		return "vitest"
	}
	if strings.Contains(content, "from '@jest") || strings.Contains(content, "jest.") || strings.Contains(content, "from 'jest") {
		return "jest"
	}
	if strings.Contains(content, "from 'mocha'") || strings.Contains(content, "require('mocha')") {
		return "mocha"
	}
	if strings.Contains(content, "from 'jasmine'") || strings.Contains(content, "jasmine.") {
		return "jasmine"
	}

	// Fallback: common test globals suggest jest (most common JS test framework)
	if strings.Contains(content, "describe(") && strings.Contains(content, "expect(") {
		return "jest"
	}

	return "unknown"
}

// detectPythonFramework looks for framework-specific imports.
func detectPythonFramework(absPath string) string {
	content := readHead(absPath, 4096)
	if content == "" {
		return "unknown"
	}
	if strings.Contains(content, "import pytest") || strings.Contains(content, "from pytest") {
		return "pytest"
	}
	if strings.Contains(content, "import unittest") || strings.Contains(content, "from unittest") {
		return "unittest"
	}
	if strings.Contains(content, "import nose") || strings.Contains(content, "from nose") {
		return "nose2"
	}
	return "pytest" // pytest is conventional default for Python
}

// detectJavaFramework looks for framework-specific imports.
func detectJavaFramework(absPath string) string {
	content := readHead(absPath, 4096)
	if content == "" {
		return "unknown"
	}
	if strings.Contains(content, "org.testng") {
		return "testng"
	}
	if strings.Contains(content, "org.junit.jupiter") || strings.Contains(content, "org.junit.Test") {
		if strings.Contains(content, "org.junit.jupiter") {
			return "junit5"
		}
		return "junit4"
	}
	return "junit5" // default for Java
}

// readHead reads the first n bytes of a file for content-based detection.
func readHead(path string, n int) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	buf := make([]byte, n)
	nr, err := f.Read(buf)
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

	// Sort by file count descending for stable, useful output.
	sort.Slice(frameworks, func(i, j int) bool {
		return frameworks[i].FileCount > frameworks[j].FileCount
	})

	return frameworks
}

// inferFrameworkType maps framework names to their broad category.
func inferFrameworkType(name string) models.FrameworkType {
	switch name {
	case "jest", "vitest", "mocha", "jasmine", "go-testing", "pytest", "unittest", "nose2", "junit4", "junit5", "testng":
		return models.FrameworkTypeUnit
	case "playwright", "cypress", "puppeteer", "webdriverio", "testcafe", "selenium":
		return models.FrameworkTypeE2E
	default:
		return models.FrameworkTypeUnknown
	}
}
