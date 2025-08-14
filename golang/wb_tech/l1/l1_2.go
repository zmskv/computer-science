package l1

import (
	"fmt"
	"sync"
)

func Square(number int) int {
	return number * number

}
func Example_L1_2() {
	slice := []int{2, 4, 6, 8, 10}

	var wg sync.WaitGroup

	wg.Add(len(slice))
	for _, number := range slice {
		go func(number int) {
			defer wg.Done()
			fmt.Println(Square(number))
		}(number)
	}
	wg.Wait()
}
