package media

import (
	"context"
	pkgerror "go-gin-clean/pkg/error"
	"io"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type StorageServiceInterface interface {
	UploadFile(ctx context.Context, filename string, size int64, file multipart.FileHeader, folder string) (*string, error)
	DeleteFile(ctx context.Context, path string) error
}

type LocalStorageService struct {
	basePath string
}

func NewLocalStorageService(basePath string) *LocalStorageService {
	if basePath == "" {
		basePath = "assets"
	}
	return &LocalStorageService{basePath: basePath}
}

func (s *LocalStorageService) UploadFile(ctx context.Context, filename string, size int64, fileHeader multipart.FileHeader, filePath string) (*string, error) {
	fileName := filepath.Base(filename)
	if fileName == "." || fileName == ".." || strings.Contains(fileName, "/") || strings.Contains(fileName, "\\") {
		return nil, pkgerror.ErrInvalidInput
	}

	const maxFileSize = 10 << 20
	if size > maxFileSize {
		return nil, pkgerror.ErrFileTooLarge
	}

	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}
	ext := strings.ToLower(filepath.Ext(fileName))
	if !allowedExts[ext] {
		return nil, pkgerror.ErrUnsupportedFileType
	}

	dirPath := filepath.Join(s.basePath, filePath)
	fullPath := filepath.Join(dirPath, fileName)

	if !strings.HasPrefix(filepath.Clean(fullPath), filepath.Clean(s.basePath)+string(filepath.Separator)) {
		return nil, pkgerror.ErrInvalidInput
	}

	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return nil, pkgerror.ErrCreateFileSpace
	}

	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, pkgerror.ErrUploadFile
	}
	defer dst.Close()

	content, err := fileHeader.Open()
	if err != nil {
		return nil, pkgerror.ErrUploadFile
	}
	defer content.Close()

	if _, err := io.Copy(dst, content); err != nil {
		return nil, pkgerror.ErrUploadFile
	}

	publicURL := path.Join("/assets", filePath, fileName)

	return &publicURL, nil
}

func (s *LocalStorageService) DeleteFile(ctx context.Context, fileURL string) error {
	if !strings.HasPrefix(fileURL, "/assets/") {
		return pkgerror.ErrInvalidInput
	}
	relativePath := strings.TrimPrefix(fileURL, "/assets/")
	fullPath := filepath.Join(s.basePath, relativePath)

	if !strings.HasPrefix(filepath.Clean(fullPath), filepath.Clean(s.basePath)+string(filepath.Separator)) {
		return pkgerror.ErrInvalidInput
	}

	if err := os.Remove(fullPath); err != nil {
		return pkgerror.ErrDeleteFile
	}

	return nil
}
