package main

import (
	"fmt"
)

func TestX() {
}

func TestY() (int8, int16) {
}


func Test(a, b int8, c int) int64 {
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
