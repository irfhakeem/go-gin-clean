package storage

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"go-gin-clean/internal/application/port"
	"go-gin-clean/pkg/config"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryStorage struct {
	cfg *config.CloudinaryConfig
}

func NewCloudinaryStorage(cfg *config.CloudinaryConfig) port.Storage {
	return &CloudinaryStorage{cfg: cfg}
}

func (c *CloudinaryStorage) credentials() (*cloudinary.Cloudinary, error) {
	cld, err := cloudinary.NewFromURL(c.cfg.CloudinaryURL)
	if err != nil {
		return nil, fmt.Errorf("initialize cloudinary: %w", err)
	}
	cld.Config.URL.Secure = true
	return cld, nil
}

func (c *CloudinaryStorage) UploadFile(ctx context.Context, filename string, size int64, fileHeader multipart.FileHeader, filePath string) (*string, error) {
	credentials, err := c.credentials()
	if err != nil {
		return nil, err
	}
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	ext := filepath.Ext(filename)
	publicID := strings.TrimSuffix(filename, ext)

	uploadResult, err := credentials.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder:         filePath,
		PublicID:       publicID,
		UniqueFilename: api.Bool(false),
		Overwrite:      api.Bool(true),
		ResourceType:   "auto",
	})
	if err != nil {
		return nil, err
	}

	if uploadResult.SecureURL == "" {
		return nil, nil
	}

	return &uploadResult.SecureURL, nil
}
func (c *CloudinaryStorage) DeleteFile(ctx context.Context, path string) error {
	credentials, err := c.credentials()
	if err != nil {
		return err
	}

	publicID := extractPublicID(path)
	_, err = credentials.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID:     publicID,
		ResourceType: "auto",
	})
	if err != nil {
		return err
	}

	return nil
}

func extractPublicID(fileURL string) string {
	if idx := strings.Index(fileURL, "/upload/"); idx != -1 {
		path := fileURL[idx+len("/upload/"):]
		if strings.HasPrefix(path, "v") {
			if slashIdx := strings.Index(path, "/"); slashIdx != -1 {
				path = path[slashIdx+1:]
			}
		}
		ext := filepath.Ext(path)
		return strings.TrimSuffix(path, ext)
	}

	ext := filepath.Ext(fileURL)
	return strings.TrimSuffix(fileURL, ext)
}
