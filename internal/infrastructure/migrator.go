package infrastructure
import "embed"
//go:embed migrations/*.sql
var MigrationFiles embed.FS
func MigrationNames() ([]string, error) {
	entries, err := MigrationFiles.ReadDir("migrations")
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names, nil
}
