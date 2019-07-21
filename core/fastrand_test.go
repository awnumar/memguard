package core

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"io"
	mrand "math/rand"
	"sync"
	"testing"
	"time"
)

func rb(n int) []byte {
	b := make([]byte, n)
	FastRandRead(b)
	return b
}

// TestRead tests that Read produces output with sufficiently high entropy.
func TestFastRandRead(t *testing.T) {
	const size = 10e3

	var b bytes.Buffer
	zip, _ := gzip.NewWriterLevel(&b, gzip.BestCompression)
	if _, err := zip.Write(rb(size)); err != nil {
		t.Fatal(err)
	}
	if err := zip.Close(); err != nil {
		t.Fatal(err)
	}
	if b.Len() < size {
		t.Error("supposedly high entropy bytes have been compressed!")
	}
}

// TestReadConcurrent tests that concurrent calls to 'Read' will not result
// result in identical entropy being produced. Note that for this test to work,
// the points at which 'counter' and 'innerCounter' get incremented need to be
// reduced substantially, to a value like '64'. (larger than the number of
// threads, but not by much).
//
// Note that while this test is capable of catching failures, it's not
// guaranteed to.
func TestReadConcurrent(t *testing.T) {
	threads := 32

	// Spin up threads which will all be collecting entropy from 'Read' in
	// parallel.
	closeChan := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(threads)
	entropys := make([]map[string]struct{}, threads)
	for i := 0; i < threads; i++ {
		entropys[i] = make(map[string]struct{})
		go func(i int) {
			for {
				select {
				case <-closeChan:
					wg.Done()
					return
				default:
				}

				// Read 32 bytes.
				buf := make([]byte, 32)
				FastRandRead(buf)
				bufStr := string(buf)
				_, exists := entropys[i][bufStr]
				if exists {
					t.Error("got the same entropy twice out of the reader")
				}
				entropys[i][bufStr] = struct{}{}
			}
		}(i)
	}

	// Let the threads spin for a bit, then shut them down.
	time.Sleep(time.Millisecond * 1250)
	close(closeChan)
	wg.Wait()

	// Compare the entropy collected and verify that no set of 32 bytes was
	// output twice.
	allEntropy := make(map[string]struct{})
	for _, entropy := range entropys {
		for str := range entropy {
			_, exists := allEntropy[str]
			if exists {
				t.Error("got the same entropy twice out of the reader")
			}
			allEntropy[str] = struct{}{}
		}
	}
}

// TestRandConcurrent checks that there are no race conditions when using the
// rngs concurrently.
func TestRandConcurrent(t *testing.T) {
	// Spin up one goroutine for each exported function. Each goroutine calls
	// its function in a tight loop.

	funcs := []func(){
		// Read some random data into a large byte slice.
		func() { FastRandRead(make([]byte, 16e3)) },

		// Call io.Copy on the global reader.
		func() { io.CopyN(new(bytes.Buffer), FastRandReader, 16e3) },
	}

	closeChan := make(chan struct{})
	var wg sync.WaitGroup
	for i := range funcs {
		wg.Add(1)
		go func(i int) {
			for {
				select {
				case <-closeChan:
					wg.Done()
					return
				default:
				}

				funcs[i]()
			}
		}(i)
	}

	// Allow goroutines to run for a moment.
	time.Sleep(100 * time.Millisecond)

	// Close the channel and wait for everything to clean up.
	close(closeChan)
	wg.Wait()
}

// BenchmarkRead benchmarks the speed of Read for small slices.
func BenchmarkRead32(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(32)
	buf := make([]byte, 32)
	for i := 0; i < b.N; i++ {
		FastRandRead(buf)
	}
}

// BenchmarkRead512kb benchmarks the speed of Read for larger slices.
func BenchmarkRead512kb(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(512e3)
	buf := make([]byte, 512e3)
	for i := 0; i < b.N; i++ {
		FastRandRead(buf)
	}
}

