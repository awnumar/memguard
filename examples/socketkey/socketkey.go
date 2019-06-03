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

package socketkey

import (
	"bytes"
	"fmt"
	"net"
	"os"

	"github.com/awnumar/memguard"
)

// Save the data here so we can compare it later. Obviously this leaks the secret.
var data []byte

/*
SocketKey is a streaming multi-threaded client->server transfer of secure data over a socket.
*/
func SocketKey(size int) {
	// Create a server to listen on.
	listener, err := net.Listen("tcp", "127.0.0.1:4128")
	if err != nil {
		memguard.SafePanic(err)
	}
	defer listener.Close()

	// Catch signals and close the listener before terminating safely.
	memguard.CatchSignal(func(s os.Signal) {
		fmt.Println("Received signal:", s.String())
		listener.Close()
	})

	// Create a client to connect to our server.
	go func() {
		// Connect to our server
		addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:4128")
		if err != nil {
			memguard.SafePanic(err)
		}
		conn, err := net.DialTCP("tcp", nil, addr)
		if err != nil {
			memguard.SafePanic(err)
		}
		defer conn.Close()

		// Create a buffer filled with random bytes
		buf := memguard.NewBufferRandom(size)
		if buf == nil {
			memguard.SafePanic("invalid size")
		}
		defer buf.Destroy()

		// Save a copy of the key for comparison later.
		data = make([]byte, buf.Size())
		copy(data, buf.Bytes())

		fmt.Printf("Sending key: %#v\n", buf.Bytes())

		// Send the data to the server
		var total, written int
		for total = 0; total < size; total += written {
			written, err = conn.Write(buf.Bytes()[total:])
			if err != nil {
				memguard.SafePanic(err)
			}
		}
	}()

	// Accept connections from clients
	conn, err := listener.Accept()
	if err != nil {
		memguard.SafePanic(err)
	}

	// Create a secure buffer.
	buf := memguard.NewBuffer(size)
	if buf == nil {
		memguard.SafePanic("invalid size")
	}
	defer buf.Destroy()

	// Read bytes from the client into our buffer.
	var total, read int
	for total = 0; total < size; total += read {
		read, err = conn.Read(buf.Bytes()[total:])
		if err != nil {
			memguard.SafePanic(err)
		}
	}
	conn.Close()

	fmt.Printf("Received key: %#v\n", buf.Bytes())

	// Compare the key to make sure it wasn't corrupted.
	if !bytes.Equal(data, buf.Bytes()) {
		memguard.SafePanic(fmt.Sprint("sent != received ::", data, buf.Bytes()))
	}

	// Seal the key into an encrypted Enclave object.
	key := buf.Seal()
	// <-- buf is destroyed by this point

	fmt.Printf("Encrypted key: %#v\n", key)

	// Decrypt the key into a new buffer.
	buf, err = key.Open()
	if err != nil {
		memguard.SafePanic(err)
	}

	fmt.Printf("Decrypted key: %#v\n", buf.Bytes())

	// Destroy the buffer.
	buf.Destroy()

	// Purge the session and wipe the keys before exiting.
	memguard.SafeExit(0)
}
