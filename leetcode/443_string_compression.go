package leetcode

import (
	"strconv"
)

func compress(chars []byte) int {
	read_ptr := 0
	w_ptr := 0

	for read_ptr < len(chars) {
		char := chars[read_ptr]
		cnt := 0

		for read_ptr < len(chars) && chars[read_ptr] == char {
			read_ptr++
			cnt++
		}
		chars[w_ptr] = char
		w_ptr++
		if cnt > 1 {
			cntStr := strconv.Itoa(cnt)
			for i := 0; i < len(cntStr); i++ {
				chars[w_ptr] = cntStr[i]
				w_ptr++
			}
		}
	}
	return w_ptr
}


// ассимптотическая сложность O(n), где n - длина массива chars
// сложность по памяти О(1)