// BenchmarkRead4Threads32 benchmarks the speed of Read when it's being using
// across four threads.
func BenchmarkRead4Threads32(b *testing.B) {
	b.ReportAllocs()
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			buf := make([]byte, 32)
			<-start
			for i := 0; i < b.N; i++ {
				FastRandRead(buf)
			}
			wg.Done()
		}()
	}
	b.SetBytes(4 * 32)

	// Signal all threads to begin
	b.ResetTimer()
	close(start)
	// Wait for all threads to exit
	wg.Wait()
}

// BenchmarkRead4Threads512kb benchmarks the speed of Read when it's being using
// across four threads with 512kb read sizes.
func BenchmarkRead4Threads512kb(b *testing.B) {
	b.ReportAllocs()
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			buf := make([]byte, 512e3)
			<-start
			for i := 0; i < b.N; i++ {
				FastRandRead(buf)
			}
			wg.Done()
		}()
	}
	b.SetBytes(4 * 512e3)

	// Signal all threads to begin
	b.ResetTimer()
	close(start)
	// Wait for all threads to exit
	wg.Wait()
}

// BenchmarkRead64Threads32 benchmarks the speed of Read when it's being using
// across 64 threads.
func BenchmarkRead64Threads32(b *testing.B) {
	b.ReportAllocs()
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			buf := make([]byte, 32)
			<-start
			for i := 0; i < b.N; i++ {
				FastRandRead(buf)
			}
			wg.Done()
		}()
	}
	b.SetBytes(64 * 32)

	// Signal all threads to begin
	b.ResetTimer()
	close(start)
	// Wait for all threads to exit
	wg.Wait()
}

// BenchmarkRead64Threads512kb benchmarks the speed of Read when it's being using
// across 64 threads with 512kb read sizes.
func BenchmarkRead64Threads512kb(b *testing.B) {
	b.ReportAllocs()
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			buf := make([]byte, 512e3)
			<-start
			for i := 0; i < b.N; i++ {
				FastRandRead(buf)
			}
			wg.Done()
		}()
	}
	b.SetBytes(64 * 512e3)

	// Signal all threads to begin
	b.ResetTimer()
	close(start)
	// Wait for all threads to exit
	wg.Wait()
}

// BenchmarkReadCrypto benchmarks the speed of (crypto/rand).Read for small
// slices. This establishes a lower limit for BenchmarkRead32.
func BenchmarkReadCrypto32(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(32)
	buf := make([]byte, 32)
	for i := 0; i < b.N; i++ {
		rand.Read(buf)
	}
}

// BenchmarkReadCrypto512kb benchmarks the speed of (crypto/rand).Read for larger
// slices. This establishes a lower limit for BenchmarkRead512kb.
func BenchmarkReadCrypto512kb(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(512e3)
	buf := make([]byte, 512e3)
	for i := 0; i < b.N; i++ {
		rand.Read(buf)
	}
}

// BenchmarkReadCrypto4Threads32 benchmarks the speed of rand.Read when its
// being used across 4 threads with 32 byte read sizes.
func BenchmarkReadCrypto4Threads32(b *testing.B) {
	b.ReportAllocs()
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			buf := make([]byte, 32)
			<-start
			for i := 0; i < b.N; i++ {
				_, err := rand.Read(buf)
				if err != nil {
					b.Fatal(err)
				}
			}
			wg.Done()
		}()
	}
	b.SetBytes(4 * 32)

	// Signal all threads to begin
	b.ResetTimer()
	close(start)
	// Wait for all threads to exit
	wg.Wait()
}

// BenchmarkReadCrypto4Threads512kb benchmarks the speed of rand.Read when its
// being used across 4 threads with 512 kb read sizes.
func BenchmarkReadCrypto4Threads512kb(b *testing.B) {
	b.ReportAllocs()
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			buf := make([]byte, 512e3)
			<-start
			for i := 0; i < b.N; i++ {
				_, err := rand.Read(buf)
				if err != nil {
					b.Fatal(err)
				}
			}
			wg.Done()
		}()
	}
	b.SetBytes(4 * 512e3)

	// Signal all threads to begin
	b.ResetTimer()
	close(start)
	// Wait for all threads to exit
	wg.Wait()
}

