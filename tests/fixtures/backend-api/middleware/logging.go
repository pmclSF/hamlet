package middleware

import "fmt"

type LogEntry struct {
	Method string
	Path   string
	Status int
}

func FormatLogEntry(method, path string, status int) string {
	return fmt.Sprintf("%s %s %d", method, path, status)
}

func ShouldLog(path string) bool {
	return path != "/health" && path != "/metrics"
}
