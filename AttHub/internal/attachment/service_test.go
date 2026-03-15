package attachment

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"atthub/internal/db"
	"atthub/internal/publicid"
)

func TestServiceImportSearchUpdateDelete(t *testing.T) {
	tempDir := t.TempDir()
	sqliteDB, err := db.OpenSQLite(filepath.Join(tempDir, "test.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() {
		_ = sqliteDB.Close()
	})

	repo := NewRepository(sqliteDB)
	storage := NewLocalStorage(filepath.Join(tempDir, "files"))
	service := NewService(repo, storage)

	url := "https://example.com/article"
	note := "golang sqlite backup"
	created, err := service.Import(context.Background(), ImportInput{
		FileReader: strings.NewReader("<html><body>Hello AttachmentHub</body></html>"),
		Filename:   "article.html",
		URL:        &url,
		Note:       &note,
	})
	if err != nil {
		t.Fatalf("import attachment: %v", err)
	}
	if len(created.PublicID) != publicid.Length {
		t.Fatalf("expected %d-char public id, got %q", publicid.Length, created.PublicID)
	}

	fetchedByPublicID, err := service.GetByPublicID(context.Background(), created.PublicID)
	if err != nil {
		t.Fatalf("get by public id: %v", err)
	}
	if fetchedByPublicID.ID != created.ID {
		t.Fatalf("expected fetched id=%d, got %d", created.ID, fetchedByPublicID.ID)
	}

	items, total, err := service.Search(context.Background(), "sqlite", "", 1, 20)
	if err != nil {
		t.Fatalf("search attachment: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total=1, got %d", total)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 search result, got %d", len(items))
	}

	itemsByFilename, totalByFilename, err := service.Search(context.Background(), "", "article", 1, 20)
	if err != nil {
		t.Fatalf("search by filename: %v", err)
	}
	if totalByFilename != 0 || len(itemsByFilename) != 0 {
		t.Fatalf("expected original filename search not to match, total=%d len=%d", totalByFilename, len(itemsByFilename))
	}

	itemsByStoredFilename, totalByStoredFilename, err := service.Search(context.Background(), "", created.StoredName, 1, 20)
	if err != nil {
		t.Fatalf("search by stored filename: %v", err)
	}
	if totalByStoredFilename != 1 || len(itemsByStoredFilename) != 1 {
		t.Fatalf("expected stored filename search to match 1 result, total=%d len=%d", totalByStoredFilename, len(itemsByStoredFilename))
	}

	updatedNote := "updated note"
	clearURL := ""
	updated, err := service.UpdateMetadata(context.Background(), created.ID, MetadataPatch{
		URL:  &clearURL,
		Note: &updatedNote,
	})
	if err != nil {
		t.Fatalf("update metadata: %v", err)
	}
	if updated.SourceURL != nil {
		t.Fatalf("expected source URL to be cleared")
	}
	if updated.Note == nil || *updated.Note != updatedNote {
		t.Fatalf("expected note=%q, got %#v", updatedNote, updated.Note)
	}

	if err := service.Delete(context.Background(), created.ID); err != nil {
		t.Fatalf("delete attachment: %v", err)
	}
	if _, err := service.GetByID(context.Background(), created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
	if _, err := service.GetByPublicID(context.Background(), created.PublicID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound by public id after delete, got %v", err)
	}

	storedFilePath := filepath.Join(tempDir, "files", created.StoredName)
	if _, err := os.Stat(storedFilePath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected stored file to be deleted, stat err=%v", err)
	}
}

func TestServiceRejectsUnsupportedFileType(t *testing.T) {
	tempDir := t.TempDir()
	sqliteDB, err := db.OpenSQLite(filepath.Join(tempDir, "test.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() {
		_ = sqliteDB.Close()
	})

	repo := NewRepository(sqliteDB)
	storage := NewLocalStorage(filepath.Join(tempDir, "files"))
	service := NewService(repo, storage)

	_, err = service.Import(context.Background(), ImportInput{
		FileReader: strings.NewReader("plain text"),
		Filename:   "note.txt",
	})
	if !errors.Is(err, ErrUnsupportedFileType) {
		t.Fatalf("expected ErrUnsupportedFileType, got %v", err)
	}
}
