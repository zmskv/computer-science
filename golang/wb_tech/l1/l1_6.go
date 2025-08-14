package l1

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

func stopByCondition() {
	go func() {
		for i := 1; i <= 5; i++ {
			fmt.Println("work work work!")
			time.Sleep(300 * time.Millisecond)
		}
	}()
	time.Sleep(2 * time.Second)
}

func stopByChannel() {
	stop := make(chan struct{})

	go func() {
		for {
			select {
			case <-stop:
				fmt.Println("channel stopped")
				return
			default:
				fmt.Println("work work work!")
				time.Sleep(300 * time.Millisecond)
			}
		}
	}()

	time.Sleep(1 * time.Second)
	close(stop)
	time.Sleep(500 * time.Millisecond)
}

func stopByContext() {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("context stopped")
				return
			default:
				fmt.Println("work work work!")
				time.Sleep(300 * time.Millisecond)
			}
		}
	}()

	time.Sleep(1 * time.Second)
	cancel()
	time.Sleep(500 * time.Millisecond)
}

func stopByGoexit() {
	go func() {
		fmt.Println("work work work!")
		runtime.Goexit()
	}()
	time.Sleep(500 * time.Millisecond)
	fmt.Println("goroutine stopped")
}

func Example_L1_6() {
	fmt.Println("\nStop by condition")
	stopByCondition()

	fmt.Println("\nStop by channel")
	stopByChannel()

	fmt.Println("\nStop by context")
	stopByContext()

	fmt.Println("\nStop by Goexit")
	stopByGoexit()
}
