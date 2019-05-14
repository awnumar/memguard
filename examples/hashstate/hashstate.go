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

package hashstate

import (
	"fmt"
	"unsafe"

	"github.com/awnumar/memguard"
	"github.com/awnumar/memguard/examples/hashstate/blake2b"
)

func HashState() {
	// Some safety work
	memguard.CatchInterrupt()
	defer memguard.Purge()

	// Initialise a new hash state structure
	s := new(blake2b.Xof)

	// Construct a buffer of the same size
	b := memguard.NewBuffer(int(unsafe.Sizeof(*s)))
	defer b.Destroy()

	// Type cast the buffer to the struct and overload the original
	s = (*blake2b.Xof)(unsafe.Pointer(&b.Bytes()[0]))

	// Generate a cryptographically-secure seed value
	seed := memguard.NewBufferRandom(32)
	defer seed.Destroy()

	// Initialise the hash state
	h, err := blake2b.NewXOF(blake2b.OutputLengthUnknown, seed.Bytes(), s)
	if err != nil {
		memguard.SafePanic(err)
	}

	// Output the hash state
	fmt.Printf("Internal hash state: %#v\n", h)

	// Generate some output
	if _, err := h.Read(seed.Bytes()); err != nil {
		memguard.SafePanic(err)
	}
	fmt.Println("Output:", seed.Bytes())
}
