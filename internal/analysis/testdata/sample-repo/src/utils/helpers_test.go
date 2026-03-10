package utils

import "testing"

func TestAdd(t *testing.T) {
	t.Parallel()
	if Add(2, 3) != 5 {
		t.Error("expected 5")
	}
}
