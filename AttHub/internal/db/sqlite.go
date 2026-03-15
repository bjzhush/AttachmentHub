package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"atthub/internal/publicid"
	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS attachments (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	public_id TEXT,
	original_name TEXT NOT NULL,
	stored_name TEXT NOT NULL UNIQUE,
	file_ext TEXT NOT NULL,
	content_type TEXT NOT NULL,
	file_size INTEGER NOT NULL,
	sha256 TEXT NOT NULL,
	source_url TEXT NULL,
	note TEXT NULL,
	created_at INTEGER NOT NULL DEFAULT (unixepoch()),
	updated_at INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE INDEX IF NOT EXISTS idx_attachments_updated_at ON attachments(updated_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_attachments_source_url ON attachments(source_url);
CREATE INDEX IF NOT EXISTS idx_attachments_note ON attachments(note);
CREATE UNIQUE INDEX IF NOT EXISTS idx_attachments_public_id ON attachments(public_id);
`

func OpenSQLite(dbPath string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	for _, pragma := range []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA synchronous=NORMAL;",
		"PRAGMA foreign_keys=ON;",
	} {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return nil, fmt.Errorf("apply sqlite pragma %q: %w", pragma, err)
		}
	}

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("initialize schema: %w", err)
	}

	if err := migratePublicID(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate public id: %w", err)
	}

	return db, nil
}

func migratePublicID(db *sql.DB) error {
	hasColumn, err := hasAttachmentColumn(db, "public_id")
	if err != nil {
		return err
	}

	if !hasColumn {
		if _, err := db.Exec("ALTER TABLE attachments ADD COLUMN public_id TEXT"); err != nil {
			return fmt.Errorf("add public_id column: %w", err)
		}
	}

	rows, err := db.Query(`
		SELECT id
		FROM attachments
		WHERE public_id IS NULL OR trim(public_id) = ''
		ORDER BY id`)
	if err != nil {
		return fmt.Errorf("query rows for public id backfill: %w", err)
	}
	defer rows.Close()

	type pair struct {
		ID       int64
		PublicID string
	}
	updates := make([]pair, 0, 64)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return fmt.Errorf("scan row for public id backfill: %w", err)
		}
		publicID, err := publicid.FromInt64(id)
		if err != nil {
			return fmt.Errorf("generate public id for %d: %w", id, err)
		}
		updates = append(updates, pair{ID: id, PublicID: publicID})
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate rows for public id backfill: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin public id migration tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("UPDATE attachments SET public_id = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("prepare public id update statement: %w", err)
	}
	defer stmt.Close()

	for _, item := range updates {
		if _, err := stmt.Exec(item.PublicID, item.ID); err != nil {
			return fmt.Errorf("update public id for %d: %w", item.ID, err)
		}
	}

	if _, err := tx.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_attachments_public_id ON attachments(public_id)"); err != nil {
		return fmt.Errorf("ensure public id unique index: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit public id migration: %w", err)
	}

	return nil
}

func hasAttachmentColumn(db *sql.DB, column string) (bool, error) {
	rows, err := db.Query("PRAGMA table_info(attachments)")
	if err != nil {
		return false, fmt.Errorf("query table info: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var colType string
		var notNull int
		var defaultValue sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &colType, &notNull, &defaultValue, &pk); err != nil {
			return false, fmt.Errorf("scan table info: %w", err)
		}

		if strings.EqualFold(name, column) {
			return true, nil
		}
	}

	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("iterate table info: %w", err)
	}

	return false, nil
}
