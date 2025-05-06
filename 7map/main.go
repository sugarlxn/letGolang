package main

import (
	"fmt"
)

func main() {

	fmt.Println("good day today!!")

	var temperature = map[string]float64{
		"Monday":    20.5,
		"Tuesday":   22.0,
		"Wednesday": 19.8,
		"Thursday":  21.2,
		"Friday":    23.5,
	}

	// Check if the key exists
	if temp, ok := temperature["Monday1"]; ok {
		fmt.Println("Temperature on Monday1:", temp)
	} else {
		fmt.Println("Monday1 not found")
	}

	fmt.Println("Temperature map:", temperature)

}
