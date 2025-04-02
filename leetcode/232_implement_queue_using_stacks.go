package leetcode

type MyQueue struct {
	stackPush []int
	stackPop  []int
}

func Constructor() MyQueue {
	return MyQueue{}
}

func (this *MyQueue) Push(x int) {
	this.stackPush = append(this.stackPush, x)
}

func (this *MyQueue) transfer() {
	if len(this.stackPop) == 0 {
		for len(this.stackPush) > 0 {
			top := this.stackPush[len(this.stackPush)-1]
			this.stackPush = this.stackPush[:len(this.stackPush)-1]
			this.stackPop = append(this.stackPop, top)
		}
	}
}

func (this *MyQueue) Pop() int {
	this.transfer()
	if len(this.stackPop) == 0 {
		return -1
	}

	top := this.stackPop[len(this.stackPop)-1]
	this.stackPop = this.stackPop[:len(this.stackPop)-1]
	return top
}

func (this *MyQueue) Peek() int {
	this.transfer()
	if len(this.stackPop) == 0 {
		return -1
	}
	return this.stackPop[len(this.stackPop)-1]
}

func (this *MyQueue) Empty() bool {
	return len(this.stackPush) == 0 && len(this.stackPop) == 0
}
