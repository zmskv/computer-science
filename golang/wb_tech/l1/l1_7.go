package l1

import (
	"fmt"
	"sync"
)

func Example_L1_7() {

	var wg sync.WaitGroup
	var mu sync.Mutex

	m := make(map[int]int)
	for i := 0; i <= 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			mu.Lock()
			m[i] = i * i
			mu.Unlock()
		}(i)
	}

	wg.Wait()
	fmt.Println(m)

}
