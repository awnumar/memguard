package deadlock

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"runtime"
	"time"

	"github.com/awnumar/memguard"
)

// OpenEnclave ...
func OpenEnclave(ctx context.Context) {
	n := 10
	data := make([][]byte, n)
	enclaves := make([]*memguard.Enclave, n)
	for i := range data {
		data[i] = make([]byte, 32)
		buf := make([]byte, 32)
		io.ReadFull(rand.Reader, buf)
		copy(data[i], buf)
		enclaves[i] = memguard.NewEnclave(buf)
	}

	threads := 20
	for i := 0; i < threads; i++ {
		j := 0
		go func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					{
						fmt.Printf("open enclave %d \n", j)
						immediateOpen(ctx, enclaves[j], data[j])
						j = (j + 1) % n
					}
				}
			}
		}(ctx)
	}
	<-ctx.Done()
	time.Sleep(time.Second)

	buf := make([]byte, 1<<20)
	fmt.Println(string(buf[:runtime.Stack(buf, true)]))
}

func openVerify(lock *memguard.Enclave, exp []byte) error {
	lb, err := lock.Open()
	if err != nil {
		return err
	}
	defer lb.Destroy()
	if !bytes.Equal(lb.Bytes(), exp) {
		return errors.New("open verify fail")
	}
	return nil
}

func immediateOpen(ctx context.Context, lock *memguard.Enclave, exp []byte) {
	start := time.Now()
	c1 := make(chan error, 1)
	go func() {
		err := openVerify(lock, exp)
		c1 <- err
	}()
	var dur time.Duration
	select {
	case err := <-c1:
		{
			dur = time.Since(start)
			if err != nil {
				fmt.Printf("### open fail: %s \n", err)
			}
		}
	case <-ctx.Done():
		{
			dur = time.Since(start)
			fmt.Printf("### timeout \n")
		}
	}
	fmt.Printf("%d, %d \n", start.UnixNano(), dur.Nanoseconds())
	//_ = dur
}
