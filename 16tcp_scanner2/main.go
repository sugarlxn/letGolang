package main

import (
	"fmt"
	"net"
	"sync"
)

func work(ports chan int, wg *sync.WaitGroup) {
	for port := range ports {
		// 执行扫描操作
		fmt.Println("Scanning port:", port)
		address := fmt.Sprintf("172.31.105.55:%d", port)
		conn, err := net.Dial("tcp", address)

		//端口打开成功
		if err == nil {
			conn.Close()
			fmt.Printf("Port %d is open\n", port)
		}
		wg.Done()
	}
}

func main() {

	ports := make(chan int, 100)
	var wg sync.WaitGroup
	//创建100个协程
	for range cap(ports) {
		go work(ports, &wg)
	}

	// 扫描所有端口
	for i := 1; i <= 1024; i++ {
		wg.Add(1)
		ports <- i
	}

	wg.Wait()
	close(ports)
	fmt.Println("All ports scanned.")
}
