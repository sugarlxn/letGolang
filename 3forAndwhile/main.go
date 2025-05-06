package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// TODO
func main() {
	fmt.Println("good day todat")

	var command string = "walk outside"
	var exit = strings.Contains(command, "outside")
	fmt.Println("leave the cave:", exit)

	var num int = rand.Intn(10) + 1
	fmt.Println(num)

	var g1 bool = true
	fmt.Println(g1)

	if command == "walk outside" {
		fmt.Println("good")
	} else if command == "outside" {
		fmt.Println("no bad")
	} else {
		fmt.Println("no")
	}

	var count int = 10
	for count > 0 {
		fmt.Println(count)
		time.Sleep(time.Second)
		count--
		if count < 5 {
			break
		}
	}
	fmt.Println("down")

}
