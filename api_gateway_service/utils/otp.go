package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

const (
	OTPTTL              = 10 * time.Minute
	OTPExpiresInMinutes = 10
)

// GenerateOTP produces a cryptographically random 6-digit numeric string
// (range "000000"–"999999"). Uses crypto/rand, not math/rand, so the
// output is suitable for security-sensitive flows like signup verification
// and password resets.
func GenerateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", fmt.Errorf("generating OTP: %w", err)
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
