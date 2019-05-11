package memcall

import "errors"

// Structure for typed specification of memory protection constants.

// MemoryProtectionFlag specifies some particular memory protection flag.
type MemoryProtectionFlag struct {
	// NOACCESS  := 1 (00000001)
	// READ      := 2 (00000010)
	// WRITE     := 4 (00000100) // unused
	// READWRITE := 6 (00000110)

	flag byte
}

// NoAccess specifies that the memory should be marked unreadable and immutable.
var NoAccess = MemoryProtectionFlag{1}

// ReadOnly specifies that the memory should be marked read-only (immutable).
var ReadOnly = MemoryProtectionFlag{2}

// ReadWrite specifies that the memory should be made readable and writable.
var ReadWrite = MemoryProtectionFlag{6}

// ErrInvalidFlag indicates that a given memory protection flag is undefined.
var ErrInvalidFlag = errors.New("<memguard::memcall> memory protection flag is undefined")
