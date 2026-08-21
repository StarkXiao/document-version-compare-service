package application
import (
	"document-version-compare-service/internal/domain"
	"document-version-compare-service/internal/repository"
	"sort"
	"strings"
	"time"
)
type AuditService struct{ store repository.Store }
func NewAuditService(store repository.Store) *AuditService { return &AuditService{store: store} }
func (s *AuditService) List(documentID string) []domain.AuditLog {
	return s.store.ListAudits(documentID)
}
type ActivityReport struct {
	DocumentID       string          `json:"document_id"`
	GeneratedAt      time.Time       `json:"generated_at"`
	VersionCount     int             `json:"version_count"`
	CommentCount     int             `json:"comment_count"`
	OpenCommentCount int             `json:"open_comment_count"`
	ParagraphCount   int             `json:"paragraph_count"`
	WordCount        int             `json:"word_count"`
	Contributors     []string        `json:"contributors"`
	Timeline         []ActivityPoint `json:"timeline"`
	Risks            []ContentRisk   `json:"risks"`
}
type ActivityPoint struct {
	At       time.Time `json:"at"`
	Action   string    `json:"action"`
	ActorID  string    `json:"actor_id"`
	EntityID string    `json:"entity_id"`
}
type ContentRisk struct {
	Level     string `json:"level"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	Paragraph int    `json:"paragraph,omitempty"`
}
func (s *AuditService) Report(documentID string) (ActivityReport, error) {
	versions := s.store.ListVersions(documentID)
	if len(versions) == 0 {
		return ActivityReport{}, domain.ErrNotFound
	}
	report := ActivityReport{DocumentID: documentID, GeneratedAt: time.Now().UTC(), VersionCount: len(versions)}
	contributors := map[string]struct{}{}
	for _, version := range versions {
		contributors[version.CreatedBy] = struct{}{}
	}
	comments := s.store.ListComments(documentID)
	report.CommentCount = len(comments)
	for _, comment := range comments {
		contributors[comment.AuthorID] = struct{}{}
		if comment.Status != domain.CommentResolved {
			report.OpenCommentCount++
		}
	}
	current := versions[len(versions)-1]
	paragraphs := s.store.Paragraphs(current.ID)
	report.ParagraphCount = len(paragraphs)
	for _, paragraph := range paragraphs {
		report.WordCount += len(domain.Tokenize(paragraph.Content))
	}
	report.Contributors = keys(contributors)
	report.Timeline = activityPoints(s.store.ListAudits(documentID))
	report.Risks = InspectContent(current.Content, paragraphs, comments)
	return report, nil
}
func keys(values map[string]struct{}) []string {
	items := make([]string, 0, len(values))
	for value := range values {
		if value != "" {
			items = append(items, value)
		}
	}
	sort.Strings(items)
	return items
}
func activityPoints(logs []domain.AuditLog) []ActivityPoint {
	points := make([]ActivityPoint, 0, len(logs))
	for _, log := range logs {
		points = append(points, ActivityPoint{At: log.CreatedAt, Action: log.Action, ActorID: log.ActorID, EntityID: log.EntityID})
	}
	sort.Slice(points, func(i, j int) bool { return points[i].At.Before(points[j].At) })
	return points
}
func InspectContent(content string, paragraphs []domain.Paragraph, comments []domain.Comment) []ContentRisk {
	risks := make([]ContentRisk, 0)
	stats := domain.AnalyzeText(content)
	if stats.Paragraphs == 0 {
		return append(risks, ContentRisk{Level: "error", Code: "empty_document", Message: "文档没有可比对的段落。"})
	}
	if stats.Characters > 500000 {
		risks = append(risks, ContentRisk{Level: "warning", Code: "large_document", Message: "文档过大，比对任务可能需要更长时间。"})
	}
	if stats.Paragraphs > 2000 {
		risks = append(risks, ContentRisk{Level: "warning", Code: "many_paragraphs", Message: "段落数较多，建议拆分文档后再协作。"})
	}
	if len(comments) > 0 {
		open := 0
		for _, comment := range comments {
			if comment.Status != domain.CommentResolved {
				open++
			}
		}
		if open > 10 {
			risks = append(risks, ContentRisk{Level: "warning", Code: "unresolved_comments", Message: "存在较多未解决批注。"})
		}
	}
	for _, paragraph := range paragraphs {
		trimmed := strings.TrimSpace(paragraph.Content)
		if len([]rune(trimmed)) > 10000 {
			risks = append(risks, ContentRisk{Level: "warning", Code: "long_paragraph", Message: "段落过长，建议以空行分段以改善比对效果。", Paragraph: paragraph.Position})
		}
		if repeatedLine(trimmed) {
			risks = append(risks, ContentRisk{Level: "info", Code: "repeated_line", Message: "段落包含重复行，可能影响阅读与比对。", Paragraph: paragraph.Position})
		}
		if placeholderText(trimmed) {
			risks = append(risks, ContentRisk{Level: "info", Code: "placeholder", Message: "段落中存在待补充占位内容。", Paragraph: paragraph.Position})
		}
	}
	return risks
}
func repeatedLine(text string) bool {
	lines := strings.Split(text, "\n")
	seen := make(map[string]bool)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if seen[line] {
			return true
		}
		seen[line] = true
	}
	return false
}
func placeholderText(text string) bool {
	lower := strings.ToLower(text)
	markers := []string{"todo", "tbd", "待补充", "待确认", "xxx", "[填写]"}
	for _, marker := range markers {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}
type VersionTrend struct {
	FromVersion int     `json:"from_version"`
	ToVersion   int     `json:"to_version"`
	Added       int     `json:"added"`
	Deleted     int     `json:"deleted"`
	Modified    int     `json:"modified"`
	Moved       int     `json:"moved"`
	Unchanged   int     `json:"unchanged"`
	ChangeRate  float64 `json:"change_rate"`
}
type ChangeReport struct {
	DocumentID         string         `json:"document_id"`
	GeneratedAt        time.Time      `json:"generated_at"`
	Trends             []VersionTrend `json:"trends"`
	TotalChanges       int            `json:"total_changes"`
	MostChangedVersion int            `json:"most_changed_version"`
}
func (s *AuditService) Changes(documentID string) (ChangeReport, error) {
	versions := s.store.ListVersions(documentID)
	if len(versions) == 0 {
		return ChangeReport{}, domain.ErrNotFound
	}
	report := ChangeReport{DocumentID: documentID, GeneratedAt: time.Now().UTC()}
	for index := 1; index < len(versions); index++ {
		previous, current := versions[index-1], versions[index]
		items := domain.DiffParagraphs(s.store.Paragraphs(previous.ID), s.store.Paragraphs(current.ID))
		trend := VersionTrend{FromVersion: previous.Number, ToVersion: current.Number}
		for _, item := range items {
			switch item.Type {
			case domain.Added:
				trend.Added++
			case domain.Deleted:
				trend.Deleted++
			case domain.Modified:
				trend.Modified++
			case domain.Moved:
				trend.Moved++
			case domain.Unchanged:
				trend.Unchanged++
			}
		}
		changed := trend.Added + trend.Deleted + trend.Modified + trend.Moved
		denominator := changed + trend.Unchanged
		if denominator > 0 {
			trend.ChangeRate = float64(changed) / float64(denominator)
		}
		report.TotalChanges += changed
		if report.MostChangedVersion == 0 || changed > trendChanges(report.Trends, report.MostChangedVersion) {
			report.MostChangedVersion = current.Number
		}
		report.Trends = append(report.Trends, trend)
	}
	return report, nil
}
func trendChanges(trends []VersionTrend, version int) int {
	for _, trend := range trends {
		if trend.ToVersion == version {
			return trend.Added + trend.Deleted + trend.Modified + trend.Moved
		}
	}
	return 0
}
func (report ChangeReport) HasSubstantialChange(rate float64) bool {
	for _, trend := range report.Trends {
		if trend.ChangeRate >= rate {
			return true
		}
	}
	return false
}
func (report ChangeReport) LatestTrend() (VersionTrend, bool) {
	if len(report.Trends) == 0 {
		return VersionTrend{}, false
	}
	return report.Trends[len(report.Trends)-1], true
}
func (report ActivityReport) IsHealthy() bool {
	for _, risk := range report.Risks {
		if risk.Level == "error" {
			return false
		}
	}
	return true
}
func (report ActivityReport) RiskCounts() map[string]int {
	counts := map[string]int{"error": 0, "warning": 0, "info": 0}
	for _, risk := range report.Risks {
		counts[risk.Level]++
	}
	return counts
}
func (report ActivityReport) RecentActions(limit int) []ActivityPoint {
	if limit <= 0 || len(report.Timeline) == 0 {
		return nil
	}
	if limit > len(report.Timeline) {
		limit = len(report.Timeline)
	}
	start := len(report.Timeline) - limit
	points := append([]ActivityPoint(nil), report.Timeline[start:]...)
	return points
}
func (report ChangeReport) AverageChangeRate() float64 {
	if len(report.Trends) == 0 {
		return 0
	}
	total := 0.0
	for _, trend := range report.Trends {
		total += trend.ChangeRate
	}
	return total / float64(len(report.Trends))
}
