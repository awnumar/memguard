package stream

import (
	"fmt"
	"testing"
)

func TestSlowRandByte(t *testing.T) {
	randByte := SlowRandByte()
	fmt.Println("Random byte:", randByte)
}
