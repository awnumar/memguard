package secretbox

// #cgo pkg-config: libsodium
// #include <stdlib.h>
// #include <sodium.h>
import "C"
import "github.com/GoKillers/libsodium-go/support"

func CryptoSecretBoxDetached(m []byte, n []byte, k []byte) ([]byte, []byte, int) {
	support.CheckSize(n, CryptoSecretBoxNonceBytes(), "nonce")
	support.CheckSize(k, CryptoSecretBoxKeyBytes(), "key")
	c := make([]byte, len(m))
	mac := make([]byte, CryptoSecretBoxMacBytes())
	exit := int(C.crypto_secretbox_detached(
		(*C.uchar)(&c[0]),
		(*C.uchar)(&mac[0]),
		(*C.uchar)(&m[0]),
		(C.ulonglong)(len(m)),
		(*C.uchar)(&n[0]),
		(*C.uchar)(&k[0])))

	return c, mac, exit
}

func CryptoSecretBoxOpenDetached(c []byte, mac []byte, n []byte, k []byte) ([]byte, int) {
	support.CheckSize(mac, CryptoSecretBoxMacBytes(), "mac")
	support.CheckSize(n, CryptoSecretBoxNonceBytes(), "nonce")
	support.CheckSize(k, CryptoSecretBoxKeyBytes(), "key")
	m := make([]byte, len(c))
	exit := int(C.crypto_secretbox_open_detached(
		(*C.uchar)(&m[0]),
		(*C.uchar)(&c[0]),
		(*C.uchar)(&mac[0]),
		(C.ulonglong)(len(c)),
		(*C.uchar)(&n[0]),
		(*C.uchar)(&k[0])))

	return m, exit
}

func CryptoSecretBoxEasy(m []byte, n []byte, k []byte) ([]byte, int) {
	support.CheckSize(n, CryptoSecretBoxNonceBytes(), "nonce")
	support.CheckSize(k, CryptoSecretBoxKeyBytes(), "key")
	c := make([]byte, len(m)+CryptoSecretBoxMacBytes())
	exit := int(C.crypto_secretbox_easy(
		(*C.uchar)(&c[0]),
		(*C.uchar)(&m[0]),
		(C.ulonglong)(len(m)),
		(*C.uchar)(&n[0]),
		(*C.uchar)(&k[0])))

	return c, exit
}

func CryptoSecretBoxOpenEasy(c []byte, n []byte, k []byte) ([]byte, int) {
	support.CheckSize(n, CryptoSecretBoxNonceBytes(), "nonce")
	support.CheckSize(k, CryptoSecretBoxKeyBytes(), "key")
	m := make([]byte, len(c)-CryptoSecretBoxMacBytes())
	exit := int(C.crypto_secretbox_open_easy(
		(*C.uchar)(&m[0]),
		(*C.uchar)(&c[0]),
		(C.ulonglong)(len(c)),
		(*C.uchar)(&n[0]),
		(*C.uchar)(&k[0])))

	return m, exit
}
