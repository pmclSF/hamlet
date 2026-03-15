package handlers

import "testing"

func TestCheckHealth(t *testing.T) {
	t.Parallel()
	h := CheckHealth("1.0.0")
	if h.Status != "healthy" {
		t.Errorf("status = %q, want healthy", h.Status)
	}
	if h.Version != "1.0.0" {
		t.Errorf("version = %q, want 1.0.0", h.Version)
	}
}
