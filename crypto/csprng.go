package crypto

import "crypto/rand"

// GetRandBytes returns a slice of a specified length, filled with cryptographically-secure random bytes.
func GetRandBytes(n int) ([]byte, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}
	return buf, nil
}
