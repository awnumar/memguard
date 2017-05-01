// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi_test

import (
	"runtime"
	"syscall"
	"testing"
	"unsafe"

	"github.com/alexbrainman/winapi"
)

func TestGlobalMemoryStatusEx(t *testing.T) {
	var m winapi.MEMORYSTATUSEX
	m.Length = uint32(unsafe.Sizeof(m))
	err := winapi.GlobalMemoryStatusEx(&m)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("MEMORYSTATUSEX is %+v", m)
}

func TestGetProcessHandleCount(t *testing.T) {
	h, err := syscall.GetCurrentProcess()
	if err != nil {
		t.Fatal(err)
	}
	var count uint32
	err = winapi.GetProcessHandleCount(h, &count)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Handle count is %v", count)
}

func TestGetVersionEx(t *testing.T) {
	var vi winapi.OSVERSIONINFOEX
	vi.OSVersionInfoSize = uint32(unsafe.Sizeof(vi))
	err := winapi.GetVersionEx(&vi)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("OSVERSIONINFOEX is %+v", vi)
	t.Logf("OSVERSIONINFOEX.CSDVersion is %v", syscall.UTF16ToString(vi.CSDVersion[:]))
}

func testTlsThread(t *testing.T, tlsidx uint32) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	threadId := winapi.GetCurrentThreadId()

	want := uintptr(threadId)
	err := winapi.TlsSetValue(tlsidx, want)
	if err != nil {
		t.Fatal(err)
	}
	have, err := winapi.TlsGetValue(tlsidx)
	if err != nil {
		t.Fatal(err)
	}
	if want != have {
		t.Errorf("threadid=%d: unexpected tls data %d, want %d", threadId, have, want)
	}
}

func TestTls(t *testing.T) {
	tlsidx, err := winapi.TlsAlloc()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := winapi.TlsFree(tlsidx)
		if err != nil {
			t.Fatal(err)
		}
	}()

	const threadCount = 20

	done := make(chan bool)
	for i := 0; i < threadCount; i++ {
		go func() {
			defer func() {
				done <- true
			}()
			testTlsThread(t, tlsidx)
		}()
	}
	for i := 0; i < threadCount; i++ {
		<-done
	}
}
