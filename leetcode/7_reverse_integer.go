package leetcode

func reverse(x int) int {
	ans := 0

	for x != 0 {
		digit := x % 10
		x /= 10

		if ans > 2147483647/10 || (ans == 2147483647/10 && digit > 7) {
			return 0
		}
		if ans < -2147483648/10 || (ans == -2147483648/10 && digit < -8) {
			return 0
		}

		ans = ans*10 + digit
	}

	return ans
}
