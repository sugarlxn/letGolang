package main

import (
	"fmt"
)

func change_value(slice []int) {
	slice[1] = 100 // Change the value at index 0
}

type word_slice []string

func (w word_slice) new_word() {
	for i := range w {
		w[i] = "new_" + w[i]
	}
}

func (w word_slice) dump() {
	fmt.Println("len:", len(w), "cap:", cap(w), "slice:", w)
}

func dump(slice []string) {
	fmt.Println("len:", len(slice), "cap:", cap(slice), "slice:", slice)
}

func main() {

	// Example of a slice
	var slice1 = []int{1, 2, 3, 4, 5}
	var slice2 = []int{6, 7, 8, 9, 10}

	// Concatenate slices
	slice3 := append(slice1, slice2...)

	slice4 := slice3[0:4] // Slice from index 1 to 4 (exclusive)

	slice4[0] = 100 // Change the value at index 1 of slice4

	fmt.Println("slice3:", slice3)
	fmt.Println("Slice 4:", slice4)

	change_value(slice3)
	fmt.Println("slice3 after change_value:", slice3)
	// var name type = value
	var good word_slice = []string{"mar", "uranus", "neptune", "pluto"}
	good.new_word()
	fmt.Println("good:", good)

	dwarfs := []string{"ceres", "pluto", "haumea", "makemake", "eris"}
	dump(dwarfs)
	dump(dwarfs[1:2])

	good.dump()

}
