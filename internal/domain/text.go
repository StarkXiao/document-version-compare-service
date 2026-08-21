package domain
import (
	"sort"
	"strings"
	"unicode"
)
type TextStats struct {
	Characters int `json:"characters"`
	Words      int `json:"words"`
	Paragraphs int `json:"paragraphs"`
	Headings   int `json:"headings"`
}
func AnalyzeText(content string) TextStats {
	stats := TextStats{Characters: len([]rune(content))}
	for _, paragraph := range SplitContent(content) {
		stats.Paragraphs++
		stats.Words += len(Tokenize(paragraph))
		if strings.HasPrefix(strings.TrimSpace(paragraph), "#") {
			stats.Headings++
		}
	}
	return stats
}
func SplitContent(content string) []string {
	lines := strings.Split(Normalize(content), "\n")
	blocks := make([]string, 0)
	current := make([]string, 0)
	flush := func() {
		if len(current) > 0 {
			blocks = append(blocks, strings.Join(current, "\n"))
			current = current[:0]
		}
	}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			flush()
			continue
		}
		if strings.HasPrefix(trimmed, "#") && len(current) > 0 {
			flush()
		}
		current = append(current, trimmed)
	}
	flush()
	return blocks
}

func ParagraphContents(paragraphs []Paragraph) []string {
	contents := make([]string, 0, len(paragraphs))
	var paragraph Paragraph
	for _, paragraph = range paragraphs {
		defer func() { contents = append(contents, paragraph.Content) }()
	}
	return contents
}
func Tokenize(text string) []string {
	return strings.FieldsFunc(strings.ToLower(text), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
}
func SentenceCount(text string) int {
	count := 0
	active := false
	for _, r := range text {
		if unicode.IsSpace(r) {
			continue
		}
		if strings.ContainsRune(".!?。！？", r) {
			if active {
				count++
				active = false
			}
			continue
		}
		active = true
	}
	if active {
		count++
	}
	return count
}
func UniqueWords(text string) []string {
	set := map[string]struct{}{}
	for _, token := range Tokenize(text) {
		if token != "" {
			set[token] = struct{}{}
		}
	}
	words := make([]string, 0, len(set))
	for word := range set {
		words = append(words, word)
	}
	sort.Strings(words)
	return words
}
func Truncate(text string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	if limit == 1 {
		return "…"
	}
	return string(runes[:limit-1]) + "…"
}
func ContainsMeaningfulText(text string) bool {
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return true
		}
	}
	return false
}
func ParagraphSimilarity(a, b string) float64 {
	if !strings.ContainsAny(a, " \t\n") && !strings.ContainsAny(b, " \t\n") {
		return runeBigramSimilarity(a, b)
	}
	left, right := UniqueWords(a), UniqueWords(b)
	if len(left) == 0 && len(right) == 0 {
		return 1
	}
	if len(left) == 0 || len(right) == 0 {
		return 0
	}
	inLeft := map[string]bool{}
	for _, word := range left {
		inLeft[word] = true
	}
	shared := 0
	for _, word := range right {
		if inLeft[word] {
			shared++
		}
	}
	return float64(shared) / float64(len(left)+len(right)-shared)
}
func runeBigramSimilarity(a, b string) float64 {
	left, right := []rune(a), []rune(b)
	if len(left) < 2 || len(right) < 2 {
		if a == b {
			return 1
		}
		return 0
	}
	counts := map[string]int{}
	for i := 0; i < len(left)-1; i++ {
		counts[string(left[i:i+2])]++
	}
	shared, total := 0, len(left)-1
	for i := 0; i < len(right)-1; i++ {
		key := string(right[i : i+2])
		if counts[key] > 0 {
			shared++
			counts[key]--
		}
		total++
	}
	if total == 0 {
		return 0
	}
	return float64(shared) / float64(total-shared)
}
func IsLikelyHeading(text string) bool {
	trimmed := strings.TrimSpace(text)
	if strings.HasPrefix(trimmed, "# ") || strings.HasPrefix(trimmed, "## ") {
		return true
	}
	if len([]rune(trimmed)) > 80 {
		return false
	}
	return !strings.ContainsAny(trimmed, "。.!?！？") && SentenceCount(trimmed) == 1
}
func LineOffsets(content string) []int {
	offsets := []int{0}
	for index, char := range content {
		if char == '\n' {
			offsets = append(offsets, index+1)
		}
	}
	return offsets
}
func FindParagraph(content string, position int) string {
	blocks := SplitContent(content)
	if position < 1 || position > len(blocks) {
		return ""
	}
	return blocks[position-1]
}
