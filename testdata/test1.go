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

func TestCmp(a, b int64) bool {
	return a > b
}

/*
func Test(a, b int64, c int32) int64 {
	var aa, bc int64 = b * int64(c), a
	aa, bc = bc, aa
	var x int64
	if x = a + 2; x > 0 {
		var xaa int64 = 10 + aa
		aa = xaa
	} else {
		aa = bc + 66
		bc = bc + 10
	}

	return (aa + bc/a) % a
}
*/
func TestNested(a int64) int64 {
	if a > 0 {

		if a < 2 {
			//a = 100
			var z int64 = a + 1
		}

		a = a + 1
	}
	return a
}

/*
func TestScope(a int64) int64 {
	if a:=1; a>0 {
		a = 2
	}
	return a
}
*/

//func main(a int32) {
//Test(1,2,3)
//	1+a
//}

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
