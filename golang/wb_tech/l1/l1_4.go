package l1

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func WorkerWithContext(id int, ctx context.Context, ch <-chan any) {
	for {
		select {
		case <-ctx.Done():
			log.Printf("Worker %d: stopped", id)
			return
		case data := <-ch:
			fmt.Printf("Worker %d: %v\n", id, data)
		}
	}
}

func GracefulShutdown(cancel context.CancelFunc) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	cancel()
}

func Example_L1_4(n int) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan any)

	for i := 1; i <= n; i++ {
		go WorkerWithContext(i, ctx, ch)
	}

	go GracefulShutdown(cancel)

	counter := 1
	for {
		select {
		case <-ctx.Done():
			close(ch)
			log.Println("Main: stopped")
			return
		default:
			ch <- counter
			counter++
			time.Sleep(500 * time.Millisecond)
		}
	}
}
