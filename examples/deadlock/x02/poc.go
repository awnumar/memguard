package x02

import (
	"fmt"

	"github.com/awnumar/memguard"
	"golang.org/x/sys/unix"
)

func POC() {
	key := memguard.NewEnclaveRandom(32)

	var oldLimit unix.Rlimit
	zeroLimit := unix.Rlimit{Cur: 0, Max: oldLimit.Max}
	if err := unix.Prlimit(0, unix.RLIMIT_MEMLOCK, &zeroLimit, &oldLimit); err != nil {
		panic(fmt.Errorf("error lowering memlock rlimit: %s", err))
	}

	keyBytes, err := key.Open()
	if err != nil {
		panic(err)
	}
	defer keyBytes.Destroy()
}
