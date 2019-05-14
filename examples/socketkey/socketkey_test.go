package socketkey

import "testing"

func TestSocketKey(t *testing.T) {
	SocketKey(32)
}

func BenchmarkSocketKey32(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		SocketKey(32)
	}
}
