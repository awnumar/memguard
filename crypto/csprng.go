package crypto

// GetRandBytes returns a slice of a specified length, filled with cryptographically-secure random bytes.
func GetRandBytes(n uint) ([]byte, error) {
	buf := make([]byte, n)
	if err := MemScr(buf); err != nil {
		return nil, err
	}
	return buf, nil
}
