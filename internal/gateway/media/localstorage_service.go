package media

import (
	"context"
	"go-gin-clean/pkg/errors"
	"io"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"strings"
)

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
		return nil, errors.ErrInvalidInput
	}

	const maxFileSize = 10 << 20 // 10MB
	if size > maxFileSize {
		return nil, errors.ErrFileTooLarge
	}

	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}
	ext := strings.ToLower(filepath.Ext(fileName))
	if !allowedExts[ext] {
		return nil, errors.ErrUnsupportedFileType
	}

	dirPath := filepath.Join(s.basePath, filePath)
	fullPath := filepath.Join(dirPath, fileName)

	if !strings.HasPrefix(filepath.Clean(fullPath), filepath.Clean(s.basePath)+string(filepath.Separator)) {
		return nil, errors.ErrInvalidInput
	}

	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return nil, errors.ErrCreateFileSpace
	}

	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, errors.ErrUploadFile
	}
	defer dst.Close()

	content, err := fileHeader.Open()
	if err != nil {
		return nil, errors.ErrUploadFile
	}
	defer content.Close()

	if _, err := io.Copy(dst, content); err != nil {
		return nil, errors.ErrUploadFile
	}

	publicURL := path.Join("/assets", filePath, fileName)

	return &publicURL, nil
}

func (s *LocalStorageService) DeleteFile(ctx context.Context, fileURL string) error {
	// Convert URL path to file system path
	// Assuming fileURL is like "/assets/path/file.jpg"
	if !strings.HasPrefix(fileURL, "/assets/") {
		return errors.ErrInvalidInput
	}
	relativePath := strings.TrimPrefix(fileURL, "/assets/")
	fullPath := filepath.Join(s.basePath, relativePath)

	// Ensure the path is within basePath
	if !strings.HasPrefix(filepath.Clean(fullPath), filepath.Clean(s.basePath)+string(filepath.Separator)) {
		return errors.ErrInvalidInput
	}

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return errors.ErrDeleteFile // or a specific not found error, but reuse
		}
		return errors.ErrDeleteFile
	}

	return nil
}
