package main

import (
	"fmt"
)

var programs = make(map[string]func())

// Hardcode the whitelisted programs.
func init() {
}

func main() {
	for k, v := range programs {
		fmt.Printf("Running %s\n", k)
		v()
	}
}
