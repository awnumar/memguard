package x01

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

const duration = 10 * time.Second

func TestPanicsPoC(t *testing.T) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	go func() {
		select {
		case <-sigs:
			cancel()
		}
	}()
	OpenEnclave(ctx)
}

// #############

// panic: runtime error: index out of range [0] with length 0

// goroutine 2060 [running]:
// github.com/awnumar/memguard/core.(*Coffer).View(0xc0000962a0, 0x0, 0x0, 0x0)
// 	/home/awn/src/go/src/github.com/awnumar/memguard/core/coffer.go:112 +0x3a9
// github.com/awnumar/memguard/core.Open(0xc00000e0e0, 0xc00001c290, 0xc00007a591, 0xc00007a590)
// 	/home/awn/src/go/src/github.com/awnumar/memguard/core/enclave.go:101 +0xa5
// github.com/awnumar/memguard.(*Enclave).Open(0xc000010040, 0xc000070770, 0x1, 0x1)
// 	/home/awn/src/go/src/github.com/awnumar/memguard/enclave.go:43 +0x50
// github.com/awnumar/memguard/examples/panics.openVerify(0xc000010040, 0xc000016600, 0x20, 0x20, 0x0, 0x0)
// 	/home/awn/src/go/src/github.com/awnumar/memguard/examples/panics/poc.go:55 +0x5c
// github.com/awnumar/memguard/examples/panics.immediateOpen.func1(0xc000010040, 0xc000016600, 0x20, 0x20, 0xc00024af60)
// 	/home/awn/src/go/src/github.com/awnumar/memguard/examples/panics/poc.go:70 +0x5b
// created by github.com/awnumar/memguard/examples/panics.immediateOpen
// 	/home/awn/src/go/src/github.com/awnumar/memguard/examples/panics/poc.go:69 +0xdf
// FAIL	github.com/awnumar/memguard/examples/panics	0.835s
// FAIL

// #############

// WARNING: DATA RACE
// Write at 0x0000007aa588 by goroutine 114:
//   github.com/awnumar/memguard/core.Purge()
//       /home/awn/src/go/src/github.com/awnumar/memguard/core/exit.go:54 +0x82
//   github.com/awnumar/memguard/core.Panic()
//       /home/awn/src/go/src/github.com/awnumar/memguard/core/exit.go:85 +0x2f
//   github.com/awnumar/memguard/core.NewBuffer()
//       /home/awn/src/go/src/github.com/awnumar/memguard/core/buffer.go:75 +0x8ce
//   github.com/awnumar/memguard/core.Open()
//       /home/awn/src/go/src/github.com/awnumar/memguard/core/enclave.go:95 +0x6b
//   github.com/awnumar/memguard.(*Enclave).Open()
//       /home/awn/src/go/src/github.com/awnumar/memguard/enclave.go:43 +0x4f
//   github.com/awnumar/memguard/examples/unsound.openVerify()
//       /home/awn/src/go/src/github.com/awnumar/memguard/examples/unsound/poc.go:55 +0x5b
//   github.com/awnumar/memguard/examples/unsound.immediateOpen.func1()
//       /home/awn/src/go/src/github.com/awnumar/memguard/examples/unsound/poc.go:70 +0x5a

// Previous read at 0x0000007aa588 by goroutine 50:
//   github.com/awnumar/memguard/core.Purge.func1()
//       /home/awn/src/go/src/github.com/awnumar/memguard/core/exit.go:22 +0x52
//   github.com/awnumar/memguard/core.Purge()
//       /home/awn/src/go/src/github.com/awnumar/memguard/core/exit.go:50 +0x44
//   github.com/awnumar/memguard/core.Panic()
//       /home/awn/src/go/src/github.com/awnumar/memguard/core/exit.go:85 +0x2f
//   github.com/awnumar/memguard.(*Enclave).Open()
//       /home/awn/src/go/src/github.com/awnumar/memguard/enclave.go:46 +0xa7
//   github.com/awnumar/memguard/examples/unsound.openVerify()
//       /home/awn/src/go/src/github.com/awnumar/memguard/examples/unsound/poc.go:55 +0x5b
//   github.com/awnumar/memguard/examples/unsound.immediateOpen.func1()
//       /home/awn/src/go/src/github.com/awnumar/memguard/examples/unsound/poc.go:70 +0x5a

// Goroutine 114 (running) created at:
//   github.com/awnumar/memguard/examples/unsound.immediateOpen()
//       /home/awn/src/go/src/github.com/awnumar/memguard/examples/unsound/poc.go:69 +0xde
//   github.com/awnumar/memguard/examples/unsound.OpenEnclave.func1()
//       /home/awn/src/go/src/github.com/awnumar/memguard/examples/unsound/poc.go:40 +0x238

// Goroutine 50 (running) created at:
//   github.com/awnumar/memguard/examples/unsound.immediateOpen()
//       /home/awn/src/go/src/github.com/awnumar/memguard/examples/unsound/poc.go:69 +0xde
//   github.com/awnumar/memguard/examples/unsound.OpenEnclave.func1()
//       /home/awn/src/go/src/github.com/awnumar/memguard/examples/unsound/poc.go:40 +0x238

// #############

// panic: <memcall> could not acquire lock on 0x7f7ef201d000, limit reached? [Err: cannot allocate memory]

// goroutine 1992 [running]:
// github.com/awnumar/memguard/core.Panic(0x607140, 0xc000223a10)
// 	/home/awn/src/go/src/github.com/awnumar/memguard/core/exit.go:86 +0x48
// github.com/awnumar/memguard/core.NewBuffer(0x20, 0x7d61c0, 0x64, 0xd)
// 	/home/awn/src/go/src/github.com/awnumar/memguard/core/buffer.go:75 +0x8cf
// github.com/awnumar/memguard/core.Open(0xc0000b80c0, 0xc000016520, 0xc0000783f1, 0xc0000783f0)
// 	/home/awn/src/go/src/github.com/awnumar/memguard/core/enclave.go:95 +0x6c
// github.com/awnumar/memguard.(*Enclave).Open(0xc0000a6048, 0xc0000eef70, 0x1, 0x1)
// 	/home/awn/src/go/src/github.com/awnumar/memguard/enclave.go:43 +0x50
// github.com/awnumar/memguard/examples/unsound.openVerify(0xc0000a6048, 0xc0000da080, 0x20, 0x20, 0x0, 0x0)
// 	/home/awn/src/go/src/github.com/awnumar/memguard/examples/unsound/poc.go:55 +0x5c
// github.com/awnumar/memguard/examples/unsound.immediateOpen.func1(0xc0000a6048, 0xc0000da080, 0x20, 0x20, 0xc000236a80)
// 	/home/awn/src/go/src/github.com/awnumar/memguard/examples/unsound/poc.go:70 +0x5b
// created by github.com/awnumar/memguard/examples/unsound.immediateOpen
// 	/home/awn/src/go/src/github.com/awnumar/memguard/examples/unsound/poc.go:69 +0xdf
// FAIL	github.com/awnumar/memguard/examples/unsound	0.853s
// FAIL
