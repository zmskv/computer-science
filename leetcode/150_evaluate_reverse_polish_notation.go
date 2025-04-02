package main

import "strconv"

func evalRPN(tokens []string) int {
	stack := make([]int, 0)

	for i := 0; i < len(tokens); i++ {
		value, err := strconv.Atoi(tokens[i])
		if err != nil {
			a := stack[len(stack)-1]
			b := stack[len(stack)-2]
			switch tokens[i] {
			case "+":
				r := a + b
				stack = stack[:len(stack)-2]
				stack = append(stack, r)
			case "-":
				r := b - a
				stack = stack[:len(stack)-2]
				stack = append(stack, r)
			case "*":
				r := a * b
				stack = stack[:len(stack)-2]
				stack = append(stack, r)
			case "/":
				r := b / a
				stack = stack[:len(stack)-2]
				stack = append(stack, r)
			}
		} else {
			stack = append(stack, value)
		}

	}
	return stack[0]
}

// ассимптотическая сложность O(n), где n - количество символов в выражении
// сложность по памяти O(n), где n - длина стека