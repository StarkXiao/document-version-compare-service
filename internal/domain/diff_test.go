package domain

import "testing"

func TestDiffParagraphsReportsChanges(t *testing.T) {
	old := SplitParagraphs("old", "alpha old text\n\nunchanged paragraph\n\nremove entirely")
	next := SplitParagraphs("new", "alpha new text\n\nunchanged paragraph\n\nadd entirely")
	items := DiffParagraphs(old, next)
	seen := map[ChangeType]int{}
	for _, item := range items {
		seen[item.Type]++
	}
	if seen[Modified] == 0 || seen[Added] == 0 || seen[Deleted] == 0 {
		t.Fatalf("unexpected diff result: %#v", seen)
	}
}

func TestSplitContentKeepsMarkdownHeadings(t *testing.T) {
	blocks := SplitContent("# 标题\n正文\n\n## 第二节\n内容")
	if len(blocks) != 2 {
		t.Fatalf("got %d blocks", len(blocks))
	}
	if blocks[0] != "# 标题\n正文" {
		t.Fatalf("unexpected block: %q", blocks[0])
	}
}
