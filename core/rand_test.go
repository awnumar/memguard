package core

import (
	"bytes"
	"testing"
)

func TestScramble(t *testing.T) {
	b := make([]byte, 32)
	Scramble(b)
	if bytes.Equal(b, make([]byte, 32)) {
		t.Error("buffer not scrambled")
	}
	c := make([]byte, 32)
	Scramble(c)
	if bytes.Equal(b, make([]byte, 32)) {
		t.Error("buffer not scrambled")
	}
	if bytes.Equal(b, c) {
		t.Error("random repeated")
	}
}
