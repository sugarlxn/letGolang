package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {

	s := strings.Join(os.Args[1:], " ")
	fmt.Println(s)
	//读入用户输入
	fmt.Println("what is your name?")
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	fmt.Println("your name is ", text)

}
