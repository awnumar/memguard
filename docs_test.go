package memguard

import "fmt"

func ExampleNew() {
	key, err := New(32, false)
	if err != nil {
		fmt.Println(err)
		SafeExit(1)
	}
	defer key.Destroy()
}

func ExampleCatchInterrupt() {
	CatchInterrupt(func() {
		fmt.Println("Exiting...")
	})
}
