package l1

import "fmt"

type Human struct {
	name string
	age  int
}

func (h Human) GetName() string {
	return h.name
}

func (h Human) GetAge() int {
	return h.age
}

type Action struct {
	Human
}

func Example_L1_1() {
	a := Action{Human: Human{name: "Pasha", age: 20}}

	fmt.Println(a.GetName())
	fmt.Println(a.GetAge())
}
