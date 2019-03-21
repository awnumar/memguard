package examples

import "fmt"

var examples = make(map[string]func())

func init() {
	examples["socketkey"] = socketkey
}

func main() {
	for k, v := range examples {
		fmt.Println("Running", k)
		v()
	}
}
