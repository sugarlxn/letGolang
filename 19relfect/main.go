package main

import (
	"fmt"
	"reflect"
)

type User struct {
	Name string
	Age  int
}

func main() {

	author := "lxn"
	fmt.Println("typeof author:", reflect.TypeOf(author))
	fmt.Println("valueof author:", reflect.ValueOf(author))

	user := User{Name: "lxn", Age: 18}
	fmt.Println("typeof user:", reflect.TypeOf(user))
	fmt.Println("valueof user:", reflect.ValueOf(user))

}
