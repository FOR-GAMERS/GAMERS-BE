package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

func GenerateSecurePassword() string {
	const (
		lowercase = "abcdefghijklmnopqrstuvwxyz"
		uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		digits    = "0123456789"
		special   = "!@#$%^&*"
		length    = 16
	)

	allChars := lowercase + uppercase + digits + special

	var password strings.Builder

	password.WriteString(randomChar(lowercase))
	password.WriteString(randomChar(uppercase))
	password.WriteString(randomChar(special))

	for i := 3; i < length; i++ {
		password.WriteString(randomChar(allChars))
	}

	return shuffleString(password.String())
}

func GenerateRandomTag() (string, error) {
	b := make([]byte, 3)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	num := int(b[0])<<16 | int(b[1])<<8 | int(b[2])
	num = num % 100000

	return fmt.Sprintf("%05d", num), nil
}

func randomChar(chars string) string {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
	return string(chars[n.Int64()])
}

func shuffleString(s string) string {
	runes := []rune(s)
	for i := len(runes) - 1; i > 0; i-- {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		j := n.Int64()
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
