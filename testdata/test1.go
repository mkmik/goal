package main

import (
	"fmt"
)

/*
func TestX() {
}

func TestY() (int8, int16) {
}
*/

func Test(a, b int32, c int32) int32 {
	var aa, bc int32 = b*c, a
	aa, bc = bc, aa
	return (aa + bc / a) % a
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
