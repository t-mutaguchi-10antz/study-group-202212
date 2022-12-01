package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	ctx0 := context.Background()

	ctx1, cancel1 := context.WithCancel(ctx0)

	go func(ctx1 context.Context) {
		<-ctx1.Done()
		fmt.Println("canceled -> 1-1")
	}(ctx1)

	go func(ctx1 context.Context) {
		<-ctx1.Done()
		fmt.Println("canceled -> 1-2")
	}(ctx1)

	cancel1()

	time.Sleep(time.Second)
}
