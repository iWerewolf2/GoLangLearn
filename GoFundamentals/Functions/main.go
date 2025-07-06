package main

import (
	"fmt"
)

func main() {

	x := add(10, 20)
	fmt.Println(x)

	if x > 10 {
		fmt.Println("x is greater than 10")
	} else {
		fmt.Println("x is less than or equal to 10")
	}

}

func add(x int, y int) int {
	return x + y
}
