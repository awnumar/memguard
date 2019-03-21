package memguard

import "fmt"

func ExampleCatchInterrupt() {
	CatchInterrupt(func() {
		fmt.Println("Exiting...")
	})
}
