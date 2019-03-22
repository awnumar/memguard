package examples

import "testing"

/* [Apache-2.0] SocketKey :: Awn Umar <awn@spacetime.dev> */

func TestSocketKey(t *testing.T) {
	SocketKey(32)
}

func benchmarkSocketKey(size int, b *testing.B) {
	for n := 0; n < b.N; n++ {
		SocketKey(size)
	}
}

func BenchmarkSocketKey32(b *testing.B)       { benchmarkSocketKey(32, b) }
func BenchmarkSocketKey64(b *testing.B)       { benchmarkSocketKey(64, b) }
func BenchmarkSocketKey256(b *testing.B)      { benchmarkSocketKey(256, b) }
func BenchmarkSocketKey4096(b *testing.B)     { benchmarkSocketKey(4096, b) }
func BenchmarkSocketKey1048576(b *testing.B)  { benchmarkSocketKey(1048576, b) }
func BenchmarkSocketKey16777216(b *testing.B) { benchmarkSocketKey(16777216, b) }
