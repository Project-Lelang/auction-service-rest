package util

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// GenerateOTP generates a random numeric OTP of the given length.
func GenerateOTP(length int) (string, error) {
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(length)), nil)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%0*d", length, n), nil
}
