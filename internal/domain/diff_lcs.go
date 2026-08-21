package domain
import "strings"
type Alignment struct {
	OldIndex int
	NewIndex int
	Score    float64
}
func AlignParagraphs(old, next []Paragraph) []Alignment {
	if len(old) == 0 || len(next) == 0 {
		return nil
	}
	table := make([][]int, len(old)+1)
	for i := range table {
		table[i] = make([]int, len(next)+1)
	}
	for i := len(old) - 1; i >= 0; i-- {
		for j := len(next) - 1; j >= 0; j-- {
			if old[i].Hash == next[j].Hash {
				table[i][j] = table[i+1][j+1] + 1
			} else if table[i+1][j] >= table[i][j+1] {
				table[i][j] = table[i+1][j]
			} else {
				table[i][j] = table[i][j+1]
			}
		}
	}
	alignments := []Alignment{}
	for i, j := 0, 0; i < len(old) && j < len(next); {
		if old[i].Hash == next[j].Hash {
			alignments = append(alignments, Alignment{OldIndex: i, NewIndex: j, Score: 1})
			i++
			j++
		} else if table[i+1][j] >= table[i][j+1] {
			i++
		} else {
			j++
		}
	}
	return alignments
}
func AlignModified(old, next []Paragraph, exact []Alignment) []Alignment {
	usedOld, usedNext := map[int]bool{}, map[int]bool{}
	for _, match := range exact {
		usedOld[match.OldIndex] = true
		usedNext[match.NewIndex] = true
	}
	matches := []Alignment{}
	for i, left := range old {
		if usedOld[i] {
			continue
		}
		best, bestScore := -1, 0.0
		for j, right := range next {
			if usedNext[j] {
				continue
			}
			score := ParagraphSimilarity(left.Content, right.Content)
			if score > bestScore {
				best, bestScore = j, score
			}
		}
		if best >= 0 && bestScore >= 0.35 {
			usedOld[i] = true
			usedNext[best] = true
			matches = append(matches, Alignment{OldIndex: i, NewIndex: best, Score: bestScore})
		}
	}
	return matches
}
func DiffParagraphs(old, next []Paragraph) []ComparisonItem {
	exact := AlignParagraphs(old, next)
	modified := AlignModified(old, next, exact)
	oldMatch, newMatch := map[int]Alignment{}, map[int]Alignment{}
	for _, match := range exact {
		oldMatch[match.OldIndex] = match
		newMatch[match.NewIndex] = match
	}
	for _, match := range modified {
		oldMatch[match.OldIndex] = match
		newMatch[match.NewIndex] = match
	}
	items := []ComparisonItem{}
	for i, left := range old {
		match, ok := oldMatch[i]
		if !ok {
			items = append(items, item(Deleted, left.Position, 0, left.Content, ""))
			continue
		}
		right := next[match.NewIndex]
		kind := Unchanged
		if left.Hash != right.Hash {
			kind = Modified
		} else if left.Position != right.Position {
			kind = Moved
		}
		before, after := left.Content, right.Content
		if kind == Modified {
			before, after = WordChanges(before, after)
		}
		items = append(items, item(kind, left.Position, right.Position, before, after))
	}
	for j, right := range next {
		if _, ok := newMatch[j]; !ok {
			items = append(items, item(Added, 0, right.Position, "", right.Content))
		}
	}
	return items
}
func CommonPrefixWords(left, right string) int {
	a, b := strings.Fields(left), strings.Fields(right)
	limit := len(a)
	if len(b) < limit {
		limit = len(b)
	}
	count := 0
	for count < limit && a[count] == b[count] {
		count++
	}
	return count
}
func CommonSuffixWords(left, right string) int {
	a, b := strings.Fields(left), strings.Fields(right)
	i, j := len(a)-1, len(b)-1
	count := 0
	for i >= 0 && j >= 0 && a[i] == b[j] {
		count++
		i--
		j--
	}
	return count
}
func ChangedWordCount(left, right string) int {
	a, b := strings.Fields(left), strings.Fields(right)
	prefix := CommonPrefixWords(left, right)
	suffix := CommonSuffixWords(strings.Join(a[prefix:], " "), strings.Join(b[prefix:], " "))
	return len(a) + len(b) - 2*prefix - 2*suffix
}
