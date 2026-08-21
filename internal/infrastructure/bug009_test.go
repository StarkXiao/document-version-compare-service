package infrastructure

import (
	"testing"
	"testing/fstest"
)

func TestBug009MigrationVersionsUseNumericOrder(t *testing.T) {
	files := fstest.MapFS{
		"migrations/2_second.sql": &fstest.MapFile{Data: []byte("select 2")},
		"migrations/10_tenth.sql": &fstest.MapFile{Data: []byte("select 10")},
	}
	migrations, err := NewSQLMigrator(nil, files, "migrations").Load()
	if err != nil { t.Fatal(err) }
	if migrations[0].Version != "2" || migrations[1].Version != "10" {
		t.Fatalf("order = %#v", migrations)
	}
}
