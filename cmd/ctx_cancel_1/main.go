package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	ctx0 := context.Background()

	ctx1, _ := context.WithCancel(ctx0)

	go func(ctx1 context.Context) {
		ctx2, cancel2 := context.WithCancel(ctx1)

		go func(ctx2 context.Context) {
			ctx3, _ := context.WithCancel(ctx2)

			go func(ctx3 context.Context) {
				<-ctx3.Done()
				fmt.Println("canceled -> 3")
			}(ctx3)

			<-ctx2.Done()
			fmt.Println("canceled -> 2")
		}(ctx2)

		cancel2()

		<-ctx1.Done()
		fmt.Println("canceled -> 1")

	}(ctx1)

	time.Sleep(time.Second)
}
