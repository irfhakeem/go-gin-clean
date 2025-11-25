package utils

import (
	"math/rand"
	"strconv"
	"time"
)

func GenerateNumber(len int, start int, end int) string {
	if len <= 0 || start >= end {
		return ""
	}

	if len > 6 {
		len = 6
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	var result string
	for i := 0; i < len; i++ {
		digit := rng.Intn(end-start+1) + start
		result += strconv.Itoa(digit)
	}

	return result
}

func ParseStrToPKID(pkidStr string) (int64, error) {
	return strconv.ParseInt(pkidStr, 10, 64)
}
