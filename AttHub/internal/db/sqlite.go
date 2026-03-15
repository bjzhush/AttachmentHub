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
CREATE INDEX IF NOT EXISTS idx_attachments_sha256 ON attachments(sha256);
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
	if err := migrateAttachmentTimestamps(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate attachment timestamps: %w", err)
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
		SELECT id, sha256, stored_name, public_id
		FROM attachments
		ORDER BY id`)
	if err != nil {
		return fmt.Errorf("query rows for public id migration: %w", err)
	}
	defer rows.Close()

	type record struct {
		ID         int64
		SHA256     string
		StoredName string
		PublicID   sql.NullString
	}

	records := make([]record, 0, 64)
	for rows.Next() {
		var item record
		if err := rows.Scan(&item.ID, &item.SHA256, &item.StoredName, &item.PublicID); err != nil {
			return fmt.Errorf("scan row for public id migration: %w", err)
		}
		records = append(records, item)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate rows for public id migration: %w", err)
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

	for _, item := range records {
		current := ""
		if item.PublicID.Valid {
			current = item.PublicID.String
		}

		normalized, normalizeErr := publicid.Normalize(current)
		if normalizeErr == nil {
			// Keep already valid hash style IDs.
			if normalized != current {
				if _, err := stmt.Exec(normalized, item.ID); err != nil {
					return fmt.Errorf("normalize public id for %d: %w", item.ID, err)
				}
			}
			continue
		}

		assigned := false
		for attempt := 0; attempt < 16; attempt++ {
			newPublicID, err := publicid.FromAttachment(item.ID, item.SHA256, item.StoredName, attempt)
			if err != nil {
				return fmt.Errorf("generate public id for %d: %w", item.ID, err)
			}

			var exists int
			if err := tx.QueryRow(`
				SELECT COUNT(1)
				FROM attachments
				WHERE public_id = ? AND id <> ?`,
				newPublicID, item.ID,
			).Scan(&exists); err != nil {
				return fmt.Errorf("check public id uniqueness for %d: %w", item.ID, err)
			}
			if exists > 0 {
				continue
			}

			if _, err := stmt.Exec(newPublicID, item.ID); err != nil {
				if isUniquePublicIDError(err) {
					continue
				}
				return fmt.Errorf("update public id for %d: %w", item.ID, err)
			}

			assigned = true
			break
		}
		if !assigned {
			return fmt.Errorf("update public id for %d: max retries reached", item.ID)
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

func migrateAttachmentTimestamps(db *sql.DB) error {
	hasCreatedAt, err := hasAttachmentColumn(db, "created_at")
	if err != nil {
		return err
	}
	hasUpdatedAt, err := hasAttachmentColumn(db, "updated_at")
	if err != nil {
		return err
	}

	if !hasCreatedAt {
		if _, err := db.Exec("ALTER TABLE attachments ADD COLUMN created_at INTEGER"); err != nil {
			return fmt.Errorf("add created_at column: %w", err)
		}
	}
	if !hasUpdatedAt {
		if _, err := db.Exec("ALTER TABLE attachments ADD COLUMN updated_at INTEGER"); err != nil {
			return fmt.Errorf("add updated_at column: %w", err)
		}
	}

	if _, err := db.Exec(`
		UPDATE attachments
		SET created_at = unixepoch()
		WHERE created_at IS NULL OR created_at <= 0`); err != nil {
		return fmt.Errorf("backfill created_at: %w", err)
	}
	if _, err := db.Exec(`
		UPDATE attachments
		SET updated_at = created_at
		WHERE updated_at IS NULL OR updated_at <= 0`); err != nil {
		return fmt.Errorf("backfill updated_at: %w", err)
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

func isUniquePublicIDError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "unique constraint failed") &&
		strings.Contains(message, "attachments.public_id")
}
