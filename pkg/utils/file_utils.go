package utils

import "strings"

func IsValidExtension(filename string, allowedExtensions []string) bool {
	filename = strings.ToLower(filename)
	for _, ext := range allowedExtensions {
		if strings.HasSuffix(filename, ext) {
			return true
		}
	}
	return false
}
