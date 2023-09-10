package deadlock

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/awnumar/memguard"
	"lukechampine.com/frand"
)

func OpenEnclave(ctx context.Context) {
	n := 10
	data := make([][]byte, n)
	enclaves := make([]*memguard.Enclave, n)
	for i := range data {
		data[i] = make([]byte, 32)
		buf := make([]byte, 32)
		if _, err := io.ReadFull(rand.Reader, buf); err != nil {
			panic("failed to read random data")
		}
		copy(data[i], buf)
		enclaves[i] = memguard.NewEnclave(buf)
	}

	threads := 20
	for i := 0; i < threads; i++ {
		go func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					j := frand.Intn(n)
					immediateOpen(ctx, enclaves[j], data[j])
				}
			}
		}(ctx)
	}
	<-ctx.Done()
	time.Sleep(time.Second)

	// buf := make([]byte, 1<<20)
	// fmt.Println(string(buf[:runtime.Stack(buf, true)]))
}

func openVerify(lock *memguard.Enclave, exp []byte) error {
	lb, err := lock.Open()
	if err != nil {
		return err
	}
	defer lb.Destroy()
	if !bytes.Equal(lb.Bytes(), exp) {
		fmt.Println(lb.Bytes(), exp)
		return errors.New("open verify fail")
	}
	return nil
}

func immediateOpen(ctx context.Context, lock *memguard.Enclave, exp []byte) {
	c1 := make(chan error, 1)
	go func() {
		err := openVerify(lock, exp)
		c1 <- err
	}()

	select {
	case err := <-c1:
		if err != nil {
			panic(err)
		}
	case <-ctx.Done():
	}
}
