package crypto

import "crypto/subtle"

// Copy is identical to Go's builtin copy function except the copying is done in constant time. This is to mitigate against side-channel attacks.
func Copy(dst, src []byte) {
	if len(dst) > len(src) {
		subtle.ConstantTimeCopy(1, dst[:len(src)], src)
	} else if len(dst) < len(src) {
		subtle.ConstantTimeCopy(1, dst, src[:len(dst)])
	} else {
		subtle.ConstantTimeCopy(1, dst, src)
	}
}

// Equal does a constant-time comparison of two byte slices. This is to mitigate against side-channel attacks.
func Equal(x, y []byte) bool {
	return subtle.ConstantTimeCompare(x, y) == 1
}
