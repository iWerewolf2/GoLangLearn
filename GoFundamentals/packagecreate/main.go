package main

import (
	"fmt"

	"utils"
)

func main() {
	utils.SayHello()
	x := utils.Sum(10, 20)
	fmt.Println(x)
}
