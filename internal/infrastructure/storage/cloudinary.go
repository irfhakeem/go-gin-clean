package storage

import (
	"context"
	"log"
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

func (c *CloudinaryStorage) credentials() *cloudinary.Cloudinary {
	var err error
	cld, err := cloudinary.NewFromURL(c.cfg.CloudinaryURL)
	if err != nil {
		log.Fatalf("Failed to initialize Cloudinary: %v", err)
	}
	cld.Config.URL.Secure = true
	return cld
}

func (c *CloudinaryStorage) UploadFile(ctx context.Context, filename string, size int64, fileHeader multipart.FileHeader, filePath string) (*string, error) {
	credentials := c.credentials()

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

func (c *CloudinaryStorage) DeleteFile(ctx context.Context, publicID string) error {
	credentials := c.credentials()
	_, err := credentials.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID:     publicID,
		ResourceType: "auto",
	})
	if err != nil {
		return err
	}

	return nil
}
