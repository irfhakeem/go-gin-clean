package utils

import (
	"math/rand"
	"regexp"
	"slices"
	"strings"
)

func SanitizeString(s string) string {
	spaceRegex := regexp.MustCompile(`\s+`)
	s = spaceRegex.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

func Contains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}

func GenerateSlug(s string) string {
	s = strings.ToLower(s)

	// Replace spaces with hyphens
	s = strings.ReplaceAll(s, " ", "-")

	// Remove special characters except hyphens
	reg := regexp.MustCompile(`[^a-z0-9\-]`)
	s = reg.ReplaceAllString(s, "")

	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile(`-+`)
	s = reg.ReplaceAllString(s, "-")

	// Trim hyphens from start and end
	s = strings.Trim(s, "-")

	return s
}

func GenerateUsername(name string) string {
	name = strings.ToLower(name)
	name = strings.TrimSpace(name)

	reg := regexp.MustCompile(`[^a-z0-9\s]`)
	name = reg.ReplaceAllString(name, "")

	nameArr := strings.Split(name, " ")
	n := len(nameArr)
	if n > 1 {
		name = nameArr[n-2] + nameArr[n-1][:2]
	} else {
		name = nameArr[0]
	}

	name = name + GenerateNumber(4, 0, 9)
	return name
}

func GeneratePassword(length int, includeSpecialChars bool) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const specialChars = "!@#$%^&*()-_=+[]{}|;:,.<>?/"
	var charset string

	if includeSpecialChars {
		charset = letters + specialChars
	} else {
		charset = letters
	}

	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}

	return string(result)
}
