package generichash

// #cgo pkg-config: libsodium
// #include <stdlib.h>
// #include <sodium.h>
import "C"
import "github.com/GoKillers/libsodium-go/support"

func CryptoGenericHashBytesMin() int {
	return int(C.crypto_generichash_bytes_min())
}

func CryptoGenericHashBytesMax() int {
	return int(C.crypto_generichash_bytes_max())
}

func CryptoGenericHashBytes() int {
	return int(C.crypto_generichash_bytes())
}

func CryptoGenericHashKeyBytesMin() int {
	return int(C.crypto_generichash_keybytes_min())
}

func CryptoGenericHashKeyBytesMax() int {
	return int(C.crypto_generichash_keybytes_max())
}

func CryptoGenericHashKeyBytes() int {
	return int(C.crypto_generichash_keybytes())
}

func CryptoGenericHashPrimitive() string {
	return C.GoString(C.crypto_generichash_primitive())
}

func CryptoGenericHashStateBytes() int {
	return int(C.crypto_generichash_statebytes())
}

// I took care of the typedef confusions. This should work okay.
func CryptoGenericHash(outlen int, in []byte, key []byte) ([]byte, int) {
	support.CheckSizeInRange(outlen, CryptoGenericHashBytesMin(), CryptoGenericHashBytesMax(), "out")
	support.CheckSizeInRange(len(key), CryptoGenericHashKeyBytesMin(), CryptoGenericHashKeyBytesMax(), "key")
	out := make([]byte, outlen)
	exit := int(C.crypto_generichash(
		(*C.uchar)(&out[0]),
		(C.size_t)(outlen),
		(*C.uchar)(&in[0]),
		(C.ulonglong)(len(in)),
		(*C.uchar)(&key[0]),
		(C.size_t)(len(key))))

	return out, exit
}

// I took care of the typedef confusions. This should work okay.
func CryptoGenericHashInit(key []byte, outlen int) (*C.struct_crypto_generichash_blake2b_state, int) {
	support.CheckSizeInRange(outlen, CryptoGenericHashBytesMin(), CryptoGenericHashBytesMax(), "out")
	state := new(C.struct_crypto_generichash_blake2b_state)
	exit := int(C.crypto_generichash_init(
		(*C.struct_crypto_generichash_blake2b_state)(state),
		(*C.uchar)(&key[0]),
		(C.size_t)(len(key)),
		(C.size_t)(outlen)))

	return state, exit
}

// I took care of the typedef confusions. This should work okay.
func CryptoGenericHashUpdate(state *C.struct_crypto_generichash_blake2b_state, in []byte) (*C.struct_crypto_generichash_blake2b_state, int) {
	exit := int(C.crypto_generichash_update(
		(state),
		(*C.uchar)(&in[0]),
		(C.ulonglong)(len(in))))

	return state, exit
}

func CryptoGenericHashFinal(state *C.struct_crypto_generichash_blake2b_state, outlen int) (*C.struct_crypto_generichash_blake2b_state, []byte, int) {
	support.CheckSizeInRange(outlen, CryptoGenericHashBytesMin(), CryptoGenericHashBytesMax(), "out")
	out := make([]byte, outlen)
	exit := int(C.crypto_generichash_final(
		state,
		(*C.uchar)(&out[0]),
		(C.size_t)(outlen)))

	return state, out, exit
}
