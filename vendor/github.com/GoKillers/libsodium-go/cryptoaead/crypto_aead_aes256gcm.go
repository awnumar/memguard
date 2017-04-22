package cryptoaead

// #cgo pkg-config: libsodium
// #include <stdlib.h>
// #include <sodium.h>
import "C"
import "github.com/GoKillers/libsodium-go/support"

func CryptoAEADAES256GCMKeyBytes() int {
	return int(C.crypto_aead_aes256gcm_keybytes())
}

func CryptoAEADAES256GCMNSecBytes() int {
	return int(C.crypto_aead_aes256gcm_keybytes())
}

func CryptoAEADAES256GCMNPubBytes() int {
	return int(C.crypto_aead_aes256gcm_npubbytes())
}

func CryptoAEADAES256GCMABytes() int {
	return int(C.crypto_aead_aes256gcm_abytes())
}

func CryptoAEADAES256GCMStateBytes() int {
	return int(C.crypto_aead_aes256gcm_statebytes())
}

func CryptoAESAES256GCMIsAvailable() int {
	return int(C.crypto_aead_aes256gcm_is_available())
}

func CryptoAEADAES256GCMEncrypt(m []byte, ad []byte, nsec []byte, npub []byte, k []byte) ([]byte, int) {
	support.CheckSize(k, CryptoAEADAES256GCMKeyBytes(), "secret key")
	support.CheckSize(npub, CryptoAEADAES256GCMNPubBytes(), "public nonce")
	c := make([]byte, len(m)+CryptoAEADAES256GCMABytes())
	cLen := len(c)
	cLenLongLong := (C.ulonglong(cLen))
	exit := int(C.crypto_aead_aes256gcm_encrypt(
		(*C.uchar)(&c[0]),
		&cLenLongLong,
		(*C.uchar)(&m[0]),
		(C.ulonglong)(len(m)),
		(*C.uchar)(&ad[0]),
		(C.ulonglong)(len(ad)),
		(*C.uchar)(&nsec[0]),
		(*C.uchar)(&npub[0]),
		(*C.uchar)(&k[0])))

	return c, exit
}

func CryptoAEADAES256GCMDecrypt(nsec []byte, c []byte, ad []byte, npub []byte, k []byte) ([]byte, int) {
	support.CheckSize(k, CryptoAEADAES256GCMKeyBytes(), "secret key")
	support.CheckSize(npub, CryptoAEADAES256GCMNPubBytes(), "public nonce")
	m := make([]byte, len(c)-CryptoAEADAES256GCMABytes())
	mLen := len(m)
	mLenLongLong := (C.ulonglong)(mLen)

	exit := int(C.crypto_aead_aes256gcm_decrypt(
		(*C.uchar)(&m[0]),
		&mLenLongLong,
		(*C.uchar)(&nsec[0]),
		(*C.uchar)(&c[0]),
		(C.ulonglong)(len(c)),
		(*C.uchar)(&ad[0]),
		(C.ulonglong)(len(ad)),
		(*C.uchar)(&npub[0]),
		(*C.uchar)(&k[0])))

	return m, exit
}
