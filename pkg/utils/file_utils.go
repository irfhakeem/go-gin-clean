package utils

import (
	"io"
	"mime/multipart"
	"net/http"
	"slices"
	"strings"
)

func IsValidExtension(filename string, allowedExtensions []string) bool {
	filename = strings.ToLower(filename)
	for _, ext := range allowedExtensions {
		if strings.HasSuffix(filename, ext) {
			return true
		}
	}
	return false
}

func IsValidImageHeader(fileHeader *multipart.FileHeader, maxSizeBytes int64, allowedMIMETypes []string) bool {
	if fileHeader == nil || fileHeader.Size <= 0 || fileHeader.Size > maxSizeBytes {
		return false
	}

	f, err := fileHeader.Open()
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return false
	}

	contentType := http.DetectContentType(buf[:n])
	return slices.Contains(allowedMIMETypes, contentType)
}
