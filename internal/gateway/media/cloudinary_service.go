package media

import (
	"context"
	"go-gin-clean/pkg/config"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryService struct {
	cfg *config.CloudinaryConfig
}

func NewCloudinaryService(cfg *config.CloudinaryConfig) *CloudinaryService {
	return &CloudinaryService{cfg: cfg}
}

func (c *CloudinaryService) credentials() *cloudinary.Cloudinary {
	var err error
	cld, err := cloudinary.NewFromURL(c.cfg.CloudinaryURL)
	if err != nil {
		log.Fatalf("Failed to initialize Cloudinary: %v", err)
	}
	cld.Config.URL.Secure = true
	return cld
}

func (c *CloudinaryService) UploadFile(ctx context.Context, filename string, size int64, fileHeader multipart.FileHeader, filePath string) (*string, error) {
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

func (c *CloudinaryService) DeleteFile(ctx context.Context, publicID string) error {
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
