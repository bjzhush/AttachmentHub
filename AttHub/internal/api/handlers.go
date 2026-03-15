package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"atthub/internal/attachment"
	"atthub/internal/publicid"

	"github.com/go-chi/chi/v5"
)

const (
	defaultPageSize      = 50
	searchResultPageSize = 50
)

type Handler struct {
	service        *attachment.Service
	logger         *slog.Logger
	maxUploadBytes int64
}

type attachmentResponse struct {
	ID           int64   `json:"id"`
	PublicID     string  `json:"public_id"`
	OriginalName string  `json:"original_name"`
	StoredName   string  `json:"stored_name"`
	FileExt      string  `json:"file_ext"`
	ContentType  string  `json:"content_type"`
	FileSize     int64   `json:"file_size"`
	SHA256       string  `json:"sha256"`
	URL          *string `json:"url"`
	Note         *string `json:"note"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

type searchResponse struct {
	Items    []attachmentResponse `json:"items"`
	Total    int                  `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type patchPayload struct {
	URL  *string `json:"url"`
	Note *string `json:"note"`
}

func (h *Handler) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) importAttachment(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.maxUploadBytes)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "request body too large") {
			writeJSON(w, http.StatusRequestEntityTooLarge, errorResponse{Error: "uploaded file is too large"})
			return
		}
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid multipart form request"})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "file is required"})
		return
	}
	defer file.Close()

	url := optionalFromForm(r.FormValue("url"))
	note := optionalFromForm(r.FormValue("note"))

	created, err := h.service.Import(r.Context(), attachment.ImportInput{
		FileReader: file,
		Filename:   header.Filename,
		URL:        url,
		Note:       note,
	})
	if err != nil {
		h.writeAttachmentError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toAttachmentResponse(created))
}

func (h *Handler) searchAttachments(w http.ResponseWriter, r *http.Request) {
	keyword := strings.TrimSpace(r.URL.Query().Get("keyword"))
	filename := strings.TrimSpace(r.URL.Query().Get("filename"))

	page := 1
	pageSize := defaultPageSize
	if keyword == "" && filename == "" {
		page = parseIntWithDefault(r.URL.Query().Get("page"), 1)
		if page < 1 {
			page = 1
		}
	} else {
		// Search mode intentionally returns only first page without pagination.
		page = 1
		pageSize = searchResultPageSize
	}

	items, total, err := h.service.Search(r.Context(), keyword, filename, page, pageSize)
	if err != nil {
		h.logger.Error("search attachments failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "search failed"})
		return
	}

	responseItems := make([]attachmentResponse, 0, len(items))
	for _, item := range items {
		responseItems = append(responseItems, toAttachmentResponse(item))
	}

	writeJSON(w, http.StatusOK, searchResponse{
		Items:    responseItems,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (h *Handler) getAttachment(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIDParam(w, r)
	if !ok {
		return
	}

	item, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, attachment.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "attachment not found"})
			return
		}
		h.logger.Error("get attachment failed", "id", id, "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to fetch attachment"})
		return
	}

	writeJSON(w, http.StatusOK, toAttachmentResponse(item))
}

func (h *Handler) getAttachmentByPublicID(w http.ResponseWriter, r *http.Request) {
	publicID, ok := parsePublicIDParam(w, r, "publicID")
	if !ok {
		return
	}

	item, err := h.service.GetByPublicID(r.Context(), publicID)
	if err != nil {
		if errors.Is(err, attachment.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "attachment not found"})
			return
		}
		h.logger.Error("get attachment by public id failed", "public_id", publicID, "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to fetch attachment"})
		return
	}

	writeJSON(w, http.StatusOK, toAttachmentResponse(item))
}

func (h *Handler) patchAttachment(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIDParam(w, r)
	if !ok {
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1024*1024))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "failed to read request body"})
		return
	}

	var payload patchPayload
	if len(body) > 0 {
		decoder := json.NewDecoder(bytes.NewReader(body))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&payload); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON payload"})
			return
		}
	}

	if payload.URL == nil && payload.Note == nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "at least one of url/note must be provided"})
		return
	}

	updated, err := h.service.UpdateMetadata(r.Context(), id, attachment.MetadataPatch{
		URL:  payload.URL,
		Note: payload.Note,
	})
	if err != nil {
		if errors.Is(err, attachment.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "attachment not found"})
			return
		}
		h.logger.Error("update attachment failed", "id", id, "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to update attachment"})
		return
	}

	writeJSON(w, http.StatusOK, toAttachmentResponse(updated))
}

func (h *Handler) deleteAttachment(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIDParam(w, r)
	if !ok {
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		if errors.Is(err, attachment.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "attachment not found"})
			return
		}
		h.logger.Error("delete attachment failed", "id", id, "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to delete attachment"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) openAttachmentByPublicID(w http.ResponseWriter, r *http.Request) {
	publicID, ok := parsePublicIDParam(w, r, "publicID")
	if !ok {
		return
	}

	item, path, err := h.service.ResolveFileByPublicID(r.Context(), publicID)
	if err != nil {
		if errors.Is(err, attachment.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		h.logger.Error("resolve attachment file failed", "public_id", publicID, "error", err)
		http.Error(w, "failed to open attachment", http.StatusInternalServerError)
		return
	}

	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			http.NotFound(w, r)
			return
		}
		h.logger.Error("stat attachment file failed", "public_id", publicID, "path", path, "error", err)
		http.Error(w, "failed to read attachment", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", item.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", item.OriginalName))
	http.ServeFile(w, r, path)
}

func (h *Handler) webApp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.WriteString(w, webAppHTML)
}

func (h *Handler) writeAttachmentError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, attachment.ErrUnsupportedFileType):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "unsupported file format, only PDF/HTML is accepted"})
	case errors.Is(err, attachment.ErrEmptyFile):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "empty file is not allowed"})
	default:
		h.logger.Error("attachment operation failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "attachment operation failed"})
	}
}

func parseIDParam(w http.ResponseWriter, r *http.Request) (int64, bool) {
	idRaw := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idRaw, 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid attachment id"})
		return 0, false
	}
	return id, true
}

func parsePublicIDParam(w http.ResponseWriter, r *http.Request, key string) (string, bool) {
	value := chi.URLParam(r, key)
	normalized, err := publicid.Normalize(value)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid attachment public id"})
		return "", false
	}
	return normalized, true
}

func parseIntWithDefault(raw string, fallback int) int {
	if strings.TrimSpace(raw) == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func optionalFromForm(raw string) *string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil
	}
	return &value
}

func toAttachmentResponse(item attachment.Attachment) attachmentResponse {
	return attachmentResponse{
		ID:           item.ID,
		PublicID:     item.PublicID,
		OriginalName: item.OriginalName,
		StoredName:   item.StoredName,
		FileExt:      item.FileExt,
		ContentType:  item.ContentType,
		FileSize:     item.FileSize,
		SHA256:       item.SHA256,
		URL:          item.SourceURL,
		Note:         item.Note,
		CreatedAt:    time.Unix(item.CreatedAt, 0).UTC().Format(time.RFC3339),
		UpdatedAt:    time.Unix(item.UpdatedAt, 0).UTC().Format(time.RFC3339),
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