// BenchmarkReadCrypto64Threads32 benchmarks the speed of rand.Read when its
// being used across 4 threads with 32 byte read sizes.
func BenchmarkReadCrypto64Threads32(b *testing.B) {
	b.ReportAllocs()
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			buf := make([]byte, 32)
			<-start
			for i := 0; i < b.N; i++ {
				_, err := rand.Read(buf)
				if err != nil {
					b.Fatal(err)
				}
			}
			wg.Done()
		}()
	}
	b.SetBytes(64 * 32)

	// Signal all threads to begin
	b.ResetTimer()
	close(start)
	// Wait for all threads to exit
	wg.Wait()
}

// BenchmarkReadCrypto64Threads512k benchmarks the speed of rand.Read when its
// being used across 4 threads with 512 kb read sizes.
func BenchmarkReadCrypto64Threads512kb(b *testing.B) {
	b.ReportAllocs()
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			buf := make([]byte, 512e3)
			<-start
			for i := 0; i < b.N; i++ {
				_, err := rand.Read(buf)
				if err != nil {
					b.Fatal(err)
				}
			}
			wg.Done()
		}()
	}
	b.SetBytes(64 * 512e3)

	// Signal all threads to begin
	b.ResetTimer()
	close(start)
	// Wait for all threads to exit
	wg.Wait()
}

// BenchmarkReadMath benchmarks the speed of (math/rand).Read for small
// slices. This establishes an upper limit for BenchmarkRead32.
func BenchmarkReadMath32(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(32)
	buf := make([]byte, 32)
	for i := 0; i < b.N; i++ {
		mrand.Read(buf)
	}
}

// BenchmarkReadMath512kb benchmarks the speed of (math/rand).Read for larger
// slices. This establishes an upper limit for BenchmarkRead512kb.
func BenchmarkReadMath512kb(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(512e3)
	buf := make([]byte, 512e3)
	for i := 0; i < b.N; i++ {
		mrand.Read(buf)
	}
}

// BenchmarkReadMath4Threads32 benchmarks the speed of ReadMath when it's being using
// across four threads.
func BenchmarkReadMath4Threads32(b *testing.B) {
	b.ReportAllocs()
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			buf := make([]byte, 32)
			<-start
			for i := 0; i < b.N; i++ {
				mrand.Read(buf)
			}
			wg.Done()
		}()
	}
	b.SetBytes(4 * 32)

	// Signal all threads to begin
	b.ResetTimer()
	close(start)
	// Wait for all threads to exit
	wg.Wait()
}

// BenchmarkReadMath4Threads512kb benchmarks the speed of ReadMath when it's being using
// across four threads with 512kb read sizes.
func BenchmarkReadMath4Threads512kb(b *testing.B) {
	b.ReportAllocs()
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			buf := make([]byte, 512e3)
			<-start
			for i := 0; i < b.N; i++ {
				mrand.Read(buf)
			}
			wg.Done()
		}()
	}
	b.SetBytes(4 * 512e3)

	// Signal all threads to begin
	b.ResetTimer()
	close(start)
	// Wait for all threads to exit
	wg.Wait()
}

// BenchmarkReadMath64Threads32 benchmarks the speed of ReadMath when it's being using
// across 64 threads.
func BenchmarkReadMath64Threads32(b *testing.B) {
	b.ReportAllocs()
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			buf := make([]byte, 32)
			<-start
			for i := 0; i < b.N; i++ {
				mrand.Read(buf)
			}
			wg.Done()
		}()
	}
	b.SetBytes(64 * 32)

	// Signal all threads to begin
	b.ResetTimer()
	close(start)
	// Wait for all threads to exit
	wg.Wait()
}

// BenchmarkReadMath64Threads512kb benchmarks the speed of ReadMath when it's being using
// across 64 threads with 512kb read sizes.
func BenchmarkReadMath64Threads512kb(b *testing.B) {
	b.ReportAllocs()
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			buf := make([]byte, 512e3)
			<-start
			for i := 0; i < b.N; i++ {
				mrand.Read(buf)
			}
			wg.Done()
		}()
	}
	b.SetBytes(64 * 512e3)

	// Signal all threads to begin
	b.ResetTimer()
	close(start)
	// Wait for all threads to exit
	wg.Wait()
}
