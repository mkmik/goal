package main

import (
	"fmt"
)

// main func
func main() {
	var x int = 1 + 2
	y := x * 13
	y = y + 1

	y(0)[1] = y + 2
	fmt.Println(x, y)
}
