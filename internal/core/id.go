package core

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

const idCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const idLength = 10

// GenerateID generates a random alphanumeric ID of length 10.
func GenerateID() string {
	b := make([]byte, idLength)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(idCharset))))
		if err != nil {
			// Fallback to less secure random if crypto/rand fails (unlikely)
			return generateFallbackID()
		}
		b[i] = idCharset[num.Int64()]
	}
	return string(b)
}

func generateFallbackID() string {
	// Simple fallback, though in practice crypto/rand shouldn't fail
	return "fallback" + fmt.Sprintf("%d", time.Now().UnixNano())[:2]
}
