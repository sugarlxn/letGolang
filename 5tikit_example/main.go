package main

import (
	"fmt"
	"math/rand"
	"strings"
)

func main() {
	fmt.Println("tikit example")
	var count int = 10
	var spaceline = [3]string{"space Adventure", "spaceX", "Virgin Galactic"}
	var trip_type = [2]string{"one way", "round trip"}

	fmt.Printf("%-15v %5v %-15v $ %-5v\n", "Space line", "Days", "Trip type", "Price")
	fmt.Println(strings.Repeat("=", 45))
	for count > 0 {
		var line_num int = rand.Intn(3)
		var speed int = rand.Intn(15) + 15
		var trip_num int = rand.Intn(2)
		var price int = rand.Intn(1000) + 1000
		var days int = 621e+5 / speed / 24 / 60 / 60
		if trip_num == 1 {
			price = price * 2
		}
		fmt.Printf("%-15v %5v %-15v $ %-5v\n", spaceline[line_num], days, trip_type[trip_num], price)
		count--
	}

	fmt.Println(strings.Repeat("=", 45))
}
