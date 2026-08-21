package domain

import "testing"

func TestBug006DeferredParagraphCapture(t *testing.T) {
	got := ParagraphContents([]Paragraph{{Content: "first"}, {Content: "second"}})
	if len(got) != 2 || got[0] != "first" || got[1] != "second" {
		t.Fatalf("contents = %#v", got)
	}
}
