package crypto

// MemClr takes a buffer and wipes it with zeroes.
func MemClr(buf []byte) {
	for i := range buf {
		buf[i] = byte(0)
	}
}

// MemSet takes a buffer and overwrites it with a given byte.
func MemSet(buf []byte, b byte) {
	for i := range buf {
		buf[i] = b
	}
}

// MemScr takes a buffer and overwrites it with random bytes.
func MemScr(buf []byte) error {
	b, err := GetRandBytes(len(buf))
	if err != nil {
		return err
	}
	for i := range b {
		buf[i] = b[i]
	}
	return nil
}
