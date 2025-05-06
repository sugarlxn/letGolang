package main

import (
	"fmt"
	"time"
)

func selectMain() {
	ch1 := make(chan string)
	ch2 := make(chan string)
	done := make(chan bool)

	go func() {
		time.Sleep(2 * time.Second)
		ch1 <- "message from ch1"
		done <- false
	}()

	go func() {
		time.Sleep(1 * time.Second)
		ch2 <- "message from ch2"
	}()

	for {
		select {
		case msg1 := <-ch1:
			fmt.Println("Received", msg1)
		case msg2 := <-ch2:
			fmt.Println("Received", msg2)
		case run := <-done:
			fmt.Println("run", run)
			return
		}
	}
}
