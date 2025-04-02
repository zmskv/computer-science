package leetcode

func plusOne(digits []int) []int {
	n := len(digits)
	if digits[n-1] != 9 {
		digits[n-1]++
		return digits
	}
	for i := n - 1; i >= 0; i-- {
		if digits[i] == 9 {
			if i == 0 {
				digits[i] = 1
				digits = append(digits, 0)
				return digits
			}
			digits[i] = 0
			continue
		} else {
			digits[i]++
			break
		}
	}
	return digits
}


// сложность O(n), где n - количество разрядов числа
// сложность по памяти O(1)