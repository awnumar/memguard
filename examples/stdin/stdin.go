package stdin

import (
	"os"

	"github.com/awnumar/memguard"
)

// ReadKeyFromStdin reads a key from standard inputs and returns it sealed inside an Enclave object.
func ReadKeyFromStdin() *memguard.Enclave {
	key := memguard.NewBufferFromReaderUntil(os.Stdin, '\n')
	if key == nil {
		memguard.SafePanic("no input received")
	}
	return key.Seal()
}
