/*
	Copyright 2019 Awn Umar <awn@spacetime.dev>

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

		http://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package casting

import (
	"unsafe"

	"github.com/awnumar/memguard"
)

// Secure is some generic example struct containing sensitive information.
type Secure struct {
	Key       [32]byte
	Salt      [2]uint64
	Counter   uint64
	Something bool
}

// ByteArray10 allocates and returns a region of memory represented as a fixed-size 10 byte array.
func ByteArray10() (*memguard.LockedBuffer, *[10]byte) {
	// Allocate 10 bytes of memory
	b := memguard.NewBuffer(10)

	// Return the LockedBuffer along with the cast pointer
	return b, (*[10]byte)(unsafe.Pointer(&b.Bytes()[0]))
}

// Uint64Array4 allocates a 32 byte memory region and returns it represented as a sequence of four unsigned 64 bit integer values.
func Uint64Array4() (*memguard.LockedBuffer, *[4]uint64) {
	// Allocate the correct amount of memory
	b := memguard.NewBuffer(32)

	// Return the LockedBuffer along with the cast pointer
	return b, (*[4]uint64)(unsafe.Pointer(&b.Bytes()[0]))
}

// SecureStruct allocates a region of memory the size of a struct type and returns a pointer to that memory represented as that struct type.
func SecureStruct() (*memguard.LockedBuffer, *Secure) {
	// Initialise an instance of the struct type
	s := new(Secure)

	// Allocate a LockedBuffer of the correct size
	b := memguard.NewBuffer(int(unsafe.Sizeof(*s)))

	// Return the LockedBuffer along with the initialised struct
	return b, (*Secure)(unsafe.Pointer(&b.Bytes()[0]))
}

// SecureStructArray allocates enough memory to hold an array of Secure structs and returns them.
func SecureStructArray() (*memguard.LockedBuffer, *[2]Secure) {
	// Initialise an instance of the struct type
	s := new(Secure)

	// Allocate a LockedBuffer of four times the size of the struct type
	b := memguard.NewBuffer(int(unsafe.Sizeof(*s)) * 2)

	// Cast a pointer to the start of the memory into a pointer of a fixed size array of Secure structs of length four
	secureArray := (*[2]Secure)(unsafe.Pointer(&b.Bytes()[0]))

	// Return the LockedBuffer along with the array
	return b, secureArray
}

// SecureStructSlice takes a length and returns a slice of Secure struct values of that length.
func SecureStructSlice(size int) (*memguard.LockedBuffer, []Secure) {
	if size < 1 {
		return nil, nil
	}

	// Initialise an instance of the struct type
	s := new(Secure)

	// Allocate the enough memory to store the struct values
	b := memguard.NewBuffer(int(unsafe.Sizeof(*s)) * size)

	// Construct the slice from its parameters
	var sl = struct {
		addr uintptr
		len  int
		cap  int
	}{uintptr(unsafe.Pointer(&b.Bytes()[0])), size, size}

	// Return the LockedBuffer along with the constructed slice
	return b, *(*[]Secure)(unsafe.Pointer(&sl))
}
