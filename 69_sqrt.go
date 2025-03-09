package main

func mySqrt(x int) int {

	if x == 0 || x == 1 {
		return x
	}

	lt, rt := 0, x

	for lt <= rt {
		mid := lt + (rt-lt)/2

		if mid*mid < x {
			lt = mid + 1
		} else if mid*mid > x {
			rt = mid - 1
		} else if mid*mid == x {
			return mid
		}
	}

	return rt

}
