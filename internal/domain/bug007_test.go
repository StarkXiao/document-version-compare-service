package domain

import "testing"

func TestBug007LineOffsetsUseBytePositions(t *testing.T) {
	got := ParagraphLineOffsets("中\n文")
	if len(got) != 2 || got[1] != 4 {
		t.Fatalf("offsets = %#v", got)
	}
}
