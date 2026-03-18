package logging

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseLevel(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input string
		want  Level
	}{
		{"quiet", LevelQuiet},
		{"q", LevelQuiet},
		{"debug", LevelDebug},
		{"d", LevelDebug},
		{"", LevelDefault},
		{"unknown", LevelDefault},
	}
	for _, tc := range cases {
		if got := ParseLevel(tc.input); got != tc.want {
			t.Errorf("ParseLevel(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestQuietLevel_SuppressesInfo(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	logger := newLogger(&buf, LevelQuiet)
	logger.Info("should not appear")
	logger.Warn("should appear")
	output := buf.String()
	if strings.Contains(output, "should not appear") {
		t.Error("quiet mode should suppress info messages")
	}
	if !strings.Contains(output, "should appear") {
		t.Error("quiet mode should allow warn messages")
	}
}

func TestDefaultLevel_ShowsInfoAndWarn(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	logger := newLogger(&buf, LevelDefault)
	logger.Info("info msg")
	logger.Debug("debug msg")
	logger.Warn("warn msg")
	output := buf.String()
	if !strings.Contains(output, "info msg") {
		t.Error("default mode should show info messages")
	}
	if strings.Contains(output, "debug msg") {
		t.Error("default mode should suppress debug messages")
	}
	if !strings.Contains(output, "warn msg") {
		t.Error("default mode should show warn messages")
	}
}

func TestDebugLevel_ShowsAll(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	logger := newLogger(&buf, LevelDebug)
	logger.Debug("debug msg")
	logger.Info("info msg")
	logger.Warn("warn msg")
	logger.Error("error msg")
	output := buf.String()
	for _, want := range []string{"debug msg", "info msg", "warn msg", "error msg"} {
		if !strings.Contains(output, want) {
			t.Errorf("debug mode should show %q, got: %s", want, output)
		}
	}
}

func TestStructuredAttributes(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	logger := newLogger(&buf, LevelDefault)
	logger.Info("coverage ingested", "path", "coverage/lcov.info", "files", 42)
	output := buf.String()
	if !strings.Contains(output, "path=coverage/lcov.info") {
		t.Errorf("structured log should include path attribute, got: %s", output)
	}
	if !strings.Contains(output, "files=42") {
		t.Errorf("structured log should include files attribute, got: %s", output)
	}
}

// TestGlobalInit verifies that Init/InitWithWriter changes the global logger.
// This test is NOT parallel because it modifies global state.
func TestGlobalInit(t *testing.T) {
	var buf bytes.Buffer
	InitWithWriter(&buf, LevelQuiet)
	L().Info("suppressed")
	if buf.Len() > 0 {
		t.Error("info should be suppressed in quiet mode")
	}
	buf.Reset()
	InitWithWriter(&buf, LevelDefault)
	L().Info("visible")
	if !strings.Contains(buf.String(), "visible") {
		t.Error("after re-init to default, info should be visible")
	}
	// Restore default for other tests.
	Init(LevelDefault)
}
