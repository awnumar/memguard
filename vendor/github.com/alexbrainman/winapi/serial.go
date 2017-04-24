// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

const (
	ONESTOPBIT  = 0
	TWOSTOPBITS = 2

	NOPARITY    = 0
	ODDPARITY   = 1
	EVENPARITY  = 2
	MARKPARITY  = 3
	SPACEPARITY = 4
)

type DCB struct {
	DCBlength uint32
	BaudRate  uint32
	Flags     uint32
	_         uint16
	XonLim    uint16
	XoffLim   uint16
	ByteSize  byte
	Parity    byte
	StopBits  byte
	XonChar   byte
	XoffChar  byte
	ErrorChar byte
	EofChar   byte
	EvtChar   byte
	_         uint16
}

type COMMTIMEOUTS struct {
	ReadIntervalTimeout         uint32
	ReadTotalTimeoutMultiplier  uint32
	ReadTotalTimeoutConstant    uint32
	WriteTotalTimeoutMultiplier uint32
	WriteTotalTimeoutConstant   uint32
}

//sys	GetCommState(handle syscall.Handle, dcb *DCB) (err error)
//sys	SetCommState(handle syscall.Handle, dcb *DCB) (err error)
//sys	GetCommTimeouts(handle syscall.Handle, timeouts *COMMTIMEOUTS) (err error)
//sys	SetCommTimeouts(handle syscall.Handle, timeouts *COMMTIMEOUTS) (err error)
//sys	SetupComm(handle syscall.Handle, inqueue uint32, outqueue uint32) (err error)
//sys	SetCommMask(handle syscall.Handle, mask uint32) (err error)
