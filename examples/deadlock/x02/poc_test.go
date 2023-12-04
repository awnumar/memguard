//go:build linux

package x02

import (
	"log"
	"testing"
)

func TestPOC(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("panic occurred:", err)
		}
	}()

	POC()
}
