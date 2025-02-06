package helper

import (
	cr "crypto/rand"
	"encoding/base64"
	"fmt"
	"math/rand/v2"
)

var randDict = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")

func RandomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = randDict[rand.IntN(len(randDict))]
	}
	return string(b)
}

// Create a random URL safe base64 string without padding, n specifies amount of bytes
func RandomBase64(n uint) string {
	if n == 0 {
		return ""
	}
	token := make([]byte, n)
	_, err := cr.Read(token)
	if err != nil {
		panic(fmt.Sprintf("RandomBase64 crypto.rand.Read() has error: %v. This should never happen!", err))
	}
	return base64.RawURLEncoding.EncodeToString(token)
}
