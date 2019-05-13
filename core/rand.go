package core

import (
	"crypto/rand"
	"io"
	"sync"

	"github.com/awnumar/memguard/memcall"

	"golang.org/x/crypto/blake2b"
)

var pool = new(entropy)

type entropy struct {
	sync.Mutex
	src  blake2b.XOF
	init bool
}

func initEntropyPool() {
	// Allocate a secure buffer for the seed.
	seed, err := memcall.Alloc(32)
	if err != nil {
		Panic(err)
	}

	// Lock the buffer into memory.
	if err := memcall.Lock(seed); err != nil {
		Panic(err)
	}

	// Initialise the seed with cryptographically-secure random bytes.
	if _, err := rand.Read(seed); err != nil {
		Panic(err)
	}

	// Initialise the hash state with the seed.
	source, err := blake2b.NewXOF(blake2b.OutputLengthUnknown, seed)
	if err != nil {
		Panic(err)
	}

	// Wipe the seed and then deallocate it.
	Wipe(seed)
	for i := range seed {
		if seed[i] != 0 {
			// make sure
			Panic("seed not wiped!")
		}
	}
	if err := memcall.Unlock(seed); err != nil {
		Panic(err)
	}
	if err := memcall.Free(seed); err != nil {
		Panic(err)
	}

	// Set the entropy pool's fields appropriately.
	pool.src = source
	pool.init = true
}

// Scramble fills a given buffer with cryptographically-secure random bytes.
func Scramble(buf []byte) {
	pool.Lock()
	defer pool.Unlock()

	// Initialise pool if not done so already
	if !pool.init {
		initEntropyPool()
	}

	if _, err := io.ReadFull(pool.src, buf); err != nil {
		// Limit reached? Reinitialise pool
		initEntropyPool()
		if _, err := io.ReadFull(pool.src, buf); err != nil {
			// Something else is going wrong here
			Panic(err)
		}
	}
}
