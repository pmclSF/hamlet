package middleware

import "testing"

func TestFormatLogEntry(t *testing.T) {
	t.Parallel()
	entry := FormatLogEntry("GET", "/api/users", 200)
	if entry != "GET /api/users 200" {
		t.Errorf("entry = %q, want GET /api/users 200", entry)
	}
}

func TestShouldLog(t *testing.T) {
	t.Parallel()
	if ShouldLog("/health") {
		t.Error("should not log /health")
	}
	if !ShouldLog("/api/users") {
		t.Error("should log /api/users")
	}
}
