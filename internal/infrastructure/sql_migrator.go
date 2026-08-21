package infrastructure
import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"time"
)
type SQLMigrator struct {
	DB        *sql.DB
	Files     fs.FS
	Directory string
	Clock     func() time.Time
}
type Migration struct {
	Version string
	Name    string
	SQL     string
}
type MigrationReport struct {
	Applied    []string
	Skipped    []string
	StartedAt  time.Time
	FinishedAt time.Time
}
func NewSQLMigrator(db *sql.DB, files fs.FS, directory string) *SQLMigrator {
	return &SQLMigrator{DB: db, Files: files, Directory: directory, Clock: time.Now}
}
func (m *SQLMigrator) Load() ([]Migration, error) {
	entries, err := fs.ReadDir(m.Files, m.Directory)
	if err != nil {
		return nil, err
	}
	migrations := make([]Migration, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		raw, err := fs.ReadFile(m.Files, m.Directory+"/"+entry.Name())
		if err != nil {
			return nil, err
		}
		parts := strings.SplitN(entry.Name(), "_", 2)
		if len(parts) < 2 {
			return nil, fmt.Errorf("migration %q requires numeric prefix", entry.Name())
		}
		migrations = append(migrations, Migration{Version: parts[0], Name: entry.Name(), SQL: string(raw)})
	}
	sort.Slice(migrations, func(i, j int) bool { return migrations[i].Version < migrations[j].Version })
	return migrations, nil
}
func (m *SQLMigrator) EnsureTable(ctx context.Context) error {
	_, err := m.DB.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS schema_migrations (version VARCHAR(64) PRIMARY KEY, applied_at TIMESTAMP NOT NULL)")
	return err
}
func (m *SQLMigrator) applied(ctx context.Context) (map[string]bool, error) {
	rows, err := m.DB.QueryContext(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		out[version] = true
	}
	return out, rows.Err()
}
func (m *SQLMigrator) Apply(ctx context.Context) (MigrationReport, error) {
	report := MigrationReport{StartedAt: m.Clock().UTC()}
	if err := m.EnsureTable(ctx); err != nil {
		return report, err
	}
	migrations, err := m.Load()
	if err != nil {
		return report, err
	}
	done, err := m.applied(ctx)
	if err != nil {
		return report, err
	}
	for _, migration := range migrations {
		if done[migration.Version] {
			report.Skipped = append(report.Skipped, migration.Name)
			continue
		}
		if err := m.applyOne(ctx, migration); err != nil {
			return report, err
		}
		report.Applied = append(report.Applied, migration.Name)
	}
	report.FinishedAt = m.Clock().UTC()
	return report, nil
}
func (m *SQLMigrator) applyOne(ctx context.Context, migration Migration) error {
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, migration.SQL); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("apply %s: %w", migration.Name, err)
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)", migration.Version, m.Clock().UTC()); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
func (m *SQLMigrator) Pending(ctx context.Context) ([]Migration, error) {
	migrations, err := m.Load()
	if err != nil {
		return nil, err
	}
	if err = m.EnsureTable(ctx); err != nil {
		return nil, err
	}
	done, err := m.applied(ctx)
	if err != nil {
		return nil, err
	}
	out := []Migration{}
	for _, migration := range migrations {
		if !done[migration.Version] {
			out = append(out, migration)
		}
	}
	return out, nil
}
