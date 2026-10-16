package utils

import (
	cryptorand "crypto/rand"
	"regexp"
	"slices"
	"strings"

	"github.com/google/uuid"
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

	s = strings.ReplaceAll(s, " ", "-")

	reg := regexp.MustCompile(`[^a-z0-9\-]`)
	s = reg.ReplaceAllString(s, "")

	reg = regexp.MustCompile(`-+`)
	s = reg.ReplaceAllString(s, "-")

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
		last := nameArr[n-1]
		if len(last) > 2 {
			last = last[:2]
		}
		name = nameArr[n-2] + last
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
	buf := make([]byte, length)
	if _, err := cryptorand.Read(buf); err == nil {
		for i := range result {
			result[i] = charset[int(buf[i])%len(charset)]
		}
	} else {
		for i := range result {
			result[i] = charset[i%len(charset)]
		}
	}
	return string(result)
}

func ParseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}
