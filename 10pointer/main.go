package main

import (
	"fmt"
)

type Person struct {
	Name string
	Age  int
}

func (p *Person) IncreaseAge() {
	if p == nil {
		return
	}
	p.Age++
}

func (p Person) String() string {
	return fmt.Sprintf("Name: %s, Age: %d", p.Name, p.Age)
}

func main() {

	fmt.Println("good day today!!")
	p := Person{"Alice", 30}
	p.IncreaseAge()
	fmt.Println(p)

	var nowshere *int
	fmt.Println(nowshere)  // nil
	fmt.Println(*nowshere) // panic: runtime error: invalid memory address or nil pointer dereference

}
