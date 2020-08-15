package utils

import "math/rand"

const allowedChar = "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// RandHash generate a random hash for size n
func RandHash(n int) string {
	b := make([]byte, n)

	for i := range b {
		b[i] = allowedChar[rand.Intn(len(allowedChar))]
	}

	return string(b)
}
