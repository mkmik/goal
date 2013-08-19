package main

import (
	"fmt"
)

func Test() int {
	return 10 + 2
}

// main func
func main() {
	var x int = 1 + 2
	var y int = x * 13
	y = y + 1
	x, y = y, x

	//fmt.Println(x, y)
}
