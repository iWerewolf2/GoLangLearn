package main

import (
	"fmt"
)

func main() {

	var name string
	fmt.Println("What is you name?")
	fmt.Scanf("%s\n", &name)

	var age int
	fmt.Println("How old are you?")
	fmt.Scanf("%d\n", &age)

	fmt.Printf("Hi, %s, your age - %d\n", name, age)

}
