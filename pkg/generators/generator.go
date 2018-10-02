package generators

import (
	"math/rand"
	"time"
)

const (
	alphaNumericCharset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	alphabeticalCharset = "abcdefghijklmnopqrstuvwxyz"
)

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

// InsecureAlphabetString generates an INSECURE alphabetical string of len characters
func InsecureAlphabetString(lenth int) string {
	return InsecureStringWithCharset(lenth, alphabeticalCharset)
}

// InsecureStringWithCharset generates an INSECURE random string of len with the supplied
// charset
func InsecureStringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
