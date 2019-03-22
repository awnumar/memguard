/*
	Copyright 2019 Awn Umar

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

package examples

import (
	"fmt"
	"net"
	"time"

	"github.com/awnumar/memguard"
)

/*
SocketKey is a streaming multi-threaded client->server transfer of secure data over a socket.
*/
func SocketKey(size int) {
	// Create a secure buffer.
	buf, err := memguard.NewBuffer(size)
	if err != nil {
		memguard.SafePanic(err)
	}
	defer buf.Destroy()

	// Create a server to listen on.
	listener, err := net.Listen("tcp", "127.0.0.1:4128")
	if err != nil {
		memguard.SafePanic(err)
	}
	defer listener.Close()

	// Catch interrupts to close the listener.
	memguard.CatchInterrupt(func() {
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
		buf, err := memguard.NewBufferRandom(size)
		if err != nil {
			memguard.SafePanic(err)
		}
		defer buf.Destroy()

		fmt.Println("Sending key:", buf, buf.Bytes())

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

	// Read 32 bytes from the client into our buffer.
	var total, read int
	for total = 0; total < size; total += read {
		read, err = conn.Read(buf.Bytes()[total:])
		if err != nil {
			memguard.SafePanic(err)
		}
	}
	conn.Close()

	fmt.Println("Received key:", buf, buf.Bytes())

	// Seal the key into an encrypted Enclave object.
	key, err := buf.Seal()
	if err != nil {
		memguard.SafePanic(err)
	}
	// <-- buf is destroyed by this point

	fmt.Println("Encrypted key:", key)

	// Decrypt the key into a new buffer.
	buf, err = key.Open()
	if err != nil {
		memguard.SafePanic(err)
	}

	fmt.Println("Decrypted key:", buf, buf.Bytes())

	time.Sleep(30 * time.Second)

	// Purge the session and wipe the keys before exiting.
	memguard.SafeExit(0)
}
