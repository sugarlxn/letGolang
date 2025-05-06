package main

import (
	"fmt"
	"strings"
)

// 接口的定义 按照约定，接口的名称以er结尾
type Talker interface {
	talk() string // 接口方法名 返回类型
}

// 接口的实现
type laser int

func (l laser) talk() string {
	return strings.Repeat("laser", int(l))

}

type dog struct{}

func (d dog) talk() string {
	return "woof"
}

// 接口类型的使用
func shout(t Talker) string {
	return strings.ToUpper(t.talk())
}

// 复用 fmt 中的 Stringer 接口
// 接口定义
//
//	type Stringer interface {
//		String() string
//	}
type Person struct {
	Name string
	Age  int
}

func (p Person) String() string {
	return fmt.Sprintf("Name: %s, Age: %d", p.Name, p.Age)
}

func main() {
	//golang 多态的表现形式
	var t Talker = laser(1) // 接口变量赋值
	fmt.Println(t.talk())   // 调用接口方法
	t = dog{}               // 接口变量赋值
	fmt.Println(t.talk())   // 调用接口方法

	t = laser(3) // 接口变量赋值
	// 接口的使用
	fmt.Println(shout(t))     // 调用接口方法
	fmt.Println(shout(dog{})) // 调用接口方法 dog{} 满足 Talker 接口

	// 接口的使用
	p := Person{"Alice", 30}
	fmt.Println(p) // 调用接口方法 p 符合 Stringer 接口 可以直接打印
}
