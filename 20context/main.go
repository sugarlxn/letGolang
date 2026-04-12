package main

import (
	"context"
	"fmt"
	"time"
)

func handler(ctx context.Context, duration time.Duration) {
	select {
	case <-time.After(duration):
		fmt.Println("handler completed")
	case <-ctx.Done():
		fmt.Println("handler canceled:", ctx.Err())
	}
}

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	go handler(ctx, 500*time.Millisecond)

	select {
	case <-ctx.Done():
		fmt.Println("main timeout:", ctx.Err())
	case <-time.After(2 * time.Second):
		fmt.Println("main completed")
	}
}
