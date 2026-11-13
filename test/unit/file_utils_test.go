package unit_test

import (
	"bytes"
	"mime/multipart"
	"net/textproto"
	"testing"

	"go-gin-clean/pkg/utils"
)

func TestFileValidation(t *testing.T) {
	t.Run("AC-012: Valid JPEG passes MIME and size checks", func(t *testing.T) {
		// Real JPEG magic bytes: FF D8 FF
		jpegContent := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}
		header := createMultipartHeader("avatar.jpg", jpegContent)

		allowedMIMEs := []string{"image/jpeg", "image/png"}
		maxSize := int64(5 * 1024 * 1024)

		if !utils.IsValidImageHeader(header, maxSize, allowedMIMEs) {
			t.Fatalf("expected valid JPEG to pass IsValidImageHeader")
		}
	})

	t.Run("AC-012: Disguised text file with .jpg extension fails magic byte check", func(t *testing.T) {
		// Plain text script disguised with .jpg
		fakeJPEG := []byte("#!/bin/bash\necho 'malicious script'\n")
		header := createMultipartHeader("payload.jpg", fakeJPEG)

		allowedMIMEs := []string{"image/jpeg", "image/png"}
		maxSize := int64(5 * 1024 * 1024)

		if utils.IsValidImageHeader(header, maxSize, allowedMIMEs) {
			t.Fatalf("vulnerability: disguised non-image file was accepted as valid image")
		}
	})

	t.Run("AC-012: Oversized file fails size check", func(t *testing.T) {
		jpegContent := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}
		header := createMultipartHeader("large.jpg", jpegContent)
		header.Size = 10 * 1024 * 1024 // 10 MB

		allowedMIMEs := []string{"image/jpeg", "image/png"}
		maxSize := int64(5 * 1024 * 1024) // 5 MB max

		if utils.IsValidImageHeader(header, maxSize, allowedMIMEs) {
			t.Fatalf("expected oversized file to fail size check")
		}
	})
}

func createMultipartHeader(filename string, content []byte) *multipart.FileHeader {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="avatar"; filename="`+filename+`"`)
	part, _ := writer.CreatePart(h)
	_, _ = part.Write(content)
	_ = writer.Close()

	reader := multipart.NewReader(body, writer.Boundary())
	form, _ := reader.ReadForm(int64(len(content) + 1024))
	return form.File["avatar"][0]
}
