package domain
import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func ParagraphLineOffsets(content string) []int { return LineOffsets(content) }
func Hash(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}
func Normalize(content string) string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(content, "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " \t")
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}
func SplitParagraphs(versionID, content string) []Paragraph {
	content = Normalize(content)
	if content == "" {
		return nil
	}
	blocks := SplitContent(content)
	result := make([]Paragraph, 0, len(blocks))
	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}
		position := len(result) + 1
		result = append(result, Paragraph{ID: versionID + "-p" + itoa(position), VersionID: versionID, Position: position, Content: block, Hash: Hash(block)})
	}
	return result
}
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := [20]byte{}
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte(n%10) + '0'
		n /= 10
	}
	return string(buf[i:])
}
