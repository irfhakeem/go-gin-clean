package media

import (
	"go-gin-clean/internal/core/domain/errors"
	"go-gin-clean/internal/core/ports"
	"io"
	"os"
	"path"
	"path/filepath"
)

type LocalStorageService struct {
}

func NewLocalStorageService() ports.MediaService {
	return &LocalStorageService{}
}

func (s *LocalStorageService) UploadFile(filename string, size int64, content io.Reader, filePath string) (*string, error) {
	basePath := "assets"
	dirPath := filepath.Join(basePath, filePath)
	fullPath := filepath.Join(dirPath, filename)

	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return nil, errors.ErrCreateFileSpace
	}

	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, errors.ErrUploadFile
	}
	defer dst.Close()

	if _, err := io.Copy(dst, content); err != nil {
		return nil, errors.ErrUploadFile
	}

	publicURL := path.Join("/assets", filePath, filename)

	return &publicURL, nil
}

func (s *LocalStorageService) DeleteFile(fileURL string) error {
	if err := os.Remove(fileURL); err != nil {
		return errors.ErrDeleteFile
	}

	return nil
}
