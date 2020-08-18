package utils

import (
	"math/rand"
	"time"
)

const allowedChar = "1234567890abcdefghijklmnopqrstuvwxyz"

// RandHash generate a random hash for size n
func RandHash(n int) string {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	b := make([]byte, n)

	for i := range b {
		b[i] = allowedChar[r.Intn(len(allowedChar))]
	}

	return string(b)
}
