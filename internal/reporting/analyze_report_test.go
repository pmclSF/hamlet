package reporting

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/pmclSF/hamlet/internal/models"
)

func TestRenderAnalyzeReport_SmokeSections(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		Repository: models.RepositoryMetadata{
			Name:              "test-repo",
			RootPath:          "/workspace/test-repo",
			Languages:         []string{"javascript"},
			PackageManagers:   []string{"npm"},
			CISystems:         []string{"github-actions"},
			Branch:            "main",
			CommitSHA:         "abc123def456",
			SnapshotTimestamp: time.Now(),
		},
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 42},
			{Name: "playwright", Type: models.FrameworkTypeE2E, FileCount: 5},
		},
		TestFiles: []models.TestFile{
			{Path: "src/__tests__/auth.test.js", Framework: "jest"},
			{Path: "src/__tests__/login.test.js", Framework: "jest"},
		},
		GeneratedAt: time.Now(),
	}

	var buf bytes.Buffer
	RenderAnalyzeReport(&buf, snap)
	output := buf.String()

	sections := []string{
		"Hamlet",
		"test-repo",
		"Frameworks",
		"jest",
		"playwright",
		"42 files",
		"Test Files",
		"Discovered:  2",
		"Signals",
		"Risk",
		"Next steps:",
	}

	for _, s := range sections {
		if !strings.Contains(output, s) {
			t.Errorf("report missing expected section/content: %q", s)
		}
	}
}

func TestRenderAnalyzeReport_EmptySnapshot(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		Repository: models.RepositoryMetadata{
			Name:     "empty-repo",
			RootPath: "/tmp/empty",
		},
		GeneratedAt: time.Now(),
	}

	var buf bytes.Buffer
	RenderAnalyzeReport(&buf, snap)
	output := buf.String()

	if !strings.Contains(output, "no test frameworks detected") {
		t.Error("expected empty framework message")
	}
	if !strings.Contains(output, "Discovered:  0") {
		t.Error("expected zero test file count")
	}
}

func TestRenderAnalyzeReport_WithSignals(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		Repository: models.RepositoryMetadata{
			Name:     "sig-repo",
			RootPath: "/tmp/sig",
		},
		Signals: []models.Signal{
			{Type: "weakAssertion", Category: models.CategoryQuality, Severity: models.SeverityMedium},
			{Type: "weakAssertion", Category: models.CategoryQuality, Severity: models.SeverityMedium},
			{Type: "flakyTest", Category: models.CategoryHealth, Severity: models.SeverityHigh},
		},
		GeneratedAt: time.Now(),
	}

	var buf bytes.Buffer
	RenderAnalyzeReport(&buf, snap)
	output := buf.String()

	if !strings.Contains(output, "quality") {
		t.Error("expected quality category in signals section")
	}
	if !strings.Contains(output, "health") {
		t.Error("expected health category in signals section")
	}
}
