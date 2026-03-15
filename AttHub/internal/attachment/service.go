package attachment

import (
	"context"
	"fmt"
	"io"
)

type Service struct {
	repo    *Repository
	storage *LocalStorage
}

func NewService(repo *Repository, storage *LocalStorage) *Service {
	return &Service{
		repo:    repo,
		storage: storage,
	}
}

type ImportInput struct {
	FileReader io.Reader
	Filename   string
	URL        *string
	Note       *string
}

type ResetResult struct {
	DeletedAttachments int64
	RemovedFiles       int64
}

func (s *Service) Import(ctx context.Context, input ImportInput) (Attachment, error) {
	storedFile, err := s.storage.Save(input.FileReader, input.Filename)
	if err != nil {
		return Attachment{}, err
	}

	created, err := s.repo.Create(ctx, CreateInput{
		OriginalName: storedFile.OriginalName,
		StoredName:   storedFile.StoredName,
		FileExt:      storedFile.FileExt,
		ContentType:  storedFile.ContentType,
		FileSize:     storedFile.FileSize,
		SHA256:       storedFile.SHA256,
		SourceURL:    input.URL,
		Note:         input.Note,
	})
	if err != nil {
		_ = s.storage.Delete(storedFile.StoredName)
		return Attachment{}, fmt.Errorf("save attachment metadata: %w", err)
	}

	return created, nil
}

func (s *Service) Search(ctx context.Context, keyword string, filename string, page int, pageSize int) ([]Attachment, int, error) {
	return s.repo.Search(ctx, keyword, filename, page, pageSize)
}

func (s *Service) GetByID(ctx context.Context, id int64) (Attachment, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetByPublicID(ctx context.Context, publicID string) (Attachment, error) {
	return s.repo.GetByPublicID(ctx, publicID)
}

func (s *Service) UpdateMetadata(ctx context.Context, id int64, patch MetadataPatch) (Attachment, error) {
	return s.repo.UpdateMetadata(ctx, id, patch)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	deleted, err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// File cleanup is best-effort after the metadata is removed.
	_ = s.storage.Delete(deleted.StoredName)
	return nil
}

func (s *Service) ResolveFileByPublicID(ctx context.Context, publicID string) (Attachment, string, error) {
	item, err := s.repo.GetByPublicID(ctx, publicID)
	if err != nil {
		return Attachment{}, "", err
	}
	return item, s.storage.ResolvePath(item.StoredName), nil
}

func (s *Service) ResetAll(ctx context.Context) (ResetResult, error) {
	deletedAttachments, err := s.repo.ResetAll(ctx)
	if err != nil {
		return ResetResult{}, err
	}

	removedFiles, err := s.storage.Clear()
	if err != nil {
		return ResetResult{}, err
	}

	return ResetResult{
		DeletedAttachments: deletedAttachments,
		RemovedFiles:       removedFiles,
	}, nil
}
