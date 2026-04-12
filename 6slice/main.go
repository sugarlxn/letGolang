package main

import (
	"fmt"
	"reflect"
	"unsafe"
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

func badAppend(s []int) {
	v := reflect.ValueOf(s)
	fmt.Println("kind:", v.Kind(), "len:", v.Len(), "cap:", v.Cap())
	s = append(s, 100) // This will not change the original slice outside this function
	fmt.Println("badapppend", len(s), cap(s), s)
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

	arr := [8]int{1, 2, 3}
	fmt.Println("arr:", arr, "len:", len(arr), "cap:", cap(arr))
	fmt.Println("arr[1:3]", arr[1:3])

	fmt.Println("valueof arr:", reflect.ValueOf(arr), "typeof arr:", reflect.TypeOf(arr))
	fmt.Println("typefor equal  typeof but [type] type must be a type: ", reflect.TypeFor[word_slice]())

	//slice 是一个结构体 type slice struct { Data uintptr; Len int; Cap int }
	//可以使用reflect包来查看slice的底层结构
	fmt.Println("valueof arr:", reflect.ValueOf(arr), "typeof arr:", reflect.TypeOf(arr))
	v := reflect.ValueOf(arr)
	fmt.Println(v.Kind(), v.Len(), v.Cap(), v.Index(0))
	//也可以通过reflect.SliceHeader来查看slice的底层结构
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&arr))
	fmt.Printf("arr header: Data=%x, Len=%d, Cap=%d\n", sliceHeader.Data, sliceHeader.Len, sliceHeader.Cap)

	good_slice := make([]int, 4, 8)
	good_slice[0] = 1
	good_slice[1] = 2
	good_slice[2] = 3
	good_slice[3] = 4
	fmt.Println("good_slice:", good_slice, "len:", len(good_slice), "cap:", cap(good_slice))

	badAppend(good_slice)
	good_slice = good_slice[:cap(good_slice)] // Extend the slice to its capacity
	fmt.Println("good_slice after badAppend:", good_slice)

}
