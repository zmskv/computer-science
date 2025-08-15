package l1

import "fmt"

func detectType(v interface{}) {
	switch v.(type) {
	case int:
		fmt.Println("type: int")
	case string:
		fmt.Println("type: string")
	case bool:
		fmt.Println("type: bool")
	case chan int:
		fmt.Println("type: int")
	case chan string:
		fmt.Println("type: string")
	default:
		fmt.Println("underfined type")
	}
}

func Example_L1_14() {
	detectType(123)
	detectType("pasha")
	detectType(true)
	detectType(make(chan int))
}
