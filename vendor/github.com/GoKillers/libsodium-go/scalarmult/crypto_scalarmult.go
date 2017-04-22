package scalarmult

// #cgo pkg-config: libsodium
// #include <stdlib.h>
// #include <sodium.h>
import "C"
import "github.com/GoKillers/libsodium-go/support"

func CryptoScalarmultBytes() int {
	return int(C.crypto_scalarmult_bytes())
}

func CryptoScalarmultScalarBytes() int {
	return int(C.crypto_scalarmult_scalarbytes())
}

func CryptoScalarmultPrimitive() string {
	return C.GoString(C.crypto_scalarmult_primitive())
}

func CryptoScalarmultBase(n []byte) ([]byte, int) {
	support.CheckSize(n, CryptoScalarmultScalarBytes(), "secret key")
	q := make([]byte, CryptoScalarmultBytes())
	var exit C.int

	exit = C.crypto_scalarmult_base(
		(*C.uchar)(&q[0]),
		(*C.uchar)(&n[0]))

	return q, int(exit)
}

func CryptoScalarMult(n []byte, p []byte) ([]byte, int) {
	support.CheckSize(n, CryptoScalarmultScalarBytes(), "secret key")
	support.CheckSize(p, CryptoScalarmultScalarBytes(), "public key")
	q := make([]byte, CryptoScalarmultBytes())
	var exit C.int
	exit = C.crypto_scalarmult(
		(*C.uchar)(&q[0]),
		(*C.uchar)(&n[0]),
		(*C.uchar)(&p[0]))

	return q, int(exit)

}
