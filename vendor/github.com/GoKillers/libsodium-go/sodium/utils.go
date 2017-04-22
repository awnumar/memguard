package sodium

import "fmt"
import "unsafe"

// #cgo pkg-config: libsodium
// #include <stdlib.h>
// #include <sodium.h>
import "C"

func MemZero(buff1 []byte) {
	if len(buff1) > 0 {
		C.sodium_memzero(unsafe.Pointer(&buff1[0]), C.size_t(len(buff1)))
	}
}

func MemCmp(buff1, buff2 []byte, length int) int {
	if length >= len(buff1) || length >= len(buff2) {
		panic(fmt.Sprintf("Attempt to compare more bytes (%d) than provided "+
			"(%d, %d)", length, len(buff1), len(buff2)))
	}
	return int(C.sodium_memcmp(unsafe.Pointer(&buff1[0]),
		unsafe.Pointer(&buff2[0]),
		C.size_t(length)))
}

func Bin2hex(bin []byte) string {
	maxlen := len(bin)*2 + 1
	binPtr := (*C.uchar)(unsafe.Pointer(&bin[0]))
	buf := (*C.char)(C.malloc(C.size_t(maxlen)))
	defer C.free(unsafe.Pointer(buf))

	C.sodium_bin2hex(buf, C.size_t(maxlen), binPtr, C.size_t(len(bin)))

	return C.GoString(buf)
}
