package main

import (
	"fmt"
)

// 结构体, 结构体是值类型
type Person struct {
	name string
	age  int
}

func main() {
	fmt.Println("Let Golang!")

	const (
		i = iota
		j = iota
		k = iota
	)

	fmt.Println(i, j, k)

	var day string = "Monday"

	if day == "Monday" || day == "Tuesday" {
		fmt.Println("Today is Monday or Tuesday")
	} else if day == "Wednesday" || day == "Thursday" {
		fmt.Println("Today is Wednesday or Thursday")
	} else {
		fmt.Println("Today is not Monday or Tuesday")
	}

	switch day {
	case "Monday", "Tuesday":
		fmt.Println("Today is Monday or Tuesday")
	case "Wednesday", "Thursday":
		fmt.Println("Today is Wednesday or Thursday")
	default:
		fmt.Println("Today is not Monday or Tuesday")
	}

	sum := 0
	for i := 0; i < 10; i++ {
		sum += i
	}
	fmt.Println(sum)

	numbers := []int{1, 2, 3, 4, 5}
	for index, number := range numbers {
		fmt.Println(index, number)
	}

	str := "Hello, World!"
	for index, char := range str {
		fmt.Println(index, char)
	}

	mymap := map[string]int{
		"one":   1,
		"two":   2,
		"three": 3,
	}
	for key, value := range mymap {
		fmt.Println(key, value)
	}

	ch := make(chan int)
	go func() {
		ch <- 1
		ch <- 2
		ch <- 3
		ch <- 4
		ch <- 5
		close(ch)
	}()

	for v := range ch {
		fmt.Println(v)
	}

}
