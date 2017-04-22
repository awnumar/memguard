package cryptobox

// #cgo pkg-config: libsodium
// #include <stdlib.h>
// #include <sodium.h>
import "C"
import "github.com/GoKillers/libsodium-go/support"

func CryptoBoxDetachedAfterNm(mac []byte, m []byte, n []byte, k []byte) ([]byte, int) {
	support.CheckSize(mac, CryptoBoxMacBytes(), "mac")
	support.CheckSize(n, CryptoBoxNonceBytes(), "nonce")
	support.CheckSize(k, CryptoBoxBeforeNmBytes(), "shared secret key")
	c := make([]byte, len(m)+CryptoBoxMacBytes())
	exit := int(C.crypto_box_detached_afternm(
		(*C.uchar)(&c[0]),
		(*C.uchar)(&mac[0]),
		(*C.uchar)(&m[0]),
		(C.ulonglong)(len(m)),
		(*C.uchar)(&n[0]),
		(*C.uchar)(&k[0])))

	return c, exit
}

func CryptoBoxDetached(mac []byte, m []byte, n []byte, pk []byte, sk []byte) ([]byte, int) {
	support.CheckSize(mac, CryptoBoxMacBytes(), "mac")
	support.CheckSize(n, CryptoBoxNonceBytes(), "nonce")
	support.CheckSize(pk, CryptoBoxPublicKeyBytes(), "public key")
	support.CheckSize(sk, CryptoBoxSecretKeyBytes(), "sender's secret key")
	c := make([]byte, len(m)+CryptoBoxMacBytes())
	exit := int(C.crypto_box_detached(
		(*C.uchar)(&c[0]),
		(*C.uchar)(&mac[0]),
		(*C.uchar)(&m[0]),
		(C.ulonglong)(len(m)),
		(*C.uchar)(&n[0]),
		(*C.uchar)(&pk[0]),
		(*C.uchar)(&sk[0])))

	return c, exit
}

func CryptoBoxEasyAfterNm(m []byte, n []byte, k []byte) ([]byte, int) {
	support.CheckSize(n, CryptoBoxNonceBytes(), "nonce")
	support.CheckSize(k, CryptoBoxBeforeNmBytes(), "shared secret key")
	c := make([]byte, len(m)+CryptoBoxMacBytes())
	exit := int(C.crypto_box_easy_afternm(
		(*C.uchar)(&c[0]),
		(*C.uchar)(&m[0]),
		(C.ulonglong)(len(m)),
		(*C.uchar)(&n[0]),
		(*C.uchar)(&k[0])))

	return c, exit
}

func CryptoBoxEasy(m []byte, n []byte, pk []byte, sk []byte) ([]byte, int) {
	support.CheckSize(n, CryptoBoxNonceBytes(), "nonce")
	support.CheckSize(pk, CryptoBoxPublicKeyBytes(), "public key")
	support.CheckSize(sk, CryptoBoxSecretKeyBytes(), "secret key")
	c := make([]byte, len(m)+CryptoBoxMacBytes())
	exit := int(C.crypto_box_easy(
		(*C.uchar)(&c[0]),
		(*C.uchar)(&m[0]),
		(C.ulonglong)(len(m)),
		(*C.uchar)(&n[0]),
		(*C.uchar)(&pk[0]),
		(*C.uchar)(&sk[0])))

	return c, exit
}

func CryptoBoxOpenDetachedAfterNm(c []byte, mac []byte, n []byte, k []byte) ([]byte, int) {
	support.CheckSize(mac, CryptoBoxMacBytes(), "mac")
	support.CheckSize(n, CryptoBoxNonceBytes(), "nonce")
	support.CheckSize(k, CryptoBoxBeforeNmBytes(), "shared secret key")
	m := make([]byte, len(c)-CryptoBoxMacBytes())
	exit := int(C.crypto_box_open_detached_afternm(
		(*C.uchar)(&m[0]),
		(*C.uchar)(&c[0]),
		(*C.uchar)(&mac[0]),
		(C.ulonglong)(len(c)),
		(*C.uchar)(&n[0]),
		(*C.uchar)(&k[0])))

	return m, exit
}

func CryptoBoxOpenDetached(c []byte, mac []byte, n []byte, pk []byte, sk []byte) ([]byte, int) {
	support.CheckSize(mac, CryptoBoxMacBytes(), "mac")
	support.CheckSize(n, CryptoBoxNonceBytes(), "nonce")
	support.CheckSize(pk, CryptoBoxPublicKeyBytes(), "public key")
	support.CheckSize(sk, CryptoBoxSecretKeyBytes(), "secret key")
	m := make([]byte, len(c)-CryptoBoxMacBytes())
	exit := int(C.crypto_box_detached(
		(*C.uchar)(&m[0]),
		(*C.uchar)(&c[0]),
		(*C.uchar)(&mac[0]),
		(C.ulonglong)(len(c)),
		(*C.uchar)(&n[0]),
		(*C.uchar)(&pk[0]),
		(*C.uchar)(&sk[0])))

	return m, exit
}

func CryptoBoxOpenEasyAfterNm(c []byte, n []byte, k []byte) ([]byte, int) {
	support.CheckSize(n, CryptoBoxNonceBytes(), "nonce")
	support.CheckSize(k, CryptoBoxBeforeNmBytes(), "shared secret key")
	m := make([]byte, len(c)-CryptoBoxMacBytes())
	exit := int(C.crypto_box_open_easy_afternm(
		(*C.uchar)(&m[0]),
		(*C.uchar)(&c[0]),
		(C.ulonglong)(len(c)),
		(*C.uchar)(&n[0]),
		(*C.uchar)(&k[0])))

	return m, exit
}

func CryptoBoxOpenEasy(c []byte, n []byte, pk []byte, sk []byte) ([]byte, int) {
	support.CheckSize(n, CryptoBoxNonceBytes(), "nonce")
	support.CheckSize(pk, CryptoBoxPublicKeyBytes(), "public key")
	support.CheckSize(sk, CryptoBoxSecretKeyBytes(), "secret key")
	m := make([]byte, len(c)-CryptoBoxMacBytes())
	exit := int(C.crypto_box_easy(
		(*C.uchar)(&m[0]),
		(*C.uchar)(&c[0]),
		(C.ulonglong)(len(c)),
		(*C.uchar)(&n[0]),
		(*C.uchar)(&pk[0]),
		(*C.uchar)(&sk[0])))

	return m, exit
}
