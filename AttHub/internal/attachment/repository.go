package attachment

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"atthub/internal/publicid"
)

var ErrNotFound = errors.New("attachment not found")

const maxPublicIDAttempts = 16

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, input CreateInput) (Attachment, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Attachment{}, fmt.Errorf("begin create transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		INSERT INTO attachments (
			original_name, stored_name, file_ext, content_type, file_size, sha256, source_url, note, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, unixepoch(), unixepoch())`,
		input.OriginalName,
		input.StoredName,
		input.FileExt,
		input.ContentType,
		input.FileSize,
		input.SHA256,
		normalizeOptionalString(input.SourceURL),
		normalizeOptionalString(input.Note),
	)
	if err != nil {
		return Attachment{}, fmt.Errorf("insert attachment: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Attachment{}, fmt.Errorf("read inserted id: %w", err)
	}

	assigned := false
	for attempt := 0; attempt < maxPublicIDAttempts; attempt++ {
		publicID, err := publicid.FromAttachment(id, input.SHA256, input.StoredName, attempt)
		if err != nil {
			return Attachment{}, fmt.Errorf("generate public id: %w", err)
		}

		if _, err := tx.ExecContext(ctx, `
			UPDATE attachments
			SET public_id = ?
			WHERE id = ?`,
			publicID, id,
		); err != nil {
			if isUniquePublicIDError(err) {
				continue
			}
			return Attachment{}, fmt.Errorf("assign public id: %w", err)
		}

		assigned = true
		break
	}
	if !assigned {
		return Attachment{}, fmt.Errorf("assign public id: max retries reached for id=%d", id)
	}

	if err := tx.Commit(); err != nil {
		return Attachment{}, fmt.Errorf("commit create transaction: %w", err)
	}

	return r.GetByID(ctx, id)
}

func (r *Repository) GetByID(ctx context.Context, id int64) (Attachment, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, public_id, original_name, stored_name, file_ext, content_type, file_size, sha256, source_url, note, created_at, updated_at
		FROM attachments
		WHERE id = ?`, id)

	attachment, err := scanAttachment(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Attachment{}, ErrNotFound
		}
		return Attachment{}, fmt.Errorf("get attachment by id: %w", err)
	}

	return attachment, nil
}

func (r *Repository) GetByPublicID(ctx context.Context, publicID string) (Attachment, error) {
	normalized, err := publicid.Normalize(publicID)
	if err != nil {
		return Attachment{}, ErrNotFound
	}

	row := r.db.QueryRowContext(ctx, `
		SELECT id, public_id, original_name, stored_name, file_ext, content_type, file_size, sha256, source_url, note, created_at, updated_at
		FROM attachments
		WHERE public_id = ?`, normalized)

	attachment, err := scanAttachment(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Attachment{}, ErrNotFound
		}
		return Attachment{}, fmt.Errorf("get attachment by public id: %w", err)
	}

	return attachment, nil
}

func (r *Repository) Search(ctx context.Context, keyword string, filename string, page int, pageSize int) ([]Attachment, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	keyword = strings.TrimSpace(keyword)
	filename = strings.TrimSpace(filename)

	var total int
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(1)
		FROM attachments
		WHERE (
			? = ''
			OR instr(lower(COALESCE(source_url, '')), lower(?)) > 0
			OR instr(lower(COALESCE(note, '')), lower(?)) > 0
		)
		AND (
			? = ''
			OR instr(lower(COALESCE(stored_name, '')), lower(?)) > 0
		)`,
		keyword, keyword, keyword,
		filename, filename,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count attachments: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, public_id, original_name, stored_name, file_ext, content_type, file_size, sha256, source_url, note, created_at, updated_at
		FROM attachments
		WHERE (
			? = ''
			OR instr(lower(COALESCE(source_url, '')), lower(?)) > 0
			OR instr(lower(COALESCE(note, '')), lower(?)) > 0
		)
		AND (
			? = ''
			OR instr(lower(COALESCE(stored_name, '')), lower(?)) > 0
		)
		ORDER BY updated_at DESC, id DESC
		LIMIT ? OFFSET ?`,
		keyword, keyword, keyword, filename, filename, pageSize, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("search attachments: %w", err)
	}
	defer rows.Close()

	items := make([]Attachment, 0, pageSize)
	for rows.Next() {
		item, err := scanAttachment(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan attachment row: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate search rows: %w", err)
	}

	return items, total, nil
}

func (r *Repository) UpdateMetadata(ctx context.Context, id int64, patch MetadataPatch) (Attachment, error) {
	current, err := r.GetByID(ctx, id)
	if err != nil {
		return Attachment{}, err
	}

	newURL := current.SourceURL
	newNote := current.Note

	if patch.URL != nil {
		newURL = normalizeOptionalString(patch.URL)
	}
	if patch.Note != nil {
		newNote = normalizeOptionalString(patch.Note)
	}

	if _, err := r.db.ExecContext(ctx, `
		UPDATE attachments
		SET source_url = ?, note = ?, updated_at = unixepoch()
		WHERE id = ?`,
		newURL, newNote, id,
	); err != nil {
		return Attachment{}, fmt.Errorf("update metadata: %w", err)
	}

	return r.GetByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id int64) (Attachment, error) {
	item, err := r.GetByID(ctx, id)
	if err != nil {
		return Attachment{}, err
	}

	if _, err := r.db.ExecContext(ctx, "DELETE FROM attachments WHERE id = ?", id); err != nil {
		return Attachment{}, fmt.Errorf("delete attachment record: %w", err)
	}

	return item, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanAttachment(scanner rowScanner) (Attachment, error) {
	var item Attachment
	var sourceURL sql.NullString
	var note sql.NullString

	if err := scanner.Scan(
		&item.ID,
		&item.PublicID,
		&item.OriginalName,
		&item.StoredName,
		&item.FileExt,
		&item.ContentType,
		&item.FileSize,
		&item.SHA256,
		&sourceURL,
		&note,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return Attachment{}, err
	}

	if sourceURL.Valid {
		value := sourceURL.String
		item.SourceURL = &value
	}

	if note.Valid {
		value := note.String
		item.Note = &value
	}

	return item, nil
}

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}

	normalized := strings.TrimSpace(*value)
	if normalized == "" {
		return nil
	}

	return &normalized
}

func isUniquePublicIDError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "unique constraint failed") &&
		strings.Contains(message, "attachments.public_id")
}
