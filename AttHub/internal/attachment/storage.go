package attachment

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrUnsupportedFileType = errors.New("unsupported file type, only PDF/HTML is allowed")
	ErrEmptyFile           = errors.New("empty file is not allowed")
)

type StoredFile struct {
	OriginalName string
	StoredName   string
	FileExt      string
	ContentType  string
	FileSize     int64
	SHA256       string
}

type LocalStorage struct {
	rootDir string
}

func NewLocalStorage(rootDir string) *LocalStorage {
	return &LocalStorage{rootDir: rootDir}
}

func (s *LocalStorage) Save(reader io.Reader, originalName string) (StoredFile, error) {
	ext := strings.ToLower(filepath.Ext(originalName))
	if !isAllowedExtension(ext) {
		return StoredFile{}, ErrUnsupportedFileType
	}

	if err := os.MkdirAll(s.rootDir, 0o755); err != nil {
		return StoredFile{}, fmt.Errorf("prepare storage directory: %w", err)
	}

	tempFile, err := os.CreateTemp(s.rootDir, "upload-*"+ext)
	if err != nil {
		return StoredFile{}, fmt.Errorf("create temporary file: %w", err)
	}
	tempPath := tempFile.Name()

	hash := sha256.New()
	size, copyErr := io.Copy(io.MultiWriter(tempFile, hash), reader)
	closeErr := tempFile.Close()
	if copyErr != nil {
		_ = os.Remove(tempPath)
		return StoredFile{}, fmt.Errorf("save uploaded file: %w", copyErr)
	}
	if closeErr != nil {
		_ = os.Remove(tempPath)
		return StoredFile{}, fmt.Errorf("close temporary file: %w", closeErr)
	}
	if size == 0 {
		_ = os.Remove(tempPath)
		return StoredFile{}, ErrEmptyFile
	}

	contentType, err := detectContentType(tempPath)
	if err != nil {
		_ = os.Remove(tempPath)
		return StoredFile{}, fmt.Errorf("detect file type: %w", err)
	}
	if !isAllowedMime(ext, contentType) {
		_ = os.Remove(tempPath)
		return StoredFile{}, ErrUnsupportedFileType
	}

	storedName, err := generateStoredName(ext)
	if err != nil {
		_ = os.Remove(tempPath)
		return StoredFile{}, fmt.Errorf("generate stored file name: %w", err)
	}

	finalPath := filepath.Join(s.rootDir, storedName)
	if err := os.Rename(tempPath, finalPath); err != nil {
		_ = os.Remove(tempPath)
		return StoredFile{}, fmt.Errorf("move file to storage: %w", err)
	}

	return StoredFile{
		OriginalName: originalName,
		StoredName:   storedName,
		FileExt:      ext,
		ContentType:  contentType,
		FileSize:     size,
		SHA256:       hex.EncodeToString(hash.Sum(nil)),
	}, nil
}

func (s *LocalStorage) Delete(storedName string) error {
	filePath := filepath.Join(s.rootDir, storedName)
	err := os.Remove(filePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("delete stored file %q: %w", storedName, err)
	}
	return nil
}

func (s *LocalStorage) ResolvePath(storedName string) string {
	return filepath.Join(s.rootDir, storedName)
}

// Clear removes all stored files under the storage root.
// It keeps ".gitkeep" if present.
func (s *LocalStorage) Clear() (int64, error) {
	if err := os.MkdirAll(s.rootDir, 0o755); err != nil {
		return 0, fmt.Errorf("prepare storage directory: %w", err)
	}

	entries, err := os.ReadDir(s.rootDir)
	if err != nil {
		return 0, fmt.Errorf("read storage directory: %w", err)
	}

	var removed int64
	for _, entry := range entries {
		name := entry.Name()
		if name == ".gitkeep" {
			continue
		}
		target := filepath.Join(s.rootDir, name)
		if err := os.RemoveAll(target); err != nil {
			return removed, fmt.Errorf("clear storage entry %q: %w", target, err)
		}
		removed++
	}

	return removed, nil
}

func detectContentType(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open stored file: %w", err)
	}
	defer file.Close()

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("read stored file header: %w", err)
	}

	return http.DetectContentType(buf[:n]), nil
}

func isAllowedExtension(ext string) bool {
	switch ext {
	case ".pdf", ".html", ".htm":
		return true
	default:
		return false
	}
}

func isAllowedMime(ext string, mime string) bool {
	switch ext {
	case ".pdf":
		return mime == "application/pdf" || mime == "application/octet-stream"
	case ".html", ".htm":
		return strings.HasPrefix(mime, "text/html") ||
			strings.HasPrefix(mime, "application/xhtml+xml") ||
			strings.HasPrefix(mime, "text/plain")
	default:
		return false
	}
}

func generateStoredName(ext string) (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]) + ext, nil
}
