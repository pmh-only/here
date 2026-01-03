package main

import (
	"crypto/rand"
	"math/big"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz1234567890")

func randStringRunes(n int) string {
	b := make([]rune, n)
	letterRunesLen := big.NewInt(int64(len(letterRunes)))

	for i := range b {
		num, err := rand.Int(rand.Reader, letterRunesLen)
		if err != nil {
			b[i] = 'a'
			continue
		}
		b[i] = letterRunes[num.Int64()]
	}
	return string(b)
}
