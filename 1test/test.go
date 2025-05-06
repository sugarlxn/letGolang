package main

import (
	"fmt"
	"math/rand"
)

func main() {
	fmt.Println("Hello, World!")
	fmt.Print("good day today!\n")

	fmt.Printf("today is %v, %v, %v\n", "2025-05-04", 234, "lxn")
	fmt.Printf("%-15v $ %4v\n", "ice cream", 4.99)
	fmt.Printf("%-15v $ %4v\n", "coffee", 3.99)
	fmt.Printf("%-15v $ %4v\n", "tea", 2.99)
	fmt.Printf("%-15v $ %4v\n", "water", 1.99)
	fmt.Printf("%-15v $ %4v\n", "apple", 0.99)
	fmt.Printf("%-15v $ %4v\n", "banana", 0.99)
	fmt.Printf("%-15v $ %4v\n", "orange", 0.99)
	fmt.Printf("%-15v $ %4v\n", "pear", 0.99)

	var a int = 10
	const speed int = 100

	a = a * 2
	a *= 2
	a++

	fmt.Println(a)
	fmt.Println(speed)

	var num int = rand.Intn(10) + 1
	fmt.Println(num)

	var num2 int = rand.Intn(10) + 1
	fmt.Println(num2)

	var distance int = 56e+6
	fmt.Println(distance)
	var time int = 28
	var speed2 int = distance / time / 24
	fmt.Println("Malacandra distance is", distance, "km,", "we can go there by speed:", speed2, "km/h", "with time:", time, "days")
}
