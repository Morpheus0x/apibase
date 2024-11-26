package helper

import (
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
