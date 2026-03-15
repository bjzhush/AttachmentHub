package api

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"atthub/internal/attachment"
	"atthub/internal/config"
	"atthub/internal/db"
)

func TestImportDuplicateReturnsExistingAttachment(t *testing.T) {
	tempDir := t.TempDir()
	sqliteDB, err := db.OpenSQLite(filepath.Join(tempDir, "test.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() {
		_ = sqliteDB.Close()
	})

	service := attachment.NewService(
		attachment.NewRepository(sqliteDB),
		attachment.NewLocalStorage(filepath.Join(tempDir, "files")),
	)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	router := NewRouter(service, config.Config{MaxUploadBytes: 10 * 1024 * 1024}, logger)

	content := []byte("<html><body>same content</body></html>")

	first := uploadHTML(t, router, "a.html", content)
	if first.Code != http.StatusCreated {
		t.Fatalf("expected first upload status %d, got %d", http.StatusCreated, first.Code)
	}

	var firstPayload attachmentResponse
	if err := json.NewDecoder(first.Body).Decode(&firstPayload); err != nil {
		t.Fatalf("decode first response: %v", err)
	}
	if firstPayload.PublicID == "" {
		t.Fatalf("first upload missing public id")
	}

	second := uploadHTML(t, router, "b.html", content)
	if second.Code != http.StatusOK {
		t.Fatalf("expected duplicate upload status %d, got %d", http.StatusOK, second.Code)
	}

	var secondPayload attachmentResponse
	if err := json.NewDecoder(second.Body).Decode(&secondPayload); err != nil {
		t.Fatalf("decode second response: %v", err)
	}
	if secondPayload.ID != firstPayload.ID {
		t.Fatalf("expected duplicate response to return existing id=%d, got %d", firstPayload.ID, secondPayload.ID)
	}
	if secondPayload.PublicID != firstPayload.PublicID {
		t.Fatalf("expected duplicate response to return existing public id=%q, got %q", firstPayload.PublicID, secondPayload.PublicID)
	}
}

func uploadHTML(t *testing.T, router http.Handler, filename string, content []byte) *httptest.ResponseRecorder {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("write form file content: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/attachments/import", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
}
