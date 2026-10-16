package utils

import (
	"crypto/rand"
	"math/big"
	"strconv"
)

func GenerateNumber(length int, start int, end int) string {
	if length <= 0 || start >= end {
		return ""
	}

	if length > 6 {
		length = 6
	}

	rangeVal := big.NewInt(int64(end - start + 1))
	var result string
	for range length {
		n, err := rand.Int(rand.Reader, rangeVal)
		digit := start
		if err == nil {
			digit += int(n.Int64())
		}
		result += strconv.Itoa(digit)
	}

	return result
}
