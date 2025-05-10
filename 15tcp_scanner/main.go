package main

import (
	"fmt"
	"net"
	"sync"
)

func main() {

	var wg sync.WaitGroup

	for i := 1; i <= 65535; i++ {
		wg.Add(1)
		go func(port int) {
			defer wg.Done()
			// 执行扫描操作
			address := fmt.Sprintf("172.31.105.55:%d", port)
			conn, err := net.Dial("tcp", address)
			if err != nil {
				// fmt.Printf("Port %d is closed\n", port)
				return
			}
			conn.Close()
			fmt.Printf("Port %d is open\n", port)

		}(i)
	}

	//等待
	wg.Wait()
	fmt.Println("All ports scanned.")
}
