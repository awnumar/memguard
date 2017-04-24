// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

const TLS_OUT_OF_INDEXES = 0xffffffff

//sys	TlsAlloc() (index uint32, err error) [failretval==TLS_OUT_OF_INDEXES] = kernel32.TlsAlloc
//sys	TlsFree(index uint32) (err error) = kernel32.TlsFree
//sys	TlsSetValue(index uint32, value uintptr) (err error) = kernel32.TlsSetValue
//sys	TlsGetValue(index uint32) (value uintptr, err error) = kernel32.TlsGetValue

//sys	GetCurrentThreadId() (id uint32) = kernel32.GetCurrentThreadId
