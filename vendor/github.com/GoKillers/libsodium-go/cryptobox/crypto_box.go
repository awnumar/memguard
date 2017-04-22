package cryptobox

// #cgo pkg-config: libsodium
// #include <stdlib.h>
// #include <sodium.h>
import "C"
import "github.com/GoKillers/libsodium-go/support"

func CryptoBoxSeedBytes() int {
	return int(C.crypto_box_seedbytes())
}

func CryptoBoxPublicKeyBytes() int {
	return int(C.crypto_box_publickeybytes())
}

func CryptoBoxSecretKeyBytes() int {
	return int(C.crypto_box_secretkeybytes())
}

func CryptoBoxNonceBytes() int {
	return int(C.crypto_box_noncebytes())
}

func CryptoBoxMacBytes() int {
	return int(C.crypto_box_macbytes())
}

func CryptoBoxPrimitive() string {
	return C.GoString(C.crypto_box_primitive())
}

func CryptoBoxBeforeNmBytes() int {
	return int(C.crypto_box_beforenmbytes())
}

func CryptoBoxZeroBytes() int {
	return int(C.crypto_box_zerobytes())
}

func CryptoBoxBoxZeroBytes() int {
	return int(C.crypto_box_boxzerobytes())
}

func CryptoBoxSeedKeyPair(seed []byte) ([]byte, []byte, int) {
	support.CheckSize(seed, CryptoBoxSeedBytes(), "seed")
	sk := make([]byte, CryptoBoxSecretKeyBytes())
	pk := make([]byte, CryptoBoxPublicKeyBytes())
	exit := int(C.crypto_box_seed_keypair(
		(*C.uchar)(&pk[0]),
		(*C.uchar)(&sk[0]),
		(*C.uchar)(&seed[0])))

	return sk, pk, exit
}

func CryptoBoxKeyPair() ([]byte, []byte, int) {
	sk := make([]byte, CryptoBoxSecretKeyBytes())
	pk := make([]byte, CryptoBoxPublicKeyBytes())
	exit := int(C.crypto_box_keypair(
		(*C.uchar)(&pk[0]),
		(*C.uchar)(&sk[0])))

	return sk, pk, exit
}

func CryptoBoxBeforeNm(pk []byte, sk []byte) ([]byte, int) {
	support.CheckSize(pk, CryptoBoxPublicKeyBytes(), "public key")
	support.CheckSize(sk, CryptoBoxSecretKeyBytes(), "sender's secret key")
	k := make([]byte, CryptoBoxBeforeNmBytes())
	exit := int(C.crypto_box_beforenm(
		(*C.uchar)(&k[0]),
		(*C.uchar)(&pk[0]),
		(*C.uchar)(&sk[0])))

	return k, exit
}

func CryptoBox(m []byte, n []byte, pk []byte, sk []byte) ([]byte, int) {
	support.CheckSize(n, CryptoBoxNonceBytes(), "nonce")
	support.CheckSize(pk, CryptoBoxPublicKeyBytes(), "public key")
	support.CheckSize(sk, CryptoBoxSecretKeyBytes(), "sender's secret key")
	c := make([]byte, len(m))
	exit := int(C.crypto_box(
		(*C.uchar)(&c[0]),
		(*C.uchar)(&m[0]),
		(C.ulonglong)(len(m)),
		(*C.uchar)(&n[0]),
		(*C.uchar)(&pk[0]),
		(*C.uchar)(&sk[0])))

	return c, exit
}

func CryptoBoxOpen(c []byte, n []byte, pk []byte, sk []byte) ([]byte, int) {
	support.CheckSize(n, CryptoBoxNonceBytes(), "nonce")
	support.CheckSize(pk, CryptoBoxPublicKeyBytes(), "public key")
	support.CheckSize(sk, CryptoBoxPublicKeyBytes(), "secret key")
	m := make([]byte, len(c))
	exit := int(C.crypto_box_open(
		(*C.uchar)(&m[0]),
		(*C.uchar)(&c[0]),
		(C.ulonglong)(len(c)),
		(*C.uchar)(&n[0]),
		(*C.uchar)(&pk[0]),
		(*C.uchar)(&sk[0])))

	return m, exit
}

func CryptoBoxAfterNm(m []byte, n []byte, k []byte) ([]byte, int) {
	support.CheckSize(n, CryptoBoxNonceBytes(), "nonce")
	support.CheckSize(k, CryptoBoxBeforeNmBytes(), "shared secret key")
	c := make([]byte, len(m))
	exit := int(C.crypto_box_afternm(
		(*C.uchar)(&c[0]),
		(*C.uchar)(&m[0]),
		(C.ulonglong)(len(m)),
		(*C.uchar)(&n[0]),
		(*C.uchar)(&k[0])))

	return c, exit
}

func CryptoBoxOpenAfterNm(c []byte, n []byte, k []byte) ([]byte, int) {
	support.CheckSize(n, CryptoBoxNonceBytes(), "nonce")
	support.CheckSize(k, CryptoBoxBeforeNmBytes(), "shared secret key")
	m := make([]byte, len(c))
	exit := int(C.crypto_box_open_afternm(
		(*C.uchar)(&m[0]),
		(*C.uchar)(&c[0]),
		(C.ulonglong)(len(c)),
		(*C.uchar)(&n[0]),
		(*C.uchar)(&k[0])))

	return m, exit
}
