package core

const (
	errCodeNoErr = iota
	errCodeBufferExpired
	errCodeBufferTooSmall
	errCodeCofferExpired
	errCodeDecryptionFailed
	errCodeInvalidKeyLength
	errCodeNullBuffer
	errCodeNullEnclave

	// Error string context.
	ctx = "memguard/core: "
)

type errCode uint

type memguardErr struct {
	s string
	c errCode
}

func (e memguardErr) Error() string {
	return e.s
}

var errors = [...]memguardErr {
	{ctx + "no error", errCodeNoErr},
	{ctx + "buffer has been purged from memory and can no longer be used", errCodeBufferExpired},
	{ctx + "the given buffer is too small to hold the plaintext", errCodeBufferTooSmall},
	{ctx + "attempted usage of destroyed key object", errCodeCofferExpired},
	{ctx + "decryption failed", errCodeDecryptionFailed},
	{ctx + "key must be exactly 32 bytes", errCodeInvalidKeyLength},
	{ctx + "buffer size must be greater than zero", errCodeNullBuffer},
	{ctx + "enclave size must be greater than zero", errCodeNullEnclave},
}

func isXError(e error, c errCode) bool {
	ei, ok := e.(memguardErr)
	return ok && ei.c == c
}

// IsBufferExpired returns a boolean indicating whether the error is
// known to report that an operation was attempted on or with a buffer
// that has been destroyed.
func IsBufferExpired(e error) bool {
	return isXError(e, errCodeBufferExpired)
}

// IsBufferTooSmall returns a boolean indicating whether the error is
// known to report that the decryption function, Open, was given an
// output buffer that is too small to hold the plaintext.
func IsBufferTooSmall(e error) bool {
	return isXError(e, errCodeBufferTooSmall)
}

// IsCofferExpired returns a boolean indicating whether the error is
// known to report that an operation was attempted using a secure key
// container that has been wiped and destroyed.
func IsCofferExpired(e error) bool {
	return isXError(e, errCodeCofferExpired)
}

// IsDecryptionFailed returns a boolean indicating whether the error is
// known to report that decryption failed. This can occur if the given
// key is incorrect or if the ciphertext is invalid.
func IsDecryptionFailed(e error) bool {
	return isXError(e, errCodeDecryptionFailed)
}

// IsInvalidKeyLength returns a boolean indicating whether the error is
// known to report that encryption or decryption with a key that is not
// exactly 32 bytes was attempted.
func IsInvalidKeyLength(e error) bool {
	return isXError(e, errCodeInvalidKeyLength)
}

// IsNullBuffer returns a boolean indicating whether the error is known
// to report that a buffer with size not greater than zero was
// attempted to be constructed.
func IsNullBuffer(e error) bool {
	return isXError(e, errCodeNullBuffer)
}

// IsNullEnclave returns a boolean indicating whether the error is
// known to report that an enclave of size less than one was attempted
// to be constructed.
func IsNullEnclave(e error) bool {
	return isXError(e, errCodeNullEnclave)
}
