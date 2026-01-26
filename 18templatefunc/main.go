package main

import "fmt"

// template func of sum  泛型
func sum[T int | float64 | float32](a, b T) T {
	return a + b
}

func main() {
	fmt.Println(sum[int](3, 5))
	fmt.Println(sum[float64](3.5, 5.2))
	fmt.Println(sum[float32](3.5, 5.2))
}
