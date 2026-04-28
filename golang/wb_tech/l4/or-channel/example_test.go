package or_test

import (
	"fmt"
	"time"

	or "or-channel"
)

func ExampleOr() {
	sig := func(after time.Duration) <-chan interface{} {
		c := make(chan interface{})
		go func() {
			defer close(c)
			time.Sleep(after)
		}()
		return c
	}

	<-or.Or(
		sig(2*time.Hour),
		sig(5*time.Minute),
		sig(10*time.Millisecond),
		sig(1*time.Hour),
		sig(1*time.Minute),
	)

	fmt.Println("done")
}
