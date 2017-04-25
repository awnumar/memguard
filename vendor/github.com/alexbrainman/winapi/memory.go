// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

const (
	MEM_COMMIT      = 0x00001000
	MEM_RESERVE     = 0x00002000
	MEM_DECOMMIT    = 0x00004000
	MEM_RELEASE     = 0x00008000
	MEM_RESET       = 0x00080000
	MEM_TOP_DOWN    = 0x00100000
	MEM_WRITE_WATCH = 0x00200000
	MEM_PHYSICAL    = 0x00400000
	MEM_RESET_UNDO  = 0x01000000
	MEM_LARGE_PAGES = 0x20000000

	PAGE_READWRITE = 0x0004
	PAGE_NOACCESS  = 0x0001
)

//sys	VirtualAlloc(address uintptr, size uintptr, alloctype uint32, protect uint32) (value uintptr, err error) = kernel32.VirtualAlloc
//sys	VirtualFree(address uintptr, size uintptr, freetype uint32) (err error) = kernel32.VirtualFree
//sys	VirtualProtect(address uintptr, size uintptr, newprotect uint32, oldprotect *uint32) (err error) = kernel32.VirtualProtect
