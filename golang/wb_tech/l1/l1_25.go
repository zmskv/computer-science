package l1

import (
	"fmt"
	"time"
)

func Sleep(duration time.Duration) {
	ch := make(chan struct{})
	go func() {
		timer := time.NewTimer(duration)
		<-timer.C
		close(ch)
	}()
	<-ch
}

func Example_L1_25() {
	fmt.Println("Start")
	Sleep(2 * time.Second)
	fmt.Println("End")
}
