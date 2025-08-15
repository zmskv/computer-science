package l1

import (
	"fmt"
	"sync"
)

type SafeCounter struct {
	Count int
	mu    sync.Mutex
}

func (c *SafeCounter) Inc() {
	c.mu.Lock()
	c.Count++
	c.mu.Unlock()
}

func (c *SafeCounter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Count
}

func Example_L1_18() {
	var counter SafeCounter
	var wg sync.WaitGroup

	n := 5
	limit := 1000

	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < limit; j++ {
				counter.Inc()
			}
		}()
	}

	wg.Wait()

	fmt.Println("Value:", counter.Value())
}
