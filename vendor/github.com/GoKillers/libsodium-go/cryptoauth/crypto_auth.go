package cryptoauth

// #cgo pkg-config: libsodium
// #include <stdlib.h>
// #include <sodium.h>
import "C"
import "github.com/GoKillers/libsodium-go/support"

func CryptoAuthBytes() int {
	return int(C.crypto_auth_bytes())
}

func CryptoAuthKeyBytes() int {
	return int(C.crypto_auth_keybytes())
}

func CryptoAuthPrimitive() string {
	return C.GoString(C.crypto_auth_primitive())
}

func CryptoAuth(in []byte, key []byte) ([]byte, int) {
	support.CheckSize(key, CryptoAuthKeyBytes(), "key")
	inlen := len(in)
	out := make([]byte, inlen + CryptoAuthBytes())

	exit := int(C.crypto_auth(
		(*C.uchar)(&out[0]),
		(*C.uchar)(&in[0]),
		(C.ulonglong)(inlen),
		(*C.uchar)(&key[0])))

		return out, exit
}

func CryptoAuthVerify(hmac []byte, in []byte, key []byte) int {
	support.CheckSize(key, CryptoAuthKeyBytes(), "key")
	inlen := len(in)

	exit := int(C.crypto_auth_verify(
		(*C.uchar)(&hmac[0]),
		(*C.uchar)(&in[0]),
		(C.ulonglong)(inlen),
		(*C.uchar)(&key[0])))

		return exit
}
