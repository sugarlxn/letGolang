package main

import (
	"fmt"
	"time"
)

func sleepyGopyer(id int, c chan int) {
	fmt.Printf("Gopyer %d is sleeping\n", id)
	time.Sleep(2 * time.Second)
	c <- id
}

// 遍历读取channel 从通道中读取值直到它关闭为止
func printGopher(upstream chan string) {
	for msg := range upstream {
		fmt.Println(msg)
	}
}

// 使用channel 在goroutines中传递数据
func main() {

	fmt.Println("good day today!!")
	c := make(chan int)
	for i := 0; i < 5; i++ {
		go sleepyGopyer(i, c)
	}
	//使用channel 等待goroutines完成 goroutine 同步
	for i := 0; i < 5; i++ {
		id := <-c
		fmt.Printf("Gopyer %d is done\n", id)
	}
}
