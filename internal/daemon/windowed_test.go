package daemon

import "testing"

func TestScaledWindowSize(t *testing.T) {
	t.Parallel()
	gotW, gotH := scaledWindowSize(8, 8, 10)
	if gotW != 80 || gotH != 80 {
		t.Fatalf("expected 80x80, got %dx%d", gotW, gotH)
	}
}
