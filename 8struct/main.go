package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// struct 中的变量首字母大写才能导出 json 包要求struct 的字段首字母大写，类似驼峰型命名规范
type Location struct {
	Label string
	X     int
	Y     int
}

type Temperature struct {
	Label string
	Temp  float64
}

// 优先使用对象组合 而不是 类的继承 Favor object composition over class inheritance
// struct 组合 composition
type Weather struct {
	Location    Location
	Temperature Temperature
}

// struct 嵌入 embedding 使用嵌入的方式来组合结构体 Location 和 Temperature 的方法会自动转发
// 直接使用 Weather2 的方法 Weather2.String() 来调用 Location 的方法， 也可以指定某个类型的方法 Weather2.Location.String()
type Weather2 struct {
	Location
	Temperature
}

// Weather2 结构体的 String 方法 要比 Location 自动转发的方法string() 优先级高
func (w2 Weather2) String() string {
	return fmt.Sprintf("Location: %s, Temperature: %.2f", w2.Location.String(), w2.Temperature.Temp)
}

// 关联类型的方法
func (l Location) String() string {
	return fmt.Sprintf("Label: %s, X: %d, Y: %d", l.Label, l.X, l.Y)
}

// 约定 构造函数 函数名称 ： New + 结构体名称 or new + 结构体名称
func NewLocation(label string, x, y int) Location {
	return Location{
		Label: label,
		X:     x,
		Y:     y,
	}
}

// 转发
func (w Weather) String() string {
	return fmt.Sprintf("Location: %s, Temperature: %.2f", w.Location.String(), w.Temperature.Temp)
}

// 值传递 无效更改
func changeLocation(l Location) {
	l.X = 100
	l.Y = 200
}

// exit on error
func ExitOnError(err error) {
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func main() {
	fmt.Println("good day today!!")

	location1 := Location{Label: "Location1", X: 10, Y: 20}
	//深拷贝
	location2 := location1
	location2.Label = "Location2"

	fmt.Printf("Location1: %+v\n", location1)
	fmt.Printf("Location2: %+v\n", location2)
	//struct 转化为 json
	bytes, err := json.Marshal(location1)
	ExitOnError(err)
	fmt.Println("JSON:", string(bytes))

	fmt.Println(location1.String())
}
