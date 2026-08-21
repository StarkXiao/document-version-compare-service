package domain
import "strings"
func Compare(old, next []Paragraph) []ComparisonItem {
	items := make([]ComparisonItem, 0, len(old)+len(next))
	matchedOld := make(map[int]int)
	matchedNew := make(map[int]int)
	for i := range old {
		for j := range next {
			if _, used := matchedNew[j]; !used && old[i].Hash == next[j].Hash {
				matchedOld[i], matchedNew[j] = j, i
				break
			}
		}
	}
	for i, paragraph := range old {
		if j, ok := matchedOld[i]; ok {
			kind := Unchanged
			if paragraph.Position != next[j].Position {
				kind = Moved
			}
			items = append(items, item(kind, paragraph.Position, next[j].Position, paragraph.Content, next[j].Content))
			continue
		}
		if i < len(next) {
			items = append(items, item(Modified, paragraph.Position, next[i].Position, paragraph.Content, next[i].Content))
			matchedNew[i] = -1
		} else {
			items = append(items, item(Deleted, paragraph.Position, 0, paragraph.Content, ""))
		}
	}
	for j, paragraph := range next {
		if _, seen := matchedNew[j]; !seen {
			items = append(items, item(Added, 0, paragraph.Position, "", paragraph.Content))
		}
	}
	return items
}
func item(kind ChangeType, oldPos, newPos int, oldText, newText string) ComparisonItem {
	return ComparisonItem{ID: Hash(string(kind) + oldText + newText)[:16], Type: kind, OldPosition: oldPos, NewPosition: newPos, OldContent: oldText, NewContent: newText}
}
func WordChanges(oldText, newText string) (string, string) {
	oldWords, newWords := strings.Fields(oldText), strings.Fields(newText)
	oldSet, newSet := map[string]bool{}, map[string]bool{}
	for _, word := range oldWords {
		oldSet[word] = true
	}
	for _, word := range newWords {
		newSet[word] = true
	}
	for i, word := range oldWords {
		if !newSet[word] {
			oldWords[i] = "[-" + word + "-]"
		}
	}
	for i, word := range newWords {
		if !oldSet[word] {
			newWords[i] = "{+" + word + "+}"
		}
	}
	return strings.Join(oldWords, " "), strings.Join(newWords, " ")
}
