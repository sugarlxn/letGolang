package main

import "fmt"

// 泛型结构
type Pair[T any, U any] struct {
	first  T
	second U
}

// 泛型结构体
type Company[T int | string, S int | string] struct {
	Name  string
	Id    T
	Stuff []S
}

// 泛型切片
type GenerateSlice[T int | int32 | int64] []T

// 泛型hashmap
type GenericMap[K comparable, V int | string | byte] map[K]V

// 泛型接口
type Sayable[T int | string] interface {
	say() T
}

// template func of sum  泛型
func sum[T int | float64 | float32](a, b T) T {
	return a + b
}

func main() {
	fmt.Println(sum[int](3, 5))
	fmt.Println(sum[float64](3.5, 5.2))
	fmt.Println(sum[float32](3.5, 5.2))
}
