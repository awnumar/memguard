package main

import (
	"fmt"
	"net"

	"github.com/awnumar/memguard"
)

func socketkey() {
	// Create a secure buffer.
	buf, err := memguard.NewBuffer(32)
	if err != nil {
		memguard.SafePanic(err)
	}

	// Create a server to listen on.
	listener, err := net.Listen("tcp", "127.0.0.1:4128")
	if err != nil {
		memguard.SafePanic(err)
	}

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
		buf, err := memguard.NewBufferRandom(32)
		if err != nil {
			memguard.SafePanic(err)
		}
		defer buf.Destroy()

		fmt.Println("Sending key:", buf, buf.Bytes())

		// Send the data to the server
		var total, written int
		for total = 0; total < 32; total += written {
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
	for total = 0; total < 32; total += read {
		read, err = conn.Read(buf.Bytes()[total:])
		if err != nil {
			memguard.SafePanic(err)
		}
	}
	conn.Close()

	// Output the value to standard output.
	fmt.Println("Received key:", buf, buf.Bytes())

	// Seal the key into an encrypted Enclave object.
	key, err := buf.Seal()
	if err != nil {
		memguard.SafePanic(err)
	}
	// <-- buf is destroyed by this point

	fmt.Println("Encrypted key:", key)

	// Decrypt the key into a buffer.
	buf, err = key.Open()
	if err != nil {
		memguard.SafePanic(err)
	}

	fmt.Println("Decrypted key:", buf, buf.Bytes())

	// Purge the session and wipe the keys before exiting.
	memguard.SafeExit(0)
}
