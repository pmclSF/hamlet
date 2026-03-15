package golden

import "testing"

func TestAdd(t *testing.T) {
	t.Parallel()
	if Add(1, 2) != 3 {
		t.Fatalf("unexpected sum")
	}
}
