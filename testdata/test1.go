package main

import (
//	"fmt"
)

/*
func TestX() {
}

func TestY() (int8, int16) {
}
*/

func Test(a, b int64, c int32) int64 {
	var aa, bc int64 = b*int64(c), a
//	var aa int64 = 2
	aa, bc = bc, aa
	return (aa + bc / a) % a
}

func main() {
}

/*

// main func
func main() {
	var x int = 1 + 2
	var y int = x * 13
	y = y + 1
	x, y = y, x

	//fmt.Println(x, y)
}
*/
