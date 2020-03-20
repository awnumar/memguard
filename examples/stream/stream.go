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

package stream

import (
	"io"
	"os"

	"github.com/awnumar/memguard"
)

// SlowRandByte writes 16KiB of random data to a stream and then operates on it in chunks, returning a random number between 0 and 255.
func SlowRandByte() byte {
	// Get 16KiB bytes of random data.
	// In the real world we might be reading from a socket instead.
	// Also we are free to write data in arbitrarily sized chunks.
	data := memguard.NewBufferRandom(1024 * 16)
	data.Melt() // Allow mutation so stream writer can wipe source buffer.

	// Create a stream object.
	s := memguard.NewStream() // Implements io.Reader and io.Writer interfaces.

	// Write the data to it.
	_, _ = s.Write(data.Bytes()) // Should never error or write less data.
	data.Destroy()               // No longer need the source buffer. (Has been wiped.)

	// Create a buffer to work on this data in chunks.
	buf := memguard.NewBuffer(os.Getpagesize())
	defer buf.Destroy()

	// Read the data back in chunks.
	var parity byte
	for {
		n, err := s.Read(buf.Bytes()) // Reads directly into guarded allocation.
		if err != nil {
			if err == io.EOF {
				break // end of data
			}
			memguard.SafePanic(err) // other error
		}

		// Do some example computation on this data.
		for i := 0; i < n; i++ {
			parity = parity ^ buf.Bytes()[i]
		}
	}

	// Return the result.
	return parity
}
