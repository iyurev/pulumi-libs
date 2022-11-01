package gcp

import "testing"

func TestCutOutDot(t *testing.T) {
	s := "firts.second.local."
	t.Logf("%s\n", CutOutDot(s))
}
